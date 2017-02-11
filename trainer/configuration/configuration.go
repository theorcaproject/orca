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
	"errors"
	"os"
	"time"
	"encoding/json"
	"fmt"
	"orca/util"
	Logger "orca/trainer/logs"
	"io/ioutil"
	"orca/trainer/state"
	"orca/trainer/model"
)

type ConfigurationStore struct {
	ApplicationConfigurations    	map[string]*model.ApplicationConfiguration;
	GlobalSettings			GlobalSettings
	AuditDatabaseUri             	string

	trainerConfigurationFilePath 	string
}

func (store *ConfigurationStore) Init(trainerConfigurationFilePath string){
	store.trainerConfigurationFilePath = trainerConfigurationFilePath

	store.ApplicationConfigurations = make(map[string]*model.ApplicationConfiguration);
	store.GlobalSettings = GlobalSettings{
		ApiPort:5001,
		AppChangeTimeout:300,
		ServerChangeTimeout:300,
		ServerTimeout:300,
	}
}

func (store *ConfigurationStore) DumpConfig(){
	Logger.InitLogger.Infof("Loading config file from %+v", store.ApplicationConfigurations)
}

func (store* ConfigurationStore) Load(){
	store.loadApplicationConfigurationsFromFile(store.trainerConfigurationFilePath)

	/* If the schedule has not been defined, we should set it to defaults */
	for _, app := range store.ApplicationConfigurations {
		if app.DeploymentSchedule.IsEmpty() {
			app.DeploymentSchedule.SetAll(-1)
		}
	}
}

func (store* ConfigurationStore) Save(){
	store.saveConfigToFile(store.trainerConfigurationFilePath)
}

func (store* ConfigurationStore) loadApplicationConfigurationsFromFile(filename string) {
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

func (store* ConfigurationStore) Add(name string, config *model.ApplicationConfiguration) *model.ApplicationConfiguration{
	state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
		Message: fmt.Sprintf("Adding application %s to orca", name),
		Details:map[string]string{
	}})

	store.ApplicationConfigurations[name] = config;
	return config
}

func (store* ConfigurationStore) saveConfigToFile(filename string) {
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
		return app, nil;
	}

	return nil, errors.New("Could not find application");
}

func (store *ConfigurationStore) GetAllConfiguration() (map[string]*model.ApplicationConfiguration) {
	return store.ApplicationConfigurations
}

func (store *ConfigurationStore) ApplySchedules() {
	for _, config := range store.ApplicationConfigurations {
		if config.DisableSchedule {
			continue
		}

		if config.DeploymentSchedule.Get(time.Now()) == -1 {
			continue
		}

		if config.DeploymentSchedule.Get(time.Now()) > config.MinDeployment {
			config.DesiredDeployment = config.DeploymentSchedule.Get(time.Now())
		} else {
			config.DesiredDeployment = config.MinDeployment
		}
	}
}
