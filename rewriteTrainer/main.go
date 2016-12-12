package main

import (
	"gatoor/orca/rewriteTrainer/state/needs"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/rewriteTrainer/state/configuration"
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/rewriteTrainer/config"
	"gatoor/orca/rewriteTrainer/api"
	"gatoor/orca/rewriteTrainer/cloud"
	"gatoor/orca/rewriteTrainer/db"
	"gatoor/orca/rewriteTrainer/scheduler"
	"gatoor/orca/rewriteTrainer/planner"
	"time"
	"flag"
	"gatoor/orca/base"
	"gatoor/orca/rewriteTrainer/audit"
)


const CHECKIN_WAIT_TIME = 5

func main() {
	var configurationRoot = flag.String("configroot", "/orca/config/", "Configuration Root Directory")
	flag.Parse()

	audit.Audit.Init()
	audit.Audit.AddEvent(map[string]string{
		"message": "Orca Trainer Started",
	})
	var baseConfiguration config.JsonConfiguration

	Logger.InitLogger.Info("Starting trainer...")
	initState()
	initConfig(&baseConfiguration, *configurationRoot)
	cloud.Init()
	db.Init("")
	initApi(&baseConfiguration)
	waitForCheckin()
	scheduler.Start()
	planner.InitialPlan()
	Logger.InitLogger.Info("Trainer started")
	ticker := time.NewTicker(time.Second * 60)
	for {
		<- ticker.C
	}
}

func waitForCheckin() {
	Logger.InitLogger.Infof("Waiting %ds for existsing clients to check in", CHECKIN_WAIT_TIME)
	time.Sleep(time.Duration(CHECKIN_WAIT_TIME * time.Second))
	Logger.InitLogger.Info("Done waiting")
}

func initState() {
	state_configuration.GlobalConfigurationState.Init()
	state_cloud.GlobalCloudLayout.Init()
	state_needs.GlobalAppsNeedState = state_needs.AppsNeedState{}
}

func initConfig(baseConfiguration *config.JsonConfiguration, configurationRoot string) {
	baseConfiguration.Init(configurationRoot)
	baseConfiguration.Load()
	baseConfiguration.Check()
	baseConfiguration.ApplyToState()

	audit.Audit.AddEvent(map[string]string{
		"message": "Configuration has been loaded from filesystem",
	})
}


func initApi(baseConfiguration *config.JsonConfiguration) {
	var a api.Api
	a.ConfigManager = baseConfiguration
	a.Init()
}