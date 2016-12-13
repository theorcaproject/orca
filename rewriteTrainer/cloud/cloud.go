package cloud

import (
	"gatoor/orca/base"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"strings"
	"strconv"
	"gatoor/orca/rewriteTrainer/state/configuration"
)

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

type ProviderEventType string
type InstanceStatus string

type ProviderEvent struct {
	HostId base.HostId
	Type ProviderEventType
}

type Provider interface {
	Init()
	SpawnInstances([]base.InstanceType) bool
	SpawnInstance(base.InstanceType) base.HostId
	SpawnInstanceSync(base.InstanceType) base.HostId
	SpawnInstanceLike(base.HostId) base.HostId
	GetIp(base.HostId) base.IpAddr
	GetResources(base.InstanceType) base.InstanceResources
	GetInstanceType(base.HostId) base.InstanceType
	SuitableInstanceTypes(base.InstanceResources) []base.InstanceType
	CheckInstance(base.HostId) InstanceStatus
	TerminateInstance(base.HostId) bool
	GetSpawnLog() []base.HostId
	RemoveFromSpawnLog(base.HostId)

	UpdateLoadBalancers(hostId base.HostId, app base.AppName, version base.Version, event string)
}

var CurrentProvider Provider

func Init() {
	AWSLogger.Infof("Initializing CloudProvider of type %s", state_configuration.GlobalConfigurationState.CloudProvider.Type)
	if state_configuration.GlobalConfigurationState.CloudProvider.Type == PROVIDER_AWS {
		CurrentProvider = &AWSProvider{ProviderConfiguration: state_configuration.GlobalConfigurationState.CloudProvider}
	} else {
		CurrentProvider = &TestProvider{ProviderConfiguration: state_configuration.GlobalConfigurationState.CloudProvider}
	}
	CurrentProvider.Init()

	spawnToMinInstances()
	AWSLogger.Infof("Instances: %+v", state_cloud.GlobalAvailableInstances)
}

func spawnToMinInstances() {
	if len(state_cloud.GlobalAvailableInstances) < int(state_configuration.GlobalConfigurationState.CloudProvider.MinInstances) {
		AWSLogger.Infof("Not enough instances available. Spawning more, available:%d min:%d", len(state_cloud.GlobalAvailableInstances), state_configuration.GlobalConfigurationState.CloudProvider.MinInstances)
		for i := len(state_cloud.GlobalAvailableInstances); i < int(state_configuration.GlobalConfigurationState.CloudProvider.MinInstances); i++ {
			CurrentProvider.SpawnInstanceSync("t2.micro")
		}
	} else {
		AWSLogger.Infof("Enough instances available, going on")
	}
}


type TestProvider struct {
	ProviderConfiguration base.ProviderConfiguration

	Type base.ProviderType
	InstanceTypes []base.InstanceType
	SpawnList []base.HostId
	KillList []base.HostId
}

var testInstanceResouces = map[base.InstanceType]base.InstanceResources{
	"test": {TotalCpuResource: 10, TotalMemoryResource: 10, TotalNetworkResource: 10},
}



func (a *TestProvider) Init() {
	a.Type = PROVIDER_TEST
	a.InstanceTypes = []base.InstanceType{"test", "otherstuff"}
}

func (a *TestProvider) UpdateLoadBalancers(hostId base.HostId, app base.AppName, version base.Version, event string){

}

func (a *TestProvider) GetResources(ty base.InstanceType) base.InstanceResources {

	if !strings.Contains(string(ty), "metrics=") {
		return base.InstanceResources{UsedMemoryResource:0, UsedCpuResource:0, UsedNetworkResource:0, TotalMemoryResource: 10, TotalNetworkResource:10, TotalCpuResource:10}
	}
	arr := strings.Split(strings.Split(string(ty), "metrics=")[1], "_")
	cpu, _ := strconv.Atoi(arr[0])
	mem, _ := strconv.Atoi(arr[1])
	net, _ := strconv.Atoi(arr[2])

	return base.InstanceResources{UsedMemoryResource:0, UsedCpuResource:0, UsedNetworkResource:0, TotalMemoryResource: base.MemoryResource(mem), TotalNetworkResource:base.NetworkResource(net), TotalCpuResource:base.CpuResource(cpu)}
}

func (a *TestProvider) SpawnInstance(ty base.InstanceType) base.HostId {
	AWSLogger.Infof("Trying to spawn a single instance of type '%s'", ty)
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	a.SpawnList = append(a.SpawnList, base.HostId(string(ty)))
	return "TODO"
}

func (a *TestProvider) GetSpawnLog() []base.HostId{
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	return a.SpawnList
}

func (a *TestProvider) RemoveFromSpawnLog(base.HostId) {
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
}

func (a *TestProvider) TerminateInstance(hostId base.HostId) bool{
	AWSLogger.Errorf("NOT IMPLEMENTED TerminateInstance")
	a.KillList = append(a.KillList, hostId)
	return true
}

func (a *TestProvider) GetInstanceType(hostId base.HostId) base.InstanceType{
	return base.InstanceType(hostId)
}

func (a *TestProvider) SpawnInstanceSync(ty base.InstanceType) base.HostId {
	AWSLogger.Infof("Trying to spawn a single instance of type '%s'", ty)
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	return ""
}

func (a *TestProvider) SpawnInstanceLike(hostId base.HostId) base.HostId{
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	//a.SpawnList = append(a.SpawnList, "new_" + hostId)
	return "new_" + hostId
}

func (a *TestProvider) SpawnInstances(tys []base.InstanceType) bool{
	AWSLogger.Infof("Trying to spawn %d instances", len(tys))
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	return true
}

func (a *TestProvider) GetIp(hostId base.HostId) base.IpAddr {
	return ""
}

func (a *TestProvider) SuitableInstanceTypes(resources base.InstanceResources) []base.InstanceType {
	res := []base.InstanceType{}
	return res
}

func (a *TestProvider) CheckInstance(hostId base.HostId) InstanceStatus {
	if strings.Contains(string(hostId), "healthy") {
		return INSTANCE_STATUS_HEALTHY
	}
	if strings.Contains(string(hostId), "spawning") {
		return INSTANCE_STATUS_SPAWNING
	}
	if strings.Contains(string(hostId), "spawn_triggered") {
		return INSTANCE_STATUS_SPAWN_TRIGGERED
	}
	if strings.Contains(string(hostId), "dead") {
		return INSTANCE_STATUS_DEAD
	}
	return INSTANCE_STATUS_HEALTHY
}

