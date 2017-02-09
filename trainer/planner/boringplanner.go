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

package planner

import (
	"orca/trainer/configuration"
	"orca/trainer/state"
	"github.com/twinj/uuid"
	"fmt"
	"orca/trainer/model"
)

type BoringPlanner struct {
}

func (*BoringPlanner) Init() {

}

func hostIsSuitable(host *model.Host, app *model.ApplicationConfiguration) bool {
	if host.State != "running" {
		return false
	}

	/*
	We should not take the version into consideration here, if we do then we will break
	the min on older versions during an upgrade. Upgrades should be done on new hosts without impacting
	the old versions.
	*/
	if host.HasApp(app.Name) {
		return false
	}
	if host.Network != app.GetLatestConfiguration().Network {
		return false
	}
	//for _, grp := range host.SecurityGroups {
		//for _, appGrps := range app.GetLatestConfiguration().SecurityGroups {
		//	if app
		//}
		//if grp ==  {
		//	return true
		//}
	//}
	return false
}

func (planner *BoringPlanner) Plan(configurationStore configuration.ConfigurationStore, currentState state.StateStore) ([]PlanningChange) {
	fmt.Println("Starting BoringPlanner")
	ret := make([]PlanningChange, 0)

	requiresMinServer := false
	requiresSpotServer := false
	serverNetwork := ""
	var serverSecurityGroups []model.SecurityGroup


	for name, applicationConfiguration := range configurationStore.GetAllConfiguration() {
		if !applicationConfiguration.Enabled {
			continue
		}

		currentCount := 0
		for _, hostEntity := range currentState.GetAllHosts() {
			if hostEntity.HasAppWithSameVersion(name, applicationConfiguration.GetLatestVersion()) {
				currentCount += 1
			}
		}

		if currentCount < applicationConfiguration.MinDeployment {
			foundServer := false
			for _, hostEntity := range currentState.GetAllHosts() {
				if hostIsSuitable(hostEntity, applicationConfiguration) {
					fmt.Println(fmt.Sprintf("Found host for min deployment of app %s", applicationConfiguration.Name))
					change := PlanningChange{
						Type: "add_application",
						ApplicationName: name,
						HostId: hostEntity.Id,
						Id:uuid.NewV4().String(),
					}

					ret = append(ret, change)
					foundServer = true
					break
				}
			}

			if !foundServer {
				requiresMinServer = true
				serverNetwork = applicationConfiguration.GetLatestConfiguration().Network
				serverSecurityGroups = applicationConfiguration.GetLatestConfiguration().SecurityGroups
			}
		}

		//spawn to desired
		if currentCount >= applicationConfiguration.MinDeployment && currentCount < applicationConfiguration.DesiredDeployment {
			foundServer := false
			for _, hostEntity := range currentState.GetAllHosts() {
				if hostIsSuitable(hostEntity, applicationConfiguration) {
					fmt.Println(fmt.Sprintf("Found host for desired deployment of app %s", applicationConfiguration.Name))
					change := PlanningChange{
						Type: "add_application",
						ApplicationName: name,
						HostId: hostEntity.Id,
						Id:uuid.NewV4().String(),
					}

					ret = append(ret, change)
					foundServer = true
					break
				}
			}
			if !foundServer {
				requiresSpotServer = true
				serverNetwork = applicationConfiguration.GetLatestConfiguration().Network
				serverSecurityGroups = applicationConfiguration.GetLatestConfiguration().SecurityGroups
			}
		}

		//If the needs are greater than required, then scale them back
		if currentCount > applicationConfiguration.DesiredDeployment && currentCount > applicationConfiguration.MinDeployment{

			/* Can we kill of some extra desired machines? */
			if (applicationConfiguration.DesiredDeployment - applicationConfiguration.MinDeployment) > 0 {
				/* Find potential spot instances */
				terminateCandidateFound := false
				for _, hostEntity := range currentState.GetAllHosts() {
					if hostEntity.HasAppWithSameVersion(name, applicationConfiguration.GetLatestVersion()) {
						if hostEntity.SpotInstance {
							change := PlanningChange{
								Type: "remove_application",
								ApplicationName: name,
								HostId: hostEntity.Id,
								Id:uuid.NewV4().String(),
							}

							ret = append(ret, change)
							terminateCandidateFound = true
							break
						}
					}
				}

				if !terminateCandidateFound {
					for _, hostEntity := range currentState.GetAllHosts() {
						if hostEntity.HasAppWithSameVersion(name, applicationConfiguration.GetLatestVersion()) {
							change := PlanningChange{
								Type: "remove_application",
								ApplicationName: name,
								HostId: hostEntity.Id,
								Id:uuid.NewV4().String(),
							}

							ret = append(ret, change)
							break
						}
					}
				}
			} else {
				for _, hostEntity := range currentState.GetAllHosts() {
					if hostEntity.HasAppWithSameVersion(name, applicationConfiguration.GetLatestVersion()) {
						change := PlanningChange{
							Type: "remove_application",
							ApplicationName: name,
							HostId: hostEntity.Id,
							Id:uuid.NewV4().String(),
						}

						ret = append(ret, change)
						break
					}
				}
			}
		}

		//If we are running older version of the application, we can nuke them if the new versions needs are meet
		if currentCount >= applicationConfiguration.DesiredDeployment && currentCount >= applicationConfiguration.MinDeployment {
			for _, hostEntity := range currentState.GetAllHosts() {
				if hostEntity.HasApp(name) && !hostEntity.HasAppWithSameVersion(name, applicationConfiguration.GetLatestVersion()) {
					change := PlanningChange{
						Type: "remove_application",
						ApplicationName: name,
						HostId: hostEntity.Id,
						Id:uuid.NewV4().String(),
					}

					ret = append(ret, change)
				}
			}
		}
	}

	if requiresMinServer {
		fmt.Println("Planner: missing min server...")
		change := PlanningChange{
			Type: "new_server",
			Id:uuid.NewV4().String(),
			RequiresReliableInstance: true,
			Network: serverNetwork,
			SecurityGroups: serverSecurityGroups,
		}

		ret = append(ret, change)
	}

	if requiresSpotServer {
		fmt.Println("Planner: missing spot server...")
		change := PlanningChange{
			Type: "new_server",
			Id:uuid.NewV4().String(),
			RequiresReliableInstance: false,
			Network: serverNetwork,
			SecurityGroups: serverSecurityGroups,
		}

		ret = append(ret, change)
	}

	/* Second stage of planning: Terminate any instances that are left behind */
	if len(ret) == 0 {
		for _, hostEntity := range currentState.GetAllHosts() {
			if len(hostEntity.Apps) == 0 && hostEntity.State == "running" {
				change := PlanningChange{
					Type: "kill_server",
					HostId: hostEntity.Id,
					Id:uuid.NewV4().String(),
				}

				ret = append(ret, change)
			}
		}
	}

	fmt.Println(fmt.Sprintf("Planning changes: %+v", ret))
	return ret
}
