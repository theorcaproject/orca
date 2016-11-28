package config

import (
	"gatoor/orca/rewriteTrainer/base"
	"gatoor/orca/rewriteTrainer/cloud"
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/rewriteTrainer/state/needs"
	"os"
	"encoding/json"
	"gatoor/orca/util"
	"gatoor/orca/rewriteTrainer/state/configuration"
	"fmt"
)

const CONFIGURATION_FILE = "/tmp/example.json"

type JsonConfiguration struct {
	Trainer TrainerJsonConfiguration
	Habitats []HabitatJsonConfiguration
	Apps []AppJsonConfiguration
	Clouds []CloudsJsonConfiguration
}

type TrainerJsonConfiguration struct {
	Port int
}

type HabitatJsonConfiguration struct {
	Name base.HabitatName
	Version base.Version
	InstallCommands []base.OsCommand
}

type CloudsJsonConfiguration struct {
	Name base.CloudName
	Version base.Version
	Type base.CloudType
	AccessKeyId base.CloudAccessKeyId
	AccessKeySecret base.CloudAccessKeySecret
}

type AppJsonConfiguration struct {
	Name base.AppName
	Version base.Version
	Type base.AppType

	InstallCommands []base.OsCommand
	InstallFiles []base.OsCommand
	QueryStateCommand[] base.OsCommand
	RemoveCommand[] base.OsCommand
	Needs state_needs.AppNeeds

	Min base.MinInstances
	Desired base.DesiredInstances
	Max base.MaxInstances
	LoadBalancer []base.CloudProviderLoadBalancerName
	Networks []base.CloudProviderVpcName

	Cloud base.CloudName
}

type CloudJsonConfiguration struct {
	Provider cloud.Provider
	InstanceType cloud.InstanceType
	MinInstanceCount cloud.MinInstanceCount
	MaxInstanceCount cloud.MaxInstanceCount
}

func (j *JsonConfiguration) Load() {
	Logger.InitLogger.Infof("Loading config file from %s", CONFIGURATION_FILE)
	file, err := os.Open(CONFIGURATION_FILE)
	if err != nil {
		Logger.InitLogger.Fatalf("Could not open config file %s - %s", CONFIGURATION_FILE, err)
	}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&j); err != nil {
		extra := ""
		if serr, ok := err.(*json.SyntaxError); ok {
			line, col, highlight := util.HighlightBytePosition(file, serr.Offset)
			extra = fmt.Sprintf(":\nError at line %d, column %d (file offset %d):\n%s",
				line, col, serr.Offset, highlight)
		}
		Logger.InitLogger.Fatalf("error parsing JSON object in config file %s%s\n%v",
			file.Name(), extra, err)
	}
}


func (j *JsonConfiguration)  ApplyToState() {
	Logger.InitLogger.Infof("Applying config to State")
	applyHabitatConfig(j.Habitats)
	applyTrainerConfig(j.Trainer)
	applyAppsConfig(j.Apps)
	applyNeeds(j.Apps)
	applyCloudsConfig(j.Clouds)
	Logger.InitLogger.Infof("Config was applied to State")
}

func applyAppsConfig(appsConfs []AppJsonConfiguration) {
	for _, aConf := range appsConfs {
		state_configuration.GlobalConfigurationState.ConfigureApp(state_configuration.AppConfiguration{
			Name: aConf.Name,
			Type: aConf.Type,
			Version: aConf.Version,
			InstallCommands: aConf.InstallCommands,
			InstallFiles: aConf.InstallFiles,
			QueryStateCommand: aConf.QueryStateCommand,
			RemoveCommand: aConf.RemoveCommand,
			LoadBalancer:aConf.LoadBalancer,
			Networks: aConf.Networks,
			Cloud: aConf.Cloud,
		})
	}
}

func applyHabitatConfig (habitatConfs []HabitatJsonConfiguration) {
	for _, hConf := range habitatConfs {
		state_configuration.GlobalConfigurationState.ConfigureHabitat(state_configuration.HabitatConfiguration{
			Name: hConf.Name,
			Version: hConf.Version,
			InstallCommands: hConf.InstallCommands,
		})
	}
}

func applyCloudsConfig (cloudsConfs []CloudsJsonConfiguration) {
	for _, cConf := range cloudsConfs {
		state_configuration.GlobalConfigurationState.ConfigureClouds(state_configuration.CloudProviderConfiguration{
			Name: cConf.Name,
			Version: cConf.Version,
			Type: cConf.Type,
			AccessKeyId: cConf.AccessKeyId,
			AccessKeySecret: cConf.AccessKeySecret,
		})
	}
}

func applyTrainerConfig (trainerConf TrainerJsonConfiguration) {
	state_configuration.GlobalConfigurationState.Trainer.Port = trainerConf.Port
}

func applyNeeds(appConfs []AppJsonConfiguration) {
	for _, aNeeds := range appConfs {
		state_needs.GlobalAppsNeedState.UpdateNeeds(aNeeds.Name, aNeeds.Version, aNeeds.Needs)
	}
}


func (j *JsonConfiguration) Serialize() string {
	res, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	return string(res)
}