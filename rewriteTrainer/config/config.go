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

package config

import (
	"gatoor/orca/base"
	"os"
	"encoding/json"
	"gatoor/orca/util"
	"fmt"
	"gatoor/orca/rewriteTrainer/state/configuration"
	"io/ioutil"

	Logger "gatoor/orca/rewriteTrainer/log"
)


type JsonConfiguration struct {
	configRoot string

	Trainer base.TrainerConfigurationState
	Habitats []base.HabitatConfiguration
	Apps []base.AppConfiguration
	CloudProvider base.ProviderConfiguration
}

func loadConfigFromFile(filename string, conf interface{}) {
	Logger.InitLogger.Infof("Loading config file from %s", filename)
	file, err := os.Open(filename)
	if err != nil {
		Logger.InitLogger.Fatalf("Could not open config file %s - %s", filename, err)
		return
	}

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
	file.Close()
}

func saveConfigToFile(filename string, conf interface{}) {
	res, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		Logger.InitLogger.Errorf("JsonConfiguration Derialize failed: %s; %+v", err, conf)
	}
	var result = string(res)
	err = ioutil.WriteFile(filename, []byte(result), 0644)
	if err != nil {
		panic(err)
	}
}

func (j *JsonConfiguration) Check() {
}

var (
	TRAINER_CONFIGURATION_FILE = "trainer.json"
	APPS_CONFIGURATION_FILE = "applications.json"
	AVAILABLE_INSTANCES_CONFIGURATION_FILE = "instances.json"
	CLOUD_PROVIDER_CONFIGURATION_FILE = "provider.json"
)

func (j *JsonConfiguration) Init(configurationRoot string){
	j.configRoot = configurationRoot
	state_configuration.GlobalConfigurationState.ConfigurationRootPath = configurationRoot
}

func (j *JsonConfiguration) Load() {
	loadConfigFromFile(j.configRoot + TRAINER_CONFIGURATION_FILE, &j.Trainer)
	loadConfigFromFile(j.configRoot + APPS_CONFIGURATION_FILE, &j.Apps)
	loadConfigFromFile(j.configRoot + CLOUD_PROVIDER_CONFIGURATION_FILE, &j.CloudProvider)
}

func (j *JsonConfiguration) Save() {
	Logger.InitLogger.Infof("Saving all configuration files")
	saveConfigToFile(j.configRoot + TRAINER_CONFIGURATION_FILE, state_configuration.GlobalConfigurationState.Trainer)

	var buffer = make([]base.AppConfiguration, 0)
	for _, application := range state_configuration.GlobalConfigurationState.AllAppsLatest() {
		buffer = append(buffer, application)
	}

	saveConfigToFile(j.configRoot + APPS_CONFIGURATION_FILE, buffer)
	saveConfigToFile(j.configRoot + CLOUD_PROVIDER_CONFIGURATION_FILE, state_configuration.GlobalConfigurationState.CloudProvider)
}


func (j *JsonConfiguration)  ApplyToState() {
	Logger.InitLogger.Infof("Applying config to State")
	applyTrainerConfig(j.Trainer)
	applyAppsConfig(j.Apps)
	applyCloudProviderConfiguration(j.CloudProvider)
	Logger.InitLogger.Infof("Config was applied to State")
}

func applyCloudProviderConfiguration(conf base.ProviderConfiguration) {
	Logger.InitLogger.Infof("Applying CloudProvider config: %+v", conf)
	state_configuration.GlobalConfigurationState.CloudProvider.SSHKey = conf.SSHKey
	state_configuration.GlobalConfigurationState.CloudProvider.SSHUser = conf.SSHUser
	state_configuration.GlobalConfigurationState.CloudProvider.Type = conf.Type
	state_configuration.GlobalConfigurationState.CloudProvider.MaxInstances = conf.MaxInstances
	state_configuration.GlobalConfigurationState.CloudProvider.MinInstances = conf.MinInstances
	state_configuration.GlobalConfigurationState.CloudProvider.AWSConfiguration = conf.AWSConfiguration
	state_configuration.GlobalConfigurationState.CloudProvider.BaseInstanceType = conf.BaseInstanceType
}

func applyAppsConfig(appsConfs []base.AppConfiguration) {
	for _, aConf := range appsConfs {
		state_configuration.GlobalConfigurationState.ConfigureApp(base.AppConfiguration{
			Name: aConf.Name,
			Type: aConf.Type,
			TargetDeploymentCount: aConf.TargetDeploymentCount,
			MinDeploymentCount: aConf.MinDeploymentCount,
			Needs: aConf.Needs,
			ConfigurationSets: aConf.ConfigurationSets,
		})
	}
}

func applyTrainerConfig (trainerConf base.TrainerConfigurationState) {
	state_configuration.GlobalConfigurationState.Trainer.Port = trainerConf.Port
	state_configuration.GlobalConfigurationState.Trainer.Ip = trainerConf.Ip
	state_configuration.GlobalConfigurationState.Trainer.Policies.TRY_TO_REMOVE_HOSTS = trainerConf.Policies.TRY_TO_REMOVE_HOSTS
}

func (j *JsonConfiguration) Serialize() string {
	res, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		Logger.InitLogger.Errorf("JsonConfiguration Derialize failed: %s; %+v", err, j)
	}
	return string(res)
}