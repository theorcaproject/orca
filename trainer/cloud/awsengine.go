/*
Copyright Alex Mack (al9mack@gmail.com) and Michael Lawson (michael@sphinix.com)
This file is part of Orca.

Orca is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Orca is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Orca.  If not, see <http://www.gnu.org/licenses/>.
*/

package cloud

import (
	"errors"
	"fmt"
	"orca/trainer/model"
	"orca/trainer/state"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type AwsCloudEngine struct {
	awsAccessKeyId            string
	awsAccessKeySecret        string
	awsRegion                 string
	awsBaseAmi                string
	sshKey                    string
	sshKeyPath                string
	spotPrice                 float64
	instanceType              string
	spotInstanceType          string
	trainerConfigBackupBucket string
}

func (aws *AwsCloudEngine) Init(awsAccessKeyId string, awsAccessKeySecret string, awsRegion string, awsBaseAmi string,
	sshKey string, sshKeyPath string, spotPrice float64, instanceType string, spotInstanceType string, trainerConfigBackupBucket string) {

	aws.awsAccessKeySecret = awsAccessKeySecret
	aws.awsAccessKeyId = awsAccessKeyId
	aws.awsRegion = awsRegion
	aws.awsBaseAmi = awsBaseAmi
	aws.sshKey = sshKey
	aws.sshKeyPath = sshKeyPath
	aws.spotPrice = spotPrice
	aws.instanceType = instanceType
	aws.spotInstanceType = spotInstanceType
	aws.trainerConfigBackupBucket = trainerConfigBackupBucket

	//TODO: This is amazingly shitty, but because the aws api sucks and I have no patience its the approach for now
	os.Setenv("AWS_ACCESS_KEY_ID", aws.awsAccessKeyId)
	os.Setenv("AWS_SECRET_ACCESS_KEY", aws.awsAccessKeySecret)
}

func (a *AwsCloudEngine) getInstanceInfo(hostId HostId) (*ec2.Instance, error) {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(a.awsRegion)}))
	res, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{InstanceIds: aws.StringSlice([]string{string(hostId)})})
	if err != nil {
		return &ec2.Instance{}, err
	}
	if len(res.Reservations) != 1 || len(res.Reservations[0].Instances) != 1 {
		return &ec2.Instance{}, errors.New("Wrong instance count")
	}
	return res.Reservations[0].Instances[0], nil
}

func (a *AwsCloudEngine) GetIp(hostId string) string {
	info, err := a.getInstanceInfo(HostId(hostId))
	if err != nil || info == nil || (info.PublicIpAddress == nil && info.PrivateIpAddress == nil) {
		return ""
	}

	var ipAddress *string
	if info.PublicIpAddress != nil {
		ipAddress = info.PublicIpAddress
	} else if info.PrivateIpAddress != nil {
		ipAddress = info.PrivateIpAddress
	}

	return string(*ipAddress)
}

func (a *AwsCloudEngine) GetHostInfo(hostId HostId) (string, string, []model.SecurityGroup, bool, string, string) {
	info, err := a.getInstanceInfo(hostId)
	if err != nil || info == nil || (info.PublicIpAddress == nil && info.PrivateIpAddress == nil) || info.SubnetId == nil {
		return "", "", []model.SecurityGroup{}, false, "", ""
	}
	secGrps := make([]model.SecurityGroup, 0)
	for _, grp := range info.SecurityGroups {
		secGrps = append(secGrps, model.SecurityGroup{Group: string(*grp.GroupId)})
	}

	isSpot := false
	spotId := ""
	if info.InstanceLifecycle != nil {
		isSpot = string(*info.InstanceLifecycle) == "spot"
		if isSpot {
			spotId = string(*info.SpotInstanceRequestId)
		}
	}

	var ipAddress *string
	if info.PublicIpAddress != nil {
		ipAddress = info.PublicIpAddress
	} else if info.PrivateIpAddress != nil {
		ipAddress = info.PrivateIpAddress
	}

	return string(*ipAddress), string(*info.SubnetId), secGrps, isSpot, spotId, string(*info.InstanceType)
}

func (a *AwsCloudEngine) waitOnInstanceReady(hostId HostId) bool {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(a.awsRegion)}))

	if err := svc.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{InstanceIds: aws.StringSlice([]string{string(hostId)})}); err != nil {
		fmt.Println("WaitOnInstanceReady for %s failed: %s", hostId, err)
	}
	return true
}

func (engine *AwsCloudEngine) SpawnInstanceSync(change *model.ChangeServer) *model.Host {
	securityGroupsStrings := make([]string, 0)

	instanceType := InstanceType(engine.spotInstanceType)
	if change.InstanceType != ""{
		instanceType  = InstanceType(change.InstanceType)
	}

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(engine.awsRegion)}))
	conf := &ec2.RunInstancesInput{
		ImageId:      aws.String(engine.awsBaseAmi),
		InstanceType: aws.String(string(instanceType)),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		KeyName:      &engine.sshKey,
	}

	if change.Network != "" {
		conf.SubnetId = aws.String(change.Network)
	}
	if len(change.SecurityGroups) > 0 {
		for _, grp := range change.SecurityGroups {
			securityGroupsStrings = append(securityGroupsStrings, grp.Group)
		}
		conf.SecurityGroupIds = aws.StringSlice(securityGroupsStrings)
	}

	runResult, err := svc.RunInstances(conf)

	if err != nil {
		state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__ERROR,
			Message: fmt.Sprintf("AwsCloudEngine SpawnInstanceSync encountered an error '%s'", err),
		})

		return &model.Host{}
	}

	id := HostId(*runResult.Instances[0].InstanceId)
	if !engine.waitOnInstanceReady(id) {
		return &model.Host{}
	}

	host := &model.Host{
		Id:             string(id),
		SecurityGroups: change.SecurityGroups,
		Network:        change.Network,
	}
	return host
}

func (aws *AwsCloudEngine) GetInstanceType(HostId) InstanceType {
	return ""

}

func (engine *AwsCloudEngine) TerminateInstance(hostId HostId) bool {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(engine.awsRegion)}))
	_, err := svc.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: aws.StringSlice([]string{string(hostId)}),
	})

	if err != nil {
		return false
	}
	return true
}

func (aws *AwsCloudEngine) GetPem() string {
	return aws.sshKeyPath
}
func (engine *AwsCloudEngine) RegisterWithLb(hostId string, lbId string) {
	svc := elb.New(session.New(&aws.Config{Region: aws.String(engine.awsRegion)}))

	params := &elb.RegisterInstancesWithLoadBalancerInput{
		Instances: []*elb.Instance{
			{
				InstanceId: aws.String(string(hostId)),
			},
		},
		LoadBalancerName: aws.String(string(lbId)),
	}
	_, err := svc.RegisterInstancesWithLoadBalancer(params)
	if err != nil {
		return
	}
}

func (engine *AwsCloudEngine) DeRegisterWithLb(hostId string, lbId string) {
	svc := elb.New(session.New(&aws.Config{Region: aws.String(engine.awsRegion)}))

	params := &elb.DeregisterInstancesFromLoadBalancerInput{
		Instances:        []*elb.Instance{{InstanceId: aws.String(string(hostId))}},
		LoadBalancerName: aws.String(string(lbId)),
	}
	_, err := svc.DeregisterInstancesFromLoadBalancer(params)
	if err != nil {
		return
	}
}

func (engine *AwsCloudEngine) SpawnSpotInstanceSync(change *model.ChangeServer) *model.Host {
	securityGroupsStrings := make([]string, 0)

	ty := InstanceType(engine.spotInstanceType)
	if change.InstanceType != ""{
		ty = InstanceType(change.InstanceType)
	}

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(engine.awsRegion)}))
	params := ec2.RequestSpotInstancesInput{
		LaunchSpecification: &ec2.RequestSpotLaunchSpecification{
			ImageId:      aws.String(engine.awsBaseAmi),
			InstanceType: aws.String(string(ty)),
			KeyName:      &engine.sshKey,
		},

		Type:          aws.String("one-time"),
		InstanceCount: aws.Int64(1),
		SpotPrice:     aws.String(strconv.FormatFloat(float64(engine.spotPrice), 'f', 4, 32)),
	}

	if change.Network != "" {
		params.LaunchSpecification.SubnetId = aws.String(change.Network)
	}
	if len(change.SecurityGroups) > 0 {
		for _, grp := range change.SecurityGroups {
			securityGroupsStrings = append(securityGroupsStrings, grp.Group)
		}
		params.LaunchSpecification.SecurityGroupIds = aws.StringSlice(securityGroupsStrings)
	}

	runResult, err := svc.RequestSpotInstances(&params)
	if err != nil {
		state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__ERROR,
			Message: fmt.Sprintf("AwsCloudEngine SpawnSpotInstance encountered an error '%s'", err),
		})

		return &model.Host{}
	}
	if len(runResult.SpotInstanceRequests) == 1 {
		change.SpotInstanceId = (*runResult.SpotInstanceRequests[0].SpotInstanceRequestId)

		time.Sleep(2 * time.Second)
		var hostId HostId
		for {
			hostId, err := engine.GetSpotInstanceHostId(change.SpotInstanceId)
			if err != nil {
				fmt.Println(err)
				break
			}
			if hostId == "" {
				time.Sleep(2 * time.Second)
				continue
			}
			return &model.Host{
				Id:             string(hostId),
				SecurityGroups: change.SecurityGroups,
				Network:        change.Network,
				SpotInstanceId: change.SpotInstanceId,
			}
		}

		if err != nil {
			state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__ERROR,
				Message: fmt.Sprintf("Could not get instanceId for spot instance"),
			})
			return &model.Host{}
		}
		return &model.Host{
			Id:             string(hostId),
			SecurityGroups: change.SecurityGroups,
			Network:        change.Network,
			SpotInstanceId: change.SpotInstanceId,
		}
	}
	return &model.Host{}
}

func (engine *AwsCloudEngine) GetSpotInstanceHostId(spotId string) (HostId, error) {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(engine.awsRegion)}))

	params := &ec2.DescribeSpotInstanceRequestsInput{
		SpotInstanceRequestIds: []*string{
			aws.String(spotId),
		},
	}
	resp, err := svc.DescribeSpotInstanceRequests(params)

	if err != nil {
		fmt.Println(err.Error())
		return "", errors.New("Spawn spot failed")
	}
	if len(resp.SpotInstanceRequests) == 1 {
		if *resp.SpotInstanceRequests[0].Status.Code == "fulfilled" {
			return HostId(*resp.SpotInstanceRequests[0].InstanceId), nil
		}
	}
	return "", nil
}

func (engine *AwsCloudEngine) SanityCheckHosts(hosts map[string]*model.Host) {
	for _, host := range hosts {
		engine.doSanityCheck(host)
	}
}

func (engine *AwsCloudEngine) doSanityCheck(host *model.Host) {
	ip, network, securityGroups, isSpot, spotId, instanceType := engine.GetHostInfo(HostId(host.Id))
	if ip == "" || network == "" || len(securityGroups) == 0 {
		return
	}
	if host.Ip != ip || host.Network != network || host.SpotInstance != isSpot || !securityGroupsEqual(host.SecurityGroups, securityGroups) || host.SpotInstanceId != spotId  || host.InstanceType != instanceType{
		state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
			Message: fmt.Sprintf("Got different info for host %s from AWS. Host was: %s, AWS Ip: %s, Subnet: %s, SpotInstance: %t, securityGroups: %v",
				host.Id, host, ip, network, isSpot, securityGroups),
		})

		host.Ip = ip
		host.Network = network
		host.SpotInstance = isSpot
		host.SecurityGroups = securityGroups
		host.SpotInstanceId = spotId
		host.InstanceType = instanceType

	}
}

func securityGroupsEqual(groups []model.SecurityGroup, other []model.SecurityGroup) bool {
	if len(groups) != len(other) {
		return false
	}

	count := 0
	for _, group := range groups {
		for _, otherGroup := range other {
			if group.Group == otherGroup.Group {
				count += 1
			}
		}
	}
	return count == len(groups)
}

func (engine *AwsCloudEngine) WasSpotInstanceTerminatedDueToPrice(spotRequestId string) (bool, string) {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(engine.awsRegion)}))

	params := &ec2.DescribeSpotInstanceRequestsInput{
		SpotInstanceRequestIds: []*string{
			aws.String(spotRequestId),
		},
	}

	req, resp := svc.DescribeSpotInstanceRequestsRequest(params)
	req.Send()

	if req.Error != nil {
		return false, ""
	}

	for _, instance := range resp.SpotInstanceRequests {
		reason := (*instance.Status.Code)
		if reason == "instance-terminated-by-price" {
			return true, reason
		}
		if reason == "price-too-low" {
			return true, reason
		}
		if reason == "instance-terminated-no-capacity" {
			return true, reason
		}
		if reason == "instance-terminated-capacity-oversubscribed" {
			return true, reason
		}
	}

	return false, ""
}

func (engine *AwsCloudEngine) GetTag(tagKey string, newHostId string) string {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(engine.awsRegion)}))
	filters := make([]*ec2.Filter, 1)

	name := string("resource-id")
	values := make([]*string, 1)
	values[0] = &newHostId
	filters[0] = &ec2.Filter{Name: &name, Values: values}

	tags, err := svc.DescribeTags(&ec2.DescribeTagsInput{
		Filters: filters,
	})

	if err == nil {
		/* Find the name tag */
		for _, tag := range tags.Tags {
			if (*tag.Key) == tagKey {
				return (*tag.Value)
			}
		}
	}

	return ""
}

func (engine *AwsCloudEngine) SetTag(newHostId string, tagKey string, tagValue string) {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(engine.awsRegion)}))

	name := string(tagKey)
	value := tagValue

	resouces := make([]*string, 1)
	resouces[0] = &newHostId

	tags := make([]*ec2.Tag, 1)
	tags[0] = &ec2.Tag{Key: &name, Value: &value}

	svc.CreateTags(&ec2.CreateTagsInput{
		Resources: resouces,
		Tags:      tags,
	})
}

func (engine *AwsCloudEngine) AddNameTag(newHostId string, appName string) {
	currentTag := engine.GetTag("Name", newHostId)
	splices := strings.Split(currentTag, "_")
	splices = append(splices, appName)

	engine.SetTag(newHostId, "Name", strings.Join(splices, "_"))
}

func (engine *AwsCloudEngine) RemoveNameTag(newHostId string, appName string) {
	newTags := make([]string, 0)
	currentTag := engine.GetTag("Name", newHostId)
	splices := strings.Split(currentTag, "_")

	for _, tag := range splices {
		if tag != appName {
			newTags = append(newTags, tag)
		}
	}

	engine.SetTag(newHostId, "Name", strings.Join(newTags, "_"))
}

func (engine *AwsCloudEngine) BackupConfiguration(configuration string) bool {
	uploader := s3manager.NewUploader(session.New(&aws.Config{Region: aws.String(engine.awsRegion)}))
	reader := strings.NewReader(configuration)
	key := time.Now().Format("/2006/01/02/150405/") + "trainer.conf"
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(engine.trainerConfigBackupBucket),
		Key:    aws.String(key),
		Body:   reader,
	})
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

func (engine *AwsCloudEngine) CreateDataQueue(name string, rogueName string) {
	svc := sqs.New(session.New(&aws.Config{Region: aws.String(engine.awsRegion)}))

	rogueQueueARN := ""

	// create rogue queue first - we need this to configure redrive policy of main queue
	if rogueName != "" {
		_, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
			QueueName: aws.String(rogueName),
		})
		// lets only try to create a queue if we don't have one already for this
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok && aerr.Code() == sqs.ErrCodeQueueDoesNotExist {
				attributes := map[string]*string{
					"VisibilityTimeout":             aws.String("600"),
					"ReceiveMessageWaitTimeSeconds": aws.String("20"),
				}
				_, err := svc.CreateQueue(&sqs.CreateQueueInput{
					QueueName:  aws.String(rogueName),
					Attributes: attributes,
				})
				if err != nil {
					state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__ERROR,
						Message: fmt.Sprintf("Could not create SQS queue '%s': '%s'", rogueName, err),
					})
					return
				}
				state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
					Message: fmt.Sprintf("Created SQS queue '%s'", rogueName),
				})
			} else {
				state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__ERROR,
					Message: fmt.Sprintf("Could not check for SQS queue '%s': '%s'", rogueName, err),
				})
				return
			}
		}
	}

	_, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(name),
	})
	// lets only try to create a queue if we don't have one already for this
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == sqs.ErrCodeQueueDoesNotExist {
			if rogueName != "" {
				// get ARN for rogue queue for redrive configuration
				queueUrlOutput, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
					QueueName: aws.String(rogueName),
				})
				if err != nil {
					state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__ERROR,
						Message: fmt.Sprintf("Could not get SQS rogue queue URL '%s': '%s'", rogueName, err),
					})
					return
				}
				rogueQueueAttr := &sqs.GetQueueAttributesInput{
					QueueUrl: queueUrlOutput.QueueUrl,
					AttributeNames: []*string{
						aws.String("QueueArn"),
					},
				}
				resp, attrerr := svc.GetQueueAttributes(rogueQueueAttr)
				if attrerr != nil {
					state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__ERROR,
						Message: fmt.Sprintf("Could not get SQS rogue queue ARN '%s': '%s'", rogueName, err),
					})
					return
				}
				rogueQueueARN = *resp.Attributes["QueueArn"]
			}

			attributes := map[string]*string{
				"VisibilityTimeout":             aws.String("600"),
				"ReceiveMessageWaitTimeSeconds": aws.String("20"),
			}
			if rogueQueueARN != "" {
				attributes["RedrivePolicy"] = aws.String("{\"deadLetterTargetArn\": \"" + rogueQueueARN + "\",\"maxReceiveCount\": 5}")
			}

			_, err := svc.CreateQueue(&sqs.CreateQueueInput{
				QueueName:  aws.String(name),
				Attributes: attributes,
			})
			if err != nil {
				state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__ERROR,
					Message: fmt.Sprintf("Could not create SQS queue '%s': '%s'", name, err),
				})
				return
			}
			state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
				Message: fmt.Sprintf("Created SQS queue '%s'", name),
			})
		} else {
			state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__ERROR,
				Message: fmt.Sprintf("Could not check for SQS queue '%s': '%s'", name, err),
			})
		}
	}
}

func (engine *AwsCloudEngine) MonitorDataQueue(name string) int {
	result := -1
	svc := sqs.New(session.New(&aws.Config{Region: aws.String(engine.awsRegion)}))
	queueUrl, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(name),
	})
	if err != nil {
		state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__ERROR,
			Message: fmt.Sprintf("Could not monitor SQS queue '%s': '%s'", name, err),
		})
		return result
	}
	queueAttr, err := svc.GetQueueAttributes(&sqs.GetQueueAttributesInput{
		QueueUrl: queueUrl.QueueUrl,
		AttributeNames: []*string{
			aws.String("ApproximateNumberOfMessages"),
		},
	})
	if err != nil {
		fmt.Println(err.Error())
		return result
	}
	prop := queueAttr.Attributes["ApproximateNumberOfMessages"]
	result, err = strconv.Atoi(*prop)
	if err != nil {
		fmt.Println(err.Error())
		result = -1
	}
	return result
}
