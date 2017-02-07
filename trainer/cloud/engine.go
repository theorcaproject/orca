package cloud

type InstanceType string
type HostId string

type CloudEngine interface {
	SpawnInstanceSync(InstanceType) HostId
	SpawnSpotInstanceSync(InstanceType) HostId
	GetInstanceType(HostId) InstanceType
	TerminateInstance(HostId) bool

	GetIp(hostId HostId) string

	GetPem() string
	//GetIp(HostId) base.IpAddr
	//UpdateLoadBalancers(hostId HostId, app base.AppName, version base.Version, event string)
}

