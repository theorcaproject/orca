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
	"os"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/service/elb"
	"strconv"
	"time"

	"orca/trainer/model"
)

type AwsCloudEngine struct {
	awsAccessKeyId     string
	awsAccessKeySecret string
	awsRegion          string
	awsBaseAmi         string
	sshKey             string
	sshKeyPath         string
	spotPrice          float64
	instanceType       string
	spotInstanceType   string
}

func (aws *AwsCloudEngine) Init(awsAccessKeyId string, awsAccessKeySecret string, awsRegion string, awsBaseAmi string,
sshKey string, sshKeyPath string, spotPrice float64, instanceType string, spotInstanceType string) {
	aws.awsAccessKeySecret = awsAccessKeySecret
	aws.awsAccessKeyId = awsAccessKeyId
	aws.awsRegion = awsRegion
	aws.awsBaseAmi = awsBaseAmi
	aws.sshKey = sshKey
	aws.sshKeyPath = sshKeyPath
	aws.spotPrice = spotPrice
	aws.instanceType = instanceType
	aws.spotInstanceType = spotInstanceType

	//TODO: This is amazingly shitty, but because the aws api sucks and I have no patience its the approach for now
	os.Setenv("AWS_ACCESS_KEY_ID", aws.awsAccessKeyId)
	os.Setenv("AWS_SECRET_ACCESS_KEY", aws.awsAccessKeySecret)
}

func (a *AwsCloudEngine) getInstanceInfo(hostId HostId) (*ec2.Instance, error) {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(a.awsRegion)}))
	res, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{InstanceIds: aws.StringSlice([]string{string(hostId)}), })
	if err != nil {
		return &ec2.Instance{}, err
	}
	if len(res.Reservations) != 1 || len(res.Reservations[0].Instances) != 1 {
		return &ec2.Instance{}, errors.New("Wrong instance count")
	}
	fmt.Println(fmt.Sprintf("Got Host info: %+v", res.Reservations[0].Instances[0]))
	return res.Reservations[0].Instances[0], nil
}

func (a *AwsCloudEngine) GetIp(hostId string) string {
	info, err := a.getInstanceInfo(HostId(hostId))
	if err != nil || info == nil || info.PublicIpAddress == nil {
		return ""
	}

	return string(*info.PublicIpAddress)
}

func (a *AwsCloudEngine) GetHostInfo(hostId HostId) (string, string, []model.SecurityGroup) {
	info, err := a.getInstanceInfo(hostId)
	if err != nil || info == nil || info.PublicIpAddress == nil {
		return "", "", []model.SecurityGroup{}
	}
	secGrps := make([]model.SecurityGroup, 0)
	for _, grp := range info.SecurityGroups {
		secGrps = append(secGrps, model.SecurityGroup{Group: string(*grp.GroupId)})
	}
	return string(*info.PublicIpAddress), string(*info.SubnetId), secGrps
}

func (a *AwsCloudEngine) waitOnInstanceReady(hostId HostId) bool {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(a.awsRegion)}))

	if err := svc.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{InstanceIds: aws.StringSlice([]string{string(hostId)}), }); err != nil {
		fmt.Println("WaitOnInstanceReady for %s failed: %s", hostId, err)
	}
	return true
}


func (engine *AwsCloudEngine) SpawnInstanceSync(instanceType InstanceType, network string, securityGroups []model.SecurityGroup) *model.Host {
	securityGroupsStrings := make([]string, 0)
	if instanceType == "" {
		instanceType = InstanceType(engine.instanceType)
	}
	fmt.Println("AwsCloudEngine SpawnInstanceSync called with ", instanceType)
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(engine.awsRegion)}))

	conf := &ec2.RunInstancesInput{
		ImageId:      aws.String(engine.awsBaseAmi),
		InstanceType: aws.String(string(instanceType)),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		KeyName:      &engine.sshKey,
	}

	if network != "" {
		conf.SubnetId = aws.String(network)
	}
	if len(securityGroups) > 0 {
		for _, grp := range securityGroups {
			securityGroupsStrings = append(securityGroupsStrings, grp.Group)
		}
		conf.SecurityGroupIds = aws.StringSlice(securityGroupsStrings)
	}

	runResult, err := svc.RunInstances(conf)

	if err != nil {
		fmt.Println("AwsCloudEngine SpawnInstanceSync encountered an error ", err)
		return &model.Host{}
	}

	id := HostId(*runResult.Instances[0].InstanceId)
	fmt.Println("AwsCloudEngine SpawnInstanceSync got a new host, lets wait until its ready. HostID is ", id)
	if !engine.waitOnInstanceReady(id) {
		return &model.Host{}
	}

	fmt.Println("AwsCloudEngine SpawnInstanceSync finished")
	host := &model.Host{
		Id: string(id),
		SecurityGroups: securityGroups,
		Network: network,
	}
	return host
}

func (aws *AwsCloudEngine) GetInstanceType(HostId) InstanceType {
	return ""

}

func (engine *AwsCloudEngine) TerminateInstance(hostId HostId) bool {
	fmt.Println(fmt.Sprintf("TERMINATE INSTANCE: %s", hostId))
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
		Instances: []*elb.Instance{{InstanceId: aws.String(string(hostId))}},
		LoadBalancerName: aws.String(string(lbId)),
	}
	_, err := svc.DeregisterInstancesFromLoadBalancer(params)
	if err != nil {
		return
	}
}

func (engine *AwsCloudEngine) SpawnSpotInstanceSync(ty InstanceType, network string, securityGroups []model.SecurityGroup) *model.Host {
	securityGroupsStrings := make([]string, 0)
	if ty == "" {
		ty = InstanceType(engine.spotInstanceType)
	}
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(engine.awsRegion)}))

	params := ec2.RequestSpotInstancesInput{
		LaunchSpecification: &ec2.RequestSpotLaunchSpecification{
			ImageId:      aws.String(engine.awsBaseAmi),
			InstanceType: aws.String(string(ty)),
			KeyName:      &engine.sshKey,
		},

		Type: aws.String("one-time"),
		InstanceCount: aws.Int64(1),
		SpotPrice: aws.String(strconv.FormatFloat(float64(engine.spotPrice), 'f', 4, 32)),
	}

	if network != "" {
		params.LaunchSpecification.SubnetId = aws.String(network)
	}
	if len(securityGroups) >0 {
		for _, grp := range securityGroups {
			securityGroupsStrings = append(securityGroupsStrings, grp.Group)
		}
		params.LaunchSpecification.SecurityGroupIds = aws.StringSlice(securityGroupsStrings)
	}

	runResult, err := svc.RequestSpotInstances(&params)
	if err != nil {
		fmt.Println("AwsCloudEngine SpawnSpotInstance encountered an error ", err)
		return &model.Host{}
	}
	if len(runResult.SpotInstanceRequests) == 1 {
		spotId := runResult.SpotInstanceRequests[0].SpotInstanceRequestId
		time.Sleep(2* time.Second)
		var hostId HostId
		for {
			hostId, err := engine.GetSpotInstanceHostId(*spotId)
			if err != nil {
				fmt.Println(err)
				break
			}
			if hostId == "" {
				time.Sleep(2 * time.Second)
				continue
			}
			return  &model.Host{
				Id: string(hostId),
				SecurityGroups: securityGroups,
				Network: network,
			}
		}

		if err != nil {
			fmt.Println("Could not get instance id of spot request")
			return &model.Host{}
		}
		return  &model.Host{
			Id: string(hostId),
			SecurityGroups: securityGroups,
			Network: network,
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