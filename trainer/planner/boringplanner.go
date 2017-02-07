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
)

type BoringPlanner struct {
}

func (*BoringPlanner) Init() {

}

func (planner *BoringPlanner) Plan(configurationStore configuration.ConfigurationStore, currentState state.StateStore) ([]PlanningChange) {
	ret := make([]PlanningChange, 0)

	requiresMinServer := false
	requiresSpotServer := false

	for name, applicationConfiguration := range configurationStore.GetAllConfiguration() {
		currentCount := 0
		for _, hostEntity := range currentState.GetAllHosts() {
			if hostEntity.HasAppWithSameVersion(name, applicationConfiguration.GetLatestVersion()) {
				currentCount += 1
			}
		}

		if currentCount < applicationConfiguration.MinDeployment {
			foundServer := false
			for _, hostEntity := range currentState.GetAllHosts() {
				if !hostEntity.HasApp(name, applicationConfiguration.GetLatestVersion()) {
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
			}
		}

		//spawn to desired
		if currentCount >= applicationConfiguration.MinDeployment && currentCount < applicationConfiguration.DesiredDeployment {
			foundServer := false
			for _, hostEntity := range currentState.GetAllHosts() {
				if !hostEntity.HasApp(name, applicationConfiguration.GetLatestVersion()) {
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
			}
		}

		//If the needs are greater than required, then scale them back
		if currentCount > applicationConfiguration.DesiredDeployment {

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

		//If currentCount is greater than the Min, but desired is fine, then kill some nodes
		if currentCount > applicationConfiguration.MinDeployment && currentCount <= applicationConfiguration.DesiredDeployment {

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

	if requiresMinServer {
		change := PlanningChange{
			Type: "new_server",
			Id:uuid.NewV4().String(),
			RequiresReliableInstance: true,
		}

		ret = append(ret, change)
	}

	if requiresSpotServer {
		change := PlanningChange{
			Type: "new_server",
			Id:uuid.NewV4().String(),
			RequiresReliableInstance: false,
		}

		ret = append(ret, change)
	}

	/* Second stage of planning: Terminate any instances that are left behind */
	if len(ret) == 0 {
		for _, hostEntity := range currentState.GetAllHosts() {
			if len(hostEntity.Apps) == 0 {
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
