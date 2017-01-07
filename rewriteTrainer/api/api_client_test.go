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

package api


import (
	"testing"
	"net/http"
	"net/http/httptest"
	"gatoor/orca/base"
	"bytes"
	"encoding/json"
	"gatoor/orca/rewriteTrainer/state/configuration"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/rewriteTrainer/state/needs"
	"gatoor/orca/rewriteTrainer/cloud"
	"gatoor/orca/rewriteTrainer/db"
	"gatoor/orca/rewriteTrainer/needs"
	"gatoor/orca/rewriteTrainer/planner"
	"gatoor/orca/rewriteTrainer/scheduler"
)

func initTrainer() {
	state_configuration.GlobalConfigurationState.Init()
	state_cloud.GlobalCloudLayout.Init()
	state_needs.GlobalAppsNeedState = state_needs.AppsNeedState{}



	state_configuration.GlobalConfigurationState.CloudProvider.Type = "TEST"
	state_configuration.GlobalConfigurationState.CloudProvider.BaseInstanceType = "testtype"
	state_configuration.GlobalConfigurationState.CloudProvider.MinInstances = 2
	state_configuration.GlobalConfigurationState.CloudProvider.MaxInstances = 4

	cloud.Init()

	db.Init("")

	state_configuration.GlobalConfigurationState.Trainer.Ip = "0.0.0.0"
	state_configuration.GlobalConfigurationState.Trainer.Port = 5000
	state_configuration.GlobalConfigurationState.Trainer.Policies.TRY_TO_REMOVE_HOSTS = true

	state_configuration.GlobalConfigurationState.ConfigureApp(base.AppConfiguration{
		Name: "http1",
		Type: base.APP_HTTP,
		Version: 1,
		TargetDeploymentCount: 1,
		MinDeploymentCount: 1,
		DockerConfig: base.DockerConfig{},
		RawConfig: base.RawConfig{},
		LoadBalancer: "loadbalancer1",
		Network: "network1",
	})
	state_configuration.GlobalConfigurationState.ConfigureApp(base.AppConfiguration{
		Name: "http2",
		Type: base.APP_HTTP,
		Version: 1,
		TargetDeploymentCount: 2,
		MinDeploymentCount: 2,
		DockerConfig: base.DockerConfig{},
		RawConfig: base.RawConfig{},
		LoadBalancer: "loadbalancer2",
		Network: "network1",
	})
	state_configuration.GlobalConfigurationState.ConfigureApp(base.AppConfiguration{
		Name: "worker1",
		Type: base.APP_WORKER,
		Version: 1,
		TargetDeploymentCount: 1,
		MinDeploymentCount: 1,
		DockerConfig: base.DockerConfig{},
		RawConfig: base.RawConfig{},
		LoadBalancer: "",
		Network: "network2",
	})
	state_configuration.GlobalConfigurationState.ConfigureApp(base.AppConfiguration{
		Name: "worker2",
		Type: base.APP_WORKER,
		Version: 1,
		TargetDeploymentCount: 10,
		MinDeploymentCount: 5,
		DockerConfig: base.DockerConfig{},
		RawConfig: base.RawConfig{},
		LoadBalancer: "",
		Network: "network2",
	})

	state_needs.GlobalAppsNeedState.UpdateNeeds("http1", 1, needs.AppNeeds{CpuNeeds: 101, MemoryNeeds: 101, NetworkNeeds: 101})
	state_needs.GlobalAppsNeedState.UpdateNeeds("http2", 1, needs.AppNeeds{CpuNeeds: 102, MemoryNeeds: 102, NetworkNeeds: 102})
	state_needs.GlobalAppsNeedState.UpdateNeeds("worker1", 1, needs.AppNeeds{CpuNeeds: 330, MemoryNeeds: 320, NetworkNeeds: 310})
	state_needs.GlobalAppsNeedState.UpdateNeeds("worker2", 1, needs.AppNeeds{CpuNeeds: 210, MemoryNeeds: 210, NetworkNeeds: 210})
}

func clientPush(clientObj base.TrainerPushWrapper) *httptest.ResponseRecorder {
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(clientObj)
	req, _ := http.NewRequest("POST", "/push", b)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(pushStateHandler)

	handler.ServeHTTP(rr, req)
	return rr
}

func setup(t *testing.T) (base.TrainerPushWrapper, base.TrainerPushWrapper){
	initTrainer()

	clientObj := base.TrainerPushWrapper{
		HostInfo: base.HostInfo{
			HostId: "host1",
			IpAddr: "1.2.3.4",
			OsInfo: base.OsInfo{},
			Apps: []base.AppInfo{
			},
		},
		Stats: base.MetricsWrapper{},
	}
	response := clientPush(clientObj)
	var respObj = base.PushConfiguration{}
	json.Unmarshal(response.Body.Bytes(), &respObj)


	if respObj.AppConfiguration.Name != "" {
		t.Errorf("%+v", respObj)
	}


	clientObj2 := base.TrainerPushWrapper{
		HostInfo: base.HostInfo{
			HostId: "host2",
			IpAddr: "1.2.3.5",
			OsInfo: base.OsInfo{},
			Apps: []base.AppInfo{
			},
		},
		Stats: base.MetricsWrapper{},
	}
	response2 := clientPush(clientObj2)
	var respObj2 = base.PushConfiguration{}
	json.Unmarshal(response2.Body.Bytes(), &respObj2)

	if respObj2.AppConfiguration.Name != "" {
		t.Errorf("%+v", respObj2)
	}

	planner.InitialPlan()
	scheduler.TriggerRun()
	return clientObj, clientObj2
}

func getToStableState(clientObj base.TrainerPushWrapper, clientObj2 base.TrainerPushWrapper, t *testing.T) {
	clientObj.HostInfo.Apps = []base.AppInfo{{Type: base.APP_HTTP, Name: "http1", Version: 1, Status: base.STATUS_RUNNING, Id: "http1_1"},{Type: base.APP_HTTP, Name: "http2", Version: 1, Status: base.STATUS_RUNNING, Id: "http2_1"},{Type: base.APP_HTTP, Name: "worker1", Version: 1, Status: base.STATUS_RUNNING, Id: "worker1_1"},{Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_1"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_2"}}
	//this seems strange but is needed because one of the basic assumptions is that only one update is applied at any time.
	clientPush(clientObj)
	clientPush(clientObj)
	clientPush(clientObj)
	clientPush(clientObj)
	clientPush(clientObj)
	clientObj2.HostInfo.Apps = []base.AppInfo{{Type: base.APP_HTTP, Name: "http2", Version: 1, Status: base.STATUS_RUNNING, Id: "http2_11"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_11"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_22"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_33"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_44"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_55"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_66"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_77"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_88"}}
	clientPush(clientObj2)
	clientPush(clientObj2)
	clientPush(clientObj2)

	q1, _ := planner.Queue.Get("host1")
	q2, _ := planner.Queue.Get("host2")
	if len(q1) != 0 || len(q2) != 0 {
		t.Error(state_cloud.GlobalCloudLayout.Current)
		t.Error(planner.Queue.Queue)
	}
}


//2 hosts check in and 4 apps are deployed to them. Everything works
func Test_AllOk(t *testing.T) {
	clientObj, _ := setup(t)

	//get config for first app
	response3 := clientPush(clientObj)
	var respObj3 = base.PushConfiguration{}
	err := json.Unmarshal(response3.Body.Bytes(), &respObj3)
	if err != nil {
		t.Error(err)
	}
	if respObj3.DeploymentCount != 1 || respObj3.AppConfiguration.Name != "http1" {
		t.Errorf("response: %+v", respObj3)
	}

	//still deploying, get same config
	clientObj.HostInfo.Apps = []base.AppInfo{{Type: base.APP_HTTP, Name: "http1", Version: 1, Status: base.STATUS_DEPLOYING, Id: "http1_1"},}

	response4 := clientPush(clientObj)
	var respObj4 = base.PushConfiguration{}
	json.Unmarshal(response4.Body.Bytes(), &respObj4)

	if respObj4.DeploymentCount != 1 || respObj4.AppConfiguration.Name != "http1" {
		t.Errorf("queue: %+v", planner.Queue.Queue)
		t.Errorf("response before: %+v", respObj3)
		t.Errorf("response: %+v", respObj4)
	}


	//successful deployment, get next app
	clientObj.HostInfo.Apps = []base.AppInfo{{Type: base.APP_HTTP, Name: "http1", Version: 1, Status: base.STATUS_RUNNING, Id: "http1_1"},}

	response5 := clientPush(clientObj)
	var respObj5 = base.PushConfiguration{}
	json.Unmarshal(response5.Body.Bytes(), &respObj5)

	if respObj5.DeploymentCount != 1 || respObj5.AppConfiguration.Name != "http2" {
		t.Errorf("queue: %+v", planner.Queue.Queue)
		t.Errorf("response: %+v", respObj5)
	}

	//successful deployment, get next app
	clientObj.HostInfo.Apps = []base.AppInfo{{Type: base.APP_HTTP, Name: "http1", Version: 1, Status: base.STATUS_RUNNING, Id: "http1_1"},{Type: base.APP_HTTP, Name: "http2", Version: 1, Status: base.STATUS_RUNNING, Id: "http2_1"},}

	response6 := clientPush(clientObj)
	var respObj6 = base.PushConfiguration{}
	json.Unmarshal(response6.Body.Bytes(), &respObj6)

	if respObj6.DeploymentCount != 1 || respObj6.AppConfiguration.Name != "worker1" {
		t.Errorf("queue: %+v", planner.Queue.Queue)
		t.Errorf("response: %+v", respObj6)
	}

	//successful deployment, get next app
	clientObj.HostInfo.Apps = []base.AppInfo{{Type: base.APP_HTTP, Name: "http1", Version: 1, Status: base.STATUS_RUNNING, Id: "http1_1"},{Type: base.APP_HTTP, Name: "http2", Version: 1, Status: base.STATUS_RUNNING, Id: "http2_1"},{Type: base.APP_WORKER, Name: "worker1", Version: 1, Status: base.STATUS_RUNNING, Id: "worker1_1"},}

	response7 := clientPush(clientObj)
	var respObj7 = base.PushConfiguration{}
	json.Unmarshal(response7.Body.Bytes(), &respObj7)

	if respObj7.DeploymentCount != 2 || respObj7.AppConfiguration.Name != "worker2" {
		t.Errorf("queue: %+v", planner.Queue.Queue)
		t.Errorf("response: %+v", respObj7)
	}

	//successful deployment, not enough deployed
	clientObj.HostInfo.Apps = []base.AppInfo{{Type: base.APP_HTTP, Name: "http1", Version: 1, Status: base.STATUS_RUNNING, Id: "http1_1"},{Type: base.APP_HTTP, Name: "http2", Version: 1, Status: base.STATUS_RUNNING, Id: "http2_1"},{Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_1"}}

	response8 := clientPush(clientObj)
	var respObj8 = base.PushConfiguration{}
	json.Unmarshal(response8.Body.Bytes(), &respObj8)

	if respObj8.DeploymentCount != 2 || respObj8.AppConfiguration.Name != "worker2" {
		t.Errorf("queue: %+v", planner.Queue.Queue)
		t.Errorf("response: %+v", respObj8)
	}

	//successful deployment, enough deployed no next app
	clientObj.HostInfo.Apps = []base.AppInfo{{Type: base.APP_HTTP, Name: "http1", Version: 1, Status: base.STATUS_RUNNING, Id: "http1_1"},{Type: base.APP_HTTP, Name: "http2", Version: 1, Status: base.STATUS_RUNNING, Id: "http2_1"},{Type: base.APP_WORKER, Name: "worker1", Version: 1, Status: base.STATUS_RUNNING, Id: "worker1_1"},{Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_1"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_2"}}

	response9 := clientPush(clientObj)
	var respObj9 = base.PushConfiguration{}
	json.Unmarshal(response9.Body.Bytes(), &respObj9)

	if respObj9.DeploymentCount != 0 || respObj9.AppConfiguration.Name != "" {
		t.Errorf("queue: %+v", planner.Queue.Queue)
		t.Errorf("response: %+v", respObj9)
	}
}

func Test_UpdateScaleApp(t *testing.T) {
	clientObj, clientObj2 := setup(t)
	getToStableState(clientObj, clientObj2, t)

	//http1 update and scale up
	state_configuration.GlobalConfigurationState.ConfigureApp(base.AppConfiguration{
		Name: "http1",
		Type: base.APP_HTTP,
		Version: 2,
		TargetDeploymentCount: 2,
		MinDeploymentCount: 1,
		DockerConfig: base.DockerConfig{},
		RawConfig: base.RawConfig{},
		LoadBalancer: "loadbalancer1",
		Network: "network1",
	})

	scheduler.TriggerRun()

	clientObj.HostInfo.Apps = []base.AppInfo{{Type: base.APP_HTTP, Name: "http1", Version: 1, Status: base.STATUS_RUNNING, Id: "http1_1"},{Type: base.APP_HTTP, Name: "http2", Version: 1, Status: base.STATUS_RUNNING, Id: "http2_1"},{Type: base.APP_HTTP, Name: "worker1", Version: 1, Status: base.STATUS_RUNNING, Id: "worker1_1"},{Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_1"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_2"}}

	response := clientPush(clientObj)
	var responseObj = base.PushConfiguration{}
	json.Unmarshal(response.Body.Bytes(), &responseObj)

	if responseObj.DeploymentCount != 1 || responseObj.AppConfiguration.Version != 2 || responseObj.AppConfiguration.Name != "http1" {
		t.Errorf("%+v", state_cloud.GlobalCloudLayout.Desired)
		t.Error(responseObj)
	}
	//http1 was updated successfully
	clientObj.HostInfo.Apps = []base.AppInfo{{Type: base.APP_HTTP, Name: "http1", Version: 2, Status: base.STATUS_RUNNING, Id: "http1_1"},{Type: base.APP_HTTP, Name: "http2", Version: 1, Status: base.STATUS_RUNNING, Id: "http2_1"},{Type: base.APP_HTTP, Name: "worker1", Version: 1, Status: base.STATUS_RUNNING, Id: "worker1_1"},{Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_1"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_2"}}

	responseSuc := clientPush(clientObj)
	var responseObjSuc = base.PushConfiguration{}
	json.Unmarshal(responseSuc.Body.Bytes(), &responseObjSuc)

	if responseObjSuc.DeploymentCount != 0 || responseObjSuc.AppConfiguration.Version != 0 || responseObjSuc.AppConfiguration.Name != "" {
		t.Errorf("%+v", state_cloud.GlobalCloudLayout.Desired)
		t.Error(responseObjSuc)
	}

	//http1 is also assigned to host2 => scale up works
	clientObj2.HostInfo.Apps = []base.AppInfo{{Type: base.APP_HTTP, Name: "http2", Version: 1, Status: base.STATUS_RUNNING, Id: "http2_11"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_11"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_22"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_33"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_44"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_55"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_66"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_77"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_88"}}
	response2 := clientPush(clientObj2)
	var responseObj2 = base.PushConfiguration{}
	json.Unmarshal(response2.Body.Bytes(), &responseObj2)

	if responseObj2.DeploymentCount != 1 || responseObj2.AppConfiguration.Version != 2 || responseObj2.AppConfiguration.Name != "http1" {
		t.Errorf("%+v", state_cloud.GlobalCloudLayout.Desired)
		t.Error(responseObj2)
	}

	//scale up success on host 2
	clientObj2.HostInfo.Apps = []base.AppInfo{{Type: base.APP_HTTP, Name: "http2", Version: 1, Status: base.STATUS_RUNNING, Id: "http2_11"},{Type: base.APP_HTTP, Name: "http1", Version: 2, Status: base.STATUS_RUNNING, Id: "http1_11"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_11"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_22"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_33"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_44"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_55"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_66"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_77"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_88"}}
	response3 := clientPush(clientObj2)
	var responseObj3 = base.PushConfiguration{}
	json.Unmarshal(response3.Body.Bytes(), &responseObj3)

	if responseObj3.DeploymentCount != 0 || responseObj3.AppConfiguration.Version != 0 || responseObj3.AppConfiguration.Name != "" {
		t.Errorf("%+v", state_cloud.GlobalCloudLayout.Desired)
		t.Error(responseObj3)
	}


	//worker1 update fails

	state_configuration.GlobalConfigurationState.ConfigureApp(base.AppConfiguration{
		Name: "worker1",
		Type: base.APP_WORKER,
		Version: 2,
		TargetDeploymentCount: 1,
		MinDeploymentCount: 1,
		DockerConfig: base.DockerConfig{},
		RawConfig: base.RawConfig{},
		LoadBalancer: "",
		Network: "network2",
	})

	scheduler.TriggerRun()

	response4 := clientPush(clientObj)
	var responseObj4 = base.PushConfiguration{}
	json.Unmarshal(response4.Body.Bytes(), &responseObj4)

	if responseObj4.DeploymentCount != 1 || responseObj4.AppConfiguration.Version != 2 || responseObj4.AppConfiguration.Name != "worker1" {
		t.Errorf("%+v", state_cloud.GlobalCloudLayout.Desired)
		t.Error(responseObj4)
	}

	clientObj.HostInfo.Apps = []base.AppInfo{{Type: base.APP_HTTP, Name: "http1", Version: 1, Status: base.STATUS_RUNNING, Id: "http1_1"},{Type: base.APP_HTTP, Name: "http2", Version: 1, Status: base.STATUS_RUNNING, Id: "http2_1"},{Type: base.APP_HTTP, Name: "worker1", Version: 2, Status: base.STATUS_DEAD, Id: "worker1_1"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_1"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_2"}}

	response5 := clientPush(clientObj)
	var responseObj5 = base.PushConfiguration{}
	json.Unmarshal(response5.Body.Bytes(), &responseObj5)

	if responseObj5.DeploymentCount != 1 || responseObj5.AppConfiguration.Version != 1 || responseObj5.AppConfiguration.Name != "worker1" {
		t.Errorf("%+v", state_cloud.GlobalCloudLayout.Desired)
		t.Error(responseObj5)
	}
	//rollback
	clientObj.HostInfo.Apps = []base.AppInfo{{Type: base.APP_HTTP, Name: "http1", Version: 2, Status: base.STATUS_RUNNING, Id: "http1_1"},{Type: base.APP_HTTP, Name: "http2", Version: 1, Status: base.STATUS_RUNNING, Id: "http2_1"},{Type: base.APP_HTTP, Name: "worker1", Version: 2, Status: base.STATUS_RUNNING, Id: "worker1_1"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_1"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_2"}}

	response6 := clientPush(clientObj)
	var responseObj6 = base.PushConfiguration{}
	json.Unmarshal(response6.Body.Bytes(), &responseObj6)

	if responseObj6.DeploymentCount != 0 || responseObj6.AppConfiguration.Version != 0 || responseObj6.AppConfiguration.Name != "" {
		t.Errorf("%+v", state_cloud.GlobalCloudLayout.Desired)
		t.Error(responseObj6)
	}

	//scale down worker 2

	state_configuration.GlobalConfigurationState.ConfigureApp(base.AppConfiguration{
		Name: "worker2",
		Type: base.APP_WORKER,
		Version: 1,
		TargetDeploymentCount: 6,
		MinDeploymentCount: 5,
		DockerConfig: base.DockerConfig{},
		RawConfig: base.RawConfig{},
		LoadBalancer: "",
		Network: "network2",
	})

	scheduler.TriggerRun()

	clientObj2.HostInfo.Apps = []base.AppInfo{{Type: base.APP_HTTP, Name: "http2", Version: 1, Status: base.STATUS_RUNNING, Id: "http2_11"},{Type: base.APP_HTTP, Name: "http1", Version: 2, Status: base.STATUS_RUNNING, Id: "http1_11"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_11"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_22"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_33"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_44"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_55"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_66"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_77"}, {Type: base.APP_WORKER, Name: "worker2", Version: 1, Status: base.STATUS_RUNNING, Id: "worker2_88"}}
	response7 := clientPush(clientObj2)
	var responseObj7 = base.PushConfiguration{}
	json.Unmarshal(response7.Body.Bytes(), &responseObj7)

	if responseObj7.DeploymentCount != 4 || responseObj7.AppConfiguration.Version != 1 || responseObj7.AppConfiguration.Name != "worker2" {
		t.Errorf("%+v", state_cloud.GlobalCloudLayout.Desired)
		t.Error(responseObj7)
	}
}


func Test_HostDies(t *testing.T) {

}