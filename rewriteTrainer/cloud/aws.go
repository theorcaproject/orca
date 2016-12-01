package cloud

import (
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/base"
	Logger "gatoor/orca/rewriteTrainer/log"
	"os"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"sync"
	"sort"
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
	Type ProviderType
	Config AWSConfiguration
	SpawnLog SpawnLog
}

type AWSConfiguration struct {
	Key string
	Secret string
	Region string
	AMI string
	InstanceTypes []InstanceType
	InstanceCost map[InstanceType]Cost
	InstanceResources map[InstanceType]state_cloud.InstanceResources
	InstanceSafety map[InstanceType]SafeInstance
	SuitableInstanceSafetyFactor float32
}

func (a *AWSProvider) CheckCredentials() bool {
	if a.Config.Key == "" || a.Config.Secret == "" {
		AWSLogger.Errorf("No AWS Credentials set")
		return false
	}
	AWSLogger.Infof("Checking AwsCredentials: Key='%s' Secret='%s...'", a.Config.Key, a.Config.Secret[:4])

	sess, err := session.NewSession()
	if err != nil {
		AWSLogger.Errorf("AwsCredentials fail: %s", err)
		return false
	}
	svc := ec2.New(sess, &aws.Config{Region: aws.String(a.Config.Region)})

	_, err = svc.DescribeInstances(nil)
	if err != nil {
		AWSLogger.Errorf("AwsCredentials fail: %s", err)
		return false
	}

	return true
}

func (a *AWSProvider) Init() {
	a.Type = PROVIDER_AWS
	a.Config.InstanceTypes = []InstanceType{}
	a.Config.InstanceResources = make(map[InstanceType]state_cloud.InstanceResources)
	a.Config.InstanceSafety = make(map[InstanceType]SafeInstance)
	a.Config.InstanceCost = make(map[InstanceType]Cost)
	a.Config.Key = os.Getenv("AWS_ACCESS_KEY_ID")
	a.Config.Secret = os.Getenv("AWS_SECRET_ACCESS_KEY")
	if a.Config.Key == "" || a.Config.Secret == "" {
		AWSLogger.Errorf("Missing AWS credential environment variables AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY")
	}
}

func (a *AWSProvider) SpawnInstance(ty InstanceType) base.HostId {
	AWSLogger.Infof("Trying to spawn a single instance of type '%s'", ty)

	svc := ec2.New(session.New(&aws.Config{Region: aws.String(a.Config.Region)}))

	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:      aws.String(a.Config.AMI),
		InstanceType: aws.String(string(ty)),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
	})

	if err != nil {
		AWSLogger.Errorf("Could not spawn instance of type %s: %s", ty, err)
		return ""
	}

	id := base.HostId(*runResult.Instances[0].InstanceId)
	AWSLogger.Infof("Spawned a single instance of type '%s'. Id=%s. Starting Install of orca client", ty, id)
	a.SpawnLog.Add(id)
	//installOrcaClient()

	return id
}

func (a *AWSProvider) waitOnInstanceReady(hostId base.HostId) bool {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(a.Config.Region)}))

	if err := svc.WaitUntilInstanceRunning(&ec2.DescribeInstancesInput{InstanceIds: aws.StringSlice([]string{string(hostId)}), }); err != nil {
		AWSLogger.Errorf("WaitOnInstanceReady for %s failed: %s", hostId, err)
	}
	return true
}


//func installOrcaClient()

func (a AWSProvider) SpawnInstanceSync(ty InstanceType) base.HostId {
	AWSLogger.Infof("Spawning Instance synchronously, type %s", ty)
	id := a.SpawnInstance(ty)
	if id == "" {
		return ""
	}
	if !a.waitOnInstanceReady(id) {
		return ""
	}
	return id
}

func (a *AWSProvider) SpawnInstanceLike(hostId base.HostId) base.HostId{
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	return hostId
}

func (a *AWSProvider) SpawnInstances(tys []InstanceType) bool {
	AWSLogger.Infof("Trying to spawn %d instances", len(tys))
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	return true
}

func (a *AWSProvider) getInstanceInfo(hostId base.HostId) (*ec2.Instance, error) {
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(a.Config.Region)}))
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
	ip := base.IpAddr(*info.PrivateIpAddress)
	AWSLogger.Infof("Got IpAddress %s for instance %s", ip, hostId)
	return ip
}

func (a *AWSProvider) GetInstanceType(hostId base.HostId) InstanceType {
	AWSLogger.Infof("Getting InstanceType of instance %s", hostId)
	info, err := a.getInstanceInfo(hostId)
	if err != nil {
		AWSLogger.Infof("Got InstanceType for instance %s failed: %s", hostId, err)
		return ""
	}
	ty := InstanceType(*info.InstanceType)
	AWSLogger.Infof("Got InstanceType %s for instance %s", ty, hostId)
	return ty
}

func checkResources(available state_cloud.InstanceResources, needed state_cloud.InstanceResources, safety float32) bool {
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


func (a *AWSProvider) GetResources(ty InstanceType) state_cloud.InstanceResources {
	return a.Config.InstanceResources[ty]
}

type CostSort struct {
	InstanceType InstanceType
	Cost Cost
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

func sortByCost(tys []InstanceType, costMap map[InstanceType]Cost) []InstanceType {
	sorted := CostSorts{}
	for _, ty := range tys {
		if _, exists := costMap[ty]; exists {
			sorted = append(sorted, CostSort{InstanceType: ty, Cost: costMap[ty]})
		}
	}
	sort.Sort(sorted)
	res := []InstanceType{}
	for _, t := range sorted {
		res = append(res, t.InstanceType)
	}
	return res
}

func (a *AWSProvider) SuitableInstanceTypes(resources state_cloud.InstanceResources) []InstanceType {
	AWSLogger.Infof("Get Suitable Instances for needs: %+v", resources)
	suitableInstances := []InstanceType{}
	for _, ty := range a.Config.InstanceTypes {
		if checkResources(a.GetResources(ty), resources, a.Config.SuitableInstanceSafetyFactor) {
			suitableInstances = append(suitableInstances, ty)
		}
	}
	suitableInstances = sortByCost(suitableInstances, a.Config.InstanceCost)
	AWSLogger.Infof("Suitable Instances for needs %+v: %v", resources, suitableInstances)
	return suitableInstances
}

func (a *AWSProvider) CheckInstance(hostId base.HostId) InstanceStatus {
	AWSLogger.Infof("Checking instance %s", hostId)
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(a.Config.Region)}))
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
	AWSLogger.Infof("Trying to terminate instance %s", hostId)
	svc := ec2.New(session.New(&aws.Config{Region: aws.String(a.Config.Region)}))
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
