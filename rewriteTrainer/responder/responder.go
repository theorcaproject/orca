package responder

import (
	"gatoor/orca/base"
	"gatoor/orca/rewriteTrainer/planner"
	"errors"
	"gatoor/orca/rewriteTrainer/state/configuration"
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/rewriteTrainer/tracker"
	"fmt"
)

var ResponderLogger = Logger.LoggerWithField(Logger.Logger, "module", "responder")


func GetConfigForHost(hostId base.HostId) (base.PushConfiguration, error) {
	ResponderLogger.Infof("Getting config for host %s", hostId)
	appName, elem, err := getQueueElement(hostId)
	if err == nil {
		config, err := state_configuration.GlobalConfigurationState.GetApp(appName, elem.Version.Version)
		if err != nil {
			ResponderLogger.Warnf("Getting config for host %s app %s failed %s", hostId, appName, err)
			return base.PushConfiguration{}, errors.New("GlobalConfiguraitonState does not have this app")
		}
		return base.PushConfiguration{DeploymentCount: elem.Version.DeploymentCount, AppConfiguration: config}, nil
	}
	ResponderLogger.Warnf("Getting config for host %s failed %s", hostId, err)
	return base.PushConfiguration{}, err
}


func getQueueElement(hostId base.HostId) (base.AppName, planner.AppsUpdateState, error)  {
	ResponderLogger.Infof("Getting queue element for host %s", hostId)
	apps, err := planner.Queue.Get(hostId)
	if err == nil {
		for appName, appObj := range apps {
			if appObj.State == planner.STATE_APPLYING {
				ResponderLogger.Infof("Got STATE_APPLYING queue element for host '%s' app '%s'", hostId, appName)
				return appName, appObj, nil
			}
		}

		for appName, appObj := range apps {
			if appObj.State == planner.STATE_QUEUED {
				ResponderLogger.Infof("Got STATE_QUEUED queue element for host '%s' app '%s'", hostId, appName)
				planner.Queue.SetState(hostId, appName, planner.STATE_APPLYING)
				return appName, appObj, nil
			}
		}
	}
	ResponderLogger.Infof("Get queue element for host '%s' failed", hostId)
	return "", planner.AppsUpdateState{}, errors.New("No element for host ")
}


func CheckAppState(hostInfo base.HostInfo) {
	ResponderLogger.Infof("checking Apps state for host '%s'", hostInfo.HostId)
	queued, _ := planner.Queue.Get(hostInfo.HostId)

	for _, appObj := range hostInfo.Apps {
		if queuedState, exists := queued[appObj.Name]; exists { //the app is queued for an update
			if queuedState.State == planner.STATE_QUEUED { //the update is waiting for other updates, perform simple check
				simpleAppCheck(appObj, hostInfo.HostId)
			} else {
				fmt.Println("CHECK APP STATE")
				checkAppUpdate(appObj, hostInfo.HostId, queuedState)
			}
		} else { //no updates, just check if its running
			simpleAppCheck(appObj, hostInfo.HostId)
		}
	}
}


func simpleAppCheck(appObj base.AppInfo, hostId base.HostId) {
	if appObj.Status != base.STATUS_RUNNING {
		ResponderLogger.Warnf("App '%s' - '%s' on host '%s' is not running. Adding it to GlobalAppCrashes", appObj.Name, appObj.Version, hostId)
		tracker.GlobalAppsStatusTracker.Update(hostId, appObj.Name, appObj.Version, tracker.APP_EVENT_CRASH)
	} else {
		tracker.GlobalAppsStatusTracker.Update(hostId, appObj.Name, appObj.Version, tracker.APP_EVENT_CHECKIN)
	}
}

func checkAppUpdate(appObj base.AppInfo, hostId base.HostId, queuedState planner.AppsUpdateState) {
	ResponderLogger.Infof("Check update of App '%s' - '%s' on host '%s'", appObj.Name, appObj.Version, hostId)
	if queuedState.State != planner.STATE_APPLYING {
		ResponderLogger.Errorf("Got illegal state %s for update of App '%s' - '%s' on host '%s'", queuedState.State, appObj.Name, appObj.Version, hostId)
		return
	}
	if appObj.Status == base.STATUS_RUNNING {
		if appObj.Version == queuedState.Version.Version {
			ResponderLogger.Infof("Update of App '%s' - '%s' on host '%s' successful", appObj.Name, appObj.Version, hostId)
			handleSuccessfulUpdate(appObj.Name, appObj.Version)
		} else {
			ResponderLogger.Warnf("Update of App '%s' - '%s' on host '%s' rolled back to version %s", appObj.Name, queuedState.Version, hostId, appObj.Version)
			handleRollback(hostId, appObj.Name, queuedState.Version.Version)
		}
	}
	if appObj.Status == base.STATUS_DEPLOYING {
		if appObj.Version == queuedState.Version.Version {
			ResponderLogger.Infof("Update of App '%s' - '%s' on host '%s' is still applying", appObj.Name, appObj.Version, hostId)
			return
		} else {
			ResponderLogger.Warnf("Update of App '%s' - '%s' on host '%s' rolling back to version %s", appObj.Name, queuedState.Version, hostId, appObj.Version)
			handleRollback(hostId, appObj.Name, queuedState.Version.Version)
		}
	}
	if appObj.Status == base.STATUS_DEAD {
		ResponderLogger.Warnf("Update of App '%s' - '%s' on host '%s' was fatal for the app, the version that died on the host is %s", appObj.Name, queuedState.Version, hostId, appObj.Version)
		handleFatalUpdate(hostId, appObj.Name, appObj.Version)
	}
}

//TODO ADD AS STABLE VERSION remove from queue for host
func handleSuccessfulUpdate(appName base.AppName, version base.Version) {
	tracker.GlobalAppsStatusTracker.Update("", appName, version, tracker.APP_EVENT_SUCCESSFUL_UPDATE)
}


func handleRollback(hostId base.HostId, appName base.AppName, failedVersion base.Version) {
	fmt.Println("ROLLBACK")
	fmt.Println(hostId)
	fmt.Println(appName)
	fmt.Println(failedVersion)
	fmt.Println("ROLLBACK")
	tracker.GlobalAppsStatusTracker.Update(hostId, appName, failedVersion, tracker.APP_EVENT_ROLLBACK)
	planner.Queue.RemoveApp(appName, failedVersion)
}

func handleFatalUpdate(hostId base.HostId, appName base.AppName, version base.Version) {
	tracker.GlobalAppsStatusTracker.Update(hostId, appName, version, tracker.APP_EVENT_CRASH)
}

