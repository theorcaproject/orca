package configuration

import (
	"errors"
	"os"
	"encoding/json"
	"fmt"
	"gatoor/orca/util"
	Logger "gatoor/orca/trainer/logs"
	"io/ioutil"
)

type DockerConfig struct {
	Tag        string
	Repository string
	Reference  string
}

type PortMapping struct {
	HostPort      string
	ContainerPort string
}

type VolumeMapping struct {
	HostPath      string
	ContainerPath string
}

type File struct {
	HostPath           string
	Base64FileContents string
}

type EnvironmentVariable struct {
	Key   string
	Value string
}


type VersionConfig struct {
	Needs 		     string
	LoadBalancer         string
	Network              string
	PortMappings         []PortMapping
	VolumeMappings       []VolumeMapping
	EnvironmentVariables []EnvironmentVariable
	Files                []File
}

type ApplicationConfiguration struct {
	Name string
	MinDeployment int
	DesiredDeployment int
	Config map[int]VersionConfig

}

type ConfigurationStore struct {
	Configurations map[string]*ApplicationConfiguration;
}

func (store *ConfigurationStore) Init(){
	store.Configurations = make(map[string]*ApplicationConfiguration);
}

func (store *ConfigurationStore) DumpConfig(){
	fmt.Printf("Loading config file from %+v", store.Configurations)
}

func (store* ConfigurationStore) LoadFromFile(filename string) {
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

func (store* ConfigurationStore) Add(name string, config *ApplicationConfiguration) {
	store.Configurations[name] = config;
}

func (store* ConfigurationStore) SaveConfigToFile(filename string) {
	fmt.Printf("%+v", store)
	res, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		Logger.InitLogger.Errorf("JsonConfiguration Derialize failed: %s; %+v", err, store)
	}
	var result = string(res)
	fmt.Println(result)
	err = ioutil.WriteFile(filename, []byte(result), 0644)
	if err != nil {
		panic(err)
	}
}

func (store *ConfigurationStore) GetConfiguration(application string) (*ApplicationConfiguration, error) {
	if app, ok := store.Configurations[application]; !ok {
		return app, nil;
	}

	return nil, errors.New("Could not find application");
}

func (store *ConfigurationStore) GetAllConfiguration() (map[string]*ApplicationConfiguration) {
	return store.Configurations
}
