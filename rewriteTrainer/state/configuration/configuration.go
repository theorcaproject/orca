package state_configuration

import (
	"sync"
	"gatoor/orca/base"
	"errors"
	"sort"
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
	c.Trainer = TrainerConfigurationState{
		Port: 5000,
		Policies: TrainerPolicies{
			TRY_TO_REMOVE_HOSTS: true,
		},
	}
}

func (c *ConfigurationState) Snapshot() ConfigurationState {
	configurationStateMutex.Lock()
	defer configurationStateMutex.Unlock()
	res := *c
	return res
}

func (c * ConfigurationState) AllAppsLatest() map[base.AppName]base.AppConfiguration {
	apps := make(map[base.AppName]base.AppConfiguration)
	configurationStateMutex.Lock()
	confApps := c.Apps
	configurationStateMutex.Unlock()
	for appName, appObj := range confApps {
		elem, err := c.GetApp(appName, appObj.LatestVersion())
		if err == nil {
			apps[appName] = elem
		}
	}
	return apps
}

func (c *ConfigurationState) GetApp (name base.AppName, version base.Version) (base.AppConfiguration, error) {
	configurationStateMutex.Lock()
	defer configurationStateMutex.Unlock()
	if _, exists := (*c).Apps[name]; !exists {
		return base.AppConfiguration{}, errors.New("No such App")
	}
	if _, exists := (*c).Apps[name][version]; !exists {
		return base.AppConfiguration{}, errors.New("No such Version")
	}
	res := (*c).Apps[name][version]
	return res, nil
}

func (c *ConfigurationState) GetHabitat (name base.HabitatName, version base.Version) (base.HabitatConfiguration, error){
	configurationStateMutex.Lock()
	defer configurationStateMutex.Unlock()
	if _, exists := (*c).Habitats[name]; !exists {
		return base.HabitatConfiguration{}, errors.New("No such Habitat")
	}
	if _, exists := (*c).Habitats[name][version]; !exists {
		return base.HabitatConfiguration{}, errors.New("No such Version")
	}
	res := (*c).Habitats[name][version]
	return res, nil
}

func (c *ConfigurationState) ConfigureApp (conf base.AppConfiguration) {
	configurationStateMutex.Lock()
	defer configurationStateMutex.Unlock()
	if _, exists := c.Apps[conf.Name]; !exists {
		c.Apps[conf.Name] = AppConfigurationVersions{}
	}
	c.Apps[conf.Name][conf.Version] = conf
}

func (c *ConfigurationState) ConfigureHabitat (conf base.HabitatConfiguration) {
	configurationStateMutex.Lock()
	defer configurationStateMutex.Unlock()
	if _, exists := c.Habitats[conf.Name]; !exists {
		c.Habitats[conf.Name] = HabitatConfigurationVersions{}
	}
	c.Habitats[conf.Name][conf.Version] = conf
}

type TrainerPolicies struct {
	TRY_TO_REMOVE_HOSTS bool
}

type TrainerConfigurationState struct {
	Port int
	Policies TrainerPolicies
	Ip base.IpAddr
}

type AppsConfigurationState map[base.AppName]AppConfigurationVersions

type AppConfigurationVersions map[base.Version]base.AppConfiguration

func (a AppConfigurationVersions) LatestVersion() base.Version {
	var keys []int
	for k := range a {
		keys = append(keys, int(k))
	}
	if len(keys) == 0 {
		return 0
	}

	sort.Sort(sort.Reverse(sort.IntSlice(keys)))
	return base.Version(keys[0])
}

type HabitatsConfigurationState map[base.HabitatName]HabitatConfigurationVersions

type HabitatConfigurationVersions map[base.Version]base.HabitatConfiguration


