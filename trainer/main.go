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
	"flag"
	"gatoor/orca/trainer/model"
)

const MAX_ELAPSED_TIME_FOR_APP_CHANGE = 120

func main() {
	fmt.Println("starting")
	var configurationRoot = flag.String("configroot", "/orca/config", "Configuration Root Directory")
	var apiPort = flag.Int("port", 5001, "API Port")
	flag.Parse()

	store := &configuration.ConfigurationStore{};
	store.Init(*configurationRoot + "/trainer.conf")

	state_store := &state.StateStore{};
	state_store.Init()

	store.Load()

	/* Init connection to the database for auditing */
	state.Audit.Init(store.AuditDatabaseUri)

	planner := planner.BoringPlanner{}
	planner.Init()

	cloud_provider := cloud.CloudProvider{}

	ticker := time.NewTicker(time.Second * 10)
	go func () {
		for {

			<- ticker.C
			fmt.Println("Running Planning task")

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
			if(state_store.HasChanges()){
				fmt.Println("Still have unresolved changes, waiting")
				continue;
			}

			changes := planner.Plan((*store), (*state_store))
			fmt.Println("Changes from planner: %+v", changes)
			for _, change := range changes {
				if change.Type == "new_server" {
					/* Add new server */
					cloud_provider.ActionChange(&model.ChangeServer{
						Id:uuid.NewV4().String(),
						Type: "add",
						Time:time.Now().Format(time.RFC3339Nano),
					})

					continue
				}
				if change.Type == "add_application" || change.Type == "remove_application" {
					/* Add new server */
					host, _ := state_store.GetConfiguration(change.HostId)
					app, _ := store.GetConfiguration(change.ApplicationName)
					host.Changes = append(host.Changes, model.ChangeApplication{
						Id: uuid.NewV4().String(),
						Type: change.Type,
						HostId: host.Id,
						AppConfig: app.GetLatestConfiguration(),
						Name: change.ApplicationName,
						Time:time.Now().Format(time.RFC3339Nano),
					})

					continue
				}
				if change.Type == "kill_server" {
					cloud_provider.ActionChange(&model.ChangeServer{
						Id:uuid.NewV4().String(),
						Type: "remove",
						Time:time.Now().Format(time.RFC3339Nano),
					})
					continue
				}
			}
		}
	}()

	api := api.Api{}
	api.Init(*apiPort, store, state_store)

}

