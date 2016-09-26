package main

import (
    "os"
    "encoding/json"
    "time"
    "fmt"
    "bytes"
    "net/http"
    linuxproc "github.com/c9s/goprocinfo/linux"
    "gatoor/orca/util"
    base "gatoor/orca/base"
    log "gatoor/orca/base/log"
    "io/ioutil"
    "os/exec"
    "strings"
)

type Configuration struct {
    PollInterval int
    TrainerUrl string
    HostId string
}

var configuration Configuration

var hostInfo base.HostInfo

var Logger = log.Logger

var HabitatInstallLogger = log.LoggerWithField(Logger, "Stage", "HabitatInstall")
var AppInstallBaseLogger = log.LoggerWithField(Logger, "Stage", "AppInstall")
var UpdateLogger = log.LoggerWithField(Logger, "Stage", "Update")

var CmdLogger = log.LoggerWithField(log.AuditLogger, "Type", "cmd")

func main() {
    file, err := os.Open("/etc/orca/host.conf")
    if err != nil {
        Logger.Fatal(err)
    }

    decoder := json.NewDecoder(file)
    if err := decoder.Decode(&configuration); err != nil {
        extra := ""
        if serr, ok := err.(*json.SyntaxError); ok {
            line, col, highlight := util.HighlightBytePosition(file, serr.Offset)
            extra = fmt.Sprintf(":\nError at line %d, column %d (file offset %d):\n%s",
                line, col, serr.Offset, highlight)
        }
        Logger.Fatal("error parsing JSON object in config file %s%s\n%v",
            file.Name(), extra, err)
    }
    initHost()
    Logger.Info("Host initialized.")
    poll()
}

func getOsInfo() base.OsInfo {
    return base.OsInfo{"", ""}
}

func initHost() {
    hostInfo.Id = configuration.HostId
    hostInfo.HabitatInfo.Version  = "0"
    hostInfo.HabitatInfo.Status = base.STATUS_INIT
    hostInfo.OsInfo = getOsInfo()
    hostInfo.Apps = make(map[string]base.AppInfo)
}

func poll() {
    Logger.Info("Host starts polling.")
    for {
        //stat, err := linuxproc.ReadStat("/proc/stat")
        //if err != nil {
        //    Logger.Warn("stat read fail")
        //sendStats(stat)
        queryAppsState()
        logCurrentLayout()
        sendStats(nil)
        logCurrentLayout()
        time.Sleep(time.Second * time.Duration(configuration.PollInterval))
    }
}

func sendStats(stat *linuxproc.Stat) {
    wrapper := base.StatsWrapper{hostInfo, stat}
    b := new(bytes.Buffer)
    json.NewEncoder(b).Encode(wrapper)
    res, err := http.Post(configuration.TrainerUrl, "application/json; charset=utf-8", b)
    Logger.Info(fmt.Sprintf("Sent stats to %s", configuration.TrainerUrl))
    if err != nil {
        Logger.Error(fmt.Sprintf("Failed to send stats - %s", err))
    } else {
        defer res.Body.Close()
        body, err := ioutil.ReadAll(res.Body)
        if err != nil {
            Logger.Error(fmt.Sprintf("Failed to get response - %s", err))
        } else {
            handleUpdate(body)
        }
    }
}

func handleUpdate(body []byte) {
    var trainerUpdate base.TrainerUpdate
    if err := json.Unmarshal(body, &trainerUpdate); err != nil {
        Logger.Error(fmt.Sprintf("Failed to parse response - %s HTTP_BODY: %s", err, string(body)))
    } else {
        if trainerUpdate.TargetHostId != hostInfo.Id {
            UpdateLogger.Error(fmt.Sprintf("Received incorrect update, HostId is %s but received TargetId %s", hostInfo.Id, trainerUpdate.TargetHostId))
            return
        }
        UpdateLogger.Info("Starting update...")
        handleHabitatUpdate(trainerUpdate.HabitatConfiguration)
        if hostInfo.HabitatInfo.Status == base.STATUS_RUNNING {
            for _, val := range trainerUpdate.AppsConfiguration {
                handleAppUpdate(val)
            }
            removeApps(trainerUpdate.AppsConfiguration)
        } else {
            UpdateLogger.Info(fmt.Sprintf("Habitat is in %s state, skipping Apps update.", hostInfo.HabitatInfo.Status))
        }
    }
}

func handleHabitatUpdate(target base.HabitatConfiguration) bool {
    if target.Version == hostInfo.HabitatInfo.Version {
        HabitatInstallLogger.Info(fmt.Sprintf("Received same HabitatVersion  %s. Nothing to do here", target.Version))
        return false
    }
    if hostInfo.HabitatInfo.Status == base.STATUS_DEPLOYING {
        HabitatInstallLogger.Info(fmt.Sprintf("Habitat is in DEPLOYING state. Skipping", target.Version))
        return false
    }
    if target.Version > hostInfo.HabitatInfo.Version  {
        HabitatInstallLogger.Info(fmt.Sprintf("Received new HabitatVersion  %s. Current HabitatVersion  is %s.", target.Version, hostInfo.HabitatInfo.Version))
        return installHabitat(target)
    }
    return false
}

func handleAppUpdate(target base.AppConfiguration) bool {
    AppInstallLogger := log.LoggerWithField(AppInstallBaseLogger, "AppName", target.AppName)
    _, exists := hostInfo.Apps[target.AppName]
    if !exists {
        AppInstallLogger.Info(fmt.Sprintf("Received new App %s with AppVersion %s", target.AppName, target.Version))
        return installApp(target)
    }
    if hostInfo.Apps[target.AppName].Status == base.STATUS_DEPLOYING {
        AppInstallLogger.Info(fmt.Sprintf("App is in DEPLOYING state. Skipping.", target.Version))
        return false
    }
    if target.Version == hostInfo.Apps[target.AppName].CurrentVersion {
        if hostInfo.Apps[target.AppName].Status == base.STATUS_DEAD {
            AppInstallLogger.Info(fmt.Sprintf("Received same AppVersion %s. App is in DEAD state. Redeploying...", target.Version))
            return installApp(target)
        }
        AppInstallLogger.Info(fmt.Sprintf("Received same AppVersion %s. Nothing to do here", target.Version))
        return false
    }
    if target.Version > hostInfo.Apps[target.AppName].CurrentVersion {
        AppInstallLogger.Info(fmt.Sprintf("Received new AppVersion %s. Current AppVersion is %s.", target.Version, hostInfo.Apps[target.AppName].CurrentVersion))
        return installApp(target)
    }
    return false
}

func installHabitat(conf base.HabitatConfiguration) bool {
    if hostInfo.HabitatInfo.Status == base.STATUS_DEPLOYING || hostInfo.HabitatInfo.Status == base.STATUS_RUNNING {
        HabitatInstallLogger.Info(fmt.Sprintf("Habitat Status is %s, aborting install", hostInfo.HabitatInfo.Status))
    }

    HabitatInstallLogger.Info(fmt.Sprintf("Starting install of HabitatInfo.Version  %s", conf.Version))
    hostInfo.HabitatInfo.Status = base.STATUS_DEPLOYING
    for _, command := range conf.Commands {
        res := executeCommand(command)
        if !res {
            HabitatInstallLogger.Error(fmt.Sprintf("Install of HabitatInfo.Version  %s failed", conf.Version))
            hostInfo.HabitatInfo.Status = base.STATUS_DEAD
            return false
        }
    }
    HabitatInstallLogger.Info(fmt.Sprintf("Install of HabitatInfo.Version  %s successful", conf.Version))
    hostInfo.HabitatInfo.Status = base.STATUS_RUNNING
    hostInfo.HabitatInfo.Version = conf.Version
    return true
}

func installApp(conf base.AppConfiguration) bool {
    AppInstallLogger := log.LoggerWithField(AppInstallBaseLogger, "AppName", conf.AppName)
    if hostInfo.Apps[conf.AppName].Status == base.STATUS_DEPLOYING || hostInfo.Apps[conf.AppName].Status == base.STATUS_RUNNING {
        AppInstallLogger.Info(fmt.Sprintf("App Status is %s, aborting install", hostInfo.Apps[conf.AppName].Status))
    }
    AppInstallLogger.Info(fmt.Sprintf("Starting install of AppVersion %s", conf.Version))
    tempApp := hostInfo.Apps[conf.AppName]
    tempApp.Status = base.STATUS_DEPLOYING
    tempApp.Name = conf.AppName
    tempApp.Type = conf.AppType
    tempApp.CurrentVersion = conf.Version
    tempApp.QueryStateCommand = conf.QueryStateCommand
    tempApp.RemoveCommand = conf.RemoveCommand
    hostInfo.Apps[conf.AppName] = tempApp
    for _, command := range conf.InstallCommands {
        res := executeCommand(command)
        if !res {
            AppInstallLogger.Error(fmt.Sprintf("Install of AppVersion %s failed", conf.Version))
            tempApp.Status = base.STATUS_DEAD
            hostInfo.Apps[conf.AppName] = tempApp
            return false
        }
    }
    AppInstallLogger.Info(fmt.Sprintf("Install of AppVersion %s successful", conf.Version))
    tempApp.Status = base.STATUS_RUNNING
    hostInfo.Apps[conf.AppName] = tempApp
    return true
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
    CmdLogger.Info(fmt.Sprintf("Writing %d bytes to file %s", len(command.Args), command.Path))
    content := []byte(command.Args)
    err := ioutil.WriteFile(command.Path, content, 0644)
    if err != nil {
        CmdLogger.Error(fmt.Sprintf("Writing to %s failed - %s", command.Path, err))
        return false
    }
    CmdLogger.Info(fmt.Sprintf("Writing to %s complete.", command.Path))
    return true
}

func executeExecCommand(command base.Command) bool {
    var cmd *exec.Cmd
    if command.Args == "" {
        cmd = exec.Command(command.Path, command.Args)
    } else {
        cmd = exec.Command(command.Path, strings.Fields(command.Args)...)
    }
    CmdLogger.Info(fmt.Sprintf("%s %s", command.Path, command.Args))
    output, err := cmd.CombinedOutput()
    if err != nil {
        CmdLogger.Error(err)
        CmdLogger.Error(string(output))
        return false
    }
    CmdLogger.Info(string(output))
    return true
}


func queryAppsState() {
    for appName := range hostInfo.Apps {
        var AppLogger = log.LoggerWithField(AppInstallBaseLogger, "AppName", appName)
        tempApp := hostInfo.Apps[appName]
        if tempApp.Status == base.STATUS_DEAD || tempApp.Status == base.STATUS_RUNNING {
            AppLogger.Info("Querying App state...")
            AppLogger.Info(fmt.Printf("%+v", tempApp))
            res := executeCommand(tempApp.QueryStateCommand)
            if res {
                tempApp.Status = base.STATUS_RUNNING
            } else {
                tempApp.Status = base.STATUS_DEAD
            }
            hostInfo.Apps[appName] = tempApp
            AppLogger.Info(fmt.Sprintf("Updating App state to %s.", tempApp.Status))
        }
    }
}

func removeApps(conf map[string]base.AppConfiguration) {
    for appName := range hostInfo.Apps {
        _, exists := conf[appName]
        if !exists {
            var AppLogger = log.LoggerWithField(AppInstallBaseLogger, "AppName", appName)
            AppLogger.Info(fmt.Sprintf("Removing App %s from Host...", appName))
            executeCommand(hostInfo.Apps[appName].RemoveCommand)
            delete(hostInfo.Apps, appName)
            AppLogger.Info(fmt.Sprintf("Removed App %s from Host.", appName))
        }
    }
}

type AppState struct {
    Version string
    Status base.AppStatus
}

type HabitatState struct {
    Version string
    Status base.HabitatStatus
}

type AppLayout struct {
    HabitatState HabitatState
    Apps map[string]AppState
}



func logCurrentLayout() {
    Logger.Info(fmt.Printf("Current Layout: %+v", getCurrentLayout()))
}

func getCurrentLayout() AppLayout {
    layout := AppLayout{
        HabitatState{hostInfo.HabitatInfo.Version, hostInfo.HabitatInfo.Status},
        make(map[string]AppState),
    }
    for appName := range hostInfo.Apps {
        layout.Apps[appName] = AppState{hostInfo.Apps[appName].CurrentVersion, hostInfo.Apps[appName].Status,}
    }
    return layout
}