package planner


import (
	"testing"
	"orca/trainer/configuration"
	"orca/trainer/state"
	"orca/trainer/model"
	"orca/trainer/schedule"
)

func TestPlan_spawnMinHosts(t *testing.T) {
	planner := BoringPlanner{}
	planner.Init()


	config := configuration.ConfigurationStore{}
	config.Init("")
	stateStore := state.StateStore{}
	stateStore.Init()
	state.Audit.Init("")
	versionConfigApp1 := make(map[string]model.VersionConfig)
	versionConfigApp1["1"] = model.VersionConfig{
		Version: "1",
		Network: "network1",
		SecurityGroup: "secgrp1",
	}
	versionConfigApp2 := make(map[string]model.VersionConfig)
	versionConfigApp2["2"] = model.VersionConfig{
		Version: "2",
		Network: "network1",
		SecurityGroup: "secgrp2",
	}
	versionConfigApp3 := make(map[string]model.VersionConfig)
	versionConfigApp3["3"] = model.VersionConfig{
		Version: "3",
		Network: "network3",
		SecurityGroup: "secgrp1",
	}
	config.Add("app1", &model.ApplicationConfiguration{
		Name: "app1",
		MinDeployment: 1,
		DesiredDeployment: 0,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		Config: versionConfigApp1,
	})
	config.Add("app2", &model.ApplicationConfiguration{
		Name: "app2",
		MinDeployment: 2,
		DesiredDeployment: 0,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		Config: versionConfigApp2,
	})
	config.Add("app3", &model.ApplicationConfiguration{
		Name: "app3",
		MinDeployment: 3,
		DesiredDeployment: 0,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		Config: versionConfigApp3,
	})

	res := planner.Plan(config, stateStore)
	if len(res) != 1 && res[0].ApplicationName != "app1"{
		t.Errorf("%+v", res);
	}

	applied := make(map[string]bool)
	applied[res[0].Id] = true
	stateStore.HostCheckin("host1", model.HostCheckinDataPackage{
		State: []model.ApplicationStateFromHost{{"app1", model.Application{Name: "app1", State: "running", Version: "1", ChangeId: res[0].Id}}},
		ChangesApplied: applied,
	})
	res2 := planner.Plan(config, stateStore)
	if len(res2) != 1 && res2[0].ApplicationName != "app2"{
		t.Errorf("%+v", res2);
	}
}

