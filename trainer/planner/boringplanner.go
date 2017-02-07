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

func (*BoringPlanner) Plan(configurationStore configuration.ConfigurationStore, currentState state.StateStore) ([]PlanningChange) {
	ret := make([]PlanningChange, 0)

	requiresMinServer := false
	requiresSpotServer := false

	for name, applicationConfiguration := range configurationStore.GetAllConfiguration() {
		currentCount := 0
		for _, hostEntity := range currentState.GetAllHosts() {
			if hostEntity.HasApp(name, applicationConfiguration.GetLatestVersion()) {
				currentCount += 1
			}
		}

		fmt.Println(fmt.Sprintf("Planning: currentCount: %d, minDeployment: %d, desiredDeployment: %d, AllHosts: %+v", currentCount,
			applicationConfiguration.MinDeployment, applicationConfiguration.DesiredDeployment, currentState.GetAllHosts()))

		currentCount += 1
		if currentCount < applicationConfiguration.MinDeployment {
			foundServer := false
			for _, hostEntity := range currentState.GetAllHosts() {
				if !hostEntity.HasApp(name, applicationConfiguration.GetLatestVersion()){
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
				if !hostEntity.HasApp(name, applicationConfiguration.GetLatestVersion()){
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

	fmt.Println(fmt.Sprintf("Planning changes: %+v", ret))

	return ret
}
