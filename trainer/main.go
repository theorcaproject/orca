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

	/* TODO: Probably move this guy to a configuration file of its own? */
	globalSettings := configuration.GlobalSettings{}
	globalSettings.CloudProvider = (*flag.String("cloudprovider", "aws", "Cloud Provider"))
	globalSettings.AWSAccessKeyId = (*flag.String("awsaccesskeyid", "", "Amazon AWS Access Key"))
	globalSettings.AWSAccessKeySecret = (*flag.String("awsaccesskeysecret", "", "Amazon AWS Access Key Secret"))
	globalSettings.AWSRegion = (*flag.String("awsregion", "", "Amazon Region"))
	globalSettings.AWSBaseAmi = (*flag.String("awsbaseami", "", "Amazon Base AMI"))
	globalSettings.AWSSSHKey = (*flag.String("awssshkey", "", "Amazon SSH Key"))
	globalSettings.AWSSSHKeyPath = (*flag.String("awssshkeypath", "", "Amazon SSH Key Absolute Path"))
	globalSettings.PlanningAlg = (*flag.String("planner", "boringplanner", "Planning Algorithm"))
	globalSettings.InstanceUsername = (*flag.String("instanceusername", "ubuntu", "User account for the AMI"))
	globalSettings.Uri =  (*flag.String("uri", "http://localhost:5001", "Public Trainer Endpoint"))
	globalSettings.AWSSpotPrice = (*flag.Float64("spotbid", 0.5, "AWS Spot Instance Bid"))
	globalSettings.InstanceType = (*flag.String("instancetype", "t2.micro", "Regular instace type"))
	globalSettings.SpotInstanceType = (*flag.String("spotinstancetype", "c4.large", "Spot instance type"))
	globalSettings.ApiPort = 5001

	globalSettings.ServerChangeTimeout= (*flag.Int64("serverchangetimeout", 300, "Server Change Timeout"))
	globalSettings.AppChangeTimeout= (*flag.Int64("appchangetimeout", 300, "Application Change Timeout"))
	globalSettings.ServerTimeout= (*flag.Int64("servertimeout", 300, "Server Timeout"))

	flag.Parse()

	store := &configuration.ConfigurationStore{};
	store.Init(globalSettings.ConfigurationRoot + "/trainer.conf")

	state_store := &state.StateStore{};
	state_store.Init()

	store.Load()

	/* Init connection to the database for auditing */
	state.Audit.Init(store.AuditDatabaseUri)
	state.Stats.Init(store.AuditDatabaseUri)

	var plannerEngine planner.Planner;
	if globalSettings.PlanningAlg == "boringplanner"{
		plannerEngine = &planner.BoringPlanner{}

	}else if globalSettings.PlanningAlg == "diffplan" {
		plannerEngine = &planner.DiffPlan{}
	}

	cloud_provider := cloud.CloudProvider{}

	if globalSettings.CloudProvider == "aws" {
		awsEngine := cloud.AwsCloudEngine{}
		awsEngine.Init(globalSettings.AWSAccessKeyId, globalSettings.AWSAccessKeySecret, globalSettings.AWSRegion, globalSettings.AWSBaseAmi, globalSettings.AWSSSHKey, globalSettings.AWSSSHKeyPath,
			globalSettings.AWSSpotPrice, globalSettings.InstanceType, globalSettings.SpotInstanceType)
		cloud_provider.Init(&awsEngine, globalSettings.InstanceUsername, globalSettings.Uri)
	}

	startTime := time.Now()
	plannerAndTimeoutsTicker := time.NewTicker(time.Second * 10)
	go func () {
		for {
			<- plannerAndTimeoutsTicker.C
			if (startTime.Unix() + 2 * globalSettings.ServerTimeout > time.Now().Unix()) {
				continue
			}
			/* Check for timeouts */
			for _, host := range state_store.GetAllHosts() {
				for _, change := range host.Changes {
					parsedTime, _ := time.Parse(time.RFC3339Nano, change.Time)
					if (time.Now().Unix() - parsedTime.Unix()) > globalSettings.AppChangeTimeout {
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
					if (time.Now().Unix() - parsedTime.Unix()) > globalSettings.ServerChangeTimeout {
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
				if (time.Now().Unix() - parsedTime.Unix()) > globalSettings.ServerTimeout {
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
			
			/* Can we actually run the planner ? */
			if(state_store.HasChanges() || cloud_provider.HasChanges()){
				fmt.Println("Have Changes, wont plan...")
				for _, change := range cloud_provider.Changes {
					fmt.Println(fmt.Sprintf("%+v", change))
					fmt.Println(fmt.Sprintf("%s", change.Id))
				}
				continue;
			}

			changes := plannerEngine.Plan((*store), (*state_store))
			for _, change := range changes {
				if change.Type == "new_server" {
					state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
						Message: fmt.Sprintf("Planner requested a new server"),
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
	api.Init(globalSettings.ApiPort, store, state_store, &cloud_provider, &globalSettings)
}

