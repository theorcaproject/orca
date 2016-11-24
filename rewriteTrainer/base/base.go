package base


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

type HostId string
type Version string
type AppName string
type AppType string
type IpAddr string
type HabitatName string
type Status string
type DeploymentCount int


type Command struct {
	Path string
	Args string
}

type OsCommandType string

type OsCommand struct {
	Type OsCommandType
	Command Command
}