package main

import (
	"fmt"
	"gatoor/orca/trainer/configuration"
	"gatoor/orca/trainer/state"
	"gatoor/orca/trainer/api"
)

func main() {
	fmt.Println("starting")

	store := &configuration.ConfigurationStore{};
	store.Init()

	state := &state.StateStore{};
	state.Init()

	versionMap := make(map[int]configuration.VersionConfig)
	versionConfig := configuration.VersionConfig{
		Needs: "needs",
		LoadBalancer:"lb1",
		Network:"aa",
		PortMappings:         []configuration.PortMapping{{HostPort:"11",ContainerPort:"22"}},
		VolumeMappings:       []configuration.VolumeMapping{{}},
		EnvironmentVariables: []configuration.EnvironmentVariable{{}},
		Files:                []configuration.File{{}},

	}
	versionMap[20] = versionConfig
	config := &configuration.ApplicationConfiguration{
		Name: "app1",
		MinDeployment: 1,
		DesiredDeployment: 2,
		Config: versionMap,
	}

	store.Add("someapp", config)
	store.SaveConfigToFile("/orca/configuration/trainer.conf")

	store.LoadFromFile("/orca/configuration/trainer.conf")
	store.DumpConfig()

	api := api.Api{}
	api.Init(5001, store, state)
}

