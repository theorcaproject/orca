package main

import (
	"net/http"
	"gatoor/orca/base"
	"encoding/json"
	"fmt"
	"gatoor/orca/modules/cloud"
)

func sendConfig(w http.ResponseWriter) {
	var trainerUpdate base.TrainerUpdate
	trainerUpdate.TargetHostId = "TODOcalfId"
	trainerUpdate.HabitatConfiguration = buildHabitatConfiguration()
	trainerUpdate.AppsConfiguration = make(map[string]base.AppConfiguration)
	trainerUpdate.AppsConfiguration["ngin"] = buildAppConfiguration()
	json.NewEncoder(w).Encode(trainerUpdate)
	updateDesiredCloudLayout(trainerUpdate)
}

func buildHabitatConfiguration() base.HabitatConfiguration {
	conf := base.HabitatConfiguration{
		Version: "4",
		Commands: []base.OsCommand {
			{base.EXEC_COMMAND, base.Command{"apt-get", "update"},},
			{base.EXEC_COMMAND, base.Command{"apt-get", "-y install apt-transport-https ca-certificates"},},
			{base.EXEC_COMMAND, base.Command{"apt-key", "adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D"},},
			{base.FILE_COMMAND, base.Command{"/etc/apt/sources.list.d/docker.list", "deb https://apt.dockerproject.org/repo ubuntu-xenial main"},},
			{base.EXEC_COMMAND, base.Command{"apt-get", "update"},},
			{base.EXEC_COMMAND, base.Command{"apt-get", "-y install docker-engine"},},
		},
	}
	return conf
}

func buildAppConfiguration() base.AppConfiguration {
	conf := base.AppConfiguration{}
	conf.Version = "2"
	conf.AppName = "ngin"
	conf.AppType = base.APP_HTTP
	conf.InstallCommands = []base.OsCommand {
		{base.EXEC_COMMAND, base.Command{"echo", "MORE AMAZING"},},
	}
	conf.QueryStateCommand = base.OsCommand{base.EXEC_COMMAND, base.Command{"echo", "query for state"},}
	conf.RemoveCommand = base.OsCommand{base.EXEC_COMMAND, base.Command{"echo", "REMOVE"},}
	Logger.Info(fmt.Printf("%+v", conf))
	return conf
}

func updateDesiredCloudLayout(update base.TrainerUpdate) {
	if _, exists := cloudLayoutDesired[update.TargetHostId]; !exists {
		cloudLayoutDesired[update.TargetHostId] = cloud.CloudLayoutElement{"", make(map[string]string),}
	}
	elem := cloudLayoutDesired[update.TargetHostId]
	elem.HabitatVersion = update.HabitatConfiguration.Version
	for _, appConf := range update.AppsConfiguration {
		elem.AppsVersion[appConf.AppName] = appConf.Version
	}
	cloudLayoutDesired[update.TargetHostId] = elem
}

func updateCurrentCloudLayout(hostInfo base.HostInfo) {
	if _, exists := cloudLayoutCurrent[hostInfo.Id]; !exists {
		cloudLayoutCurrent[hostInfo.Id] = cloud.CloudLayoutElement{"", make(map[string]string),}
	}
	elem := cloudLayoutCurrent[hostInfo.Id]
	elem.HabitatVersion = hostInfo.HabitatInfo.Version
	for _, app := range hostInfo.Apps {
		elem.AppsVersion[app.Name] = app.CurrentVersion
	}
	cloudLayoutCurrent[hostInfo.Id] = elem
}

func deleteHostFromLayout(layout cloud.CloudLayout, id string) {
	delete(layout, id)
}

func addHostToLayout(layout cloud.CloudLayout, id string) {
	elem := cloud.CloudLayoutElement{}
	elem.HabitatVersion = "0"
	elem.AppsVersion = make(map[string]string)
	layout[id] = elem
}

func deleteAppFromLayout(layout cloud.CloudLayout, hostId string, appName string) {
	delete(layout[hostId].AppsVersion, appName)
}