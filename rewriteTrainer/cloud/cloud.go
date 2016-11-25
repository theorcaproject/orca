package cloud

import (
	"gatoor/orca/base"
	"gatoor/orca/rewriteTrainer/state/cloud"
)

type ProviderType string
type InstanceType string
type InstanceCount int

const (
	PROVIDER_TEST = "TEST"
	PROVIDER_AWS = "AWS"

	INSTANCE_STATUS_SPAWN_TRIGGERED = "INSTANCE_STATUS_SPAWN_TRIGGERED"
	INSTANCE_STATUS_SPAWNING = "INSTANCE_STATUS_SPAWNING"
	INSTANCE_STATUS_HEALTHY = "INSTANCE_STATUS_HEALTHY"
	INSTANCE_STATUS_DEAD = "INSTANCE_DEAD"

	PROVIDER_EVENT_KILLED = "PROVIDER_EVENT_KILLED"
	PROVIDER_EVENT_READY = "PROVIDER_EVENT_READY"
)


type MinInstanceCount int
type MaxInstanceCount int

type InstanceStatus string
type ProviderEventType string

type ProviderEvent struct {
	HostId base.HostId
	Type ProviderEventType
}

type ProviderConfiguration struct {
	Type ProviderType
	MinInstances InstanceCount
	MaxInstances InstanceCount
	AllowedInstanceTypes []InstanceType
	FundamentalInstanceType InstanceType
}

type Provider interface {
	SpawnInstances([]InstanceType)
	SpawnInstance(InstanceType)
	SpawnInstanceSync(InstanceType)
	SpawnInstanceLike(base.HostId) base.HostId
	GetIp(InstanceType) base.IpAddr
	GetResources(InstanceType) state_cloud.InstanceResources
	SuitableInstanceTypes(state_cloud.InstanceResources) []InstanceType
	CheckInstance(base.HostId) InstanceStatus
}

var CurrentProviderConfig ProviderConfiguration
var CurrentProvider Provider

func Init() {
	AWSLogger.Infof("Initializing CloudProvider of type %s", CurrentProviderConfig.Type)
	if CurrentProviderConfig.Type == PROVIDER_AWS {
		CurrentProvider = AWSProvider{}
	} else {
		CurrentProvider = TestProvider{}
	}

	spawnToMinInstances()
}

func spawnToMinInstances() {
	if len(state_cloud.GlobalAvailableInstances) < int(CurrentProviderConfig.MinInstances) {
		AWSLogger.Infof("Not enough instances available. Spawning more, available:%d min:%d", len(state_cloud.GlobalAvailableInstances), CurrentProviderConfig.MinInstances)
		for i := len(state_cloud.GlobalAvailableInstances); i < int(CurrentProviderConfig.MinInstances); i++ {
			CurrentProvider.SpawnInstanceSync(CurrentProviderConfig.FundamentalInstanceType)
		}
	} else {
		AWSLogger.Infof("Enough instances available, going on")
	}
}


type TestProvider struct {
	Type ProviderType
	InstanceTypes []InstanceType
	SpawnList []base.HostId
}

var testInstanceResouces = map[InstanceType]state_cloud.InstanceResources{
	"test": {TotalCpuResource: 10, TotalMemoryResource: 10, TotalNetworkResource: 10},
}



func (a TestProvider) Init() {
	a.Type = PROVIDER_TEST
	a.InstanceTypes = []InstanceType{"test", "otherstuff"}
}

func (a TestProvider) GetResources(ty InstanceType) state_cloud.InstanceResources {
	elem, _ := awsInstanceResouces[ty]
	//if !exists {
	//
	//}
	return elem
}

func (a TestProvider) SpawnInstance(ty InstanceType) {
	AWSLogger.Infof("Trying to spawn a single instance of type '%s'", ty)
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
}

func (a TestProvider) SpawnInstanceSync(ty InstanceType) {
	AWSLogger.Infof("Trying to spawn a single instance of type '%s'", ty)
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
}

func (a TestProvider) SpawnInstanceLike(hostId base.HostId) base.HostId{
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	a.SpawnList = append(a.SpawnList, "new_" + hostId)
	return "new_" + hostId
}

func (a TestProvider) SpawnInstances(tys []InstanceType) {
	AWSLogger.Infof("Trying to spawn %d instances", len(tys))
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
}

func (a TestProvider) GetIp(ty InstanceType) base.IpAddr {
	return ""
}

func (a TestProvider) SuitableInstanceTypes(resources state_cloud.InstanceResources) []InstanceType {
	res := []InstanceType{}
	return res
}

func (a TestProvider) CheckInstance(hostId base.HostId) InstanceStatus {
	if hostId == "healthy" {
		return INSTANCE_STATUS_HEALTHY
	}
	if hostId == "spawning" {
		return INSTANCE_STATUS_SPAWNING
	}
	if hostId == "spawn_triggered" {
		return INSTANCE_STATUS_SPAWN_TRIGGERED
	}
	return INSTANCE_STATUS_DEAD
}

