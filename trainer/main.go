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

package main

import (
	"fmt"
	"orca/trainer/configuration"
	"orca/trainer/state"
	"orca/trainer/api"
	"orca/trainer/planner"
	"time"
	"github.com/twinj/uuid"
	"orca/trainer/cloud"
	"flag"
	"orca/trainer/model"
)

const MAX_ELAPSED_TIME_FOR_APP_CHANGE = 120
const MAX_ELAPSED_TIME_FOR_SERVER_CHANGE = 300
const MAX_ELAPSED_TIME_FOR_HOST_CHECKIN = 60

func main() {
	fmt.Println("starting")
	var configurationRoot = flag.String("configroot", "/orca/config", "Configuration Root Directory")
	var apiPort = flag.Int("port", 5001, "API Port")

	//AWS Properties
	var cloudProvider = flag.String("cloudprovider", "aws", "Cloud Provider")
	var awsAccessKeyId = flag.String("awsaccesskeyid", "", "Amazon AWS Access Key")
	var awsAccessKeySecret = flag.String("awsaccesskeysecret", "", "Amazon AWS Access Key Secret")
	var awsRegion = flag.String("awsregion", "", "Amazon Region")
	var awsBaseAmi = flag.String("awsbaseami", "", "Amazon Base AMI")
	var awsSshKey = flag.String("awssshkey", "", "Amazon SSH Key")
	var awsSshKeyPath = flag.String("awssshkeypath", "", "Amazon SSH Key Absolute Path")
	var awsSecurityGroupId = flag.String("awssgroup", "", "Amazon Security Group")
	var plannerAlg = flag.String("planner", "boringplanner", "Planning Algorithm")
	var instanceUsername = flag.String("instanceusername", "ubuntu", "User account for the AMI")
	var uri = flag.String("uri", "http://localhost:5001", "Public Trainer Endpoint")
	var awsSpotPrice = flag.Float64("spotbid", 0.5, "AWS Spot Instance Bid")

	flag.Parse()

	store := &configuration.ConfigurationStore{};
	store.Init(*configurationRoot + "/trainer.conf")

	state_store := &state.StateStore{};
	state_store.Init()

	store.Load()

	/* Init connection to the database for auditing */
	state.Audit.Init(store.AuditDatabaseUri)

	var plannerEngine planner.Planner;
	if (*plannerAlg) == "boringplanner"{
		//WARNING: This planner is verrrry dumb, it will cost you moneyzzzzz
		//Mostly implemented to prove the system actually works, and the interface has been
		//defined well enough to support a more complicated planner
		plannerEngine = &planner.BoringPlanner{}

	}else if (*plannerAlg) == "diffplan" {
		//TODO: @Alex implement this guy
		plannerEngine = &planner.DiffPlan{}
	}

	cloud_provider := cloud.CloudProvider{}

	if (*cloudProvider) == "aws" {
		awsEngine := cloud.AwsCloudEngine{}
		awsEngine.Init((*awsAccessKeyId), (*awsAccessKeySecret), (*awsRegion), (*awsBaseAmi), (*awsSshKey),(*awsSshKeyPath), (*awsSecurityGroupId), (*awsSpotPrice))
		cloud_provider.Init(&awsEngine, (*instanceUsername), (*uri))
	}

	ticker := time.NewTicker(time.Second * 10)

	go func () {
		for {
			<- ticker.C
			/* Check for timeouts */
			for _, host := range state_store.GetAllHosts() {
				for _, change := range host.Changes {
					parsedTime, _ := time.Parse(time.RFC3339Nano, change.Time)
					if (time.Now().Unix() - parsedTime.Unix()) > MAX_ELAPSED_TIME_FOR_APP_CHANGE {
						state.Audit.Insert__AuditEvent(state.AuditEvent{Details:map[string]string{
							"message": fmt.Sprintf("Application change event %s timed out, event type was %s for application %s on host %s", change.Id, change.Type, change.Name, change.HostId),
							"application": change.Name,
							"host": change.HostId,
						}})

						state_store.RemoveChange(host.Id, change.Id)
					}
				}
			}

			for _, change := range cloud_provider.GetAllChanges() {
					parsedTime, _ := time.Parse(time.RFC3339Nano, change.Time)
					if (time.Now().Unix() - parsedTime.Unix()) > MAX_ELAPSED_TIME_FOR_SERVER_CHANGE {
						state.Audit.Insert__AuditEvent(state.AuditEvent{Details:map[string]string{
							"message": fmt.Sprintf("Server change event %s timed out, event type was %s with hostid %s", change.Id, change.Type, change.NewHostId),
							"host": change.NewHostId,
						}})

						cloud_provider.RemoveChange(change.Id)
					}
			}

			/* Look for host timeouts */
			for _, host := range state_store.GetAllHosts() {
				parsedTime, _ := time.Parse(time.RFC3339Nano, host.LastSeen)
				if (time.Now().Unix() - parsedTime.Unix()) > MAX_ELAPSED_TIME_FOR_HOST_CHECKIN {
					state.Audit.Insert__AuditEvent(state.AuditEvent{Details:map[string]string{
						"message": fmt.Sprintf("Host timed out, we have not heard from host %s since %s", host.Id, host.LastSeen),
						"host": host.Id,
					}})

					cloud_provider.ActionChange(&model.ChangeServer{
						Id:uuid.NewV4().String(),
						Type: "remove",
						Time:time.Now().Format(time.RFC3339Nano),
						NewHostId:host.Id,
					}, state_store)

					state_store.RemoveHost(host.Id)
				}
			}

			store.ApplySchedules()
			
			/* Can we actually run the planner ? */
			if(state_store.HasChanges() || cloud_provider.HasChanges()){
				continue;
			}

			changes := plannerEngine.Plan((*store), (*state_store))
			for _, change := range changes {
				if change.Type == "new_server" {
					state.Audit.Insert__AuditEvent(state.AuditEvent{Details:map[string]string{
						"message": fmt.Sprintf("Planner requested a new server"),
					}})

					/* Add new server */
					cloud_provider.ActionChange(&model.ChangeServer{
						Id:uuid.NewV4().String(),
						Type: "new_server",
						Time:time.Now().Format(time.RFC3339Nano),
						RequiresReliableInstance: change.RequiresReliableInstance,
					}, state_store)

					continue
				}
				if change.Type == "add_application" || change.Type == "remove_application" {
					/* Add new server */
					state.Audit.Insert__AuditEvent(state.AuditEvent{Details:map[string]string{
						"message": fmt.Sprintf("Planner requested application %s be %s to host %s", change.ApplicationName, change.Type, change.HostId),
						"application": change.ApplicationName,
						"host": change.HostId,
					}})

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

					if change.Type == "add_application" {
						for _, elb := range app.GetLatestConfiguration().LoadBalancer {
							cloud_provider.RegisterWithLb(host.Id, elb.Domain)
						}
					}else if change.Type == "remove_application" {
						for _, elb := range app.GetLatestConfiguration().LoadBalancer {
							cloud_provider.RegisterWithLb(host.Id, elb.Domain)
						}
					}

					continue
				}
				if change.Type == "kill_server" {
					state.Audit.Insert__AuditEvent(state.AuditEvent{Details:map[string]string{
						"message": fmt.Sprintf("Planner requested server %s be kulled in a bloodbath", change.HostId),
					}})

					cloud_provider.ActionChange(&model.ChangeServer{
						Id:uuid.NewV4().String(),
						Type: "remove",
						Time:time.Now().Format(time.RFC3339Nano),
					}, state_store)
					continue
				}
			}
		}
	}()

	api := api.Api{}
	api.Init(*apiPort, store, state_store, &cloud_provider)
}

