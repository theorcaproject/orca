package cloud

import (
	"gatoor/orca/base"
	Logger "gatoor/orca/rewriteTrainer/log"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"sync"
	"sort"
	"gatoor/orca/rewriteTrainer/installer"
	"gatoor/orca/client/types"
	"fmt"
	"gatoor/orca/rewriteTrainer/state/configuration"
	"os"
)

var AWSLogger = Logger.LoggerWithField(Logger.Logger, "module", "aws")

type SpawnLog []base.HostId
var spawnLogMutex = &sync.Mutex{}

func (s SpawnLog) Add(hostId base.HostId) {
	spawnLogMutex.Lock()
	defer spawnLogMutex.Unlock()
	s = append(s, hostId)
}

func (s SpawnLog) Remove(hostId base.HostId) {
	spawnLogMutex.Lock()
	defer spawnLogMutex.Unlock()
	i := -1
	for iter, host := range s {
		if host == hostId {
			i = iter
		}
	}
	if i >= 0 {
		s = append(s[:i], s[i+1:]...)
	}
}

type AWSProvider struct {
	ProviderConfiguration base.ProviderConfiguration

	Type base.ProviderType
	SpawnLog SpawnLog
}

func (a *AWSProvider) CheckCredentials() bool {
	if a.ProviderConfiguration.AWSConfiguration.Key == "" || a.ProviderConfiguration.AWSConfiguration.Secret == "" {
		AWSLogger.Errorf("No AWS Credentials set")
		return false
	}
	AWSLogger.Infof("Checking AwsCredentials: Key='%s' Secret='%s...'", a.ProviderConfiguration.AWSConfiguration.Key, a.ProviderConfiguration.AWSConfiguration.Secret[:4])

	sess, err := session.NewSession()
	if err != nil {
		AWSLogger.Errorf("AwsCredentials fail: %s", err)
		return false
	}
	svc := ec2.New(sess, &aws.Config{Region: aws.String(a.ProviderConfiguration.AWSConfiguration.Region)})

	_, err = svc.DescribeInstances(nil)
	if err != nil {
		AWSLogger.Errorf("AwsCredentials fail: %s", err)
		return false
	}

	return true
}

func (a *AWSProvider) Init() {
	a.Type = PROVIDER_AWS
	if a.ProviderConfiguration.AWSConfiguration.Key == "" || a.ProviderConfiguration.AWSConfiguration.Secret == "" {
		AWSLogger.Errorf("Missing AWS credential environment variables AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY")
	}

	//TODO: When the cloud provider init is called, we use the aws api based on the credentials set to populate below:
	//CurrentProviderConfig.AWSConfiguration.InstanceTypes = []InstanceType{}
	//CurrentProviderConfig.AWSConfiguration.InstanceResources = make(map[InstanceType]state_cloud.InstanceResources)
	//CurrentProviderConfig.AWSConfiguration.InstanceSafety = make(map[InstanceType]SafeInstance)
	//CurrentProviderConfig.AWSConfiguration.InstanceCost = make(map[InstanceType]Cost)
}

func (a *AWSProvider) SpawnInstance(ty base.InstanceType) base.HostId {
	AWSLogger.Warnf("Trying to spawn a single instance of type '%s' in region %s with AMI %s", ty, a.ProviderConfiguration.AWSConfiguration.Region, a.ProviderConfiguration.AWSConfiguration.AMI)

	//TODO: This is amazingly shitty, but because the aws api sucks and I have no patience its the approach for now
	os.Setenv("AWS_ACCESS_KEY_ID", a.ProviderConfiguration.AWSConfiguration.Key)
	os.Setenv("AWS_SECRET_ACCESS_KEY", a.ProviderConfiguration.AWSConfiguration.Secret)

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(a.ProviderConfiguration.AWSConfiguration.Region)}))

	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:      aws.String(a.ProviderConfiguration.AWSConfiguration.AMI),
		InstanceType: aws.String(string(ty)),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
		KeyName:      &a.ProviderConfiguration.SSHKey,
		SecurityGroupIds: aws.StringSlice([]string{string(a.ProviderConfiguration.AWSConfiguration.SecurityGroupId)}),
	})

	if err != nil {
		AWSLogger.Errorf("Could not spawn instance of type %s: %s", ty, err)
		return ""
	}

	id := base.HostId(*runResult.Instances[0].InstanceId)
	AWSLogger.Infof("Spawned a single instance of type '%s'. Id=%s", ty, id)
	a.SpawnLog.Add(id)

	return id
}

func (a *AWSProvider) waitOnInstanceReady(hostId base.HostId) bool {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(a.ProviderConfiguration.AWSConfiguration.Region)}))

	if err := svc.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{InstanceIds: aws.StringSlice([]string{string(hostId)}), }); err != nil {
		AWSLogger.Errorf("WaitOnInstanceReady for %s failed: %s", hostId, err)
	}
	return true
}


func installOrcaClient(hostId base.HostId, ip base.IpAddr, trainerIp base.IpAddr, sshKey string, sshUser string) {
	clientConf := types.Configuration{
		Type: types.DOCKER_CLIENT,
		TrainerPollInterval: 30,
		AppStatusPollInterval: 10,
		MetricsPollInterval: 10,
		TrainerUrl: fmt.Sprintf("http://%s:5000/push", trainerIp),
		Port: 5001,
		HostId: hostId,
	}
	installer.InstallNewInstance(clientConf, ip, sshKey, sshUser)
}

func (a AWSProvider) SpawnInstanceSync(ty base.InstanceType) base.HostId {
	AWSLogger.Infof("Spawning Instance synchronously, type %s", ty)
	id := a.SpawnInstance(ty)
	if id == "" {
		return ""
	}
	if !a.waitOnInstanceReady(id) {
		return ""
	}

	ipAddr := a.GetIp(id)
	sshKeyPath := state_configuration.GlobalConfigurationState.ConfigurationRootPath + "/" +a.ProviderConfiguration.SSHKey + ".pem"
	installOrcaClient(id, ipAddr, state_configuration.GlobalConfigurationState.Trainer.Ip, sshKeyPath, a.ProviderConfiguration.SSHUser)

	return id
}

func (a *AWSProvider) SpawnInstanceLike(hostId base.HostId) base.HostId{
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	return hostId
}

func (a *AWSProvider) SpawnInstances(tys []base.InstanceType) bool {
	AWSLogger.Infof("Trying to spawn %d instances", len(tys))
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	return true
}

func (a *AWSProvider) getInstanceInfo(hostId base.HostId) (*ec2.Instance, error) {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(a.ProviderConfiguration.AWSConfiguration.Region)}))
	res, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{InstanceIds: aws.StringSlice([]string{string(hostId)}), })
	if err != nil {
		return &ec2.Instance{}, err
	}
	if len(res.Reservations) != 1 || len(res.Reservations[0].Instances) != 1 {
		return &ec2.Instance{}, errors.New("Wrong instance count")
	}
	return res.Reservations[0].Instances[0], nil
}

func (a *AWSProvider) GetIp(hostId base.HostId) base.IpAddr {
	AWSLogger.Infof("Getting IpAddress of instance %s", hostId)
	info, err := a.getInstanceInfo(hostId)
	if err != nil {
		AWSLogger.Infof("Got IpAddress for instance %s failed: %s", hostId, err)
		return ""
	}
	ip := base.IpAddr(*info.PublicIpAddress)
	AWSLogger.Infof("Got IpAddress %s for instance %s", ip, hostId)
	return ip
}

func (a *AWSProvider) GetInstanceType(hostId base.HostId) base.InstanceType {
	AWSLogger.Infof("Getting InstanceType of instance %s", hostId)
	info, err := a.getInstanceInfo(hostId)
	if err != nil {
		AWSLogger.Infof("Got InstanceType for instance %s failed: %s", hostId, err)
		return ""
	}
	ty := base.InstanceType(*info.InstanceType)
	AWSLogger.Infof("Got InstanceType %s for instance %s", ty, hostId)
	return ty
}

func checkResources(available base.InstanceResources, needed base.InstanceResources, safety float32) bool {
	if float32(available.TotalCpuResource) < float32(needed.TotalCpuResource) * safety {
		return false
	}
	if float32(available.TotalMemoryResource) < float32(needed.TotalMemoryResource) * safety {
		return false
	}
	if float32(available.TotalNetworkResource) < float32(needed.TotalNetworkResource) * safety {
		return false
	}
	return true
}


func (a *AWSProvider) GetResources(ty base.InstanceType) base.InstanceResources {
	return a.ProviderConfiguration.AWSConfiguration.InstanceResources[ty]
}

type CostSort struct {
	InstanceType base.InstanceType
	Cost base.Cost
}

type CostSorts []CostSort

func (slice CostSorts) Len() int {
	return len(slice)
}

func (slice CostSorts) Less(i, j int) bool {
	return slice[i].Cost < slice[j].Cost;
}

func (slice CostSorts) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func sortByCost(tys []base.InstanceType, costMap map[base.InstanceType]base.Cost) []base.InstanceType {
	sorted := CostSorts{}
	for _, ty := range tys {
		if _, exists := costMap[ty]; exists {
			sorted = append(sorted, CostSort{InstanceType: ty, Cost: costMap[ty]})
		}
	}
	sort.Sort(sorted)
	res := []base.InstanceType{}
	for _, t := range sorted {
		res = append(res, t.InstanceType)
	}
	return res
}

func (a *AWSProvider) SuitableInstanceTypes(resources base.InstanceResources) []base.InstanceType {
	AWSLogger.Infof("Get Suitable Instances for needs: %+v", resources)
	suitableInstances := []base.InstanceType{}
	for _, ty := range a.ProviderConfiguration.AWSConfiguration.InstanceTypes {
		if checkResources(a.GetResources(ty), resources, a.ProviderConfiguration.AWSConfiguration.SuitableInstanceSafetyFactor) {
			suitableInstances = append(suitableInstances, ty)
		}
	}
	suitableInstances = sortByCost(suitableInstances, a.ProviderConfiguration.AWSConfiguration.InstanceCost)
	AWSLogger.Infof("Suitable Instances for needs %+v: %v", resources, suitableInstances)
	return suitableInstances
}

func (a *AWSProvider) CheckInstance(hostId base.HostId) InstanceStatus {
	AWSLogger.Infof("Checking instance %s", hostId)
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(a.ProviderConfiguration.AWSConfiguration.Region)}))
	res, err := svc.DescribeInstanceStatus(&ec2.DescribeInstanceStatusInput{InstanceIds: aws.StringSlice([]string{string(hostId)})})
	if err != nil {
		AWSLogger.Infof("Checking instance %s failed:%s", hostId, err)
		return INSTANCE_STATUS_DEAD
	}
	if len(res.InstanceStatuses) != 1 {
		return INSTANCE_STATUS_DEAD
	}
	status := *res.InstanceStatuses[0].InstanceState.Name
	AWSLogger.Info(status)
	return INSTANCE_STATUS_HEALTHY
}

func (a *AWSProvider) TerminateInstance(hostId base.HostId) bool {
	AWSLogger.Warnf("Trying to terminate instance %s", hostId)
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(a.ProviderConfiguration.AWSConfiguration.Region)}))
	_, err := svc.TerminateInstances(&ec2.TerminateInstancesInput{
		InstanceIds: aws.StringSlice([]string{string(hostId)}),
	})

	if err != nil {
		AWSLogger.Errorf("Could not terminate instance %s: %s", hostId, err)
		return false
	}
	AWSLogger.Infof("Terminated instance %s", hostId)
	a.SpawnLog.Remove(hostId)
	return true

}

func (a *AWSProvider) GetSpawnLog() []base.HostId {
	return a.SpawnLog
}

func (a *AWSProvider) RemoveFromSpawnLog(hostId base.HostId) {
}
