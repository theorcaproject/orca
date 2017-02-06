package planner

import (
	"gatoor/orca/trainer/configuration"
	"gatoor/orca/trainer/state"
	"github.com/twinj/uuid"
	"fmt"
)

type BoringPlanner struct {
}

func (*BoringPlanner) Init() {

}

func (*BoringPlanner) Plan(configurationStore configuration.ConfigurationStore, currentState state.StateStore) ([]PlanningChange) {
	ret := make([]PlanningChange, 0)

	for name, applicationConfiguration := range configurationStore.GetAllConfiguration() {
		fmt.Println("BoringPlanner checking application " + name + " with min %d", applicationConfiguration.MinDeployment)

		currentCount := 0
		for _, hostEntity := range currentState.GetAllHosts() {
			for _, runningApplicationState := range hostEntity.Apps {
				if runningApplicationState.Name == name && runningApplicationState.Version == applicationConfiguration.GetLatestVersion() {
					if runningApplicationState.State == "running" {
						currentCount += 1
					}
				}
			}
		}

		if currentCount < applicationConfiguration.MinDeployment {
			for _, hostEntity := range currentState.GetAllHosts() {
				if !hostEntity.HasApp(name, applicationConfiguration.GetLatestVersion()){
					change := PlanningChange{
						Type: "add_application",
						ApplicationName: name,
						HostId: hostEntity.Id,
						Id:uuid.NewV4().String(),
					}

					ret = append(ret, change)

				}
			}
		}

		if currentCount > applicationConfiguration.MinDeployment {
			for _, hostEntity := range currentState.GetAllHosts() {
				if hostEntity.HasApp(name, applicationConfiguration.GetLatestVersion()){
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
