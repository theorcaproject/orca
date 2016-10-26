package main

import (
	"net/http"
	"gatoor/orca/base"
	"encoding/json"
	"fmt"
)

func sendConfig(w http.ResponseWriter, trainerUpdate base.TrainerUpdate) {
	json.NewEncoder(w).Encode(trainerUpdate)
	orcaCloud.UpdateDesired(trainerUpdate)
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
	conf.Name = "ngin"
	conf.Type = base.APP_HTTP
	conf.InstallCommands = []base.OsCommand {
		{base.EXEC_COMMAND, base.Command{"echo", "MORE AMAZING"},},
	}
	conf.QueryStateCommand = base.OsCommand{base.EXEC_COMMAND, base.Command{"echo", "query for state"},}
	conf.RemoveCommand = base.OsCommand{base.EXEC_COMMAND, base.Command{"echo", "REMOVE"},}
	Logger.Info(fmt.Printf("%+v", conf))
	return conf
}


func buildTrainerUpdate() {

}

func determineNecessaryChanges() {

}



