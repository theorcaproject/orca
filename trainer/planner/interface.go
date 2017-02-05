package planner

import (
	"gatoor/orca/trainer/configuration"
	"gatoor/orca/trainer/state"
)

type PlanningChange struct {
	Id string
	Type string /* Create Server, Add/Remove Application */

	/* Creation or removal of application */
	HostId string
	ApplicationName string

	/* Creation or removal of server*/
	InstanceId string
	InstanceNeeds string
}

type Planner interface {
	Init()

	Plan (configurationStore configuration.ConfigurationStore, currentState state.StateStore) ([]PlanningChange)
}
