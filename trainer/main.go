package main

import (
	"fmt"
	"gatoor/orca/trainer/configuration"
	"gatoor/orca/trainer/state"
	"gatoor/orca/trainer/api"
	"gatoor/orca/trainer/planner"
	"time"
	"github.com/twinj/uuid"
	"gatoor/orca/trainer/cloud"
)

const MAX_ELAPSED_TIME_FOR_APP_CHANGE = 120

func main() {
	fmt.Println("starting")

	store := &configuration.ConfigurationStore{};
	store.Init()

	state_store := &state.StateStore{};
	state_store.Init()

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
	store.LoadFromFile("/orca/configuration/trainer.conf")
	store.DumpConfig()

	planner := planner.DiffPlan{}
	planner.Init()

	cloud_provider := cloud.CloudProvider{}

	ticker := time.NewTicker(time.Second * 1)
	go func () {
		for {
			<- ticker.C
			/* Check for timeouts */
			for _, host := range state_store.GetAllHosts() {
				for _, change := range host.Changes {
					parsedTime, _ := time.Parse(time.RFC3339Nano, change.Time)
					if (time.Now().Unix() - parsedTime.Unix()) > MAX_ELAPSED_TIME_FOR_APP_CHANGE {
						state_store.RemoveChange(host.Id, change.Id)
					}
				}
			}

			for _, change := range cloud_provider.GetAllChanges() {
					parsedTime, _ := time.Parse(time.RFC3339Nano, change.Time)
					if (time.Now().Unix() - parsedTime.Unix()) > MAX_ELAPSED_TIME_FOR_APP_CHANGE {
						cloud_provider.RemoveChange(change.Id)
					}
			}

			/* Can we actually run the planner ? */
			if(state_store.HasChanges() || cloud_provider.HasChanges()){
				continue;
			}

			changes := planner.Plan((*store), (*state_store))
			for _, change := range changes {
				if change.Type == "new_server" {
					/* Add new server */
					cloud_provider.ActionChange(&state.ChangeServer{
						Id:uuid.NewV4().String(),
						Type: "add",
					})

					continue
				}
				if change.Type == "add_application" || change.Type == "remove_application" {
					/* Add new server */
					host, _ := state_store.GetConfiguration(change.HostId)
					host.Changes = append(host.Changes, state.ChangeApplication{
						Id: uuid.NewV4().String(),
						Type: change.Type,
						HostId: host.Id,
					})

					continue
				}
				if change.Type == "kill_server" {
					cloud_provider.ActionChange(&state.ChangeServer{
						Id:uuid.NewV4().String(),
						Type: "remove",
					})
					continue
				}
			}
		}
	}()

	api := api.Api{}
	api.Init(5001, store, state_store)
}

