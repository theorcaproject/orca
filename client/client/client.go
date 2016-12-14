/*
Copyright Alex Mack
This file is part of Orca.

Orca is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Orca is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Orca.  If not, see <http://www.gnu.org/licenses/>.
*/

package client

import (
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/base"
	"gatoor/orca/client/docker"
	"gatoor/orca/client/raw"
	"gatoor/orca/client/types"
	"gatoor/orca/client/testClient"
)


var ClientLogger = Logger.LoggerWithField(Logger.Logger, "module", "client")

var AppsState types.AppsState
var AppsConfiguration types.AppsConfiguration
var AppsMetricsById types.AppsMetricsById
//var RetryCounter *types.RetryCounter
var Configuration types.Configuration
var cli Client

type Client interface {
	Type() types.ClientType
	Init()

	InstallApp(base.AppConfiguration, *types.AppsState, *types.Configuration) bool
	RunApp(base.AppId, base.AppConfiguration, *types.AppsState, *types.Configuration) bool
	QueryApp(base.AppId, base.AppConfiguration, *types.AppsState, *types.Configuration) bool
	StopApp(base.AppId, base.AppConfiguration, *types.AppsState, *types.Configuration) bool
	DeleteApp(base.AppConfiguration, *types.AppsState, *types.Configuration) bool
	//
	//HostMetrics()
	AppMetrics(base.AppId, base.AppConfiguration, *types.AppsState, *types.Configuration, *types.AppsMetricsById) bool
}

func Init() {
	ClientLogger.Info("Initializing Client...")
	if Configuration.Type == types.DOCKER_CLIENT {
		cli = &docker.Client{}
	} else if Configuration.Type == types.RAW_CLIENT {
		cli = &raw.Client{}
	} else {
		cli = &testClient.Client{}
	}
	AppsState = make(map[base.AppId]base.AppInfo)
	AppsConfiguration = make(map[base.AppId]base.AppConfiguration)
	AppsMetricsById = make(map[base.AppId]map[string]base.AppStats)
	ClientLogger.Infof("Initialized Client of Type %s", cli.Type())
}

func Handle(config base.PushConfiguration) {
	ClientLogger.Infof("Received Configuration from trainer: %+v", config)
	if config.AppConfiguration.Name == base.AppName("") {
		ClientLogger.Infof("Configuration is empty. Skipping")
		return
	}
	existingAppsVersion := AppsState.GetAllWithVersion(config.AppConfiguration.Name, config.AppConfiguration.Version)
	if len(existingAppsVersion) > 0 {
		if len(existingAppsVersion) != int(config.DeploymentCount) {
			ClientLogger.Infof("Configuration for existing app version %s:%d DeploymentCount new %d; old %s", config.AppConfiguration.Name, config.AppConfiguration.Version, config.DeploymentCount, len(existingAppsVersion))
			scaleApp(existingAppsVersion, config)
		}
		ClientLogger.Infof("Configuration for existing app version %s:%d DeploymentCount %d matches, skipping", config.AppConfiguration.Name, config.AppConfiguration.Version, config.DeploymentCount)
	} else {
		existingApps := AppsState.GetAll(config.AppConfiguration.Name)
		if len(existingApps) > 0 {
			ClientLogger.Infof("Configuration for different version of app %s. From %d to %d", config.AppConfiguration.Name, existingApps[0].Version, config.AppConfiguration.Version)
			if !updateApp(existingApps, config) {
				 ClientLogger.Warnf("Update of app %s from %d to %d failed.", config.AppConfiguration.Name, existingApps[0].Version, config.AppConfiguration.Version)
				 if !rollbackApp(AppsConfiguration.Get(existingApps[0].Id), config.DeploymentCount) {
				 	 ClientLogger.Warnf("Rollback of app %s from %d to %d failed. Doing nothing", config.AppConfiguration.Name, existingApps[0].Version, config.AppConfiguration.Version)
					 return
				 }
			}
		} else {
			newApp(config)
		}
	}
	ClientLogger.Infof("Applied Configuration from trainer")
}

func newApp(config base.PushConfiguration) bool {
	return installAndRun(config)
}

func rollbackApp(config base.AppConfiguration, count base.DeploymentCount) bool {
	ClientLogger.Infof("Starting rollback of app %s to %d", config.Name, config.Version)
	pushConf := base.PushConfiguration{DeploymentCount: count, AppConfiguration: config}
	StopAll(config.Name)
	DeleteApp(pushConf)
	res := installAndRun(pushConf)
	ClientLogger.Infof("Finished rollback of app %s to %d. Success=%t", config.Name, config.Version, res)
	return res
}

func updateApp(existingApps []base.AppInfo, config base.PushConfiguration) bool {
	ClientLogger.Infof("Starting update app %s:%d to %d", config.AppConfiguration.Name, existingApps[0].Version, config.AppConfiguration.Version)
	for _, app := range existingApps {
		StopApp(app.Id)
	}
	DeleteApp(config)
	res := installAndRun(config)
	ClientLogger.Infof("Finished update app %s:%d to %d. Success=%t", config.AppConfiguration.Name, existingApps[0].Version, config.AppConfiguration.Version, res)
	return res
}

func installAndRun(config base.PushConfiguration) bool {
	if !InstallApp(config.AppConfiguration) {
		ClientLogger.Infof("Install of app %s:%d failed, skipping run", config.AppConfiguration.Name, config.AppConfiguration.Version)
		return false
	}
	res := true
	for i := 0; i < int(config.DeploymentCount); i++ {
		if !RunApp(config.AppConfiguration) {
			res = false
		}
	}
	return res
}

func scaleApp(existingApps []base.AppInfo, config base.PushConfiguration) bool {
	 if len(existingApps) > int(config.DeploymentCount) {
		 return scaleDown(existingApps, config)
	 } else {
		 return scaleUp(existingApps, config)
	 }
}

func scaleDown(existingApps []base.AppInfo, config base.PushConfiguration) bool {
	ClientLogger.Infof("Starting scale down of app %s:%d", config.AppConfiguration.Name, config.AppConfiguration.Version)
	stopped := 0
	res := true
	for _, app := range existingApps {
		if stopped < (len(existingApps) - int(config.DeploymentCount)) {
			if !StopApp(app.Id) {
				res = false
			}
			stopped++
		}
	}
	if config.DeploymentCount == 0 {
		DeleteApp(config)
	}
	ClientLogger.Infof("Finished scale down of app %s:%d. Success=%t", config.AppConfiguration.Name, config.AppConfiguration.Version, res)
	return res
}

func scaleUp(existingApps []base.AppInfo, config base.PushConfiguration) bool {
	ClientLogger.Infof("Starting scale up of app %s:%d", config.AppConfiguration.Name, config.AppConfiguration.Version)
	res := true
	for i := 0; i < int(config.DeploymentCount) - len(existingApps); i++ {
		if !RunApp(config.AppConfiguration) {
			res = false
		}
	}
	ClientLogger.Infof("Finished scale up of app %s:%d. Success=%t", config.AppConfiguration.Name, config.AppConfiguration.Version, res)
	return res
}

func InstallApp(conf base.AppConfiguration) bool {
	ClientLogger.Infof("Starting Install of app %s:%d", conf.Name, conf.Version)
	res := cli.InstallApp(conf, &AppsState, &Configuration)
	ClientLogger.Infof("Finished Install of app %s:%d done. Success=%t", conf.Name, conf.Version, res)
	return res
}

func RunApp(conf base.AppConfiguration) bool {
	ClientLogger.Infof("Starting app %s:%d", conf.Name, conf.Version)
	id := types.GenerateId(conf.Name)
	info := base.AppInfo{Name: conf.Name, Type: conf.Type, Version: conf.Version, Id: id, Status: base.STATUS_DEPLOYING}
	AppsConfiguration.Add(id, conf)
	AppsState.Add(id, info)
	res := cli.RunApp(id, conf, &AppsState, &Configuration)
	if !res {
		AppsState.Set(id, base.STATUS_DEAD)
	}
	ClientLogger.Infof("Starting app %s:%d done. Success=%t", conf.Name, conf.Version, res)
	return res
}

func StopApp(id base.AppId) bool {
	conf := AppsConfiguration.Get(id)
	ClientLogger.Infof("Stopping app %s (%s:%d)", id, conf.Name, conf.Version)
	AppsState.Set(id, base.STATUS_DEAD)
	res := cli.StopApp(id, conf, &AppsState, &Configuration)
	ClientLogger.Infof("Stopping app %s (%s:%d) done. Success=%t", id, conf.Name, conf.Version, res)
	return res
}

func QueryApp(id base.AppId) bool {
	ClientLogger.Debugf("Ouery app state of %s", id)
	res := cli.QueryApp(id, AppsConfiguration.Get(id), &AppsState, &Configuration)
	var status base.Status
	if res {
		status = base.STATUS_RUNNING
	} else {
		status = base.STATUS_DEAD
	}
	AppsState.Set(id, status)
	ClientLogger.Debugf("Ouery app state of %s done: %s, Success=%t", id, status, res)
	return res
}

func StopAll(appName base.AppName) bool {
	apps := AppsState.GetAll(appName)
	res := true
	for _, app := range apps {
		if !StopApp(app.Id) {
			res = false
		}
	}
	return res
}

func DeleteApp(config base.PushConfiguration) bool {
	ClientLogger.Infof("Starting deletion of app %s", config.AppConfiguration.Name)
	apps := AppsState.GetAll(config.AppConfiguration.Name)
	for _, app := range apps {
		AppsState.Remove(app.Id)
	}
	res := cli.DeleteApp(config.AppConfiguration, &AppsState, &Configuration)
	ClientLogger.Infof("Finished deletion of app %s. Success=%t", config.AppConfiguration.Name, res)
	return res
}


func PollAppsState() {
	ClientLogger.Debug("Starting App Poll")
	for _, app := range AppsState.All() {
		QueryApp(app.Id)
	}
	ClientLogger.Debug("Finished App Poll")
}

func AppMetrics(id base.AppId) bool {
	ClientLogger.Debugf("Query Metrics of app %s", id)
	res := cli.AppMetrics(id, AppsConfiguration.Get(id), &AppsState, &Configuration, &AppsMetricsById)
	ClientLogger.Debugf("Query Metrics of app %s done. Success=%t", id, res)
	return res
}

func PollMetrics() {
	ClientLogger.Debug("Starting Metrics Poll")
	for _, app := range AppsState.All() {
		AppMetrics(app.Id)
	}
	ClientLogger.Debug("Finished Metrics Poll")
}


func generateCombinedMetrics() base.AppMetrics {
	combined := base.AppMetrics{}
	metrics := AppsMetricsById.All()
	ClientLogger.Infof("inner metrics %+v", metrics)
	for id, metricsbyTime := range metrics {
		app := AppsState.Get(id)
		for time, metrics := range metricsbyTime {
			combined.Add(app.Name, app.Version, time, base.AppStats{CpuUsage: metrics.CpuUsage, MemoryUsage: metrics.MemoryUsage, NetworkUsage: metrics.NetworkUsage, ResponsePerformance: metrics.ResponsePerformance})
		}
	}
	AppsMetricsById.Clear()
	return combined
}

func GetAppMetrics() base.AppMetrics {
	return generateCombinedMetrics()
}