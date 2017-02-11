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

func main() {
	fmt.Println("starting")

	configurationPath := flag.String("configroot", "/orca/configuration", "Orca Configuration Root")

	flag.Parse()

	store := &configuration.ConfigurationStore{};
	store.Init((*configurationPath) + "/trainer.conf")

	state_store := &state.StateStore{};
	state_store.Init()

	/* Load configuration */
	store.Load()

	/* Init connection to the database for auditing */
	state.Audit.Init(store.AuditDatabaseUri)
	state.Stats.Init(store.AuditDatabaseUri)

	var plannerEngine planner.Planner;
	if store.GlobalSettings.PlanningAlg == "boringplanner"{
		plannerEngine = &planner.BoringPlanner{}

	}else if store.GlobalSettings.PlanningAlg == "diffplan" {
		plannerEngine = &planner.DiffPlan{}
	}

	cloud_provider := cloud.CloudProvider{}

	if store.GlobalSettings.CloudProvider == "aws" {
		awsEngine := cloud.AwsCloudEngine{}
		awsEngine.Init(store.GlobalSettings.AWSAccessKeyId, store.GlobalSettings.AWSAccessKeySecret, store.GlobalSettings.AWSRegion, store.GlobalSettings.AWSBaseAmi, store.GlobalSettings.AWSSSHKey, store.GlobalSettings.AWSSSHKeyPath,
			store.GlobalSettings.AWSSpotPrice, store.GlobalSettings.InstanceType, store.GlobalSettings.SpotInstanceType)
		cloud_provider.Init(&awsEngine, store.GlobalSettings.InstanceUsername, store.GlobalSettings.Uri)
	}

	startTime := time.Now()
	plannerAndTimeoutsTicker := time.NewTicker(time.Second * 10)
	go func () {
		for {
			<- plannerAndTimeoutsTicker.C
			if (startTime.Unix() + 2 * 120 > time.Now().Unix()) {
				continue
			}
			/* Check for timeouts */
			for _, host := range state_store.GetAllHosts() {
				for _, change := range host.Changes {
					parsedTime, _ := time.Parse(time.RFC3339Nano, change.Time)
					if (time.Now().Unix() - parsedTime.Unix()) > store.GlobalSettings.AppChangeTimeout {
						state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__ERROR,
							Message: fmt.Sprintf("Application change event %s timed out, event type was %s for application %s on host %s", change.Id, change.Type, change.Name, change.HostId),
							Details:map[string]string{
							"application": change.Name,
							"host": change.HostId,
						}})
						state_store.RemoveChange(host.Id, change.Id)
					}
				}
			}

			for _, change := range cloud_provider.GetAllChanges() {
					if host, exists := state_store.GetAllHosts()[change.NewHostId]; exists && host.State == "initializing" {
						continue
					}
					parsedTime, _ := time.Parse(time.RFC3339Nano, change.Time)
					if (time.Now().Unix() - parsedTime.Unix()) > store.GlobalSettings.ServerChangeTimeout {
						state.Audit.Insert__AuditEvent(state.AuditEvent{
							Severity: state.AUDIT__ERROR,
							Message: fmt.Sprintf("Server change event %s timed out, event type was %s with hostid %s", change.Id, change.Type, change.NewHostId),
							Details:map[string]string{
							"host": change.NewHostId,
						}})
						cloud_provider.RemoveChange(change.Id)
					}
			}

			/* Look for host timeouts */
			for _, host := range state_store.GetAllHosts() {
				if host.State == "initializing" {
					continue
				}
				parsedTime, _ := time.Parse(time.RFC3339Nano, host.LastSeen)
				if (time.Now().Unix() - parsedTime.Unix()) > store.GlobalSettings.ServerTimeout {
					state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__ERROR,
						Message:fmt.Sprintf("Host timed out, we have not heard from host %s since %s", host.Id, host.LastSeen),
						Details:map[string]string{
						"host": host.Id,
					}})

					cloud_provider.ActionChange(&model.ChangeServer{
						Id:uuid.NewV4().String(),
						Type: "remove",
						Time:time.Now().Format(time.RFC3339Nano),
						NewHostId:host.Id,
					}, state_store)
				}
			}

			store.ApplySchedules()
			cloud_provider.Engine.SanityCheckHosts(state_store.GetAllHosts())
			
			/* Can we actually run the planner ? */
			if(state_store.HasChanges() || cloud_provider.HasChanges()){
				continue;
			}

			changes := plannerEngine.Plan((*store), (*state_store))
			for _, change := range changes {
				if change.Type == "new_server" {
					state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
						Message: fmt.Sprintf("Planner requested a new server, spot: %t subnet: %s", change.RequiresReliableInstance, change.Network),
						Details:map[string]string{
					}})

					/* Add new server */
					cloud_provider.ActionChange(&model.ChangeServer{
						Id:uuid.NewV4().String(),
						Type: "new_server",
						Time:time.Now().Format(time.RFC3339Nano),
						RequiresReliableInstance: change.RequiresReliableInstance,
						Network: change.Network,
						SecurityGroups: change.SecurityGroups,
					}, state_store)

					continue
				}
				if change.Type == "add_application" || change.Type == "remove_application" {
					/* Add new server */
					state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
						Message: fmt.Sprintf("Planner requested application %s be %s to host %s", change.ApplicationName, change.Type, change.HostId),
						Details:map[string]string{
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
							state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
								Message: fmt.Sprintf("Registering host %s with load balancer %s for application %s", change.HostId, elb.Domain, change.ApplicationName),
								Details:map[string]string{
									"application": change.ApplicationName,
									"host": change.HostId,
								}})

							cloud_provider.RegisterWithLb(host.Id, elb.Domain)
						}
					}else if change.Type == "remove_application" {
						for _, elb := range app.GetLatestConfiguration().LoadBalancer {
							state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
								Message: fmt.Sprintf("Deregistering host %s with load balancer %s for application %s", change.HostId, elb.Domain, change.ApplicationName),
								Details:map[string]string{
									"application": change.ApplicationName,
									"host": change.HostId,
								}})

							cloud_provider.RegisterWithLb(host.Id, elb.Domain)
						}
					}

					continue
				}
				if change.Type == "kill_server" {
					state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
						Message:fmt.Sprintf("Planner requested server %s be kulled in a bloodbath", change.HostId),
						Details:map[string]string{
					}})

					cloud_provider.ActionChange(&model.ChangeServer{
						Id: change.Id,
						Type: "remove",
						NewHostId:change.HostId,
						Time:time.Now().Format(time.RFC3339Nano),
					}, state_store)
					continue
				}
			}
		}
	}()

	metricsCollectionTicker := time.NewTicker(time.Second * 120)
	go func () {
		for {
			<-metricsCollectionTicker.C

			for appName, _ := range store.GetAllConfiguration() {
				metric := state.ApplicationUtilisationStatistic{}
				metric.AppName = appName
				metric.Timestamp = time.Now()

				for hostId, _ := range state_store.GetAllHosts() {
					appHostEntry, err := state_store.GetApplication(hostId,  appName)
					if err == nil{
						metric.Cpu += appHostEntry.Metrics.CpuUsage
						metric.Mbytes += appHostEntry.Metrics.MemoryUsage
						metric.Network += appHostEntry.Metrics.NetworkUsage
						metric.InstanceCount += 1
					}
				}
				state.Stats.Insert__ApplicationUtilisationStatistic(metric)
			}
		}
	}()

	api := api.Api{}
	api.Init(store.GlobalSettings.ApiPort, store, state_store, &cloud_provider)
}

