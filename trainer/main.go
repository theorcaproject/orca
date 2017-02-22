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
	"orca/trainer/state"
	"orca/trainer/api"
	"orca/trainer/planner"
	"time"
	"github.com/twinj/uuid"
	"orca/trainer/cloud"
	"flag"
	"gopkg.in/mcuadros/go-syslog.v2"
	"orca/trainer/model"
	"orca/trainer/configuration"
	"strings"
)

func main() {
	fmt.Println("starting")

	configurationPath := flag.String("configroot", "/orca/configuration", "Orca Configuration Root")

	flag.Parse()

	store := &configuration.ConfigurationStore{};
	store.Init((*configurationPath) + "/trainer.conf")

	state_store := &state.StateStore{};
	state_store.Init(store)

	/* Load configuration */
	store.Load()

	/* Init connection to the database for auditing */
	state.Audit.Init(store)
	state.Stats.Init(store)

	var plannerEngine planner.Planner;
	if store.GlobalSettings.PlanningAlg == "boringplanner" {
		plannerEngine = &planner.BoringPlanner{}

	} else if store.GlobalSettings.PlanningAlg == "diffplan" {
		plannerEngine = &planner.DiffPlan{}
	}

	cloud_provider := cloud.CloudProvider{}

	if store.GlobalSettings.CloudProvider == "aws" {
		awsEngine := cloud.AwsCloudEngine{}
		awsEngine.Init(store.GlobalSettings.AWSAccessKeyId, store.GlobalSettings.AWSAccessKeySecret, store.GlobalSettings.AWSRegion, store.GlobalSettings.AWSBaseAmi, store.GlobalSettings.AWSSSHKey, store.GlobalSettings.AWSSSHKeyPath,
			store.GlobalSettings.AWSSpotPrice, store.GlobalSettings.InstanceType, store.GlobalSettings.SpotInstanceType)
		cloud_provider.Init(&awsEngine, store.GlobalSettings.InstanceUsername, store.GlobalSettings.Uri, store.GlobalSettings.LoggingUri)
	}

	startTime := time.Now()
	plannerAndTimeoutsTicker := time.NewTicker(time.Second * 10)
	go func() {
		for {
			<-plannerAndTimeoutsTicker.C
			if (startTime.Unix() + 120 > time.Now().Unix()) {
				continue
			}

			/* Check for published configuration */
			for _, app := range store.ApplicationConfigurations {
				if !app.Enabled {
					continue
				}

				latestConfiguredVersion := app.GetLatestConfiguration()
				latestPublishedVersion := app.GetLatestPublishedConfiguration()
				if latestPublishedVersion == nil || latestConfiguredVersion.GetVersion() > latestPublishedVersion.GetVersion() {
					/* Publish */
					state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
						Message: fmt.Sprintf("Publishing application configuration %s for app %s", latestConfiguredVersion.Version, app.Name),
						AppId: app.Name,
					})

					store.RequestPublishConfiguration(app)
					continue
				}

				/* Check the params */
				for _, propertyGroupName := range app.PropertyGroups {
					if item, ok :=latestPublishedVersion.AppliedPropertyGroups[propertyGroupName.Name]; ok {
						if item != store.Properties[propertyGroupName.Name].Version {
							/* Publish */
							state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
								Message: fmt.Sprintf("Publishing app configuration %s for app %s. Properties have been updated/modified", latestConfiguredVersion.Version, app.Name),
								AppId: app.Name,
							})

							store.RequestPublishConfiguration(app)
							continue
						}
					}else{
						/* Publish */
						state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
							Message: fmt.Sprintf("Publishing app configuration %s for app %s. Properties have been updated/modified", latestConfiguredVersion.Version, app.Name),
							AppId: app.Name,
						})

						store.RequestPublishConfiguration(app)
						continue
					}
				}
			}

			/* Check for timeouts */
			for _, host := range state_store.GetAllHosts() {
				for _, change := range host.Changes {
					parsedTime, _ := time.Parse(time.RFC3339Nano, change.Time)
					if (time.Now().Unix() - parsedTime.Unix()) > store.GlobalSettings.AppChangeTimeout {
						state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__ERROR,
							Message: fmt.Sprintf("Application change event %s timed out, event type was %s for application %s on host %s", change.Id, change.Type, change.Name, change.HostId),
							AppId: change.Name,
							HostId: change.Name,
							})
						state_store.RemoveChange(host.Id, change.Id)
						host.NumberOfChangeFailuresInRow += 1

						appConfiguration, _ := store.GetConfiguration(change.Name)
						appConfigurationVersion := appConfiguration.PublishedConfig[change.AppConfig.Version]
						if appConfigurationVersion != nil {
							appConfigurationVersion.DeploymentFailures += 1
						}

					}
				}
			}

			for _, change := range cloud_provider.GetAllChanges() {
				parsedTime, _ := time.Parse(time.RFC3339Nano, change.Time)
				if (time.Now().Unix() - parsedTime.Unix()) > store.GlobalSettings.ServerChangeTimeout {
					state.Audit.Insert__AuditEvent(state.AuditEvent{
						Severity: state.AUDIT__ERROR,
						Message: fmt.Sprintf("Server change event %s timed out, event type was %s with hostid %s", change.Id, change.Type, change.NewHostId),
						HostId: change.NewHostId,
						})

					/* If we were attempting to launch a spot instance, we need to relaunch it as a reserved */
					if change.Type == "new_server" && !change.RequiresReliableInstance {
						state.Audit.Insert__AuditEvent(state.AuditEvent{
							Severity: state.AUDIT__ERROR,
							Message: fmt.Sprintf("Failed to launch spot instance, change event was %s, attempting to relaunch on-demand instance", change.Id),
							HostId: change.NewHostId,
							})

						cloud_provider.ActionChange(&model.ChangeServer{
							Id:uuid.NewV4().String(),
							Type: "new_server",
							Time:time.Now().Format(time.RFC3339Nano),
							RequiresReliableInstance: true,
							Network: change.Network,
							SecurityGroups: change.SecurityGroups,
						}, state_store)
					}

					/* In-case the system actually launched an instance, nuke it from the system */
					if change.NewHostId != "" {
						cloud_provider.ActionChange(&model.ChangeServer{
							Id:uuid.NewV4().String(),
							Type: "remove",
							Time:time.Now().Format(time.RFC3339Nano),
							NewHostId:change.NewHostId,
						}, state_store)
					}

					cloud_provider.RemoveChange(change.Id, false)
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
						HostId: host.Id,
						})

					cloud_provider.ActionChange(&model.ChangeServer{
						Id:uuid.NewV4().String(),
						Type: "remove",
						Time:time.Now().Format(time.RFC3339Nano),
						NewHostId:host.Id,
					}, state_store)
				}
			}

			/* Check load balancer states */
			/* How can we do this in a scalable way?? */

			store.ApplySchedules()
			cloud_provider.Engine.SanityCheckHosts(state_store.GetAllHosts())

			/* Can we actually run the planner ? */
			if (state_store.HasChanges() || cloud_provider.HasChanges()) {
				continue;
			}

			changes := plannerEngine.Plan((*store), (*state_store))
			for _, change := range changes {
				if change.Type == "new_server" {
					state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
						Message: fmt.Sprintf("Planner requested a new server, spot: %t subnet: %s", !change.RequiresReliableInstance, change.Network),
					})

					/* Add new server */
					cloud_provider.ActionChange(&model.ChangeServer{
						Id:uuid.NewV4().String(),
						Type: "new_server",
						Time:time.Now().Format(time.RFC3339Nano),
						RequiresReliableInstance: change.RequiresReliableInstance,
						Network: change.Network,
						SecurityGroups: change.SecurityGroups,
						GroupingTag:change.GroupingTag,
					}, state_store)

					continue
				}
				if change.Type == "add_application" {
					/* Add new server */
					state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
						Message: fmt.Sprintf("Planner requested application %s be %s to host %s", change.ApplicationName, change.Type, change.HostId),
						AppId: change.ApplicationName,
						HostId: change.HostId,
						})

					host, _ := state_store.GetConfiguration(change.HostId)
					app, _ := store.GetConfiguration(change.ApplicationName)
					host.Changes = append(host.Changes, model.ChangeApplication{
						Id: uuid.NewV4().String(),
						Type: "add_application",
						HostId: host.Id,
						AppConfig: (*app.GetLatestPublishedConfiguration()),
						Name: change.ApplicationName,
						Time:time.Now().Format(time.RFC3339Nano),
					})

					for _, elb := range app.GetLatestConfiguration().LoadBalancer {
						state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
							Message: fmt.Sprintf("Registering host %s with load balancer %s for application %s", change.HostId, elb.Domain, change.ApplicationName),
							AppId: change.ApplicationName,
							HostId: change.HostId,
							})

						cloud_provider.ActionChange(&model.ChangeServer{
							Id:uuid.NewV4().String(),
							Type: "loadbalancer_join",
							Time:time.Now().Format(time.RFC3339Nano),
							LoadBalancerName: elb.Domain,
							NewHostId: change.HostId,
						}, state_store)
					}

					continue
				}

				if change.Type == "remove_application" {
					state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
						Message: fmt.Sprintf("Planner requested application %s be %s to host %s", change.ApplicationName, change.Type, change.HostId),
						AppId: change.ApplicationName,
						HostId: change.HostId,
					})

					host, _ := state_store.GetConfiguration(change.HostId)
					app, _ := store.GetConfiguration(change.ApplicationName)
					host.Changes = append(host.Changes, model.ChangeApplication{
						Id: uuid.NewV4().String(),
						Type: "remove_application",
						HostId: host.Id,
						AppConfig: (*app.GetLatestPublishedConfiguration()),
						Name: change.ApplicationName,
						Time:time.Now().Format(time.RFC3339Nano),
					})

					for _, elb := range app.GetLatestConfiguration().LoadBalancer {
						state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
							Message: fmt.Sprintf("Deregistering host %s with load balancer %s for application %s", change.HostId, elb.Domain, change.ApplicationName),
							AppId: change.ApplicationName,
							HostId: change.HostId,
						})

						cloud_provider.ActionChange(&model.ChangeServer{
							Id:uuid.NewV4().String(),
							Type: "loadbalancer_leave",
							Time:time.Now().Format(time.RFC3339Nano),
							LoadBalancerName: elb.Domain,
							NewHostId: change.HostId,
						}, state_store)
					}

					continue

				}
				if change.Type == "kill_server" {
					state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
						Message:fmt.Sprintf("Planner requested server %s be kulled in a bloodbath", change.HostId),
					})

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
	go func() {
		for {
			<-metricsCollectionTicker.C

			for appName, _ := range store.GetAllConfiguration() {
				metric := state.ApplicationUtilisationStatistic{}
				metric.AppName = appName
				metric.Timestamp = time.Now()

				for hostId, _ := range state_store.GetAllHosts() {
					appHostEntry, err := state_store.GetApplication(hostId, appName)
					if err == nil {
						metric.Cpu += appHostEntry.Metrics.CpuUsage
						metric.Mbytes += appHostEntry.Metrics.MemoryUsage
						metric.Network += appHostEntry.Metrics.NetworkUsage
						metric.InstanceCount += 1

						state.Stats.Insert__ApplicationHostUtilisationStatistic(state.ApplicationHostUtilisationStatistic{
							Cpu: appHostEntry.Metrics.CpuUsage,
							Mbytes: appHostEntry.Metrics.MemoryUsage,
							Network: appHostEntry.Metrics.NetworkUsage,
							Host: hostId,
							AppName: appName,
							Timestamp:time.Now(),
						})
					}
				}
				state.Stats.Insert__ApplicationUtilisationStatistic(metric)
			}
		}
	}()

	/* tart logging endpoint */
	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)
	server := syslog.NewServer()
	server.SetFormat(syslog.RFC3164)
	server.SetHandler(handler)
	server.ListenTCP(fmt.Sprintf("0.0.0.0:%d", store.GlobalSettings.LoggingPort))
	server.Boot()
	go func(channel syslog.LogPartsChannel) {
		for logParts := range channel {
			if hostId, ex := logParts["hostname"]; ex {
				if message, exists := logParts["content"]; exists {
					wholeMessage := message.(string)

					if len(wholeMessage) > 0 {
						entries := strings.Split( wholeMessage, "\n")
						for i := len(entries)- 1; i >= 0; i-- {
							state.Audit.Insert__Log(state.LogEvent{
								LogLevel: "stdout", HostId: hostId.(string), AppId: "", Message: entries[i],
							})
						}
					}
				}
			}
		}
	}(channel)

	api := api.Api{}
	api.Init(store.GlobalSettings.ApiPort, store, state_store, &cloud_provider)
}