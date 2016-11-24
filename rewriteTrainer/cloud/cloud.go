package cloud

import (
	"gatoor/orca/rewriteTrainer/base"
	"gatoor/orca/rewriteTrainer/state/cloud"
)

type ProviderType string
type InstanceType string

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

type Provider interface {
	SpawnInstances([]InstanceType)
	SpawnInstance(InstanceType)
	SpawnInstanceLike(base.HostId) base.HostId
	GetIp(InstanceType) base.IpAddr
	GetResources(InstanceType) state_cloud.InstanceResources
	SuitableInstanceTypes(state_cloud.InstanceResources) []InstanceType
	CheckInstance(base.HostId) InstanceStatus
}

var CurrentProvider Provider

func init() {
	CurrentProvider = TestProvider{}
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

