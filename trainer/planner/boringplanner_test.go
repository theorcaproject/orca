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
		SecurityGroup: "secgrp1",
	}
	versionConfigApp3 := make(map[string]model.VersionConfig)
	versionConfigApp3["3"] = model.VersionConfig{
		Version: "3",
		Network: "network3",
		SecurityGroup: "secgrp3",
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
	if len(res) != 1 || res[0].Type != "new_server" {
		t.Errorf("%+v", res);
	}

	applied := make(map[string]bool)
	applied[res[0].Id] = true
	// create a appropriate host object and check in
	host1 := &model.Host{
		Id: "host1",
		Network: "network1",
		SecurityGroups: []string{"secgrp1"},
	}
	stateStore.HostInit(host1)
	stateStore.HostCheckin("host1", model.HostCheckinDataPackage{
		State: []model.ApplicationStateFromHost{{"app1", model.Application{Name: "app1", State: "running", Version: "1", ChangeId: res[0].Id}}},
		ChangesApplied: applied,
	})
	res2 := planner.Plan(config, stateStore)
	if len(res2) != 1 || res2[0].Type != "new_server" || !res2[0].RequiresReliableInstance {
		t.Errorf("%+v", stateStore);
		t.Errorf("%+v", res2);
	}


	//This app can be deployed to an existing host, new_server event for app3 because it is in the wrong network and securitygroup
	applied2 := make(map[string]bool)
	applied2[res2[0].Id] = true
	// create a appropriate host object and check in
	host2 := &model.Host{
		Id: "host2",
		Network: "network1",
		SecurityGroups: []string{"secgrp1"},
	}
	stateStore.HostInit(host2)
	stateStore.HostCheckin("host2", model.HostCheckinDataPackage{
		State: []model.ApplicationStateFromHost{{"app2", model.Application{Name: "app2", State: "running", Version: "2", ChangeId: res2[0].Id}}},
		ChangesApplied: applied,
	})

	res3 := planner.Plan(config, stateStore)
	if len(res3) != 2 || res3[0].ApplicationName != "app2" || res3[0].Type != "add_application" || res3[1].Type != "new_server" {
		t.Errorf("%+v", stateStore);
		t.Errorf("%+v", res3);
	}



}

