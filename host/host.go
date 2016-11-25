package main

import (
    "encoding/json"
    "time"
    //"bytes"
    //"net/http"
    "gatoor/orca/base"
    Logger "gatoor/orca/rewriteTrainer/log"
    "io/ioutil"
    "os/exec"
    "strings"
    "sync"
    "bytes"
    "net/http"
    "os"
    "gatoor/orca/util"
    "fmt"
)

type Configuration struct {
    PollInterval int
    TrainerUrl string
    HostId base.HostId
}


var configuration Configuration
var HostLogger = Logger.LoggerWithField(Logger.Logger, "module", "host")

var hostInfo base.HostInfo
var hostInfoMutex = &sync.Mutex{}

type AppConfig map[base.AppName]map[base.Version]base.AppConfiguration
var appConfigsMutex = &sync.Mutex{}

func (a AppConfig) Set(config base.AppConfiguration) {
    appConfigsMutex.Lock()
    defer appConfigsMutex.Unlock()
    if _, exists :=  a[config.Name]; !exists {
        a[config.Name] = make(map[base.Version]base.AppConfiguration)
    }
    a[config.Name][config.Version] = config
}

func (a AppConfig) Get(appName base.AppName, version base.Version) base.AppConfiguration {
    appConfigsMutex.Lock()
    defer appConfigsMutex.Unlock()
    return a[appName][version]
}

type StableAppVersions map[base.AppName]map[base.Version]bool
var stableVersionMutex = &sync.Mutex{}

func (s StableAppVersions) Set(appName base.AppName, version base.Version, isStable bool) {
    stableVersionMutex.Lock()
    defer stableVersionMutex.Unlock()
    if _, exists :=  s[appName]; !exists {
        s[appName] = make(map[base.Version]bool)
    }
    s[appName][version] = isStable
}

func (s StableAppVersions) IsStable(appName base.AppName, version base.Version) bool {
    stableVersionMutex.Lock()
    defer stableVersionMutex.Unlock()
    return s[appName][version]
}

func (s StableAppVersions) GetLatestStable(appName base.AppName) base.Version {
    stableVersionMutex.Lock()
    defer stableVersionMutex.Unlock()
    var latestVersion base.Version
    for version, stable := range s[appName] {
        if version > latestVersion && stable {
            latestVersion = version
        }
    }
    return latestVersion
}

type LocalAppStatus map[base.AppName]map[base.Version]base.AppStatus




var StableAppVersionsCache StableAppVersions
var AppConfigCache AppConfig

var MetricsCache base.MetricsWrapper


func parseConfig() {
    configuration.HostId = "somehost"
}

func init() {
    parseConfig()
    StableAppVersionsCache = make(map[base.AppName]map[base.Version]bool)
    AppConfigCache = make(map[base.AppName]map[base.Version]base.AppConfiguration)
    hostInfo = base.HostInfo{
        HostId: configuration.HostId,
        IpAddr: "",
        OsInfo: getOsInfo(),
        Apps: []base.AppInfo{},
    }
    MetricsCache = base.MetricsWrapper{}
    MetricsCache.HostMetrics = make(map[time.Time]base.HostStats)
    MetricsCache.AppMetrics = make(map[base.AppName]map[time.Time]base.AppStats)
}

type pollingFunc func(conf base.AppConfiguration) bool

func commandPollingFunc (conf base.AppConfiguration) bool {
    return executeCommand(conf.QueryStateCommand)
}

func pollAppsStatus() {
    HostLogger.Info("Polling Statuses...")
    for _, app := range hostInfo.Apps {
        HostLogger.Infof("Polling status for app %s:%s", app.Name, app.Version)
        pollAppStatus(app, commandPollingFunc)
    }
}


func pollAppStatus(app base.AppInfo, pollingFunc pollingFunc) {
    conf := AppConfigCache.Get(app.Name, app.Version)
    res := pollingFunc(conf)
    if res {
        app.Status = base.STATUS_RUNNING
        replaceApp(app)
        StableAppVersionsCache.Set(app.Name, app.Version, true)
    }
    if !res && app.Status != base.STATUS_DEPLOYING {
        app.Status = base.STATUS_DEAD
        replaceApp(app)
        StableAppVersionsCache.Set(app.Name, app.Version, false)
    }
    HostLogger.Infof("App %s:%s status is %s", app.Name, app.Version, app.Status)
}

func pollAppsMetrics() {
    HostLogger.Info("Polling Metrics...")
    pollHostMetrics()
    for _, app := range hostInfo.Apps {
        HostLogger.Infof("Polling metrics for app %s:%s", app.Name, app.Version)
        pollAppMetrics(app)
    }
}

func pollHostMetrics() {
    var metrics base.HostStats
    metrics.CpuUsage = 30
    metrics.MemoryUsage = 20
    metrics.NetworkUsage = 10
    MetricsCache.AddHostMetrics(metrics)
}

func pollAppMetrics(app base.AppInfo) {
    var metrics base.AppStats
    metrics.CpuUsage = 3
    metrics.MemoryUsage = 2
    metrics.NetworkUsage = 1
    MetricsCache.AddAppMetrics(app.Name, metrics)
}


func replaceApp(app base.AppInfo) {
    removeApp(app)
    addApp(app)
}


func addApp(app base.AppInfo) {
    HostLogger.Infof("Adding App %s:%s to HostInfo", app.Name, app.Version)
    hostInfoMutex.Lock()
    defer hostInfoMutex.Unlock()
    hostInfo.Apps = append(hostInfo.Apps, app)
}


func removeApp(app base.AppInfo) {
    HostLogger.Infof("Removing App %s from HostInfo", app.Name)
    hostInfoMutex.Lock()
    defer hostInfoMutex.Unlock()
    for i, appObj := range hostInfo.Apps {
        if appObj.Name == app.Name {
            hostInfo.Apps = append(hostInfo.Apps[:i], hostInfo.Apps[i+1:]...)
            return
        }
    }
}

func updateAppState(appName base.AppName, version base.Version, status base.Status) {
    hostInfoMutex.Lock()
    defer hostInfoMutex.Unlock()
    for i, app := range hostInfo.Apps {
        if app.Name == appName && app.Version == version {
            tmp := hostInfo.Apps[i]
            tmp.Status = status
            hostInfo.Apps[i] = tmp
        }
    }
}

func getAppStatus(appName base.AppName, version base.Version) base.Status {
    hostInfoMutex.Lock()
    defer hostInfoMutex.Unlock()
    for _, app := range hostInfo.Apps {
        if app.Name == appName && app.Version == version {
            return app.Status
        }
    }
    return base.STATUS_UNKNOWN
}

func main() {
    file, err := os.Open("/etc/orca/host.conf")
    if err != nil {
        HostLogger.Fatal(err)
    }

    decoder := json.NewDecoder(file)
    if err := decoder.Decode(&configuration); err != nil {
        extra := ""
        if serr, ok := err.(*json.SyntaxError); ok {
            line, col, highlight := util.HighlightBytePosition(file, serr.Offset)
            extra = fmt.Sprintf(":\nError at line %d, column %d (file offset %d):\n%s",
                line, col, serr.Offset, highlight)
        }
        HostLogger.Fatal("error parsing JSON object in config file %s%s\n%v",
            file.Name(), extra, err)
    }
    HostLogger.Info("Host initialized.")
    startSchedule()
}

func getOsInfo() base.OsInfo {
    return base.OsInfo{"", ""}
}



func startSchedule() {
    HostLogger.Info("Starting scheduled tasks")
    trainerTicker := time.NewTicker(10 * time.Second)
    go func () {
        for {
            <- trainerTicker.C
            sendToTrainer()
        }
    }()

    pollTicker := time.NewTicker(2 * time.Second)
    func () {
        for {
            <- pollTicker.C
            pollAppsStatus()
            pollAppsMetrics()

        }
    }()
}



func sendToTrainer() {
    metrics := MetricsCache.Get()
    MetricsCache.Wipe()
    wrapper := base.TrainerPushWrapper{hostInfo, metrics}
    HostLogger.Infof("Sending data to trainer: %+v", wrapper)
    b := new(bytes.Buffer)
    json.NewEncoder(b).Encode(wrapper)
    res, err := http.Post(configuration.TrainerUrl, "application/json; charset=utf-8", b)
    fmt.Println(".....")
    fmt.Println(b.String())
    fmt.Println(".....")
    if err != nil {
        HostLogger.Errorf("Could not send data to trainer: %+v", err)
    } else {
        defer res.Body.Close()
        body, err := ioutil.ReadAll(res.Body)
        if err != nil {
            HostLogger.Errorf("Could not read reponse from trainer: %+v", err)
        } else {
            handleTrainerResponse(body)
        }
    }
}


func handleTrainerResponse(body []byte) {
    var config base.PushConfiguration
    if err := json.Unmarshal(body, &config); err != nil {
        HostLogger.Errorf("Failed to parse response - %s HTTP_BODY: %s", err, string(body))
    } else {
        installApp(config.AppConfiguration, config.DeploymentCount)
    }
}


func installApp(conf base.AppConfiguration, deploymentCount base.DeploymentCount) {
    if string(conf.Name) == "" || string(conf.Version) == "" {
        return
    }
    HostLogger.Infof("Installing App %s:%s", conf.Name, conf.Version)
    status := getAppStatus(conf.Name, conf.Version)
    if status == base.STATUS_DEPLOYING {
        HostLogger.Warnf("App %s:%s is deploying, aborting install", conf.Name, conf.Version)
        return
    }

    if status == base.STATUS_RUNNING {
        HostLogger.Infof("App %s:%s is running, skipping install, trigger run to scale to new deploymentCount %d", conf.Name, conf.Version, deploymentCount)
        runApp(conf, deploymentCount)
        return
    }

    if status == base.STATUS_DEAD {
        HostLogger.Infof("App %s:%s is %s, skipping install, trigger run", conf.Name, conf.Version, status)
        runApp(conf, deploymentCount)
        return
    }

    AppConfigCache.Set(conf)
    appObj := base.AppInfo{
        Type: conf.Type,
        Name: conf.Name,
        Version: conf.Version,
        Status: base.STATUS_DEPLOYING,
    }

    uninstallApp(conf.Name)
    removeApp(appObj)
    doInstallApp(conf, deploymentCount)
}

func uninstallApp(appName base.AppName) {
    for _, app := range hostInfo.Apps {
        if app.Name == appName {
            conf := AppConfigCache.Get(app.Name, app.Version)
            res := executeCommand(conf.RemoveCommand)
            if res {
                HostLogger.Infof("Uninstalled app %s:%s", app.Name, app.Version)
                removeApp(base.AppInfo{Type: app.Type, Name: app.Name, Version: app.Version, Status: base.STATUS_DEAD})
            } else {
                HostLogger.Infof("Uninstall of app %s:%s failed - config: %+v", app.Name, app.Version, conf)
                updateAppState(app.Name, app.Version, base.STATUS_DEAD)
                StableAppVersionsCache.Set(app.Name, app.Version, false)
            }
        }
    }
}

func doInstallApp(appConfig base.AppConfiguration, deploymentCount base.DeploymentCount) {
     for _, command := range appConfig.InstallCommands {
         res := executeCommand(command)
         if !res {
             HostLogger.Errorf("Install of App %s:%s failed - config: %+v", appConfig.Name, appConfig.Version, appConfig)
             updateAppState(appConfig.Name, appConfig.Version, base.STATUS_DEAD)
             StableAppVersionsCache.Set(appConfig.Name, appConfig.Version, false)
             rollbackApp(appConfig, deploymentCount)
             return
         }
         HostLogger.Infof("Install of App %s:%s successful", appConfig.Name, appConfig.Version)
         runApp(appConfig, deploymentCount)
     }
}

func rollbackApp(app base.AppConfiguration, deploymentCount base.DeploymentCount) {
    lastVer := StableAppVersionsCache.GetLatestStable(app.Name)
    if string(lastVer) == "" {
        HostLogger.Warnf("No suitable rollback candidate for %s:%s", app.Name, app.Version)
        return
    }
    HostLogger.Warnf("Initiating rollback of app %s:%s - will install version %s", app.Name, app.Version, lastVer)
    lastConf := AppConfigCache.Get(app.Name, lastVer)
    installApp(lastConf, deploymentCount)
}

func runApp(app base.AppConfiguration, deploymentCount base.DeploymentCount) {
    if deploymentCount == 0 {
        HostLogger.Infof("Got DeploymentCount 0 for App %s:%s, triggering uninstall", app.Name, app.Version)
        uninstallApp(app.Name)
        return
    }
    currentCount := 1
    for _, appObj := range hostInfo.Apps {
        if app.Name == appObj.Name {
            currentCount++
        }
    }

    for i := currentCount; i <= int(deploymentCount); i++ {
        HostLogger.Infof("Starting App %s:%s - iteration: %d of %d", app.Name, app.Version, i, deploymentCount)
        res := executeCommand(app.RunCommand)
        if res {
            updateAppState(app.Name, app.Version, base.STATUS_RUNNING)
            StableAppVersionsCache.Set(app.Name, app.Version, true)
            HostLogger.Infof("App %s:%s started", app.Name, app.Version)
            addApp(base.AppInfo{Type: app.Type, Name: app.Name, Version: app.Version, Status: base.STATUS_RUNNING})
        } else {
            updateAppState(app.Name, app.Version, base.STATUS_DEAD)
            StableAppVersionsCache.Set(app.Name, app.Version, false)
            HostLogger.Infof("App %s:%s start failed", app.Name, app.Version)
        }
    }
}

func executeCommand(command base.OsCommand) bool {
    if command.Type == base.FILE_COMMAND {
        return executeFileCommand(command.Command)
    }
    if command.Type == base.EXEC_COMMAND {
        return executeExecCommand(command.Command)
    }
    return false
}

func executeFileCommand(command base.Command) bool {
    HostLogger.Infof("Writing %d bytes to file %s", len(command.Args), command.Path)
    content := []byte(command.Args)
    err := ioutil.WriteFile(command.Path, content, 0644)
    if err != nil {
        HostLogger.Errorf("Writing to %s failed - %s", command.Path, err)
        return false
    }
    HostLogger.Infof("Writing to %s complete.", command.Path)
    return true
}

func executeExecCommand(command base.Command) bool {
    var cmd *exec.Cmd
    if command.Args == "" {
        cmd = exec.Command(command.Path, command.Args)
    } else {
        cmd = exec.Command(command.Path, strings.Fields(command.Args)...)
    }
    HostLogger.Infof("Will execute: %s %s", command.Path, command.Args)
    output, err := cmd.CombinedOutput()
    if err != nil {
        HostLogger.Error(err)
        HostLogger.Error(string(output))
        return false
    }
    HostLogger.Info(string(output))
    return true
}

type AppState struct {
    Version base.Version
    Status base.AppStatus
}

type HabitatState struct {
    Version base.Version
    Status base.HabitatStatus
}

type AppLayout struct {
    HabitatState HabitatState
    Apps map[base.HostId]AppState
}
