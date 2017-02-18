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
	"orca/trainer/model"
	"sort"
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

	count := 0
	for _, appGrp := range app.GetLatestConfiguration().SecurityGroups {
		for _, hostGrp := range host.SecurityGroups {
			if appGrp.Group == hostGrp.Group {
				count += 1
			}
		}
	}
	if count == len(app.GetLatestConfiguration().SecurityGroups) {
		return true
	}
	return false
}

/* Well this is rather nasty aint it */
func hostHasCorrectAffinity(host *model.Host, app *model.ApplicationConfiguration, configurationStore configuration.ConfigurationStore) bool {
	if len(host.Apps) == 0 {
		return true
	}

	for _, otherApps := range host.Apps {
		otherAppConfiguration, _ := configurationStore.GetConfiguration(otherApps.Name)
		affinity := otherAppConfiguration.Config[otherApps.Version].GroupingTag
		if affinity != app.GetLatestConfiguration().GroupingTag {
			return false
		}
	}

	return true
}

func isMinSatisfied(applicationConfiguration *model.ApplicationConfiguration, currentState *state.StateStore) bool {
	instanceCount := 0;
	for _, hostEntity := range currentState.GetAllHosts() {
		if hostEntity.SpotInstance {
			continue
		}

		/* Only use reserved instances when working with the min count */
		if hostEntity.HasAppWithSameVersion(applicationConfiguration.Name, applicationConfiguration.GetLatestVersion()) {
			instanceCount += 1
		}
	}

	return instanceCount >= applicationConfiguration.MinDeployment
}

func (planner *BoringPlanner) Plan_SatisfyMinNeeds(configurationStore configuration.ConfigurationStore, currentState state.StateStore) ([]PlanningChange) {
	ret := make([]PlanningChange, 0)

	requiresMinServer := false
	serverNetwork := ""
	var serverSecurityGroups []model.SecurityGroup

	for name, applicationConfiguration := range configurationStore.GetAllConfiguration() {
		if !applicationConfiguration.Enabled {
			continue
		}

		if !isMinSatisfied(applicationConfiguration, &currentState) {
			foundServer := false
			for _, hostEntity := range currentState.GetAllHosts() {
				/* Only use reserved instances when working with the min count */
				if hostIsSuitable(hostEntity, applicationConfiguration) && !hostEntity.SpotInstance && hostHasCorrectAffinity(hostEntity, applicationConfiguration, configurationStore) {
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
	}

	if requiresMinServer {
		change := PlanningChange{
			Type: "new_server",
			Id:uuid.NewV4().String(),
			RequiresReliableInstance: true,
			Network: serverNetwork,
			SecurityGroups: serverSecurityGroups,
		}

		ret = append(ret, change)
	}

	return ret
}

func (planner *BoringPlanner) Plan_SatisfyDesiredNeeds(configurationStore configuration.ConfigurationStore, currentState state.StateStore) ([]PlanningChange) {
	ret := make([]PlanningChange, 0)

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

		//spawn to desired
		if currentCount >= applicationConfiguration.MinDeployment && currentCount < applicationConfiguration.DesiredDeployment {
			foundServer := false
			for _, hostEntity := range currentState.GetAllHosts() {
				if hostIsSuitable(hostEntity, applicationConfiguration) && hostHasCorrectAffinity(hostEntity, applicationConfiguration, configurationStore) {
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
	}

	if requiresSpotServer {
		change := PlanningChange{
			Type: "new_server",
			Id:uuid.NewV4().String(),
			RequiresReliableInstance: false,
			Network: serverNetwork,
			SecurityGroups: serverSecurityGroups,
		}

		ret = append(ret, change)
	}

	return ret
}

func (planner *BoringPlanner) Plan_RemoveOldVersions(configurationStore configuration.ConfigurationStore, currentState state.StateStore) ([]PlanningChange) {
	ret := make([]PlanningChange, 0)
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
	return ret
}

type Hosts []*model.Host
type ByApplicationCount struct{ Hosts }

func (s Hosts) Len() int {
	return len(s)
}
func (s Hosts) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByApplicationCount) Less(i, j int) bool {
	return len(s.Hosts[i].Apps) < len(s.Hosts[j].Apps)
}

func (planner *BoringPlanner) Plan_RemoveOldDesired(configurationStore configuration.ConfigurationStore, currentState state.StateStore) ([]PlanningChange) {
	ret := make([]PlanningChange, 0)

	sortedHosts := currentState.ListOfHosts()
	sort.Sort(ByApplicationCount{sortedHosts})

	for name, applicationConfiguration := range configurationStore.GetAllConfiguration() {
		if !applicationConfiguration.Enabled {
			continue
		}

		currentCount := 0
		for _, hostEntity := range sortedHosts {
			if hostEntity.HasAppWithSameVersion(name, applicationConfiguration.GetLatestVersion()) {
				currentCount += 1
			}
		}

		/* Can we kill of some extra desired machines? */
		if currentCount > applicationConfiguration.DesiredDeployment && currentCount > applicationConfiguration.MinDeployment {
			if (applicationConfiguration.DesiredDeployment - applicationConfiguration.MinDeployment) > 0 {
				/* Find potential spot instances */
				terminateCandidateFound := false
				for _, hostEntity := range sortedHosts {
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
				for _, hostEntity := range sortedHosts {
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
	}
	return ret
}

func (planner *BoringPlanner) Plan_KullUnusedServers(configurationStore configuration.ConfigurationStore, currentState state.StateStore) ([]PlanningChange) {
	ret := make([]PlanningChange, 0)

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
	return ret
}

func (planner *BoringPlanner) Plan_KullBrokenServers(configurationStore configuration.ConfigurationStore, currentState state.StateStore) ([]PlanningChange) {
	ret := make([]PlanningChange, 0)

	for _, hostEntity := range currentState.GetAllHosts() {
		if hostEntity.State == "running" {
			/* This server is messing up, changes are failing for some reason */
			if hostEntity.NumberOfChangeFailuresInRow >= configurationStore.GlobalSettings.HostChangeFailureLimit {
				change := PlanningChange{
					Type: "kill_server",
					HostId: hostEntity.Id,
					Id:uuid.NewV4().String(),
				}

				ret = append(ret, change)
			}
		}
	}
	return ret
}

func (planner *BoringPlanner) Plan_KullBrokenApplications(configurationStore configuration.ConfigurationStore, currentState state.StateStore) ([]PlanningChange) {
	ret := make([]PlanningChange, 0)

	for _, hostEntity := range currentState.GetAllHosts() {
		for _, application := range hostEntity.Apps {
			if application.State != "running" {
				/* This application is messing up, if we have gotten to this stage then the mins and desired have already been dealt with for it */
				if hostEntity.NumberOfChangeFailuresInRow >= configurationStore.GlobalSettings.HostChangeFailureLimit {
					change := PlanningChange{
						Type: "remove_application",
						ApplicationName: application.Name,
						HostId: hostEntity.Id,
						Id:uuid.NewV4().String(),
					}

					ret = append(ret, change)
				}
			}
		}
	}
	return ret
}

func (planner *BoringPlanner) Plan_OptimiseLayout(configurationStore configuration.ConfigurationStore, currentState state.StateStore) ([]PlanningChange) {
	ret := make([]PlanningChange, 0)

	sortedHosts := currentState.ListOfHosts()
	sort.Sort(ByApplicationCount{sortedHosts})

	for _, hostEntity := range sortedHosts {

		for _, app := range hostEntity.Apps {
			appConfiguration, err := configurationStore.GetConfiguration(app.Name)
			if err != nil {
				continue
			}

			/* Now search, can we move this application to any other machine ?*/
			for _, potentialHost := range currentState.ListOfHosts() {
				if hostEntity.SpotInstance != potentialHost.SpotInstance {
					continue
				}

				if potentialHost.Id != hostEntity.Id && !potentialHost.HasAppWithSameVersion(app.Name, app.Version) && len(potentialHost.Apps) >= len(hostEntity.Apps) {
					if hostIsSuitable(potentialHost, appConfiguration) && hostHasCorrectAffinity(potentialHost, appConfiguration, configurationStore) {
						change := PlanningChange{
							Type: "add_application",
							ApplicationName: app.Name,
							HostId: potentialHost.Id,
							Id:uuid.NewV4().String(),
						}

						ret = append(ret, change)

						change2 := PlanningChange{
							Type: "remove_application",
							ApplicationName: app.Name,
							HostId: hostEntity.Id,
							Id:uuid.NewV4().String(),
						}

						ret = append(ret, change2)

						return ret
					}
				}
			}
		}
	}
	return ret
}

func extend(existing []PlanningChange, changes []PlanningChange) []PlanningChange {
	ret := make([]PlanningChange, 0)
	for _, change := range existing {
		ret = append(ret, change)
	}

	for _, change := range changes {
		ret = append(ret, change)
	}
	return ret
}

func (planner *BoringPlanner) Plan(configurationStore configuration.ConfigurationStore, currentState state.StateStore) ([]PlanningChange) {
	ret := make([]PlanningChange, 0)

	/* First step, deal with servers that are broken ? */
	ret = extend(ret, planner.Plan_KullBrokenServers(configurationStore, currentState))
	if len(ret) > 0 {
		return ret
	}

	/* First step, lets check that our min needs are satisfied? */
	ret = extend(ret, planner.Plan_SatisfyMinNeeds(configurationStore, currentState))
	if len(ret) > 0 {
		return ret
	}

	/* Ok, now that the mins are running, lets kull of old version of the app */
	ret = extend(ret, planner.Plan_RemoveOldVersions(configurationStore, currentState))
	if len(ret) > 0 {
		return ret
	}

	/* Ok, now that the min is sorted, lets scale down desired instances */
	ret = extend(ret, planner.Plan_RemoveOldDesired(configurationStore, currentState))
	if len(ret) > 0 {
		return ret
	}

	/* Grand, lets scale up the desired */
	ret = extend(ret, planner.Plan_SatisfyDesiredNeeds(configurationStore, currentState))
	if len(ret) > 0 {
		return ret
	}

	/* Second stage of planning: Terminate any instances that are left behind */
	ret = extend(ret, planner.Plan_KullBrokenApplications(configurationStore, currentState))
	if len(ret) > 0 {
		return ret
	}

	ret = extend(ret, planner.Plan_KullUnusedServers(configurationStore, currentState))
	if len(ret) > 0 {
		return ret
	}

	/* Third stage of planning: Move applications around to see if we can optimise it to be cheaper */
	ret = extend(ret, planner.Plan_OptimiseLayout(configurationStore, currentState))
	return ret
}
