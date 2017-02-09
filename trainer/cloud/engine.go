package cloud

import "orca/trainer/model"

type InstanceType string
type HostId string

type CloudEngine interface {
	SpawnInstanceSync(InstanceType, string, []model.SecurityGroup) *model.Host
	SpawnSpotInstanceSync(InstanceType, string, []model.SecurityGroup) *model.Host
	GetInstanceType(HostId) InstanceType
	TerminateInstance(HostId) bool
	GetHostInfo(HostId) (string, string, []model.SecurityGroup, bool)

	GetIp(hostId string) string

	GetPem() string
	RegisterWithLb(hostId string, elb string)
	DeRegisterWithLb(hostId string, elb string)
	SanityCheckHosts(map[string]*model.Host)
	//GetIp(HostId) base.IpAddr
	//UpdateLoadBalancers(hostId HostId, app base.AppName, version base.Version, event string)
}

