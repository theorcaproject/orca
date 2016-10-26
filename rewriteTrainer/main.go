package main

import (
	"gatoor/orca/rewriteTrainer/state/needs"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/rewriteTrainer/config"
	"gatoor/orca/rewriteTrainer/state/configuration"
	"gatoor/orca/rewriteTrainer/api"
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/rewriteTrainer/planner"
	"os"
	"strconv"
)




type SampleSt struct {
	Name string
	Version string
}

func main() {
	Logger.InitLogger.Info("Starting trainer...")
	Logger.InitLogger.Info(os.Args)

	if (len(os.Args) == 2){
		apiPort, err := strconv.Atoi(os.Args[1])
		Logger.InitLogger.Info("Api port is %d err: %s" ,apiPort, err)
		initState()
		initConfig()
		initApi(apiPort)

	}else{
		Logger.InitLogger.Info("Please supply an api port")
	}
}

func initPlanner() {
	planner.Init()
}

func initState() {
	state_configuration.GlobalConfigurationState.Init()
	state_cloud.GlobalCloudLayout.Init()
	state_needs.GlobalAppsNeedState = state_needs.AppsNeedState{}
}

func initConfig() {
	var baseConfiguration config.JsonConfiguration
	baseConfiguration.Load()
	baseConfiguration.ApplyToState()
}


func initApi(port int) {
	Logger.InitLogger.Info("init api on port %d", port)

	var a api.Api
	a.Init(port)
}