package state_configuration

import (
	"sync"
	"gatoor/orca/rewriteTrainer/base"
	"errors"
)

var GlobalConfigurationState ConfigurationState

var configurationStateMutex = &sync.Mutex{}

type ConfigurationState struct {
	Trainer TrainerConfigurationState
	Apps AppsConfigurationState
	Habitats HabitatsConfigurationState
}

func (c *ConfigurationState) Init() {
	configurationStateMutex.Lock()
	defer configurationStateMutex.Unlock()
	c.Apps = AppsConfigurationState{}
	c.Habitats = HabitatsConfigurationState{}
}

func (c *ConfigurationState) Snapshot() ConfigurationState {
	configurationStateMutex.Lock()
	defer configurationStateMutex.Unlock()
	res := *c
	return res
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
	QueryStateCommand base.OsCommand
	RemoveCommand base.OsCommand
}

type HabitatsConfigurationState map[base.HabitatName]HabitatConfigurationVersions

type HabitatConfigurationVersions map[base.Version]HabitatConfiguration

type HabitatConfiguration struct {
	Name base.HabitatName
	Version base.Version
	InstallCommands []base.OsCommand
}
