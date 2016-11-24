package base

import (
    //linuxproc "github.com/c9s/goprocinfo/linux"
)

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
type IpAddr string
type HabitatName string
type HabitatStatus string
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

type Usage float32

type HostStats struct {
    MemoryUsage Usage
    CpuUsage Usage
    NetworkUsage Usage
}

type AppStats struct {
    MemoryUsage Usage
    CpuUsage Usage
    NetworkUsage Usage
}

type HostInfo struct {
    HostId HostId
    IpAddr IpAddr
    OsInfo OsInfo
    HabitatInfo HabitatInfo
    Apps []AppInfo
}

type OsInfo struct {
    Os string
    Arch string
}

type HabitatInfo struct {
    Version Version
    Name HabitatName
    Status Status
}

type AppInfo struct {
    Type AppType
    Name AppName
    Version Version
    Status Status
}

type StatsWrapper struct {
    Host HostStats
    Apps map[AppName]AppStats
}

type PushWrapper struct {
    HostInfo HostInfo
    Stats StatsWrapper
}

type PushConfiguration struct {
    TargetHostId HostId
    DeploymentCount DeploymentCount
    AppConfiguration AppConfiguration
    HabitatConfiguration HabitatConfiguration
}

type HabitatConfiguration struct {
    Name HabitatName
    Version Version
    InstallCommands []OsCommand
}

type AppConfiguration struct {
    Name AppName
    Type AppType
    Version Version
    MinDeploymentCount DeploymentCount
    MaxDeploymentCount DeploymentCount
    InstallCommands []OsCommand
    QueryStateCommand OsCommand
    RemoveCommand OsCommand
}