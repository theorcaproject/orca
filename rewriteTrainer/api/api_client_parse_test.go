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
)

func initTrainer() {
	state_configuration.GlobalConfigurationState.Init()
	state_cloud.GlobalCloudLayout.Init()
	state_needs.GlobalAppsNeedState = state_needs.AppsNeedState{}



	cloud.CurrentProviderConfig.Type = "TEST"
	cloud.CurrentProviderConfig.BaseInstanceType = "testtype"
	cloud.CurrentProviderConfig.MinInstances = 2
	cloud.CurrentProviderConfig.MaxInstances = 4

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
	handler := http.HandlerFunc(pushHandler)

	handler.ServeHTTP(rr, req)
	return rr
}

func Test_NewHosts(t *testing.T) {
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


	expected := `{"alive": true}`
	if response.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			response.Body.String(), expected)
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


	expected2 := `{"alive": true}`
	if response2.Body.String() != expected2 {
		t.Errorf("handler returned unexpected body: got %v want %v",
			response2.Body.String(), expected)
	}
}