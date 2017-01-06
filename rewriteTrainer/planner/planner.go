/*
Copyright Alex Mack
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

package planner

import (
	"gatoor/orca/base"
	"gatoor/orca/rewriteTrainer/state/cloud"
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/rewriteTrainer/state/configuration"
	"gatoor/orca/rewriteTrainer/cloud"
	"gatoor/orca/rewriteTrainer/needs"
	"gatoor/orca/rewriteTrainer/db"
	"time"
)

var PlannerLogger = Logger.LoggerWithField(Logger.Logger, "module", "planner")

var trainerConfiguration base.TrainerConfigurationState
func Init(configuration base.TrainerConfigurationState){
	trainerConfiguration = configuration
}

func Plan() {
	PlannerLogger.Info("Stating Plan()")

	doCheckForTimeoutHosts()
	doCheckForTimedOutChanges()

	/* We do not run several scheduling cycles at once */
	if len(state_cloud.GlobalCloudLayout.Changes) > 0 {
		return
	}

	doPlanInternal()
	doPromisedWork()
	PlannerLogger.Info("Finished Plan()")
}


func doCheckForTimedOutChanges() {
	for _, change := range state_cloud.GlobalCloudLayout.Changes {
		if change.ChangeType == base.CHANGE_REQUEST__SPAWN_SERVER {
			if (change.CreatedTime.Unix() + trainerConfiguration.ChangeSpawnTimeout) < time.Now().Unix() {
				db.Audit.Insert__AuditEvent(db.AuditEvent{Details:map[string]string{
					"message": "Change request " + change.Id + " of type " + change.ChangeType + " failed. Removing change so planning can continue",
				}})
				state_cloud.GlobalCloudLayout.DeleteChange(change.Id)
			}
		} else {
			if (change.CreatedTime.Unix() + trainerConfiguration.ChangeDefaultTimeout) < time.Now().Unix() {
				db.Audit.Insert__AuditEvent(db.AuditEvent{Details:map[string]string{
					"message": "Change request " + change.Id + " of type " + change.ChangeType + " failed. Removing change so planning can continue",
				}})
				state_cloud.GlobalCloudLayout.DeleteChange(change.Id)
			}
		}
	}
}

func doPlanInternal() {
	/* TODO: Spot instances:
		1. Always try to launch a spot instance unless the node is marked as critical.
		2. If the spot instance launch fails, then pick a more expensive instance to try and launch.
		3. If a running node drops of due to a spot instance culling, then immediately launch a more expensive instance, then perform planning.
	*/

	apps := state_configuration.GlobalConfigurationState.AllAppsLatest()
	missingServerNeeds := make([]needs.AppNeeds, 0)

	/* First check that the min needs are satisfied: Mins take priority over desired as they are part of the QOS we protect */
	for _, appObject := range apps {
		latestAppObjectConfiguration := appObject.LatestConfiguration()

		deployment_count, _ := state_cloud.GlobalCloudLayout.Current.DeploymentCount(appObject.Name, latestAppObjectConfiguration.Version)
		if appObject.MinDeploymentCount > deployment_count {
			/* Dudes we have an issue!! */
			/* Find a server that could meet our needs */

			foundResource := false
			for _, host := range state_cloud.GlobalCloudLayout.Current.Layout {
				if host.HostHasResourcesForApp(appObject.Needs) && !host.HostHasApp(appObject.Name) {
					state_cloud.GlobalCloudLayout.AddChange(base.ChangeRequest{
						Host: host.HostId,
						Application:appObject.Name,
						AppVersion:latestAppObjectConfiguration.Version,
						ChangeType:base.UPDATE_TYPE__ADD,
						Cost:appObject.Needs,

						AppConfig:appObject.LatestConfiguration(),
					})

					db.Audit.Insert__AuditEvent(db.AuditEvent{Details:map[string]string{
						"message": "Min deployment count not satisfied for application " + string(appObject.Name) + ", installing application on host " + string(host.HostId),
						"application": string(appObject.Name),
						"host": string(host.HostId),
					}})

					/* With this change in mind, reduce the usedResources so that we dont overpopulate this host */
					host.AvailableResources.UsedCpuResource += base.CpuResource(base.Resource(appObject.Needs.CpuNeeds))
					host.AvailableResources.UsedMemoryResource += base.MemoryResource(base.Resource(appObject.Needs.MemoryNeeds))
					host.AvailableResources.UsedNetworkResource += base.NetworkResource(base.Resource(appObject.Needs.NetworkNeeds))
					foundResource = true
					break
				}
			}

			if !foundResource {
				/* We could not find a server suitable for what we need */
				if len(missingServerNeeds) == 0 {
					missingServerNeeds = append(missingServerNeeds, needs.AppNeeds{})
				}

				missingServerNeeds[0].CpuNeeds += appObject.Needs.CpuNeeds
				missingServerNeeds[0].MemoryNeeds += appObject.Needs.MemoryNeeds
				missingServerNeeds[0].NetworkNeeds += appObject.Needs.NetworkNeeds
			}
		} else {
			//We have a couple of cases here to account for:
			//One: Desired is not meet
			if appObject.TargetDeploymentCount > deployment_count {
				remainingDeloymentCount := appObject.TargetDeploymentCount
				for _, host := range state_cloud.GlobalCloudLayout.Current.Layout {
					if remainingDeloymentCount > 0 {
						if host.HostHasResourcesForApp(appObject.Needs) && !host.HostHasApp(appObject.Name) {
							state_cloud.GlobalCloudLayout.AddChange(base.ChangeRequest{
								Host: host.HostId,
								Application:appObject.Name,
								AppVersion:latestAppObjectConfiguration.Version,
								ChangeType:base.UPDATE_TYPE__ADD,
								Cost:appObject.Needs,

								AppConfig:appObject.LatestConfiguration(),
							})

							db.Audit.Insert__AuditEvent(db.AuditEvent{Details:map[string]string{
								"message": "Desired deployment count not satisfied for application " + string(appObject.Name) + ", installing application on host " + string(host.HostId),
								"application": string(appObject.Name),
								"host": string(host.HostId),
							}})

							/* With this change in mind, reduce the usedResources so that we dont overpopulate this host */
							host.AvailableResources.UsedCpuResource += base.CpuResource(base.Resource(appObject.Needs.CpuNeeds))
							host.AvailableResources.UsedMemoryResource += base.MemoryResource(base.Resource(appObject.Needs.MemoryNeeds))
							host.AvailableResources.UsedNetworkResource += base.NetworkResource(base.Resource(appObject.Needs.NetworkNeeds))
							remainingDeloymentCount--
						}
					}
				}

				if remainingDeloymentCount > 0 {
					for i := 0; i < int(remainingDeloymentCount); i++ {
						if len(missingServerNeeds) > i - 1 {
							missingServerNeeds = append(missingServerNeeds, needs.AppNeeds{})
						}

						missingServerNeeds[i].CpuNeeds += appObject.Needs.CpuNeeds
						missingServerNeeds[i].MemoryNeeds += appObject.Needs.MemoryNeeds
						missingServerNeeds[i].NetworkNeeds += appObject.Needs.NetworkNeeds
					}
				}
			}

			//Two: Search for old versions of the application that we need to kull */
			for _, host := range state_cloud.GlobalCloudLayout.Current.Layout {
				if appsOfType, ok := host.Apps[appObject.Name]; ok {
					if appsOfType.Version != latestAppObjectConfiguration.Version {
						state_cloud.GlobalCloudLayout.AddChange(base.ChangeRequest{
							Host: host.HostId,
							Application:appObject.Name,
							AppVersion:appsOfType.Version,
							ChangeType:base.UPDATE_TYPE__REMOVE,
						})

						db.Audit.Insert__AuditEvent(db.AuditEvent{Details:map[string]string{
							"message": "Removing old application version " + string(appObject.Name) + " from host " + string(host.HostId),
							"application": string(appObject.Name),
							"host": string(host.HostId),
						}})
					}
				}
			}

			//Three: Search for excess resources being allocated and kull them off
			running_instance_counter := 0
			for _, host := range state_cloud.GlobalCloudLayout.Current.Layout {
				if appsOfType, ok := host.Apps[appObject.Name]; ok {
					if appsOfType.Version == latestAppObjectConfiguration.Version {
						running_instance_counter += 1
						if running_instance_counter > int(appObject.TargetDeploymentCount) {
							state_cloud.GlobalCloudLayout.AddChange(base.ChangeRequest{
								Host: host.HostId,
								Application:appObject.Name,
								AppVersion:latestAppObjectConfiguration.Version,
								ChangeType:base.UPDATE_TYPE__REMOVE,
							})

							db.Audit.Insert__AuditEvent(db.AuditEvent{Details:map[string]string{
								"message": "Removing excessive application version " + string(appObject.Name) + " from host " + string(host.HostId),
								"application": string(appObject.Name),
								"host": string(host.HostId),
							}})
						}
					}
				}
			}
		}
	}

	/* So, knowing what we now know, should we scale more servers up? */
	for _, serverReqs := range missingServerNeeds {
		change := base.ChangeRequest{
			ChangeType:base.CHANGE_REQUEST__SPAWN_SERVER,
		}
		change.Cost.CpuNeeds += serverReqs.CpuNeeds
		change.Cost.MemoryNeeds += serverReqs.MemoryNeeds
		change.Cost.NetworkNeeds += serverReqs.NetworkNeeds
		state_cloud.GlobalCloudLayout.AddChange(change)

		db.Audit.Insert__AuditEvent(db.AuditEvent{Details:map[string]string{
			"message": "Scaling server requirements were not meet, requesting new instance.",
		}})
	}

	/* If we are here, with no changes, then the system is running in either optimised or an excessive fashion. Can we reduce it? */
	if len(state_cloud.GlobalCloudLayout.Changes) == 0 {
		/* First lets kill servers with no applications on them */
		for _, host := range state_cloud.GlobalCloudLayout.Current.Layout {
			if len(host.Apps) == 0 {
				state_cloud.GlobalCloudLayout.AddChange(base.ChangeRequest{
					ChangeType:base.CHANGE_REQUEST__TERMINATE_SERVER,
					Host:host.HostId,
				})

				db.Audit.Insert__AuditEvent(db.AuditEvent{Details:map[string]string{
					"message": "Scaling server requirements are oversubscribed, terminating blank instance " + string(host.HostId),
					"host": string(host.HostId),
				}})
			}
		}

		/* Now lets see if there is resource we can move of servers */
		/* Can we relaunch an instance as a spot instance???*/
	}

	/* Done with this change iteration */
}

func doPromisedWork() {
	//TODO: Each spawn should be executed in a separate thread and sync to that thread. This way success
	//and failures can be dealt with correctly while not blocking future ops. We need to try hard to find a
	//suitable host, and deal with errors that could pop up
	changes := state_cloud.GlobalCloudLayout.Changes

	for _, change := range changes {
		if change.ChangeType == base.CHANGE_REQUEST__SPAWN_SERVER {
			//TODO: Work out which instance type we should be using here
			cloud.CurrentProvider.SpawnInstanceSync(base.InstanceType("t2.micro"))
			state_cloud.GlobalCloudLayout.DeleteChange(change.Id)

		} else if change.ChangeType == base.CHANGE_REQUEST__TERMINATE_SERVER {
			cloud.CurrentProvider.TerminateInstance(change.Host)
			state_cloud.GlobalCloudLayout.DeleteChange(change.Id)
		}
	}
}

func doCheckForTimeoutHosts() {
	for _, host := range state_cloud.GlobalCloudLayout.Current.Layout {
		if (host.LastSeen.Unix() + trainerConfiguration.DeadHostTimeout) < time.Now().Unix() {
			db.Audit.Insert__AuditEvent(db.AuditEvent{Details:map[string]string{
				"message": "Host "+ string(host.HostId) +" has disapeared, removing it from the system and nuking any changes associated with it",
				"host": string(host.HostId),
			}})

			state_cloud.GlobalCloudLayout.Current.RemoveHost(host.HostId)

			for _, change := range state_cloud.GlobalCloudLayout.Changes {
				if change.Host == host.HostId {
					state_cloud.GlobalCloudLayout.DeleteChange(change.Id)
				}
			}
		}
	}
}