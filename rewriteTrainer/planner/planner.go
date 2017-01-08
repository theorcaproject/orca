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
	"sort"
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
					missingServerNeeds = append(missingServerNeeds, needs.AppNeeds{
					})
				}

				missingServerNeeds[0].CpuNeeds += appObject.Needs.CpuNeeds
				missingServerNeeds[0].MemoryNeeds += appObject.Needs.MemoryNeeds
				missingServerNeeds[0].NetworkNeeds += appObject.Needs.NetworkNeeds
				missingServerNeeds[0].SpotAllowed = false
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
							missingServerNeeds = append(missingServerNeeds, needs.AppNeeds{
								SpotAllowed:true,
							})
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
			SpotInstance: true,
		}

		bestInstanceType := findSuitableInstances(base.InstanceResources{
			TotalCpuResource: base.CpuResource(serverReqs.CpuNeeds),
			TotalMemoryResource: base.MemoryResource(serverReqs.MemoryNeeds),
			TotalNetworkResource: base.NetworkResource(serverReqs.NetworkNeeds),
		})

		/* Ok, so we found a good instance to work with, can we spot it? */
		bestInstanceTypeObject := cloud.CurrentProvider.GetAvailableInstances(bestInstanceType)
		if bestInstanceTypeObject.SupportsSpotInstance && serverReqs.SpotAllowed {
			if (bestInstanceTypeObject.LastSpotInstanceFailure.Unix() + trainerConfiguration.SpotInstanceFailureTimeThreshold) < time.Now().Unix() {
				bestInstanceTypeObject.LastSpotInstanceFailure = time.Unix(0,0)
				bestInstanceTypeObject.SpotInstanceTerminationCount = 0
				cloud.CurrentProvider.UpdateAvailableInstances(bestInstanceType, bestInstanceTypeObject)
			}

			if bestInstanceTypeObject.SpotInstanceTerminationCount >= trainerConfiguration.SpotInstanceFailureThreshold {
				change.SpotInstance = false
			}
		}else{
			change.SpotInstance = false
		}

		change.InstanceType = bestInstanceType
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

func checkResources(available base.InstanceResources, needed base.InstanceResources, safety float32) bool {
	if float32(available.TotalCpuResource) < float32(needed.TotalCpuResource) * safety {
		return false
	}
	if float32(available.TotalMemoryResource) < float32(needed.TotalMemoryResource) * safety {
		return false
	}
	if float32(available.TotalNetworkResource) < float32(needed.TotalNetworkResource) * safety {
		return false
	}
	return true
}

type CostSort struct {
	InstanceType base.InstanceType
	Cost         base.Cost
}

type CostSorts []CostSort

func (slice CostSorts) Len() int {
	return len(slice)
}

func (slice CostSorts) Less(i, j int) bool {
	return slice[i].Cost < slice[j].Cost;
}

func (slice CostSorts) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func sortByCost(tys []base.InstanceType) []base.InstanceType {
	sorted := CostSorts{}
	for _, ty := range tys {
		sorted = append(sorted, CostSort{InstanceType: ty, Cost: cloud.CurrentProvider.GetAvailableInstances(ty).InstanceCost})
	}
	sort.Sort(sorted)
	res := []base.InstanceType{}
	for _, t := range sorted {
		res = append(res, t.InstanceType)
	}
	return res
}

func findSuitableInstances(resources base.InstanceResources) base.InstanceType {
	suitableInstances := []base.InstanceType{}
	for _, ty := range cloud.CurrentProvider.GetAllAvailableInstanceTypes(){
		if checkResources(ty.InstanceResources, resources, 1) {
			suitableInstances = append(suitableInstances, ty.Type)
		}
	}
	suitableInstances = sortByCost(suitableInstances)
	return suitableInstances[0]
}


func doPromisedWork() {
	//TODO: Each spawn should be executed in a separate thread and sync to that thread. This way success
	//and failures can be dealt with correctly while not blocking future ops. We need to try hard to find a
	//suitable host, and deal with errors that could pop up
	changes := state_cloud.GlobalCloudLayout.Changes

	for _, change := range changes {
		if change.ChangeType == base.CHANGE_REQUEST__SPAWN_SERVER {
			//TODO: Work out which instance type we should be using here
			change.Host = cloud.CurrentProvider.SpawnInstanceSync(change.InstanceType, change.SpotInstance)

		} else if change.ChangeType == base.CHANGE_REQUEST__TERMINATE_SERVER {
			cloud.CurrentProvider.TerminateInstance(change.Host)
			host, err := state_cloud.GlobalCloudLayout.Current.GetHost(change.Host)
			if err == nil {
				host.HostState = state_cloud.HOST_PLANNING_TERMINATING
			}
			state_cloud.GlobalCloudLayout.Current.AddHost(change.Host, host)
			state_cloud.GlobalCloudLayout.DeleteChange(change.Id)
		}
	}
}

func doCheckForTimeoutHosts() {
	for _, host := range state_cloud.GlobalCloudLayout.Current.Layout {
		if (host.LastSeen.Unix() + trainerConfiguration.DeadHostTimeout) < time.Now().Unix() {
			if(host.HostState != state_cloud.HOST_PLANNING_TERMINATING){
				db.Audit.Insert__AuditEvent(db.AuditEvent{Details:map[string]string{
					"message": "Host "+ string(host.HostId) +" has disapeared, removing it from the system and nuking any changes associated with it",
					"host": string(host.HostId),
				}})

				/* Was this host a spot instance? If so then we might have lost it because shit is shit */
				if host.SpotInstance {
					providerInstanceMetadata := cloud.CurrentProvider.GetAvailableInstances(host.InstanceType)
					providerInstanceMetadata.LastSpotInstanceFailure = time.Now()
					providerInstanceMetadata.SpotInstanceTerminationCount += 1
					cloud.CurrentProvider.UpdateAvailableInstances(host.InstanceType, providerInstanceMetadata)
				}
			}

			state_cloud.GlobalCloudLayout.Current.RemoveHost(host.HostId)
			for _, change := range state_cloud.GlobalCloudLayout.Changes {
				if change.Host == host.HostId {
					state_cloud.GlobalCloudLayout.DeleteChange(change.Id)
				}
			}
		}
	}
}