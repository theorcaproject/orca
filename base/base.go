package base

import linuxproc "github.com/c9s/goprocinfo/linux"

type TrainerUpdate struct {
    Version string
    TargetHostId string
    HabitatConfiguration HabitatConfiguration
    AppsConfiguration map[string]AppConfiguration
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


type AppType string

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
    AppsConfiguration map[string]AppConfiguration
}

type HabitatConfiguration struct {
    Version string
    Commands []OsCommand
}

type AppConfiguration struct {
    Version string
    AppName string
    AppType AppType
    InstallCommands []OsCommand
    QueryStateCommand OsCommand
    RemoveCommand OsCommand
}

type AppStatus string
type HabitatStatus string

type HostInfo struct {
    Id string
    HabitatInfo HabitatInfo
    OsInfo OsInfo
    Apps map[string]AppInfo
}

type OsInfo struct {
    Os string
    Arch string
}

type HabitatInfo struct {
    Version string
    Status HabitatStatus
}

type AppInfo struct {
    Type AppType
    Name string
    CurrentVersion string
    Status AppStatus
    QueryStateCommand OsCommand
    RemoveCommand OsCommand
}

type StatsWrapper struct {
    HostInfo HostInfo
    Stats *linuxproc.Stat
}
