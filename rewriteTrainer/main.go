package main

import (
	"gatoor/orca/rewriteTrainer/state/needs"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/rewriteTrainer/config"
	"gatoor/orca/rewriteTrainer/state/configuration"
	"gatoor/orca/rewriteTrainer/api"
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/rewriteTrainer/planner"
)




type SampleSt struct {
	Name string
	Version string
}

func main() {
	Logger.InitLogger.Info("Starting trainer...")
	//initState()
	//initConfig()
	//initApi()
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


func initApi() {
	var a api.Api
	a.Init()
}