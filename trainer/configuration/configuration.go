/*
Copyright Alex Mack (al9mack@gmail.com) and Michael Lawson (michael@sphinix.com)
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

package configuration

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	Logger "orca/trainer/logs"
	"orca/trainer/model"
	"orca/util"
	"os"
	"reflect"
	"strings"
	"time"
)

type ConfigurationStore struct {
	ApplicationConfigurations map[string]*model.ApplicationConfiguration
	Properties                map[string]*model.PropertyGroup
	GlobalSettings            GlobalSettings

	trainerConfigurationFilePath string
}

func (store *ConfigurationStore) Init(trainerConfigurationFilePath string) {
	store.trainerConfigurationFilePath = trainerConfigurationFilePath

	defaultUserAccount := User{
		Password: "admin",
	}
	store.ApplicationConfigurations = make(map[string]*model.ApplicationConfiguration)
	store.Properties = make(map[string]*model.PropertyGroup)
	store.GlobalSettings = GlobalSettings{
		ApiPort:                5001,
		AppChangeTimeout:       300,
		ServerChangeTimeout:    300,
		ServerTimeout:          300,
		HostChangeFailureLimit: 10,
		ServerCapacity:         10,
		Users:                  map[string]User{"admin": defaultUserAccount},
		PlanningAlg:            "boringplanner",
		CloudProvider:          "aws",
		AuditDatabaseUri:       "http://localhost:9200",
		StatsDatabaseUri:       "localhost",
		ServerTTL:              86400,
		CloudProviderCommands:  make([] string, 0),
	}
}

func (store *ConfigurationStore) DumpConfig() {
	Logger.InitLogger.Infof("Loading config file from %+v", store.ApplicationConfigurations)
}

func (store *ConfigurationStore) Load() {
	store.loadApplicationConfigurationsFromFile(store.trainerConfigurationFilePath)

	/* If the schedule has not been defined, we should set it to defaults */
	for _, app := range store.ApplicationConfigurations {
		if app.DeploymentSchedule.IsEmpty() {
			app.DeploymentSchedule.SetAll(0)
		}
	}
}

func (store *ConfigurationStore) Save() {
	store.saveConfigToFile(store.trainerConfigurationFilePath)
}

func (store *ConfigurationStore) loadApplicationConfigurationsFromFile(filename string) {
	Logger.InitLogger.Infof("Loading config file from %s", filename)
	file, err := os.Open(filename)
	if err != nil {
		Logger.InitLogger.Fatalf("Could not open config file %s - %s", filename, err)
		return
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(store); err != nil {
		extra := ""
		if serr, ok := err.(*json.SyntaxError); ok {
			line, col, highlight := util.HighlightBytePosition(file, serr.Offset)
			extra = fmt.Sprintf(":\nError at line %d, column %d (file offset %d):\n%s",
				line, col, serr.Offset, highlight)
		}
		Logger.InitLogger.Fatalf("error parsing JSON object in config file %s%s\n%v",
			file.Name(), extra, err)
	} else {
		fmt.Sprintf("error: %v", err)
	}

	Logger.InitLogger.Infof("Load done")
	file.Close()
}

func (store *ConfigurationStore) Add(name string, config *model.ApplicationConfiguration) *model.ApplicationConfiguration {
	store.ApplicationConfigurations[name] = config
	return config
}

func (store *ConfigurationStore) Remove(name string) {
	delete(store.ApplicationConfigurations, name)
}

func (store *ConfigurationStore) saveConfigToFile(filename string) {
	res, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		Logger.InitLogger.Errorf("JsonConfiguration Derialize failed: %s; %+v", err, store)
	}
	var result = string(res)
	err = ioutil.WriteFile(filename, []byte(result), 0644)
	if err != nil {
		panic(err)
	}
}

func (store *ConfigurationStore) GetConfiguration(application string) (*model.ApplicationConfiguration, error) {
	if app, ok := store.ApplicationConfigurations[application]; ok {
		return app, nil
	}

	return nil, errors.New("Could not find application")
}

func AppExistsInList(items []*model.ApplicationConfiguration, appName string) bool {
	for _, app := range items {
		if strings.Compare(app.Name, appName) == 0 {
			return true
		}
	}

	return false
}

func (store *ConfigurationStore) GetAllConfigurations_MoveLeft(items []*model.ApplicationConfiguration) []*model.ApplicationConfiguration {
	ret := make([]*model.ApplicationConfiguration, 0)
	for _, app := range items {
		/* Check dependencies */
		for _, dependency := range app.Depends {
			if !AppExistsInList(ret, dependency.Name) {
				dependency_app, err := store.GetConfiguration(dependency.Name)
				if err == nil {
					ret = append(ret, dependency_app)
				}
			}
		}

		if !AppExistsInList(ret, app.Name) {
			ret = append(ret, app)
		}
	}

	if !reflect.DeepEqual(items, ret) {
		return store.GetAllConfigurations_MoveLeft(ret)
	}

	return ret
}

func (store *ConfigurationStore) GetAllConfigurationAsOrderedList() []*model.ApplicationConfiguration {
	ret := make([]*model.ApplicationConfiguration, 0)
	for _, applicationConfiguration := range store.GetAllConfiguration() {
		ret = append(ret, applicationConfiguration)
	}

	return store.GetAllConfigurations_MoveLeft(ret)
}
func (store *ConfigurationStore) GetAllConfiguration() map[string]*model.ApplicationConfiguration {
	return store.ApplicationConfigurations
}

func (store *ConfigurationStore) GetConfigAsString() string {
	res, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		Logger.InitLogger.Errorf("JsonConfiguration Derialize failed: %s; %+v", err, store)
	}
	return string(res)
}

func (store *ConfigurationStore) ApplySchedules() {
	for _, config := range store.ApplicationConfigurations {
		if config.DisableSchedule {
			continue
		}

		if config.DeploymentSchedule.Get(time.Now()) == 0 {
			continue
		}

		if config.DeploymentSchedule.Get(time.Now()) > config.MinDeployment {
			config.DesiredDeployment = config.DeploymentSchedule.Get(time.Now())
		} else {
			config.DesiredDeployment = config.MinDeployment
		}
	}
}

func (store *ConfigurationStore) RequestPublishConfiguration(config *model.ApplicationConfiguration) {
	templateForConfiguration := config.GetLatestConfiguration()

	publishedConfiguration := model.VersionConfig{
		Version:              config.GetSuitableNextVersion(),
		DockerConfig:         templateForConfiguration.DockerConfig,
		Needs:                templateForConfiguration.Needs,
		LoadBalancer:         templateForConfiguration.LoadBalancer,
		Network:              templateForConfiguration.Network,
		SecurityGroups:       templateForConfiguration.SecurityGroups,
		PortMappings:         templateForConfiguration.PortMappings,
		VolumeMappings:       templateForConfiguration.VolumeMappings,
		EnvironmentVariables: templateForConfiguration.EnvironmentVariables,
		Checks:               templateForConfiguration.Checks,
		GroupingTag:          templateForConfiguration.GroupingTag,

		AppliedPropertyGroups: make(map[string]int),
		DeploymentFailures:    0,
		DeploymentSuccess:     0,
	}

	publishedConfiguration.Files = make([]model.File, len(templateForConfiguration.Files))
	copy(publishedConfiguration.Files, templateForConfiguration.Files)

	publishedConfiguration.SecurityGroups = make([]model.SecurityGroup, len(templateForConfiguration.SecurityGroups))
	copy(publishedConfiguration.SecurityGroups, templateForConfiguration.SecurityGroups)

	publishedConfiguration.EnvironmentVariables = make([]model.EnvironmentVariable, len(templateForConfiguration.EnvironmentVariables))
	copy(publishedConfiguration.EnvironmentVariables, templateForConfiguration.EnvironmentVariables)

	for _, templateName := range config.PropertyGroups {
		templateObject := store.Properties[templateName.Name]
		if templateObject != nil {
			publishedConfiguration.ApplyPropertyGroup(templateName.Name, templateObject)
		}
	}

	if config.PublishedConfig == nil {
		config.PublishedConfig = make(map[string]*model.VersionConfig)
	}
	config.PublishedConfig[publishedConfiguration.Version] = &publishedConfiguration
	store.Save()
}

func (store *ConfigurationStore) DoesRequestPublishConfigurationMakeSense(config *model.ApplicationConfiguration) bool {
	templateForConfiguration := config.GetLatestConfiguration()
	lastPublishedConfiguration := config.GetLatestPublishedConfiguration()

	publishedConfiguration := model.VersionConfig{
		Version:              lastPublishedConfiguration.Version,
		DockerConfig:         templateForConfiguration.DockerConfig,
		Needs:                templateForConfiguration.Needs,
		LoadBalancer:         templateForConfiguration.LoadBalancer,
		Network:              templateForConfiguration.Network,
		SecurityGroups:       templateForConfiguration.SecurityGroups,
		PortMappings:         templateForConfiguration.PortMappings,
		VolumeMappings:       templateForConfiguration.VolumeMappings,
		EnvironmentVariables: templateForConfiguration.EnvironmentVariables,
		Checks:               templateForConfiguration.Checks,
		GroupingTag:          templateForConfiguration.GroupingTag,

		AppliedPropertyGroups: lastPublishedConfiguration.AppliedPropertyGroups,
		DeploymentFailures:    lastPublishedConfiguration.DeploymentFailures,
		DeploymentSuccess:     lastPublishedConfiguration.DeploymentSuccess,
	}

	publishedConfiguration.Files = make([]model.File, len(templateForConfiguration.Files))
	copy(publishedConfiguration.Files, templateForConfiguration.Files)

	publishedConfiguration.SecurityGroups = make([]model.SecurityGroup, len(templateForConfiguration.SecurityGroups))
	copy(publishedConfiguration.SecurityGroups, templateForConfiguration.SecurityGroups)

	publishedConfiguration.EnvironmentVariables = make([]model.EnvironmentVariable, len(templateForConfiguration.EnvironmentVariables))
	copy(publishedConfiguration.EnvironmentVariables, templateForConfiguration.EnvironmentVariables)

	for _, templateName := range config.PropertyGroups {
		templateObject := store.Properties[templateName.Name]
		if templateObject != nil {
			publishedConfiguration.ApplyPropertyGroup(templateName.Name, templateObject)
		}
	}

	/* The hackiest and probably easiest way is to simply stringify it */
	return strings.Compare(lastPublishedConfiguration.AsString(), publishedConfiguration.AsString()) != 0
}
