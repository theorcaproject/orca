/*
Copyright Alex Mack and Michael Lawson (michael@sphinix.com)
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
	"encoding/json"
	"fmt"
	"orca/util"
	Logger "orca/trainer/logs"
	"io/ioutil"
	"orca/trainer/state"
	"orca/trainer/model"
)

type ConfigurationStore struct {
	Configurations map[string]*model.ApplicationConfiguration;
	AuditDatabaseUri string;

	trainerConfigurationFilePath string;
}

func (store *ConfigurationStore) Init(trainerConfigurationFilePath string){
	store.trainerConfigurationFilePath = trainerConfigurationFilePath
	store.Configurations = make(map[string]*model.ApplicationConfiguration);
}

func (store *ConfigurationStore) DumpConfig(){
	fmt.Printf("Loading config file from %+v", store.Configurations)
}

func (store* ConfigurationStore) Load(){
	store.loadFromFile(store.trainerConfigurationFilePath)
}

func (store* ConfigurationStore) Save(){
	store.saveConfigToFile(store.trainerConfigurationFilePath)
}

func (store* ConfigurationStore) loadFromFile(filename string) {
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
	state.Audit.Insert__AuditEvent(state.AuditEvent{Details:map[string]string{
		"message": "Adding application " + name + " to orca",
	}})

	store.Configurations[name] = config;
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
	if app, ok := store.Configurations[application]; ok {
		return app, nil;
	}

	return nil, errors.New("Could not find application");
}

func (store *ConfigurationStore) GetAllConfiguration() (map[string]*model.ApplicationConfiguration) {
	return store.Configurations
}
