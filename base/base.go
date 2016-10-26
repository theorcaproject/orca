package base

import linuxproc "github.com/c9s/goprocinfo/linux"

type HostId string
type IpAddr string
type Version string

type TrainerUpdate struct {
    Version Version
    TargetHostId HostId
    HabitatConfiguration HabitatConfiguration
    AppsConfiguration map[HostId]AppConfiguration
}

type Command struct {
    Path string
    Args string
}

type OsCommandType string

type OsCommand struct {
    Type OsCommandType
    Command Command
}



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

type HostConfiguration struct {
    HabitatConfiguration HabitatConfiguration
    AppsConfiguration map[HostId]AppConfiguration
}

type HabitatConfiguration struct {
    Version Version
    Commands []OsCommand
}

type HabitatStatus string

type HostInfo struct {
    HostId HostId
    IpAddr IpAddr
    HabitatInfo HabitatInfo
    OsInfo OsInfo
    Apps map[HostId]AppInfo
}

type OsInfo struct {
    Os string
    Arch string
}

type HabitatInfo struct {
    Version Version
    Status HabitatStatus
}

type AppInfo struct {
    Type AppType
    Name AppName
    Version Version
    Status AppStatus
    QueryStateCommand OsCommand
    RemoveCommand OsCommand
}

type StatsWrapper struct {
    HostInfo HostInfo
    Stats *linuxproc.Stat
}
