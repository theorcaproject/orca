/*
Copyright Alex Mack
This file is part of Orca.

Orca is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Orca is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Orca.  If not, see <http://www.gnu.org/licenses/>.
*/

package state_configuration

import (
	"sync"
	"gatoor/orca/base"
	"errors"
	Logger "gatoor/orca/rewriteTrainer/log"
)

var ConfigLogger = Logger.LoggerWithField(Logger.Logger, "module", "configuration")
var GlobalConfigurationState ConfigurationState

var configurationStateMutex = &sync.Mutex{}


type ConfigurationState struct {
	ConfigurationRootPath string

	Trainer base.TrainerConfigurationState
	Apps AppsConfigurationState
	CloudProvider base.ProviderConfiguration
}

func (c *ConfigurationState) Init() {
	configurationStateMutex.Lock()
	defer configurationStateMutex.Unlock()
	c.Apps = AppsConfigurationState{}
	c.Trainer = base.TrainerConfigurationState{
		Port: 5000,
		Policies: base.TrainerPolicies{
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

func (c *ConfigurationState) AllAppsLatest() map[base.AppName]base.AppConfiguration {
	return c.Apps
}

func (c *ConfigurationState) GetApp (name base.AppName) (base.AppConfiguration, error) {
	configurationStateMutex.Lock()
	defer configurationStateMutex.Unlock()
	if _, exists := (*c).Apps[name]; !exists {
		return base.AppConfiguration{}, errors.New("No such App")
	}
	res := (*c).Apps[name]
	return res, nil
}

func (c *ConfigurationState) ConfigureApp (conf base.AppConfiguration) {
	ConfigLogger.Infof("ConfigureApp %s:%d", conf.Name)
	configurationStateMutex.Lock()
	defer configurationStateMutex.Unlock()
	if _, exists := c.Apps[conf.Name]; !exists {
		c.Apps[conf.Name] = base.AppConfiguration{}
	}
	c.Apps[conf.Name] = conf
}

type AppsConfigurationState map[base.AppName]base.AppConfiguration
type ProviderConfigurationState base.ProviderConfiguration


