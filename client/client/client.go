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
	"gatoor/orca/client/types"
)


var ClientLogger = Logger.LoggerWithField(Logger.Logger, "module", "client")

var AppsState types.AppsState
var AppsConfiguration types.AppsConfiguration
var AppsMetricsById types.AppsMetricsById
var Configuration types.Configuration
var cli Client

type Client interface {
	Type() types.ClientType
	Init()

	InstallApp(base.AppName, base.AppConfigurationSet, *types.AppsState, *types.Configuration) bool
	RunApp(base.AppId, base.AppConfigurationSet, *types.AppsState, *types.Configuration) bool
	QueryApp(base.AppId, base.AppConfigurationSet, *types.AppsState, *types.Configuration) bool
	StopApp(base.AppId, base.AppConfigurationSet, *types.AppsState, *types.Configuration) bool
	DeleteApp(base.AppConfigurationSet, *types.AppsState, *types.Configuration) bool
	//
	//HostMetrics()
	AppMetrics(base.AppId, base.AppConfigurationSet, *types.AppsState, *types.Configuration, *types.AppsMetricsById) bool
}

func Init() {
	ClientLogger.Info("Initializing Client...")
	if Configuration.Type == types.DOCKER_CLIENT {
		cli = &docker.Client{}
	}
	AppsState = make(map[base.AppId]base.AppInfo)
	AppsConfiguration = make(map[base.AppId]base.AppConfigurationSet)
	AppsMetricsById = make(map[base.AppId]map[string]base.AppStats)
	ClientLogger.Infof("Initialized Client of Type %s", cli.Type())
}

func Handle(changes []base.ChangeRequest) {
	for _, change := range changes {
		if change.ChangeType == base.UPDATE_TYPE__ADD {
			newApp(change.Application, change.AppConfig)
		}

		if change.ChangeType == base.UPDATE_TYPE__REMOVE {
			//TODO
		}
	}
}

func newApp(name base.AppName, config base.AppConfigurationSet) bool {
	return installAndRun(name, config)
}

func installAndRun(name base.AppName, config base.AppConfigurationSet) bool {
	if !InstallApp(name, config) {
		ClientLogger.Infof("Install of app %s:%d failed, skipping run", config.DockerConfig.Repository, config.Version)
		return false
	}
	return RunApp(name, config)
}

func InstallApp(name base.AppName, conf base.AppConfigurationSet) bool {
	ClientLogger.Infof("Starting Install of app %s:%d", conf.DockerConfig.Repository, conf.Version)
	res := cli.InstallApp(name, conf, &AppsState, &Configuration)
	ClientLogger.Infof("Finished Install of app %s:%d done. Success=%t", conf.DockerConfig.Repository, conf.Version, res)
	return res
}

func RunApp(name base.AppName, conf base.AppConfigurationSet) bool {
	ClientLogger.Infof("Starting app %s:%d", conf.DockerConfig.Repository, conf.Version)
	id := types.GenerateId(name)
	info := base.AppInfo{Name: name, Type: base.APP_HTTP, Version: conf.Version, Id: id, Status: base.STATUS_DEPLOYING}
	AppsConfiguration.Add(id, conf)
	AppsState.Add(id, info)
	res := cli.RunApp(id, conf, &AppsState, &Configuration)
	if !res {
		AppsState.Set(id, base.STATUS_DEAD)
	}
	ClientLogger.Infof("Starting app %s:%d done. Success=%t", name, conf.Version, res)
	return res
}

func StopApp(id base.AppId) bool {
	conf := AppsConfiguration.Get(id)
	ClientLogger.Infof("Stopping app %s", id)
	AppsState.Set(id, base.STATUS_DEAD)
	res := cli.StopApp(id, conf, &AppsState, &Configuration)
	ClientLogger.Infof("Stopping app %s done. Success=%t", id, res)
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

func DeleteApp(name base.AppName, config base.AppConfigurationSet) bool {
	ClientLogger.Infof("Starting deletion of app %s", name)
	apps := AppsState.GetAll(name)
	for _, app := range apps {
		AppsState.Remove(app.Id)
	}
	res := cli.DeleteApp(config, &AppsState, &Configuration)
	ClientLogger.Infof("Finished deletion of app %s. Success=%t", name, res)
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