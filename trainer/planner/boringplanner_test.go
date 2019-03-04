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
	"orca/trainer/schedule"
	"orca/trainer/state"
	"testing"
)

func TestPlan_spawnMinHosts(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	planner.Init(config.GlobalSettings)
	stateStore := state.StateStore{}
	stateStore.Init(&config)
	//state.Audit.Init(&config)
	versionConfigApp1 := make(map[string]*model.VersionConfig)
	versionConfigApp1["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}

	versionConfigApp2 := make(map[string]*model.VersionConfig)
	versionConfigApp2["2"] = &model.VersionConfig{
		Version:        "2",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}}}

	config.Add("app1", &model.ApplicationConfiguration{
		Name:               "app1",
		MinDeployment:      1,
		DesiredDeployment:  0,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})
	config.Add("app2", &model.ApplicationConfiguration{
		Name:               "app2",
		MinDeployment:      2,
		DesiredDeployment:  0,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp2,
		Enabled:            true,
	})

	res := planner.Plan(config, stateStore)
	if len(res) != 1 || res[0].Type != "new_server" || res[0].Network == "" {
		t.Errorf("%+v", res)
	}

	applied := make(map[string]bool)
	applied[res[0].Id] = true
	// create a appropriate host object and check in
	host1 := &model.Host{
		Id:             "host1",
		Network:        "network1",
		State:          "running",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}

	stateStore.Add("host1", host1)

	res2 := planner.Plan(config, stateStore)
	if len(res2) != 2 || res2[0].Type != "add_application" || res2[0].RequiresReliableInstance {
		t.Errorf("%+v", stateStore.GetAllHosts()["host1"])
		t.Errorf("%+v", res2)
	}
}


func TestPlan_spawnMinHostsTwoNeeded(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	planner.Init(config.GlobalSettings)
	stateStore := state.StateStore{}
	stateStore.Init(&config)
	//state.Audit.Init(&config)
	versionConfigApp1 := make(map[string]*model.VersionConfig)
	versionConfigApp1["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}

	versionConfigApp2 := make(map[string]*model.VersionConfig)
	versionConfigApp2["2"] = &model.VersionConfig{
		Version:        "2",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp2"}}}

	config.Add("app1", &model.ApplicationConfiguration{
		Name:               "app1",
		MinDeployment:      1,
		DesiredDeployment:  0,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})
	config.Add("app2", &model.ApplicationConfiguration{
		Name:               "app2",
		MinDeployment:      2,
		DesiredDeployment:  0,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp2,
		Enabled:            true,
	})

	res := planner.Plan(config, stateStore)
	if len(res) != 2 || res[0].Type != "new_server" || res[0].Network == "" {
		t.Errorf("%+v", res)
	}

	applied := make(map[string]bool)
	applied[res[0].Id] = true
	// create a appropriate host object and check in
	host1 := &model.Host{
		Id:             "host1",
		Network:        "network1",
		State:          "running",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}

	stateStore.Add("host1", host1)

	res2 := planner.Plan(config, stateStore)
	if len(res2) != 2 || res2[0].Type != "add_application" || res2[0].RequiresReliableInstance {
		t.Errorf("%+v", stateStore.GetAllHosts()["host1"])
		t.Errorf("%+v", res2)
	}
}

func TestPlan_spawnMinZeroDesiredOneHosts(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	planner.Init(config.GlobalSettings)

	stateStore := state.StateStore{}
	stateStore.Init(&config)
	//state.Audit.Init(&config)
	versionConfigApp1 := make(map[string]*model.VersionConfig)
	versionConfigApp1["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}

	config.Add("app1", &model.ApplicationConfiguration{
		Name:               "app1",
		MinDeployment:      0,
		DesiredDeployment:  1,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})

	res := planner.Plan(config, stateStore)

	//t.Errorf("%+v", res)
	if len(res) != 1 || res[0].Type != "new_server" || res[0].Network == "" {
		t.Errorf("%+v", res)
	}
}

func TestPlan_spawnMinHosts__OverrideFailing(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	planner.Init(config.GlobalSettings)

	stateStore := state.StateStore{}
	stateStore.Init(&config)
	//state.Audit.Init(&config)
	versionConfigApp1 := make(map[string]*model.VersionConfig)
	versionConfigApp1["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}

	config.Add("app1", &model.ApplicationConfiguration{
		Name:               "app1",
		MinDeployment:      1,
		DesiredDeployment:  0,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})

	host1 := &model.Host{
		Id:      "host1",
		Network: "network1",
		State:   "running",
		Apps: []model.Application{{
			State:    "failed",
			ChangeId: "",
			Name:     "app1",
			Version:  "1",
		}},
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}

	stateStore.Add("host1", host1)

	res2 := planner.Plan(config, stateStore)
	if len(res2) != 1 || res2[0].Type != "add_application" || res2[0].RequiresReliableInstance {
		t.Errorf("%+v", stateStore.GetAllHosts()["host1"])
		t.Errorf("%+v", res2)
	}
}

func TestPlan_spawnMinHosts__DontOverrideRunning(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	planner.Init(config.GlobalSettings)

	stateStore := state.StateStore{}
	stateStore.Init(&config)
	//state.Audit.Init(&config)
	versionConfigApp1 := make(map[string]*model.VersionConfig)
	versionConfigApp1["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}

	config.Add("app1", &model.ApplicationConfiguration{
		Name:               "app1",
		MinDeployment:      2,
		DesiredDeployment:  0,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})

	host1 := &model.Host{
		Id:      "host1",
		Network: "network1",
		State:   "running",
		Apps: []model.Application{{
			State:    "running",
			ChangeId: "",
			Name:     "app1",
			Version:  "1",
		}},
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}

	stateStore.Add("host1", host1)

	res := planner.Plan(config, stateStore)
	if len(res) != 1 || res[0].Type != "new_server" || res[0].Network == "" {
		t.Errorf("%+v", res)
	}
}

func TestPlan_spawnMinHosts__DontOverrideRunningOldVersion(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	planner.Init(config.GlobalSettings)

	stateStore := state.StateStore{}
	stateStore.Init(&config)
	//state.Audit.Init(&config)
	versionConfigApp1 := make(map[string]*model.VersionConfig)
	versionConfigApp1["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}

	config.Add("app1", &model.ApplicationConfiguration{
		Name:               "app1",
		MinDeployment:      2,
		DesiredDeployment:  0,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})

	host1 := &model.Host{
		Id:      "host1",
		Network: "network1",
		State:   "running",
		Apps: []model.Application{{
			State:    "running",
			ChangeId: "",
			Name:     "app1",
			Version:  "0",
		}},
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}

	stateStore.Add("host1", host1)

	res := planner.Plan(config, stateStore)
	if len(res) != 1 || res[0].Type != "new_server" || res[0].Network == "" {
		t.Errorf("%+v", res)
	}
}

func TestPlan_spawnMinHosts__DontUseSpotInstance(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	planner.Init(config.GlobalSettings)

	stateStore := state.StateStore{}
	stateStore.Init(&config)
	//state.Audit.Init(&config)
	versionConfigApp1 := make(map[string]*model.VersionConfig)
	versionConfigApp1["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}

	config.Add("app1", &model.ApplicationConfiguration{
		Name:               "app1",
		MinDeployment:      2,
		DesiredDeployment:  0,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})

	host1 := &model.Host{
		Id:             "host1",
		Network:        "network1",
		State:          "running",
		SpotInstance:   true,
		Apps:           []model.Application{},
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}

	stateStore.Add("host1", host1)

	res := planner.Plan(config, stateStore)
	if len(res) != 1 || res[0].Type != "new_server" || res[0].Network == "" {
		t.Errorf("%+v", res)
	}
}

func TestPlan_spawnDesiredHosts(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	planner.Init(config.GlobalSettings)

	stateStore := state.StateStore{}
	stateStore.Init(&config)
	//state.Audit.Init(&config)
	versionConfigApp1 := make(map[string]*model.VersionConfig)
	versionConfigApp1["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}
	config.Add("app1", &model.ApplicationConfiguration{
		Name:               "app1",
		MinDeployment:      1,
		DesiredDeployment:  4,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})

	res := planner.Plan(config, stateStore)
	if len(res) != 1 || res[0].Type != "new_server" || res[0].Network != "network1" {
		t.Errorf("%+v", res)
	}

	applied := make(map[string]bool)
	applied[res[0].Id] = true
	// create a appropriate host object and check in
	host1 := &model.Host{
		Id:             "host1",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}
	stateStore.HostInit(host1)
	stateStore.HostCheckin("host1", model.HostCheckinDataPackage{
		State:          []model.ApplicationStateFromHost{{"app1", model.Application{Name: "app1", State: "running", Version: "1", ChangeId: res[0].Id}}},
		ChangesApplied: applied,
	})
	res2 := planner.Plan(config, stateStore)
	if len(res2) != 1 || res2[0].Type != "new_server" || res2[0].RequiresReliableInstance || res2[0].Network != "network1" {
		t.Errorf("%+v", stateStore)
		t.Errorf("%+v", res2)
	}

	applied2 := make(map[string]bool)
	applied2[res[0].Id] = true
	// create a appropriate host object and check in
	stateStore.HostCheckin("host1", model.HostCheckinDataPackage{
		State:          []model.ApplicationStateFromHost{{"app1", model.Application{Name: "app1", State: "running", Version: "1", ChangeId: res[0].Id}}},
		ChangesApplied: applied2,
	})
	res3 := planner.Plan(config, stateStore)
	if len(res3) != 1 || res3[0].Type != "new_server" || res3[0].RequiresReliableInstance || res3[0].Network != "network1" {
		t.Errorf("%+v", stateStore)
		t.Errorf("%+v", res3)
	}
}

func TestPlan_DesiredOverrideFailingInstance(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	planner.Init(config.GlobalSettings)

	stateStore := state.StateStore{}
	stateStore.Init(&config)

	versionConfigApp1 := make(map[string]*model.VersionConfig)
	versionConfigApp1["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}
	config.Add("app1", &model.ApplicationConfiguration{
		Name:               "app1",
		MinDeployment:      1,
		DesiredDeployment:  4,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})

	res := planner.Plan(config, stateStore)
	if len(res) != 1 || res[0].Type != "new_server" || res[0].Network != "network1" {
		t.Errorf("%+v", res)
	}

	applied := make(map[string]bool)
	applied[res[0].Id] = true
	// create a appropriate host object and check in
	host1 := &model.Host{
		Id:             "host1",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}
	stateStore.HostInit(host1)
	stateStore.HostCheckin("host1", model.HostCheckinDataPackage{
		State:          []model.ApplicationStateFromHost{{"app1", model.Application{Name: "app1", State: "running", Version: "1", ChangeId: res[0].Id}}},
		ChangesApplied: applied,
	})
	res2 := planner.Plan(config, stateStore)
	if len(res2) != 1 || res2[0].Type != "new_server" || res2[0].RequiresReliableInstance || res2[0].Network != "network1" {
		t.Errorf("%+v", stateStore)
		t.Errorf("%+v", res2)
	}

	applied2 := make(map[string]bool)
	applied2[res[0].Id] = true
	// create a appropriate host object and check in
	stateStore.HostCheckin("host1", model.HostCheckinDataPackage{
		State:          []model.ApplicationStateFromHost{{"app1", model.Application{Name: "app1", State: "failed", Version: "1", ChangeId: res[0].Id}}},
		ChangesApplied: applied2,
	})
	res3 := planner.Plan(config, stateStore)
	if len(res3) != 1 || res3[0].Type != "add_application" {
		t.Errorf("%+v", stateStore)
		t.Errorf("%+v", res3)
	}
}

func TestPlan_DesiredOverrideOldInstance(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	planner.Init(config.GlobalSettings)

	stateStore := state.StateStore{}
	stateStore.Init(&config)

	versionConfigApp1 := make(map[string]*model.VersionConfig)
	versionConfigApp1["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}

	config.Add("app1", &model.ApplicationConfiguration{
		Name:               "app1",
		MinDeployment:      0,
		DesiredDeployment:  1,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})

	host1 := &model.Host{
		Id:             "host1",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		State:          "running",
		Apps:           []model.Application{{Name: "app1", State: "running", Version: "0"}},
	}

	stateStore.Add("host1", host1)
	res3 := planner.Plan(config, stateStore)
	if len(res3) != 1 || res3[0].Type != "add_application" {
		t.Errorf("%+v", stateStore)
		t.Errorf("%+v", res3)
	}
}

func TestPlan_DesiredOverrideOldInstance_Failing(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	planner.Init(config.GlobalSettings)

	stateStore := state.StateStore{}
	stateStore.Init(&config)

	versionConfigApp1 := make(map[string]*model.VersionConfig)
	versionConfigApp1["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}

	config.Add("app1", &model.ApplicationConfiguration{
		Name:               "app1",
		MinDeployment:      0,
		DesiredDeployment:  1,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})

	host1 := &model.Host{
		Id:             "host1",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		State:          "running",
		Apps:           []model.Application{{Name: "app1", State: "failed", Version: "0"}},
	}

	stateStore.Add("host1", host1)
	res3 := planner.Plan(config, stateStore)
	if len(res3) != 1 || res3[0].Type != "add_application" {
		t.Errorf("%+v", stateStore)
		t.Errorf("%+v", res3)
	}
}

func TestPlan_DesiredIgnoreExistingVersionRunning(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	planner.Init(config.GlobalSettings)

	stateStore := state.StateStore{}
	stateStore.Init(&config)

	versionConfigApp1 := make(map[string]*model.VersionConfig)
	versionConfigApp1["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}

	config.Add("app1", &model.ApplicationConfiguration{
		Name:               "app1",
		MinDeployment:      0,
		DesiredDeployment:  2,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})

	host1 := &model.Host{
		Id:             "host1",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		State:          "running",
		Apps:           []model.Application{{Name: "app1", State: "running", Version: "1"}},
	}

	stateStore.Add("host1", host1)
	res3 := planner.Plan(config, stateStore)
	if len(res3) != 1 || res3[0].Type != "new_server" {
		t.Errorf("%+v", stateStore)
		t.Errorf("%+v", res3)
	}
}

func TestPlan_scaleDown(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	planner.Init(config.GlobalSettings)

	stateStore := state.StateStore{}
	stateStore.Init(&config)
	//state.Audit.Init(&config)
	versionConfigApp1 := make(map[string]*model.VersionConfig)
	versionConfigApp1["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}
	config.Add("app1", &model.ApplicationConfiguration{
		Name:               "app1",
		MinDeployment:      1,
		DesiredDeployment:  2,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})

	//check in host objects with app1:
	host1 := &model.Host{
		Id:             "host1",
		Network:        "network1",
		State:          "running",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{{Name: "app1", Version: "1", State: "running"}},
	}
	stateStore.Add("host1", host1)

	host2 := &model.Host{
		Id:             "host2",
		Network:        "network1",
		State:          "running",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{{Name: "app1", Version: "1", State: "running"}},
	}
	stateStore.Add("host2", host2)

	host3 := &model.Host{
		Id:             "host3",
		Network:        "network1",
		State:          "running",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{{Name: "app1", Version: "1", State: "running"}},
	}
	stateStore.Add("host3", host3)

	res := planner.Plan(config, stateStore)
	if len(res) != 1 || res[0].Type != "remove_application" {
		t.Errorf("%+v", stateStore)
		t.Errorf("%+v", res)
	}
}

func TestPlan_scaleUp(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	planner.Init(config.GlobalSettings)

	stateStore := state.StateStore{}
	stateStore.Init(&config)
	versionConfigApp1 := make(map[string]*model.VersionConfig)
	versionConfigApp1["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}
	config.Add("app1", &model.ApplicationConfiguration{
		Name:               "app1",
		MinDeployment:      1,
		DesiredDeployment:  3,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})

	host1 := &model.Host{
		Id:             "host1",
		Network:        "network1",
		State:          "running",
		SpotInstance:   false,
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{{Name: "app1", Version: "1", State: "running"}},
	}
	stateStore.Add("host1", host1)

	host2 := &model.Host{
		Id:             "host2",
		State:          "running",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{{Name: "app1", Version: "1", State: "running"}},
	}
	stateStore.Add("host2", host2)

	host3 := &model.Host{
		Id:             "host3",
		State:          "running",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{},
	}
	stateStore.Add("host3", host3)

	res := planner.Plan(config, stateStore)
	if len(res) != 1 || res[0].Type != "add_application" {
		t.Errorf("%+v", res)
	}
}

func TestPlan_scaleUp_UsingAffinity(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	planner.Init(config.GlobalSettings)

	stateStore := state.StateStore{}
	stateStore.Init(&config)
	//state.Audit.Init(&config)

	/* App1 has an affinity with any other apps */
	versionConfigApp1 := make(map[string]*model.VersionConfig)
	versionConfigApp1["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		GroupingTag:    "tag1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}
	config.Add("app1", &model.ApplicationConfiguration{
		Name:               "app1",
		MinDeployment:      1,
		DesiredDeployment:  2,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})

	/* App2 does not have an affinity with any other apps */
	versionConfigApp2 := make(map[string]*model.VersionConfig)
	versionConfigApp2["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}
	config.Add("app2", &model.ApplicationConfiguration{
		Name:               "app2",
		MinDeployment:      1,
		DesiredDeployment:  1,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp2,
		Enabled:            true,
	})

	/* App3 has an affinity with any other apps */
	versionConfigApp3 := make(map[string]*model.VersionConfig)
	versionConfigApp3["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		GroupingTag:    "tag1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}

	config.Add("app3", &model.ApplicationConfiguration{
		Name:               "app3",
		MinDeployment:      1,
		DesiredDeployment:  1,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp3,
		Enabled:            true,
	})

	host1 := &model.Host{
		Id:             "host1",
		Network:        "network1",
		State:          "running",
		SpotInstance:   false,
		GroupingTag:    "",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{{Name: "app2", Version: "1", State: "running"}},
	}
	stateStore.Add("host1", host1)

	host2 := &model.Host{
		Id:             "host2",
		State:          "running",
		Network:        "network1",
		GroupingTag:    "tag1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{{Name: "app3", Version: "1", State: "running"}},
	}
	stateStore.Add("host2", host2)

	res := planner.Plan(config, stateStore)
	if len(res) != 1 || res[0].Type != "add_application" || res[0].HostId != "host2" {
		t.Errorf("%+v", res)
	}
}

func TestPlan_scaleUp_HitCapacity(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	config.GlobalSettings.ServerCapacity = 2

	planner.Init(config.GlobalSettings)
	stateStore := state.StateStore{}
	stateStore.Init(&config)

	versionConfigApp1 := make(map[string]*model.VersionConfig)
	versionConfigApp1["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
	}
	config.Add("app1", &model.ApplicationConfiguration{
		Name:               "app1",
		MinDeployment:      1,
		DesiredDeployment:  1,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})

	config.Add("app2", &model.ApplicationConfiguration{
		Name:               "app2",
		MinDeployment:      1,
		DesiredDeployment:  1,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})

	config.Add("app3", &model.ApplicationConfiguration{
		Name:               "app3",
		MinDeployment:      1,
		DesiredDeployment:  1,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})

	host1 := &model.Host{
		Id:             "host1",
		Network:        "network1",
		State:          "running",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{{Name: "app1", Version: "1", State: "running"}, {Name: "app2", Version: "1", State: "running"}},
	}
	stateStore.Add("host1", host1)

	host2 := &model.Host{
		Id:             "host2",
		State:          "running",
		Network:        "network1",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{},
	}
	stateStore.Add("host2", host2)

	res := planner.Plan(config, stateStore)
	if len(res) != 1 || res[0].Type != "add_application" || res[0].HostId != "host2" {
		t.Errorf("%+v", res)
	}
}

func TestPlan__Plan_RemoveOldDesired(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	stateStore := state.StateStore{}
	stateStore.Init(&config)

	planner.Init(config.GlobalSettings)
	//state.Audit.Init(&config)

	versionConfigApp1 := make(map[string]*model.VersionConfig)
	versionConfigApp1["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		SecurityGroups: make([]model.SecurityGroup, 0),
	}

	config.Add("app1", &model.ApplicationConfiguration{
		Name:               "app1",
		MinDeployment:      1,
		DesiredDeployment:  1,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})

	host1 := &model.Host{
		Id:             "host1",
		Network:        "network1",
		State:          "running",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{{Name: "app1", Version: "1", State: "running"}, {Name: "app2", Version: "1", State: "running"}},
	}
	stateStore.Add("host1", host1)

	host2 := &model.Host{
		Id:             "host2",
		Network:        "network1",
		State:          "running",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{{Name: "app1", Version: "1", State: "running"}},
	}
	stateStore.Add("host2", host2)

	host3 := &model.Host{
		Id:             "host3",
		Network:        "network1",
		State:          "running",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{{Name: "app1", Version: "1", State: "running"}},
	}
	stateStore.Add("host3", host3)

	changes := planner.Plan_RemoveOldDesired(config, stateStore)
	if len(changes) != 1 || changes[0].Type != "remove_application" && changes[0].HostId != "host2" {
		t.Errorf("%+v", changes)
	}
}

func TestPlan__Plan_KullUnusedServers(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	planner.Init(config.GlobalSettings)
	stateStore := state.StateStore{}
	stateStore.Init(&config)
	//state.Audit.Init(&config)

	host1 := &model.Host{
		Id:             "host1",
		Network:        "network1",
		State:          "running",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{{Name: "app1", Version: "1", State: "running"}, {Name: "app2", Version: "1", State: "running"}},
	}
	stateStore.Add("host1", host1)

	host2 := &model.Host{
		Id:             "host2",
		Network:        "network1",
		State:          "running",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{{Name: "app1", Version: "1", State: "running"}},
	}
	stateStore.Add("host2", host2)

	host3 := &model.Host{
		Id:             "host3",
		Network:        "network1",
		State:          "running",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{},
	}
	stateStore.Add("host3", host3)

	host4 := &model.Host{
		Id:             "host4",
		Network:        "network1",
		State:          "initialising",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{},
	}
	stateStore.Add("host4", host4)

	changes := planner.Plan_KullUnusedServers(config, stateStore)
	if len(changes) != 1 || changes[0].Type != "kill_server" && changes[0].HostId != "host3" {
		t.Errorf("%+v", changes)
	}

}

func TestPlan__Plan_BasicOptimise(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	planner.Init(config.GlobalSettings)

	stateStore := state.StateStore{}
	stateStore.Init(&config)
	//state.Audit.Init(&config)

	versionConfigApp1 := make(map[string]*model.VersionConfig)
	versionConfigApp1["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		SecurityGroups: make([]model.SecurityGroup, 0),
	}

	config.Add("app1", &model.ApplicationConfiguration{
		Name:               "app1",
		MinDeployment:      1,
		DesiredDeployment:  1,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})
	config.Add("app2", &model.ApplicationConfiguration{
		Name:               "app2",
		MinDeployment:      1,
		DesiredDeployment:  1,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})
	config.Add("app3", &model.ApplicationConfiguration{
		Name:               "app3",
		MinDeployment:      1,
		DesiredDeployment:  1,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})

	host1 := &model.Host{
		Id:             "host1",
		Network:        "network1",
		State:          "running",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{{Name: "app1", Version: "1", State: "running"}, {Name: "app2", Version: "1", State: "running"}},
	}
	stateStore.Add("host1", host1)

	host2 := &model.Host{
		Id:             "host2",
		Network:        "network1",
		State:          "running",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{{Name: "app3", Version: "1", State: "running"}},
	}
	stateStore.Add("host2", host2)

	host3 := &model.Host{
		Id:             "host4",
		Network:        "network1",
		State:          "initialising",
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{},
	}
	stateStore.Add("host3", host3)

	changes := planner.Plan_OptimiseLayout(config, stateStore)
	fmt.Println("%+v", changes)
	fmt.Printf("changes: %+v", changes)
	if len(changes) != 2 || changes[0].Type != "add_application" && changes[0].HostId != "host1" {
		t.Errorf("%+v", changes)
	}
	if len(changes) != 2 || changes[1].Type != "remove_application" && changes[1].HostId != "host2" {
		t.Errorf("%+v", changes)
	}
}

func TestPlan__Plan_ComplexOptimise_Step1(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	planner.Init(config.GlobalSettings)

	stateStore := state.StateStore{}
	stateStore.Init(&config)
	//state.Audit.Init(&config)

	versionConfigApp1 := make(map[string]*model.VersionConfig)
	versionConfigApp1["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		SecurityGroups: make([]model.SecurityGroup, 0),
	}

	config.Add("surfwizeweb", &model.ApplicationConfiguration{
		Name:               "surfwizeweb",
		MinDeployment:      1,
		DesiredDeployment:  2,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})
	config.Add("surfwizeauth", &model.ApplicationConfiguration{
		Name:               "surfwizeauth",
		MinDeployment:      1,
		DesiredDeployment:  2,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})
	config.Add("classwize", &model.ApplicationConfiguration{
		Name:               "classwize",
		MinDeployment:      1,
		DesiredDeployment:  2,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})
	config.Add("mylinewize", &model.ApplicationConfiguration{
		Name:               "mylinewize",
		MinDeployment:      1,
		DesiredDeployment:  2,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})

	host1 := &model.Host{
		Id:             "host1",
		Network:        "network1",
		State:          "running",
		SpotInstance:   true,
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{{Name: "surfwizeweb", Version: "1", State: "running"}, {Name: "surfwizeauth", Version: "1", State: "running"}, {Name: "classwize", Version: "1", State: "running"}},
	}
	stateStore.Add("host1", host1)

	host2 := &model.Host{
		Id:             "host2",
		Network:        "network1",
		State:          "running",
		SpotInstance:   true,
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{{Name: "mylinewize", Version: "1", State: "running"}, {Name: "surfwizeweb", Version: "1", State: "running"}, {Name: "surfwizeauth", Version: "1", State: "running"}},
	}
	stateStore.Add("host2", host2)

	host3 := &model.Host{
		Id:             "host4",
		Network:        "network1",
		State:          "running",
		SpotInstance:   false,
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{{Name: "mylinewize", Version: "1", State: "running"}, {Name: "surfwizeweb", Version: "1", State: "running"}, {Name: "surfwizeauth", Version: "1", State: "running"}, {Name: "classwize", Version: "1", State: "running"}},
	}
	stateStore.Add("host3", host3)

	changes := planner.Plan_OptimiseLayout(config, stateStore)
	fmt.Printf("changes: %+v", changes)
	if len(changes) != 2 || changes[0].Type != "add_application" && changes[0].ApplicationName == "classwize" && changes[0].HostId != "host2" {
		t.Errorf("%+v", changes)
	}
	if len(changes) != 2 || changes[1].Type != "remove_application" && changes[0].ApplicationName == "classwize" && changes[1].HostId != "host1" {
		t.Errorf("%+v", changes)
	}
}

func TestPlan__Plan_ComplexOptimise_Step2(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	planner.Init(config.GlobalSettings)

	stateStore := state.StateStore{}
	stateStore.Init(&config)

	versionConfigApp1 := make(map[string]*model.VersionConfig)
	versionConfigApp1["1"] = &model.VersionConfig{
		Version:        "1",
		Network:        "network1",
		SecurityGroups: make([]model.SecurityGroup, 0),
	}

	config.Add("surfwizeweb", &model.ApplicationConfiguration{
		Name:               "surfwizeweb",
		MinDeployment:      1,
		DesiredDeployment:  2,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})
	config.Add("surfwizeauth", &model.ApplicationConfiguration{
		Name:               "surfwizeauth",
		MinDeployment:      1,
		DesiredDeployment:  2,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})
	config.Add("classwize", &model.ApplicationConfiguration{
		Name:               "classwize",
		MinDeployment:      1,
		DesiredDeployment:  2,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})
	config.Add("mylinewize", &model.ApplicationConfiguration{
		Name:               "mylinewize",
		MinDeployment:      1,
		DesiredDeployment:  2,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		PublishedConfig:    versionConfigApp1,
		Enabled:            true,
	})

	host1 := &model.Host{
		Id:             "host1",
		Network:        "network1",
		State:          "running",
		SpotInstance:   true,
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{{Name: "surfwizeweb", Version: "1", State: "running"}, {Name: "surfwizeauth", Version: "1", State: "running"}},
	}
	stateStore.Add("host1", host1)

	host2 := &model.Host{
		Id:             "host2",
		Network:        "network1",
		State:          "running",
		SpotInstance:   true,
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{{Name: "mylinewize", Version: "1", State: "running"}, {Name: "surfwizeweb", Version: "1", State: "running"}, {Name: "surfwizeauth", Version: "1", State: "running"}, {Name: "classwize", Version: "1", State: "running"}},
	}
	stateStore.Add("host2", host2)

	host3 := &model.Host{
		Id:             "host3",
		Network:        "network1",
		State:          "running",
		SpotInstance:   false,
		SecurityGroups: []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:           []model.Application{{Name: "mylinewize", Version: "1", State: "running"}, {Name: "surfwizeweb", Version: "1", State: "running"}, {Name: "surfwizeauth", Version: "1", State: "running"}, {Name: "classwize", Version: "1", State: "running"}},
	}
	stateStore.Add("host3", host3)

	changes := planner.Plan_OptimiseLayout(config, stateStore)
	fmt.Printf("changes: %+v", changes)
	if len(changes) != 2 {
		t.Errorf("%+v", changes)
	}

	if changes[0].Type != "add_application" {
		t.Errorf("change is not valid %+v", changes[0])
	}
	if changes[1].Type != "remove_application" {
		t.Errorf("change is not valid %+v", changes[1])
	}
}

func TestPlan__Plan_HostWithFailedAppsAndErrors_Terminated(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	planner.Init(config.GlobalSettings)

	stateStore := state.StateStore{}
	stateStore.Init(&config)

	host1 := &model.Host{
		Id:                          "host1",
		Network:                     "network1",
		State:                       "running",
		SpotInstance:                true,
		NumberOfChangeFailuresInRow: 5,
		SecurityGroups:              []model.SecurityGroup{{Group: "secgrp1"}},
		Apps:                        []model.Application{{Name: "surfwizeweb", Version: "1", State: "failed"}, {Name: "surfwizeauth", Version: "1", State: "failed"}},
	}
	stateStore.Add("host1", host1)

	changes := planner.Plan_OptimiseLayout(config, stateStore)
	fmt.Printf("changes: %+v", changes)
	if len(changes) != 0 {
		t.Errorf("%+v", changes)
	}
}

func Test_OrderingByDependencies(t *testing.T) {
	planner := BoringPlanner{}

	config := configuration.ConfigurationStore{}
	config.Init("")
	planner.Init(config.GlobalSettings)

	stateStore := state.StateStore{}
	stateStore.Init(&config)

	config.Add("mylinewize", &model.ApplicationConfiguration{
		Name:               "mylinewize",
		MinDeployment:      1,
		DesiredDeployment:  2,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		Enabled:            true,
		Depends:            []model.Dependency{{Name: "api"}},
	})

	config.Add("surfwizeauth", &model.ApplicationConfiguration{
		Name:               "surfwizeauth",
		MinDeployment:      1,
		DesiredDeployment:  2,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		Enabled:            true,
		Depends:            []model.Dependency{{Name: "api"}},
	})

	config.Add("api", &model.ApplicationConfiguration{
		Name:               "api",
		MinDeployment:      1,
		DesiredDeployment:  2,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		Enabled:            true,
		Depends:            []model.Dependency{{Name: "acm"}},
	})

	config.Add("acm", &model.ApplicationConfiguration{
		Name:               "acm",
		MinDeployment:      1,
		DesiredDeployment:  2,
		DeploymentSchedule: schedule.DeploymentSchedule{},
		Enabled:            true,
	})

	items := config.GetAllConfigurationAsOrderedList()
	if items[0].Name != "acm" {
		t.Errorf("ordering is wrong, acm not first %+v", items)
	}
	if items[1].Name != "api" {
		t.Errorf("ordering is wrong, api not first %+v", items)
	}
	if items[2].Name != "mylinewize" && items[2].Name != "surfwizeauth" {
		t.Errorf("ordering is wrong, mylinewize not first %+v", items)
	}
	if items[3].Name != "mylinewize" && items[3].Name != "surfwizeauth" {
		t.Errorf("ordering is wrong, mylinewize not first %+v", items)
	}
	for _, app := range config.GetAllConfigurationAsOrderedList() {
		fmt.Printf("app: %s\n", app.Name)
	}
}
