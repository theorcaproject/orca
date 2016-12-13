/*
Copyright Alex Mack
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

package client

import (
	"testing"
	"gatoor/orca/client/types"
	"gatoor/orca/base"
	"time"
)

func TestInit(t *testing.T) {
	Configuration = types.Configuration{}
	Init()

	if cli.Type() != types.TEST_CLIENT {
		t.Error(cli)
	}

	Configuration.Type = types.DOCKER_CLIENT
	Init()

	if cli.Type() != types.DOCKER_CLIENT {
		t.Error(cli)
	}
}

func TestInstallApp(t *testing.T) {
	Configuration = types.Configuration{}
	Init()

	confFail := base.AppConfiguration{Name: "app_installfail", Version: 1.0, DockerConfig: base.DockerConfig{Tag: "1.0_installfail"}}
	res := InstallApp(confFail)
	if res {
		t.Error(res)
	}
	confOk := base.AppConfiguration{Name: "ok", Version: 1.0}
	resOk := InstallApp(confOk)
	if !resOk {
		t.Error(resOk)
	}
}

func TestRunApp(t *testing.T) {
	Configuration = types.Configuration{}
	Init()

	confFail := base.AppConfiguration{Name: "app_runfail", Version: 1.0, DockerConfig: base.DockerConfig{Tag: "1.0_runfail"}}
	res := RunApp(confFail)
	if res {
		t.Error(res)
	}
	for _, status := range AppsState {
		if status.Status != base.STATUS_DEAD {
			t.Error(AppsState)
		}
	}
	if len(AppsConfiguration) != 1 {
		t.Error(AppsConfiguration)
	}
	AppsState = make(map[base.AppId]base.AppInfo)
	confOk := base.AppConfiguration{Name: "ok", Version: 1.0, DockerConfig: base.DockerConfig{Tag: "1.0"}}
	resOk := RunApp(confOk)
	if !resOk {
		t.Error(resOk)
	}
	for _, status := range AppsState {
		if status.Status != base.STATUS_DEPLOYING {
			t.Error(AppsState)
		}
	}
	if len(AppsConfiguration) != 2 {
		t.Error(AppsConfiguration)
	}
}

func TestStopAll(t *testing.T) {
	Configuration = types.Configuration{}
	Init()
	confOk := base.AppConfiguration{Name: "app1", Version: 1.0, DockerConfig: base.DockerConfig{Tag: "1.0"}}
	confOk2 := base.AppConfiguration{Name: "app1", Version: 1.0, DockerConfig: base.DockerConfig{Tag: "1.0"}}
	conf3 := base.AppConfiguration{Name: "app2", Version: 1.0, DockerConfig: base.DockerConfig{Tag: "1.0_stopfail"}}
	if !RunApp(confOk) {
		t.Error(AppsState)
	}
	if !RunApp(confOk2) {
		t.Error(AppsState)
	}

	for _, status := range AppsState {
		if status.Status != base.STATUS_DEPLOYING {
			t.Error(AppsState)
		}
	}
	if !StopAll("app1") {
		t.Error(AppsState)
	}
	for _, status := range AppsState {
		if status.Status != base.STATUS_DEAD {
			t.Error(AppsState)
		}
	}
	if !RunApp(conf3) {
		t.Error(AppsState)
	}
	if StopAll("app2") {
		t.Error(AppsState)
	}
}


func TestStopApp(t *testing.T) {
	Configuration = types.Configuration{}
	Init()
	confOk := base.AppConfiguration{Name: "ok", Version: 1.0, DockerConfig: base.DockerConfig{Tag: "1.0"}}
	if !RunApp(confOk) {
		t.Error(AppsState)
	}
	var appId base.AppId
	for id, status := range AppsState {
		appId = id
		if status.Status != base.STATUS_DEPLOYING {
			t.Error(AppsState)
		}
	}
	if !StopApp(appId) {
		t.Error(AppsState)
	}
	if AppsState.Get(appId).Status != base.STATUS_DEAD  {
		t.Error(AppsState)
	}
}

func TestScaleApp_up(t *testing.T) {
	Configuration = types.Configuration{}
	Init()

	existsing := []base.AppInfo{{Name: "app1", Version: 1.0, Status: base.STATUS_RUNNING, Id: "app1_1"}, {Name: "app1", Version: 1.0, Status: base.STATUS_RUNNING, Id: "app1_2"}}
	AppsState.Add("app1_1", base.AppInfo{Name: "app1", Version: 1.0, Status: base.STATUS_RUNNING, Id: "app1_1"})
	AppsState.Add("app1_2", base.AppInfo{Name: "app1", Version: 1.0, Status: base.STATUS_RUNNING, Id: "app1_2"})
	AppsState.Add("app2_1", base.AppInfo{Name: "app2", Version: 1.0, Status: base.STATUS_RUNNING, Id: "app2_1"})
	config := base.PushConfiguration{DeploymentCount: 4, AppConfiguration: base.AppConfiguration{Name: "app1", Version: 1.0}}
	if !scaleApp(existsing, config) {
		t.Error(AppsState)
	}
	for _, status := range AppsState {
		if (status.Id != "app1_1" && status.Id != "app1_2"  && status.Id != "app2_1" ) && status.Status != base.STATUS_DEPLOYING {
			t.Error(AppsState)
		}
	}
	if AppsState.Get("app1_1").Status != base.STATUS_RUNNING {
		t.Error(AppsState)
	}
	if AppsState.Get("app1_2").Status != base.STATUS_RUNNING {
		t.Error(AppsState)
	}
	if AppsState.Get("app2_1").Status != base.STATUS_RUNNING {
		t.Error(AppsState)
	}

	if scaleApp(existsing, base.PushConfiguration{DeploymentCount: 4, AppConfiguration: base.AppConfiguration{Name: "app2", Version: 2, DockerConfig: base.DockerConfig{Tag: "1.1_runfail"}}}) {
		t.Error(AppsState)
	}
}

func TestScaleApp_down(t *testing.T) {
	Configuration = types.Configuration{}
	Init()

	existsing := []base.AppInfo{{Name: "app1", Version: 1.0, Status: base.STATUS_RUNNING, Id: "app1_1"}, {Name: "app1", Version: 1.0, Status: base.STATUS_RUNNING, Id: "app1_2"}}
	AppsState.Add("app1_1", base.AppInfo{Name: "app1", Version: 1.0, Status: base.STATUS_RUNNING, Id: "app1_1"})
	AppsState.Add("app1_2", base.AppInfo{Name: "app1", Version: 1.0, Status: base.STATUS_RUNNING, Id: "app1_2"})
	AppsState.Add("app2_1", base.AppInfo{Name: "app2", Version: 1.0, Status: base.STATUS_RUNNING, Id: "app2_1"})
	AppsState.Add("app2_2", base.AppInfo{Name: "app2", Version: 1.0, Status: base.STATUS_RUNNING, Id: "app2_2"})
	AppsConfiguration.Add("app2_1", base.AppConfiguration{Name:"app2", Version:1.0, DockerConfig: base.DockerConfig{Tag: "1.0_stopfail"}})
	AppsConfiguration.Add("app2_2", base.AppConfiguration{Name:"app2", Version:1.0, DockerConfig: base.DockerConfig{Tag: "1.0_stopfail"}})
	config := base.PushConfiguration{DeploymentCount: 0, AppConfiguration: base.AppConfiguration{Name: "app1", Version: 1.0}}
	if !scaleApp(existsing, config) {
		t.Error(AppsState)
	}
	for _, status := range AppsState {
		if ( status.Id != "app2_1" && status.Id != "app2_2" ) && status.Status != base.STATUS_DEAD {
			t.Error(AppsState)
		}
	}
	existsing2 := []base.AppInfo{{Name: "app2", Version: 1.0, Status: base.STATUS_RUNNING, Id: "app2_1"}, {Name: "app2", Version: 1.0, Status: base.STATUS_RUNNING, Id: "app2_2"}}
	if scaleApp(existsing2, base.PushConfiguration{DeploymentCount: 1, AppConfiguration: base.AppConfiguration{Name: "app2", Version: 1.0, DockerConfig: base.DockerConfig{Tag: "1.0_stopfail"}}}) {
		t.Error(AppsState)
	}
}

func TestUpdateApp(t *testing.T) {
	Configuration = types.Configuration{}
	Init()

	existsing := []base.AppInfo{{Name: "app1", Version: 1.0, Status: base.STATUS_RUNNING, Id: "app1_1"}, {Name: "app1", Version: 1.0, Status: base.STATUS_RUNNING, Id: "app1_2"}}
	AppsState.Add("app1_1", base.AppInfo{Name: "app1", Version: 1.0, Status: base.STATUS_RUNNING, Id: "app1_1"})
	AppsState.Add("app1_2", base.AppInfo{Name: "app1", Version: 1.0, Status: base.STATUS_RUNNING, Id: "app1_2"})
	AppsState.Add("app2_1", base.AppInfo{Name: "app2", Version: 1.0, Status: base.STATUS_RUNNING, Id: "app2_1"})
	config := base.PushConfiguration{DeploymentCount: 3, AppConfiguration: base.AppConfiguration{Name: "app1", Version: 2, DockerConfig: base.DockerConfig{Tag: "1.1"}}}

	if !updateApp(existsing, config) {
		t.Error(AppsState)
	}

	if len(AppsState) != 4 {
		t.Error(AppsState)
	}

	for _, status := range AppsState {
		if status.Id != "app2_1" && status.Status != base.STATUS_DEPLOYING {
			t.Error(AppsState)
		}
	}
}

func TestQueryApp(t *testing.T) {
	Configuration = types.Configuration{}
	Init()
	AppsConfiguration.Add("app1_1", base.AppConfiguration{Name: "app1", Version: 1.0, DockerConfig: base.DockerConfig{Tag: "1.0"}})
	AppsConfiguration.Add("app1_2", base.AppConfiguration{Name: "app1", Version: 1.0, DockerConfig: base.DockerConfig{Tag: "1.0_queryfail"}})
	AppsState.Add("app1_1", base.AppInfo{Name: "app1", Version: 1.0, Status: base.STATUS_DEPLOYING, Id: "app1_1"})
	if !QueryApp("app1_1") {
		t.Error(AppsState)
	}
	if AppsState.Get("app1_1").Status != base.STATUS_RUNNING {
		t.Error(AppsState)
	}
	AppsState.Add("app1_2", base.AppInfo{Name: "app1", Version: 1.0, Status: base.STATUS_DEPLOYING, Id: "app1_2"})
	if QueryApp("app1_2") {
		t.Error(AppsState)
	}
	if AppsState.Get("app1_2").Status != base.STATUS_DEAD {
		t.Error(AppsState)
	}
}

func TestNewApp_Ok(t *testing.T) {
	Configuration = types.Configuration{}
	Init()
	if len(AppsState) != 0 {
		t.Error(AppsState)
	}
	if len(AppsConfiguration) != 0 {
		t.Error(AppsConfiguration)
	}
	config := base.PushConfiguration{DeploymentCount: 4, AppConfiguration: base.AppConfiguration{Name: "app1", Version: 1.0}}

	res := newApp(config)
	if !res {
		t.Error(res)
	}
	if len(AppsState) != 4 {
		t.Error(AppsState)
	}
	if len(AppsConfiguration) != 4 {
		t.Error(AppsConfiguration)
	}
	for _, state := range AppsState {
		if state.Status != base.STATUS_DEPLOYING {
			t.Error(AppsState)
		}
	}
}

func TestNewApp_InstallFail(t *testing.T) {
	Configuration = types.Configuration{}
	Init()
	if len(AppsState) != 0 {
		t.Error(AppsState)
	}
	if len(AppsConfiguration) != 0 {
		t.Error(AppsConfiguration)
	}
	config := base.PushConfiguration{DeploymentCount: 4, AppConfiguration: base.AppConfiguration{Name: "app1", Version: 1.0, DockerConfig: base.DockerConfig{Tag: "1.0_installfail"}}}

	res := newApp(config)
	if res {
		t.Error(res)
	}
	if len(AppsState) != 0 {
		t.Error(AppsState)
	}
	if len(AppsConfiguration) != 0 {
		t.Error(AppsConfiguration)
	}
}

func TestNewApp_RunFail(t *testing.T) {
	Configuration = types.Configuration{}
	Init()
	if len(AppsState) != 0 {
		t.Error(AppsState)
	}
	if len(AppsConfiguration) != 0 {
		t.Error(AppsConfiguration)
	}
	config := base.PushConfiguration{DeploymentCount: 4, AppConfiguration: base.AppConfiguration{Name: "app1", Version: 1.0, DockerConfig: base.DockerConfig{Tag: "1.0_runfail"}}}

	res := newApp(config)

	if res {
		t.Error(res)
	}
	if len(AppsState) != 4 {
		t.Error(AppsState)
	}
	if len(AppsConfiguration) != 4 {
		t.Error(AppsConfiguration)
	}
	for _, state := range AppsState {
		if state.Status != base.STATUS_DEAD{
			t.Error(AppsState)
		}
	}
}

func TestRollbackApp(t *testing.T) {
	Configuration = types.Configuration{}
	Init()
	if len(AppsState) != 0 {
		t.Error(AppsState)
	}
	if len(AppsConfiguration) != 0 {
		t.Error(AppsConfiguration)
	}
	config := base.PushConfiguration{DeploymentCount: 4, AppConfiguration: base.AppConfiguration{Name: "app1", Version: 2.0, DockerConfig: base.DockerConfig{Tag: "2.0"}}}
	installAndRun(config)

	if len(AppsState) != 4 {
		t.Error(AppsState)
	}

	rollbackApp(base.AppConfiguration{Name: "app1", Version: 1.0}, 2)
	if len(AppsState) != 2 {
		t.Error(AppsState)
	}
	for _, state := range AppsState {
		if state.Version != 1.0 {
			t.Error(AppsState)
		}
	}
}


func prepareHandleTest(t *testing.T) {
	Configuration = types.Configuration{}
	Init()

	AppsConfiguration.Add("app_1_1", base.AppConfiguration{Name: "app1", Version: 1.0})

	AppsConfiguration.Add("app_queryfail_1", base.AppConfiguration{Name: "app_queryfail", Version: 1.0})
	AppsConfiguration.Add("app_runfail_1", base.AppConfiguration{Name: "app_runfail", Version: 1.0})
	AppsConfiguration.Add("app_installfail_1", base.AppConfiguration{Name: "app_installfail", Version: 1.0})
	AppsConfiguration.Add("app_stopfail_1", base.AppConfiguration{Name: "app_stopfail", Version: 1.0})
	AppsConfiguration.Add("app_deletefail_1", base.AppConfiguration{Name: "app_deletefail", Version: 1.0})

	AppsConfiguration.Add("app_2_1", base.AppConfiguration{Name: "app2", Version: 1.0})
	AppsConfiguration.Add("app_2_2", base.AppConfiguration{Name: "app2", Version: 1.0})


	AppsState.Add("app_1_1", base.AppInfo{Name:"app_1", Version:1.0, Status:base.STATUS_RUNNING, Id: "app_1_1"})

	AppsState.Add("app_queryfail_1", base.AppInfo{Name:"app_queryfail", Version:1.0, Status:base.STATUS_RUNNING, Id: "app_queryfail_1"})
	AppsState.Add("app_runfail_1", base.AppInfo{Name:"app_runfail", Version:1.0, Status:base.STATUS_RUNNING, Id: "app_runfail_1"})
	AppsState.Add("app_installfail_1", base.AppInfo{Name:"app_installfail", Version:1.0, Status:base.STATUS_RUNNING, Id: "app_installfail_1"})
	AppsState.Add("app_stopfail_1", base.AppInfo{Name:"app_stopfail", Version:1.0, Status:base.STATUS_RUNNING, Id: "app_stopfail_1"})
	AppsState.Add("app_deletefail_1", base.AppInfo{Name:"app_deletefail", Version:1.0, Status:base.STATUS_RUNNING, Id: "app_deletefail_1"})

	AppsState.Add("app_2_1", base.AppInfo{Name:"app_2", Version:1.0, Status:base.STATUS_RUNNING, Id: "app_2_1"})
	AppsState.Add("app_2_2", base.AppInfo{Name:"app_2", Version:1.0, Status:base.STATUS_RUNNING, Id: "app_2_2"})

	if AppsState.Get("app_2_1").Status != base.STATUS_RUNNING {
		t.Error(AppsState)
	}
	if AppsState.Get("app_installfail_1").Status != base.STATUS_RUNNING {
		t.Error(AppsState)
	}
	if AppsState.Get("app_stopfail_1").Status != base.STATUS_RUNNING {
		t.Error(AppsState)
	}
}

func TestHandle_updateApp_Ok(t *testing.T) {
        prepareHandleTest(t)

	config := base.PushConfiguration{DeploymentCount: 2, AppConfiguration: base.AppConfiguration{Name: "app_1", Version: 2.0}}
	Handle(config)

	res := AppsState.GetAll("app_1")
	if len(res) != 2 {
		t.Error(res)
	}
	if res[0].Status != base.STATUS_DEPLOYING {
		t.Error(res)
	}
}

func TestHandle_updateApp_QueryFail(t *testing.T) {
        prepareHandleTest(t)

	config := base.PushConfiguration{DeploymentCount: 2, AppConfiguration: base.AppConfiguration{Name: "app_queryfail", Version: 2.0, DockerConfig: base.DockerConfig{Tag: "2.0_queryfail"}}}
	Handle(config)

	res := AppsState.GetAll("app_queryfail")
	if len(res) != 2 {
		t.Error(res)
	}
	if res[0].Status != base.STATUS_DEPLOYING {
		t.Error(res)
	}
}


func TestHandle_updateApp_StopFail(t *testing.T) {
        prepareHandleTest(t)

	config := base.PushConfiguration{DeploymentCount: 2, AppConfiguration: base.AppConfiguration{Name: "app_stopfail", Version: 2.0, DockerConfig: base.DockerConfig{Tag: "2.0_stopfail"}}}
	Handle(config)

	res := AppsState.GetAll("app_stopfail")
	if len(res) != 2 {
		t.Error(res)
	}
	if res[0].Status != base.STATUS_DEPLOYING {
		t.Error(res)
	}
}

func TestHandle_updateApp_InstallFail_Rollback(t *testing.T) {
        prepareHandleTest(t)
	config := base.PushConfiguration{DeploymentCount: 3, AppConfiguration: base.AppConfiguration{Name: "app_installfail", Version: 2.0, DockerConfig: base.DockerConfig{Tag: "2.0_installfail"}}}
	Handle(config)

	res := AppsState.GetAll("app_installfail")
	if len(res) != 3 {
		t.Error(res)
	}
	if res[0].Status != base.STATUS_DEPLOYING || res[0].Version != 1.0 {
		t.Error(res)
	}
}

func TestHandle_updateApp_RunFail_Rollback(t *testing.T) {
        prepareHandleTest(t)
	config := base.PushConfiguration{DeploymentCount: 4, AppConfiguration: base.AppConfiguration{Name: "app_runfail", Version: 2.0, DockerConfig: base.DockerConfig{Tag: "2.0_runfail"}}}
	Handle(config)

	res := AppsState.GetAll("app_runfail")
	if len(res) != 4 {
		t.Error(res)
	}
	if res[0].Status != base.STATUS_DEPLOYING || res[0].Version != 1.0 {
		t.Error(res)
	}
}


func TestHandle_scaleUpDown(t *testing.T) {
	prepareHandleTest(t)

	config := base.PushConfiguration{DeploymentCount: 4, AppConfiguration: base.AppConfiguration{Name: "app_1", Version: 1.0, DockerConfig: base.DockerConfig{Tag: "1.0"}}}
	Handle(config)

	res := AppsState.GetAll("app_1")
	if len(res) != 4 {
		t.Error(res)
	}
	for _, elem := range res {
		if elem.Id != "app_1_1" && elem.Status != base.STATUS_DEPLOYING {
			t.Error(res)
		}
	}

	config1 := base.PushConfiguration{DeploymentCount: 1, AppConfiguration: base.AppConfiguration{Name: "app_1", Version: 1.0, DockerConfig: base.DockerConfig{Tag: "1.0"}}}
	Handle(config1)

	res1 := AppsState.GetAll("app_1")
	if len(res1) != 4 {
		t.Error(res1)
	}
	dead_count := 0
	for _, elem := range res1 {
		if elem.Status == base.STATUS_DEAD {
			dead_count++
		}
	}
	if dead_count != 3 {
		t.Error(res1)
	}
}


func TestHandle_NewApp(t *testing.T) {
	prepareHandleTest(t)

	config := base.PushConfiguration{DeploymentCount: 4, AppConfiguration: base.AppConfiguration{Name: "newapp", Version: 1.0, DockerConfig: base.DockerConfig{Tag: "1.0"}}}
	Handle(config)

	res := AppsState.GetAll("newapp")
	if len(res) != 4 {
		t.Error(res)
	}
}

func TestPollAppsState(t *testing.T) {
	prepareHandleTest(t)
	AppsConfiguration.Add("app_queryfail_1", base.AppConfiguration{Name: "app_queryfail", Version: 1.0, DockerConfig:base.DockerConfig{Tag:"1.0_queryfail"}})

	for _, app := range AppsState.All() {
		if app.Status != base.STATUS_RUNNING {
			t.Error(app)
		}
	}

	PollAppsState()

	if AppsState.Get("app_1_1").Status != base.STATUS_RUNNING {
		t.Error(AppsState)
	}

	if AppsState.Get("app_queryfail_1").Status != base.STATUS_DEAD {
		t.Error(AppsState)
	}
}

func TestAppMetrics(t *testing.T) {
	Configuration = types.Configuration{}
	Init()

	AppsConfiguration.Add("app_1", base.AppConfiguration{Name: "app1", Version: 1.0, DockerConfig: base.DockerConfig{Tag: "1.0"}})
	AppsConfiguration.Add("app_2", base.AppConfiguration{Name: "app1", Version: 1.0, DockerConfig: base.DockerConfig{Tag: "1.0_metrics=10_20_30_2"}})

	if AppMetrics("app_1") {
		t.Error()
	}
	if !AppMetrics("app_2") {
		t.Error()
	}
	for _, met := range AppsMetricsById["app_2"] {
		if met.CpuUsage != 10 || met.MemoryUsage != 20 || met.NetworkUsage != 30 || met.ResponsePerformance != 2 {
			t.Error(met)
		}
	}
}

func TestGenerateCombinedMetrics(t *testing.T) {
	Configuration = types.Configuration{}
	Init()
	AppsConfiguration.Add("app_1_1", base.AppConfiguration{Name: "app1", Version: 1.0, DockerConfig: base.DockerConfig{Tag: "1.0_metrics=10_20_30_2"}})
	AppsConfiguration.Add("app_1_2", base.AppConfiguration{Name: "app1", Version: 1.0, DockerConfig: base.DockerConfig{Tag: "1.0_metrics=10_20_30_2"}})
	AppsState.Add("app_1_1", base.AppInfo{Name:"app1", Version:1.0, Status:base.STATUS_RUNNING, Id: "app_1_1"})
	AppsState.Add("app_1_2", base.AppInfo{Name:"app1", Version:1.0, Status:base.STATUS_RUNNING, Id: "app_1_2"})

	if !AppMetrics("app_1_1") {
		t.Error()
	}
	time.Sleep(time.Duration(2 * time.Millisecond))
	if !AppMetrics("app_1_2") {
		t.Error()
	}
	res := generateCombinedMetrics()

	if len(res) != 2 || len(res["app1"][1]) != 2 {
		t.Error(res)
	}

	if len(AppsMetricsById) != 0 {
		t.Error(AppsMetricsById)
	}

}





