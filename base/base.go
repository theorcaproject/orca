package base

import (
    //linuxproc "github.com/c9s/goprocinfo/linux"
)
import (
    "time"
    "sync"
)

const (
    APP_HTTP = "http"
    APP_WORKER = "worker"

    STATUS_INIT = "init"
    STATUS_RUNNING = "running"
    STATUS_DEPLOYING = "deploying"
    STATUS_DEAD = "dead"
    STATUS_UNKNOWN = "unknown"

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

type Usage int

type HostStats struct {
    MemoryUsage Usage
    CpuUsage Usage
    NetworkUsage Usage
}

type AppStats struct {
    MemoryUsage Usage
    CpuUsage Usage
    NetworkUsage Usage
    ResponseTime int
}

type HostInfo struct {
    HostId HostId
    IpAddr IpAddr
    OsInfo OsInfo
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
    Id AppId
}

type StatsWrapper struct {
    Host HostStats
    Apps map[AppName]AppStats
}

type TrainerPushWrapper struct {
    HostInfo HostInfo
    Stats MetricsWrapper
}

type AppMetrics map[AppName]map[Version]map[string]AppStats

type MetricsWrapper struct {
    HostMetrics map[string]HostStats
    AppMetrics map[AppName]map[Version]map[string]AppStats
}

var metricsMutex = &sync.Mutex{}

func (m *MetricsWrapper) Wipe() {
    metricsMutex.Lock()
    defer metricsMutex.Unlock()
    (*m).HostMetrics = make(map[string]HostStats)
    (*m).AppMetrics = make(map[AppName]map[Version]map[string]AppStats)
}

func (m MetricsWrapper) Get() MetricsWrapper{
    metricsMutex.Lock()
    defer metricsMutex.Unlock()
    metrics := m
    return metrics
}

func (m MetricsWrapper) AddHostMetrics(hostMetrics HostStats) {
    metricsMutex.Lock()
    defer metricsMutex.Unlock()
    m.HostMetrics[time.Now().UTC().Format(time.RFC3339Nano)] = hostMetrics
}

func (m MetricsWrapper) AddAppMetrics(appName AppName, version Version, appMetrics AppStats) {
    metricsMutex.Lock()
    defer metricsMutex.Unlock()
    if _, exists := m.AppMetrics[appName]; !exists {
        m.AppMetrics[appName] = make(map[Version]map[string]AppStats)
    }
    if _, exists := m.AppMetrics[appName][version]; !exists {
        m.AppMetrics[appName][version] = make(map[string]AppStats)
    }
    m.AppMetrics[appName][version][time.Now().UTC().Format(time.RFC3339Nano)] = appMetrics
}


type PushWrapper struct {
    HostInfo HostInfo
    Stats StatsWrapper
}

type PushConfiguration struct {
    //TargetHostId HostId
    OrcaVersion string
    DeploymentCount DeploymentCount
    AppConfiguration AppConfiguration
    //HabitatConfiguration HabitatConfiguration
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
    DockerConfig DockerConfig
    RawConfig RawConfig
}

type RawConfig struct {
    InstallCommands []OsCommand
    RunCommand OsCommand
    QueryStateCommand OsCommand
    RemoveCommand OsCommand
    StopCommand OsCommand
}

type DockerConfig struct {
    Registry string
    Repository string
    Reference string
}

var appsMetricsMutex = &sync.Mutex{}

func (a *AppMetrics) Add(name AppName, version Version, time string, stats AppStats) {
    appsMetricsMutex.Lock()
    defer appsMetricsMutex.Unlock()
    if _, exists := (*a)[name]; !exists {
        (*a)[name] = make(map[Version]map[string]AppStats)
    }
    if _, exists := (*a)[name][version]; !exists {
        (*a)[name][version] = make(map[string]AppStats)
    }
    (*a)[name][version][time] = stats
}

func (a *AppMetrics) All() map[AppName]map[Version]map[string]AppStats {
    appsMetricsMutex.Lock()
    defer appsMetricsMutex.Unlock()
    res := (*a)
    return res
}

func (a *AppMetrics) Clear() {
    appsMetricsMutex.Lock()
    defer appsMetricsMutex.Unlock()
    (*a) = make(map[AppName]map[Version]map[string]AppStats)
}
