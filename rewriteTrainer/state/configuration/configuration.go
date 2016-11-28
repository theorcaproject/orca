package state_configuration

import (
	"sync"
	"gatoor/orca/rewriteTrainer/base"
	Logger "gatoor/orca/rewriteTrainer/log"
	"errors"
)
var ConfigLogger = Logger.LoggerWithField(Logger.Logger, "module", "configuration")

var GlobalConfigurationState ConfigurationState

var configurationStateMutex = &sync.Mutex{}

type ConfigurationState struct {
	Trainer TrainerConfigurationState
	Apps AppsConfigurationState
	Habitats HabitatsConfigurationState
	Clouds CloudProviderState
}

func (c *ConfigurationState) Init() {
	configurationStateMutex.Lock()
	defer configurationStateMutex.Unlock()
	c.Apps = AppsConfigurationState{}
	c.Habitats = HabitatsConfigurationState{}
	c.Clouds = CloudProviderState{}
}

func (c *ConfigurationState) Snapshot() ConfigurationState {
	configurationStateMutex.Lock()
	defer configurationStateMutex.Unlock()
	res := *c
	return res
}

func (c *ConfigurationState) GetAllApps() ([]AppConfiguration) {
	configurationStateMutex.Lock()
	defer configurationStateMutex.Unlock()

	var ret_configurations []AppConfiguration
	for _, app_body:= range (*c).Apps {
		var top_version base.Version
		var top_version_object AppConfiguration

		for version, application_config := range app_body {
			if(version > top_version){
				top_version = version
				top_version_object = application_config
			}
		}

		ret_configurations = append(ret_configurations, top_version_object)
	}
	return ret_configurations
}

func (c *ConfigurationState) GetAllClouds() ([]CloudProviderConfiguration) {
	configurationStateMutex.Lock()
	defer configurationStateMutex.Unlock()

	var ret_configurations []CloudProviderConfiguration
	for _, app_body:= range (*c).Clouds {
		var top_version base.Version
		var top_version_object CloudProviderConfiguration

		for version, application_config := range app_body {
			if(version > top_version){
				top_version = version
				top_version_object = application_config
			}
		}

		ret_configurations = append(ret_configurations, top_version_object)
	}
	return ret_configurations
}

func (c *ConfigurationState) GetApp (name base.AppName, version base.Version) (AppConfiguration, error) {
	configurationStateMutex.Lock()
	defer configurationStateMutex.Unlock()
	if _, exists := (*c).Apps[name]; !exists {
		return AppConfiguration{}, errors.New("No such App")
	}
	if _, exists := (*c).Apps[name][version]; !exists {
		return AppConfiguration{}, errors.New("No such Version")
	}
	res := (*c).Apps[name][version]
	return res, nil
}

func (c *ConfigurationState) GetHabitat (name base.HabitatName, version base.Version) (HabitatConfiguration, error){
	configurationStateMutex.Lock()
	defer configurationStateMutex.Unlock()
	if _, exists := (*c).Habitats[name]; !exists {
		return HabitatConfiguration{}, errors.New("No such Habitat")
	}
	if _, exists := (*c).Habitats[name][version]; !exists {
		return HabitatConfiguration{}, errors.New("No such Version")
	}
	res := (*c).Habitats[name][version]
	return res, nil
}

func (c *ConfigurationState) ConfigureApp (conf AppConfiguration) {
	configurationStateMutex.Lock()
	defer configurationStateMutex.Unlock()
	if _, exists := c.Apps[conf.Name]; !exists {
		c.Apps[conf.Name] = AppConfigurationVersions{}
	}

	ConfigLogger.Infof("Configured a new application version: %d", conf.Version)
	c.Apps[conf.Name][conf.Version] = conf
}

func (c *ConfigurationState) ConfigureHabitat (conf HabitatConfiguration) {
	configurationStateMutex.Lock()
	defer configurationStateMutex.Unlock()
	if _, exists := c.Habitats[conf.Name]; !exists {
		c.Habitats[conf.Name] = HabitatConfigurationVersions{}
	}
	c.Habitats[conf.Name][conf.Version] = conf
}

func (c *ConfigurationState) ConfigureClouds (conf CloudProviderConfiguration) {
	configurationStateMutex.Lock()
	defer configurationStateMutex.Unlock()
	if _, exists := c.Clouds[conf.Name]; !exists {
		c.Clouds[conf.Name] = CloudProviderConfigurationVersions{}
	}
	c.Clouds[conf.Name][conf.Version] = conf
}

type TrainerConfigurationState struct {
	Port int
}

type AppsConfigurationState map[base.AppName]AppConfigurationVersions

type AppConfigurationVersions map[base.Version]AppConfiguration

type AppConfiguration struct {
	Name base.AppName
	Type base.AppType
	Version base.Version

	InstallCommands []base.OsCommand
	InstallFiles []base.OsCommand
	QueryStateCommand[] base.OsCommand
	RemoveCommand[] base.OsCommand

	Min base.MinInstances
	Desired base.DesiredInstances
	Max base.MaxInstances
	LoadBalancer []base.CloudProviderLoadBalancerName
	Networks []base.CloudProviderVpcName
	Cloud base.CloudName
}

type HabitatsConfigurationState map[base.HabitatName]HabitatConfigurationVersions

type HabitatConfigurationVersions map[base.Version]HabitatConfiguration

type HabitatConfiguration struct {
	Name base.HabitatName
	Version base.Version
	InstallCommands []base.OsCommand
}

type CloudProviderState map[base.CloudName]CloudProviderConfigurationVersions
type CloudProviderConfigurationVersions map[base.Version]CloudProviderConfiguration
type CloudProviderConfiguration struct {
	Name base.CloudName
	Version base.Version
	Type base.CloudType
	AccessKeyId base.CloudAccessKeyId
	AccessKeySecret base.CloudAccessKeySecret
}
