package cloud

import "orca/trainer/model"

type InstanceType string
type HostId string

type CloudEngine interface {
	SpawnInstanceSync(InstanceType, string, string) *model.Host
	SpawnSpotInstanceSync(InstanceType, string, string) *model.Host
	GetInstanceType(HostId) InstanceType
	TerminateInstance(HostId) bool

	GetIp(hostId string) string

	GetPem() string
	RegisterWithLb(hostId string, elb string)
	DeRegisterWithLb(hostId string, elb string)
	//GetIp(HostId) base.IpAddr
	//UpdateLoadBalancers(hostId HostId, app base.AppName, version base.Version, event string)
}

