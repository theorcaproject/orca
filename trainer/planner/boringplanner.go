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
	"fmt"
	"orca/trainer/configuration"
	"orca/trainer/model"
	"orca/trainer/state"
	"sort"
	"time"

	"github.com/twinj/uuid"
)

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

type BoringPlanner struct {
	AppChangeTimeout       int64
	ServerChangeTimeout    int64
	ServerTimeout          int64
	HostChangeFailureLimit int64
	ServerTTL              int64
	ServerCapacity         int64
}

func (bp *BoringPlanner) Init(globalConfig configuration.GlobalSettings) {
	bp.AppChangeTimeout = globalConfig.AppChangeTimeout
	bp.ServerTimeout = globalConfig.ServerChangeTimeout
	bp.ServerChangeTimeout = globalConfig.ServerChangeTimeout
	bp.HostChangeFailureLimit = globalConfig.HostChangeFailureLimit
	bp.ServerTTL = globalConfig.ServerTTL
	bp.ServerCapacity = globalConfig.ServerCapacity
}

func hostIsSuitable(host *model.Host, app *model.ApplicationConfiguration) bool {
	if host.State != "running" {
		return false
	}

	if host.Network != app.GetLatestPublishedConfiguration().Network {
		return false
	}

	count := 0
	for _, appGrp := range app.GetLatestPublishedConfiguration().SecurityGroups {
		for _, hostGrp := range host.SecurityGroups {
			if appGrp.Group == hostGrp.Group {
				count += 1
			}
		}
	}
	if count == len(app.GetLatestPublishedConfiguration().SecurityGroups) {
		return true
	}
	return false
}

func securityGroupsMatch(a []model.SecurityGroup, b []model.SecurityGroup) bool {
	for i, x := range a {
		if x.Group != b[i].Group {
			return false
		}
	}

	if len(a) != len(b) {
		return false
	}

	return true
}

/* Well this is rather nasty aint it */
func (planner *BoringPlanner) hostHasCorrectAffinity(host *model.Host, app *model.ApplicationConfiguration) bool {
	return host.GroupingTag == app.GetLatestPublishedConfiguration().GroupingTag
}

func (planner *BoringPlanner) hostHasCapacity(host *model.Host, configurationStore configuration.ConfigurationStore) bool {
	return int64(len(host.Apps)) < planner.ServerCapacity
}

func (planner *BoringPlanner) isMinSatisfied(applicationConfiguration *model.ApplicationConfiguration, currentState *state.StateStore) bool {
	instanceCount := 0
	for _, hostEntity := range currentState.GetAllRunningHosts() {
		if hostEntity.SpotInstance {
			continue
		}

		/* Only use reserved instances when working with the min count */
		if hostEntity.HasAppWithSameVersionRunning(applicationConfiguration.Name, applicationConfiguration.GetLatestPublishedVersion()) {
			instanceCount += 1
		}
	}

	return instanceCount >= applicationConfiguration.MinDeployment
}

func (planner *BoringPlanner) canDeploy(applicationConfiguration *model.ApplicationConfiguration) bool {
	if applicationConfiguration.GetLatestPublishedConfiguration() == nil {
		return false
	}

	if applicationConfiguration.GetLatestPublishedConfiguration().DeploymentFailures >= 2 && applicationConfiguration.GetLatestPublishedConfiguration().DeploymentSuccess == 0 {
		return false
	}

	return true
}

func (planner *BoringPlanner) FindServerInChanges(changes []PlanningChange, app *model.ApplicationConfiguration) bool {
	for _, newServerChange := range changes {
		if newServerChange.Type != "new_server" {
			continue
		}

		if newServerChange.Network != app.GetLatestPublishedConfiguration().Network {
			continue
		}

		if !securityGroupsMatch(newServerChange.SecurityGroups, app.GetLatestPublishedConfiguration().SecurityGroups) {
			continue
		}

		if newServerChange.InstanceType != app.GetLatestPublishedConfiguration().InstanceType {
			continue
		}

		if newServerChange.GroupingTag != app.GetLatestPublishedConfiguration().GroupingTag {
			continue
		}

		return true
	}

	return false
}

func (planner *BoringPlanner) Plan_SatisfyMinNeeds(configurationStore configuration.ConfigurationStore, currentState state.StateStore) []PlanningChange {
	ret := make([]PlanningChange, 0)

	for _, applicationConfiguration := range configurationStore.GetAllConfigurationAsOrderedList() {
		if !applicationConfiguration.Enabled {
			continue
		}

		if !planner.canDeploy(applicationConfiguration) {
			continue
		}

		if !planner.isMinSatisfied(applicationConfiguration, &currentState) {
			foundServer := false
			for _, hostEntity := range currentState.GetAllRunningHosts() {
				/* Only use reserved instances when working with the min count */
				if !hostIsSuitable(hostEntity, applicationConfiguration) {
					continue
				}

				if hostEntity.SpotInstance {
					continue
				}

				if !planner.hostHasCorrectAffinity(hostEntity, applicationConfiguration) {
					continue
				}

				if !planner.hostHasCapacity(hostEntity, configurationStore) {
					continue
				}

				/* Does this application override the normal instance type */
				if applicationConfiguration.GetLatestPublishedConfiguration().InstanceType != "" {
					if hostEntity.InstanceType != applicationConfiguration.GetLatestPublishedConfiguration().InstanceType {
						continue
					}
				}else{
					if hostEntity.InstanceType != configurationStore.GlobalSettings.InstanceType {
						continue
					}
				}

				/* If this host already has this application version and its running avoid */
				if hostEntity.HasAppWithSameVersionRunning(applicationConfiguration.Name, applicationConfiguration.GetLatestPublishedVersion()) {
					continue
				}

				/* If this host has an older version of the app running, avoid */
				if hostEntity.HasAppWithDifferentVersion(applicationConfiguration.Name, applicationConfiguration.GetLatestPublishedVersion()) && hostEntity.HasAppRunning(applicationConfiguration.Name) {
					continue
				}

				/*
					If we are here this means:
					1) This host does not have an older version of the app running
					2) This host does not have the same version of the app running
					3) This means either the app is not installed on this host, or a failing version is
					4) This host is not a spot instance
				*/
				change := PlanningChange{
					Type:            "add_application",
					ApplicationName: applicationConfiguration.Name,
					HostId:          hostEntity.Id,
					Id:              uuid.NewV4().String(),
				}

				ret = append(ret, change)
				foundServer = true
				break
			}

			if !foundServer && !planner.FindServerInChanges(ret, applicationConfiguration) {
				/* Search through the current changes and check to see if it will work */
				change := PlanningChange{
					Type: "new_server",
					Id:   uuid.NewV4().String(),
					RequiresReliableInstance: true,
					Network:                  applicationConfiguration.GetLatestPublishedConfiguration().Network,
					SecurityGroups:           applicationConfiguration.GetLatestPublishedConfiguration().SecurityGroups,
					GroupingTag:              applicationConfiguration.GetLatestPublishedConfiguration().GroupingTag,
					InstanceType:             applicationConfiguration.GetLatestPublishedConfiguration().InstanceType,
				}

				ret = append(ret, change)
			}
		}
	}

	return ret
}

func (planner *BoringPlanner) Plan_SatisfyDesiredNeeds(configurationStore configuration.ConfigurationStore, currentState state.StateStore) []PlanningChange {
	ret := make([]PlanningChange, 0)

	requiresSpotServer := false
	serverNetwork := ""
	groupingTag := ""
	instanceType := ""
	var serverSecurityGroups []model.SecurityGroup

	for _, applicationConfiguration := range configurationStore.GetAllConfigurationAsOrderedList() {
		if !applicationConfiguration.Enabled {
			continue
		}

		if !planner.canDeploy(applicationConfiguration) {
			continue
		}

		currentCount := 0
		for _, hostEntity := range currentState.GetAllRunningHosts() {
			if hostEntity.HasAppWithSameVersionRunning(applicationConfiguration.Name, applicationConfiguration.GetLatestPublishedVersion()) {
				currentCount += 1
			}
		}

		//spawn to desired
		if currentCount >= applicationConfiguration.MinDeployment && currentCount < applicationConfiguration.DesiredDeployment {
			foundServer := false
			for _, hostEntity := range currentState.GetAllRunningHosts() {
				if !hostIsSuitable(hostEntity, applicationConfiguration) {
					continue
				}

				if !planner.hostHasCorrectAffinity(hostEntity, applicationConfiguration) {
					continue
				}

				if !planner.hostHasCapacity(hostEntity, configurationStore) {
					continue
				}

				/* Does this application override the normal instance type */
				if applicationConfiguration.GetLatestPublishedConfiguration().InstanceType != "" {
					if hostEntity.InstanceType != applicationConfiguration.GetLatestPublishedConfiguration().InstanceType {
						continue
					}
				}else{
					if hostEntity.InstanceType != configurationStore.GlobalSettings.InstanceType {
						continue
					}
				}

				if hostEntity.HasAppWithSameVersionRunning(applicationConfiguration.Name, applicationConfiguration.GetLatestPublishedVersion()) {
					continue
				}

				change := PlanningChange{
					Type:            "add_application",
					ApplicationName: applicationConfiguration.Name,
					HostId:          hostEntity.Id,
					Id:              uuid.NewV4().String(),
				}

				ret = append(ret, change)
				foundServer = true
				break
			}

			if !foundServer {
				requiresSpotServer = true
				serverNetwork = applicationConfiguration.GetLatestPublishedConfiguration().Network
				serverSecurityGroups = applicationConfiguration.GetLatestPublishedConfiguration().SecurityGroups
				groupingTag = applicationConfiguration.GetLatestPublishedConfiguration().GroupingTag
				instanceType = applicationConfiguration.GetLatestPublishedConfiguration().InstanceType
			}
		}
	}

	if requiresSpotServer {
		change := PlanningChange{
			Type: "new_server",
			Id:   uuid.NewV4().String(),
			RequiresReliableInstance: false,
			Network:                  serverNetwork,
			SecurityGroups:           serverSecurityGroups,
			GroupingTag:              groupingTag,
			InstanceType:             instanceType,
		}

		ret = append(ret, change)
	}

	return ret
}

func (planner *BoringPlanner) Plan_RemoveOldVersions(configurationStore configuration.ConfigurationStore, currentState state.StateStore) []PlanningChange {
	ret := make([]PlanningChange, 0)
	for _, applicationConfiguration := range configurationStore.GetAllConfigurationAsOrderedList() {
		if !applicationConfiguration.Enabled {
			continue
		}

		currentCount := 0
		for _, hostEntity := range currentState.GetAllRunningHosts() {
			if hostEntity.HasAppWithSameVersionRunning(applicationConfiguration.Name, applicationConfiguration.GetLatestPublishedVersion()) {
				currentCount += 1
			}
		}

		if currentCount >= applicationConfiguration.DesiredDeployment && currentCount >= applicationConfiguration.MinDeployment {
			for _, hostEntity := range currentState.GetAllRunningHosts() {
				if hostEntity.HasApp(applicationConfiguration.Name) && !hostEntity.HasAppWithSameVersionRunning(applicationConfiguration.Name, applicationConfiguration.GetLatestPublishedVersion()) {
					change := PlanningChange{
						Type:            "remove_application",
						ApplicationName: applicationConfiguration.Name,
						HostId:          hostEntity.Id,
						Id:              uuid.NewV4().String(),
					}

					ret = append(ret, change)
				}
			}
		}
	}
	return ret
}

func (planner *BoringPlanner) Plan_RemoveOldDesired(configurationStore configuration.ConfigurationStore, currentState state.StateStore) []PlanningChange {
	ret := make([]PlanningChange, 0)

	sortedHosts := currentState.ListOfHosts()
	sort.Sort(ByApplicationCount{sortedHosts})

	for _, applicationConfiguration := range configurationStore.GetAllConfigurationAsOrderedList() {
		if !applicationConfiguration.Enabled {
			continue
		}

		currentCount := 0
		for _, hostEntity := range sortedHosts {
			if hostEntity.HasAppWithSameVersionRunning(applicationConfiguration.Name, applicationConfiguration.GetLatestPublishedVersion()) {
				currentCount += 1
			}
		}

		/* Can we kill of some extra desired machines? */
		if currentCount > applicationConfiguration.DesiredDeployment && currentCount > applicationConfiguration.MinDeployment {
			if (applicationConfiguration.DesiredDeployment - applicationConfiguration.MinDeployment) > 0 {
				/* Find potential spot instances */
				terminateCandidateFound := false
				for _, hostEntity := range sortedHosts {
					if hostEntity.HasAppWithSameVersionRunning(applicationConfiguration.Name, applicationConfiguration.GetLatestPublishedVersion()) {
						if hostEntity.SpotInstance {
							change := PlanningChange{
								Type:            "remove_application",
								ApplicationName: applicationConfiguration.Name,
								HostId:          hostEntity.Id,
								Id:              uuid.NewV4().String(),
							}

							ret = append(ret, change)
							terminateCandidateFound = true
							break
						}
					}
				}

				if !terminateCandidateFound {
					for _, hostEntity := range currentState.GetAllRunningHosts() {
						if hostEntity.HasAppWithSameVersionRunning(applicationConfiguration.Name, applicationConfiguration.GetLatestPublishedVersion()) {
							change := PlanningChange{
								Type:            "remove_application",
								ApplicationName: applicationConfiguration.Name,
								HostId:          hostEntity.Id,
								Id:              uuid.NewV4().String(),
							}

							ret = append(ret, change)
							break
						}
					}
				}
			} else {
				for _, hostEntity := range sortedHosts {
					if hostEntity.HasAppWithSameVersionRunning(applicationConfiguration.Name, applicationConfiguration.GetLatestPublishedVersion()) {
						change := PlanningChange{
							Type:            "remove_application",
							ApplicationName: applicationConfiguration.Name,
							HostId:          hostEntity.Id,
							Id:              uuid.NewV4().String(),
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

func (planner *BoringPlanner) Plan_KullUnusedServers(configurationStore configuration.ConfigurationStore, currentState state.StateStore) []PlanningChange {
	ret := make([]PlanningChange, 0)

	for _, hostEntity := range currentState.GetAllRunningHosts() {
		if len(hostEntity.Apps) == 0 {
			change := PlanningChange{
				Type:   "kill_server",
				HostId: hostEntity.Id,
				Id:     uuid.NewV4().String(),
				Reason: "Planner deemed this server to be unused, Plan_KullUnusedServers",
			}

			ret = append(ret, change)
		}
	}
	return ret
}

func (planner *BoringPlanner) Plan_KullBrokenServers(configurationStore configuration.ConfigurationStore, currentState state.StateStore) []PlanningChange {
	ret := make([]PlanningChange, 0)

	for _, hostEntity := range currentState.GetAllRunningHosts() {
		/* This server is messing up, changes are failing for some reason */
		if hostEntity.NumberOfChangeFailuresInRow >= planner.HostChangeFailureLimit {
			change := PlanningChange{
				Type:   "kill_server",
				HostId: hostEntity.Id,
				Id:     uuid.NewV4().String(),
				Reason: "Planner deemed this server to be broken, Plan_KullBrokenServers",
			}

			ret = append(ret, change)
		}
	}
	return ret
}

func (planner *BoringPlanner) Plan_KullBrokenApplications(configurationStore configuration.ConfigurationStore, currentState state.StateStore) []PlanningChange {
	ret := make([]PlanningChange, 0)

	for _, hostEntity := range currentState.GetAllRunningHosts() {
		for _, application := range hostEntity.Apps {
			if application.State != "running" {
				/* This application is messing up, if we have gotten to this stage then the mins and desired have already been dealt with for it */
				if hostEntity.NumberOfChangeFailuresInRow >= planner.HostChangeFailureLimit {
					change := PlanningChange{
						Type:            "remove_application",
						ApplicationName: application.Name,
						HostId:          hostEntity.Id,
						Id:              uuid.NewV4().String(),
					}

					ret = append(ret, change)
				}
			}
		}
	}
	return ret
}

func (planner *BoringPlanner) Plan_OptimiseLayout(configurationStore configuration.ConfigurationStore, currentState state.StateStore) []PlanningChange {
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
				if potentialHost.Id == hostEntity.Id {
					continue
				}

				if hostEntity.SpotInstance != potentialHost.SpotInstance {
					continue
				}

				/* Does this application override the normal instance type */
				if appConfiguration.GetLatestPublishedConfiguration().InstanceType != "" {
					if potentialHost.InstanceType != appConfiguration.GetLatestPublishedConfiguration().InstanceType {
						continue
					}
				}else{
					if potentialHost.InstanceType != configurationStore.GlobalSettings.InstanceType {
						continue
					}
				}

				if hostIsSuitable(potentialHost, appConfiguration) &&
					!potentialHost.HasAppWithSameVersionRunning(app.Name, app.Version) &&
					planner.hostHasCorrectAffinity(potentialHost, appConfiguration) &&
					planner.hostHasCapacity(potentialHost, configurationStore) &&
					len(potentialHost.Apps) >= len(hostEntity.Apps) {

					change := PlanningChange{
						Type:            "add_application",
						ApplicationName: app.Name,
						HostId:          potentialHost.Id,
						Id:              uuid.NewV4().String(),
					}

					ret = append(ret, change)

					change2 := PlanningChange{
						Type:            "remove_application",
						ApplicationName: app.Name,
						HostId:          hostEntity.Id,
						Id:              uuid.NewV4().String(),
					}

					ret = append(ret, change2)

					return ret
				}
			}
		}
	}
	return ret
}

func (planner *BoringPlanner) Plan_KullServersExceedingTTL(configurationStore configuration.ConfigurationStore, currentState state.StateStore) []PlanningChange {
	ret := make([]PlanningChange, 0)

	for _, hostEntity := range currentState.GetAllRunningHosts() {
		firstTimeParsed, _ := time.Parse(time.RFC3339Nano, hostEntity.FirstSeen)
		if planner.ServerTTL == 0 {
			continue
		}

		if (time.Now().Unix() - firstTimeParsed.Unix()) > planner.ServerTTL {
			change := PlanningChange{
				Type:   "retire_server",
				HostId: hostEntity.Id,
				Id:     uuid.NewV4().String(),
				Reason: "Server has exceeded the TTL configured, Plan_KullServersExceedingTTL",
			}

			ret = append(ret, change)
			break /* Only kull one server at once */
		}
	}
	return ret
}

func (planner *BoringPlanner) Plan_KullServersInTerminatingState(configurationStore configuration.ConfigurationStore, currentState state.StateStore) []PlanningChange {
	ret := make([]PlanningChange, 0)

	for _, hostEntity := range currentState.GetAllHosts() {
		if hostEntity.State == "terminating" {
			change := PlanningChange{
				Type:   "kill_server",
				HostId: hostEntity.Id,
				Id:     uuid.NewV4().String(),
				Reason: "Server has exceeded the TTL configured, Plan_KullServersInTerminatingState",
			}

			ret = append(ret, change)
		}
	}
	return ret
}

func (planner *BoringPlanner) Plan_KullServersResourceExceededState(configurationStore configuration.ConfigurationStore, currentState state.StateStore) []PlanningChange {
	ret := make([]PlanningChange, 0)

	for _, hostEntity := range currentState.GetAllHosts() {
		if hostEntity.State == "resourceExceeded" {
			change := PlanningChange{
				Type:   "retire_server",
				HostId: hostEntity.Id,
				Id:     uuid.NewV4().String(),
				Reason: "Server has exceeded one of it's resources (most likely HDD), Plan_KullServersResourceExceededState",
			}

			ret = append(ret, change)
		}else if hostEntity.State == "userTerminateRequested" {
			change := PlanningChange{
				Type:   "retire_server",
				HostId: hostEntity.Id,
				Id:     uuid.NewV4().String(),
				Reason: "Someone from the interface has requested that this instance be terminated",
			}

			ret = append(ret, change)
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

func (planner *BoringPlanner) Plan(configurationStore configuration.ConfigurationStore, currentState state.StateStore) []PlanningChange {
	ret := make([]PlanningChange, 0)

	/* First step, deal with servers that are broken ? */
	ret = extend(ret, planner.Plan_KullBrokenServers(configurationStore, currentState))
	if len(ret) > 0 {
		state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
			Message: fmt.Sprintf("Plan_KullBrokenServers had events"),
		})

		return ret
	}

	/* First step, lets check that our min needs are satisfied? */
	ret = extend(ret, planner.Plan_SatisfyMinNeeds(configurationStore, currentState))
	if len(ret) > 0 {
		state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
			Message: fmt.Sprintf("Plan_SatisfyMinNeeds had events"),
		})

		return ret
	}

	/* Ok, now that the mins are running, lets kull of old version of the app */
	ret = extend(ret, planner.Plan_RemoveOldVersions(configurationStore, currentState))
	if len(ret) > 0 {
		state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
			Message: fmt.Sprintf("Plan_RemoveOldVersions had events"),
		})

		return ret
	}

	/* Ok, now that the min is sorted, lets scale down desired instances */
	ret = extend(ret, planner.Plan_RemoveOldDesired(configurationStore, currentState))
	if len(ret) > 0 {
		state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
			Message: fmt.Sprintf("Plan_RemoveOldDesired had events"),
		})

		return ret
	}

	/* Grand, lets scale up the desired */
	ret = extend(ret, planner.Plan_SatisfyDesiredNeeds(configurationStore, currentState))
	if len(ret) > 0 {
		state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
			Message: fmt.Sprintf("Plan_SatisfyDesiredNeeds had events"),
		})

		return ret
	}

	/* Second stage of planning: Terminate any instances that are left behind */
	ret = extend(ret, planner.Plan_KullBrokenApplications(configurationStore, currentState))
	if len(ret) > 0 {
		state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
			Message: fmt.Sprintf("Plan_KullBrokenApplications had events"),
		})

		return ret
	}

	ret = extend(ret, planner.Plan_KullUnusedServers(configurationStore, currentState))
	if len(ret) > 0 {
		state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
			Message: fmt.Sprintf("Plan_KullUnusedServers had events"),
		})

		return ret
	}

	/* Third stage of planning: Move applications around to see if we can optimise it to be cheaper */
	ret = extend(ret, planner.Plan_OptimiseLayout(configurationStore, currentState))
	if len(ret) > 0 {
		state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
			Message: fmt.Sprintf("Plan_OptimiseLayout had events"),
		})

		return ret
	}

	/* Kull servers that are terminating. If we reach this step, MINS/DESIRED are meet so we can kill them of */
	ret = extend(ret, planner.Plan_KullServersInTerminatingState(configurationStore, currentState))
	if len(ret) > 0 {
		state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
			Message: fmt.Sprintf("Plan_KullServersInTerminatingState had events"),
		})

		return ret
	}

	/* Kull servers that have their resources exeeded. If we reach this step, MINS/DESIRED are meet so we can kill them of */
	ret = extend(ret, planner.Plan_KullServersResourceExceededState(configurationStore, currentState))
	if len(ret) > 0 {
		state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
			Message: fmt.Sprintf("Plan_KullServersResourceExceededState had events"),
		})

		return ret
	}

	/* Last stage of planning: Kill servers that are older than 24hours or configured TTL */
	ret = extend(ret, planner.Plan_KullServersExceedingTTL(configurationStore, currentState))
	return ret
}
