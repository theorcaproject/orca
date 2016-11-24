package cloud

import (
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/rewriteTrainer/base"
	Logger "gatoor/orca/rewriteTrainer/log"
)

var AWSLogger = Logger.LoggerWithField(Logger.Logger, "module", "aws")

type AWSProvider struct {
	Type ProviderType
	InstanceTypes []InstanceType
}

var awsInstanceResouces = map[InstanceType]state_cloud.InstanceResources{
	"m1.xlarge": {TotalCpuResource: 10, TotalMemoryResource: 10, TotalNetworkResource: 10},
}



func (a AWSProvider) Init() {
	a.Type = PROVIDER_AWS
	a.InstanceTypes = []InstanceType{"m1.xlarge", "otherstuff"}
}

func (a AWSProvider) GetResources(ty InstanceType) state_cloud.InstanceResources {
	elem, _ := awsInstanceResouces[ty]
	//if !exists {
	//
	//}
	return elem
}

func (a AWSProvider) SpawnInstance(ty InstanceType) {
	AWSLogger.Infof("Trying to spawn a single instance of type '%s'", ty)
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
}

func (a AWSProvider) SpawnInstanceLike(hostId base.HostId) base.HostId{
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	return hostId
}

func (a AWSProvider) SpawnInstances(tys []InstanceType) {
	AWSLogger.Infof("Trying to spawn %d instances", len(tys))
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
	AWSLogger.Errorf("NOT IMPLEMENTED")
}

func (a AWSProvider) GetIp(ty InstanceType) base.IpAddr {
	return ""
}

func (a AWSProvider) SuitableInstanceTypes(resources state_cloud.InstanceResources) []InstanceType {
	res := []InstanceType{}
	return res
}

func (a AWSProvider) CheckInstance(hostId base.HostId) InstanceStatus {
	return INSTANCE_STATUS_DEAD
}
