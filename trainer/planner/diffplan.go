package planner

import (
	"gatoor/orca/trainer/configuration"
	"gatoor/orca/trainer/state"
)

type DiffPlan struct {
}

func (*DiffPlan) Init() {

}

func (*DiffPlan) Plan(configurationStore configuration.ConfigurationStore, currentState state.StateStore) ([]PlanningChange) {
	return []PlanningChange{}
}
