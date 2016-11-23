package base

import "gatoor/orca/base"

const (
	APP_HTTP = "http"
	APP_WORKER = "worker"

	STATUS_INIT = "init"
	STATUS_RUNNING = "running"
	STATUS_DEPLOYING = "deploying"
	STATUS_DEAD = "dead"

	FILE_COMMAND = "FILE_COMMAND"
	EXEC_COMMAND = "EXEC_COMMAND"
)

type OsCommandId string
type HostId string
type Version float32
type AppName string
type AppType string
type IpAddr string
type HabitatName string
type Status string
type MinInstances int
type DesiredInstances int
type MaxInstances int
type CloudType string
type CloudName string

type CloudProviderLoadBalancerName string
type CloudProviderLoadBalancerId string
type CloudProviderVpcName string
type CloudProviderVpcCloudIdentifier string
type CloudProviderRegionName string
type CloudProviderRegionCloudIdentifier string
type CloudProviderAvailablityZoneName string
type CloudProviderAvailablityZoneCloudIdentifier string

type Command struct {
	Path string
	Args string
}

type OsCommandType string

type OsCommand struct {
	Id OsCommandId
	Type OsCommandType
	Command Command
}