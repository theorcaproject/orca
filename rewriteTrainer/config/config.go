package config

import (
	"gatoor/orca/base"
	"gatoor/orca/rewriteTrainer/cloud"
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/rewriteTrainer/state/needs"
	"os"
	"encoding/json"
	"gatoor/orca/util"
	"fmt"
	"gatoor/orca/rewriteTrainer/state/configuration"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/rewriteTrainer/needs"
)


type JsonConfiguration struct {
	Trainer TrainerJsonConfiguration
	AvailableInstances []base.HostId
	//Habitats []HabitatJsonConfiguration
	Apps []AppJsonConfiguration
	CloudProvider cloud.ProviderConfiguration
}

type TrainerJsonConfiguration struct {
	Port int
	Ip base.IpAddr
	Policies TrainerPolicies
}

type TrainerPolicies struct {
	TRY_TO_REMOVE_HOSTS bool
}

type HabitatJsonConfiguration struct {
	Name base.HabitatName
	Version base.Version
	InstallCommands []base.OsCommand
}

type AppJsonConfiguration struct {
	Name base.AppName
	Version base.Version
	Type base.AppType
	MinDeploymentCount base.DeploymentCount
	MaxDeploymentCount base.DeploymentCount
	//InstallCommands []base.OsCommand
	//QueryStateCommand base.OsCommand
	//RemoveCommand base.OsCommand
	//RunCommand base.OsCommand
	//StopCommand base.OsCommand
	DockerConfig base.DockerConfig
	Needs needs.AppNeeds
}

type CloudJsonConfiguration struct {
	Provider cloud.Provider
	InstanceType cloud.InstanceType
	MinInstanceCount cloud.MinInstanceCount
	MaxInstanceCount cloud.MaxInstanceCount
}



func loadConfigFromFile(file *os.File, conf interface{}) {
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(conf); err != nil {
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

func (j *JsonConfiguration) Check() {


}

func (j *JsonConfiguration) Load(trainerConfigPath string, appConfigPath string, availableInstancesPath string, cloudProviderConfigPath string) {
	configFiles := make(map[string]interface{})
	configFiles[trainerConfigPath] = &j.Trainer
	configFiles[appConfigPath] = &j.Apps
	configFiles[availableInstancesPath] = &j.AvailableInstances
	configFiles[cloudProviderConfigPath] = &j.CloudProvider
	for key, interf := range configFiles {
		Logger.InitLogger.Infof("Loading config file from %s", key)
		file, err := os.Open(key)
		if err != nil {
			Logger.InitLogger.Fatalf("Could not open config file %s - %s", key, err)
		}
		loadConfigFromFile(file, interf)
		file.Close()
	}
}


func (j *JsonConfiguration)  ApplyToState() {
	Logger.InitLogger.Infof("Applying config to State")
	//applyHabitatConfig(j.Habitats)
	applyTrainerConfig(j.Trainer)
	applyAppsConfig(j.Apps)
	applyNeeds(j.Apps)
	applyAvailableInstances(j.AvailableInstances)
	applyCloudProviderConfiguration(j.CloudProvider)
	Logger.InitLogger.Infof("Config was applied to State")
}

func applyAvailableInstances(instances []base.HostId) {
	Logger.InitLogger.Infof("Applying AvailableInstances config: %+v", instances)
	for _, hostId := range instances {
		state_cloud.GlobalAvailableInstances.Update(hostId, state_cloud.InstanceResources{UsedCpuResource:0, UsedMemoryResource:0, UsedNetworkResource:0, TotalCpuResource: 20, TotalMemoryResource: 20, TotalNetworkResource: 20})
	}
}

func applyCloudProviderConfiguration(conf cloud.ProviderConfiguration) {
	Logger.InitLogger.Infof("Applying CloudProvider config: %+v", conf)
	cloud.CurrentProviderConfig.Type = conf.Type
	cloud.CurrentProviderConfig.MaxInstances = conf.MaxInstances
	cloud.CurrentProviderConfig.MinInstances = conf.MinInstances
	cloud.CurrentProviderConfig.AWSConfiguration = conf.AWSConfiguration
}

func applyAppsConfig(appsConfs []AppJsonConfiguration) {
	for _, aConf := range appsConfs {
		state_configuration.GlobalConfigurationState.ConfigureApp(base.AppConfiguration{
			Name: aConf.Name,
			Type: aConf.Type,
			Version: aConf.Version,
			TargetDeploymentCount: aConf.MinDeploymentCount,
			MinDeploymentCount: aConf.MaxDeploymentCount,
			DockerConfig: aConf.DockerConfig,
			//InstallCommands: aConf.InstallCommands,
			//QueryStateCommand: aConf.QueryStateCommand,
			//RunCommand: aConf.RunCommand,
			//StopCommand: aConf.StopCommand,
			//RemoveCommand: aConf.RemoveCommand,
		})
	}
}

func applyHabitatConfig (habitatConfs []HabitatJsonConfiguration) {
	for _, hConf := range habitatConfs {
		state_configuration.GlobalConfigurationState.ConfigureHabitat(base.HabitatConfiguration{
			Name: hConf.Name,
			Version: hConf.Version,
			InstallCommands: hConf.InstallCommands,
		})
	}
}

func applyTrainerConfig (trainerConf TrainerJsonConfiguration) {
	state_configuration.GlobalConfigurationState.Trainer.Port = trainerConf.Port
	state_configuration.GlobalConfigurationState.Trainer.Ip = trainerConf.Ip
	state_configuration.GlobalConfigurationState.Trainer.Policies.TRY_TO_REMOVE_HOSTS = trainerConf.Policies.TRY_TO_REMOVE_HOSTS
}

//TODO use WeeklyNeeds
func applyNeeds(appConfs []AppJsonConfiguration) {
	for _, aNeeds := range appConfs {
		state_needs.GlobalAppsNeedState.UpdateNeeds(aNeeds.Name, aNeeds.Version, aNeeds.Needs)
	}
}


func (j *JsonConfiguration) Serialize() string {
	res, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		Logger.InitLogger.Errorf("JsonConfiguration Derialize failed: %s; %+v", err, j)
	}
	return string(res)
}