package scheduler

import (
	"testing"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/rewriteTrainer/state/configuration"
	"gatoor/orca/rewriteTrainer/planner"
	"gatoor/orca/rewriteTrainer/base"
)


func initAll() {
	state_cloud.GlobalCloudLayout.Init()
	state_configuration.GlobalConfigurationState.Init()
}



func TestScheduler_run_NoChanges(t *testing.T) {
	initAll()

	state_cloud.GlobalCloudLayout.Current.AddHost("hostId1", state_cloud.CloudLayoutElement{
		HostId: "hostId1",
		IpAddress: "0.0.0.0",
		HabitatVersion: "0.0",
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state_cloud.GlobalCloudLayout.Current.AddApp("hostId1", "someapp", "0.2", 1)
	state_cloud.GlobalCloudLayout.Current.AddApp("hostId1", "someapp2", "0.4", 5)

	state_cloud.GlobalCloudLayout.Desired.AddHost("hostId1", state_cloud.CloudLayoutElement{
		HostId: "hostId1",
		IpAddress: "0.0.0.0",
		HabitatVersion: "0.0",
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state_cloud.GlobalCloudLayout.Desired.AddApp("hostId1", "someapp", "0.2", 1)
	state_cloud.GlobalCloudLayout.Desired.AddApp("hostId1", "someapp2", "0.4", 5)

	run()

	if !planner.Queue.Empty("hostId1") {
		t.Error("queue should be empty")
	}
	if !planner.Queue.Empty("unkown") {
		t.Error("queue should be empty")
	}
}

func TestScheduler_run_UpdateScale_SingleHost(t *testing.T) {
	initAll()

	state_cloud.GlobalCloudLayout.Current.AddHost("hostId1", state_cloud.CloudLayoutElement{
		HostId: "hostId1",
		IpAddress: "0.0.0.0",
		HabitatVersion: "0.0",
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state_cloud.GlobalCloudLayout.Current.AddApp("hostId1", "someapp", "0.2", 1)
	state_cloud.GlobalCloudLayout.Current.AddApp("hostId1", "someapp2", "0.4", 5)
	state_cloud.GlobalCloudLayout.Current.AddApp("hostId1", "someapp3", "0.6", 5)

	state_cloud.GlobalCloudLayout.Desired.AddHost("hostId1", state_cloud.CloudLayoutElement{
		HostId: "hostId1",
		IpAddress: "0.0.0.0",
		HabitatVersion: "0.0",
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state_cloud.GlobalCloudLayout.Desired.AddApp("hostId1", "someapp", "0.2", 4)
	state_cloud.GlobalCloudLayout.Desired.AddApp("hostId1", "someapp2", "0.3", 5)
	state_cloud.GlobalCloudLayout.Desired.AddApp("hostId1", "someapp3", "0.7", 5)

	run()

	if planner.Queue.Empty("hostId1") {
		t.Error("queue is empty")
	}
	if !planner.Queue.Empty("unkown") {
		t.Error("queue should be empty")
	}

	elem, _ := planner.Queue.Get("hostId1")

	if elem["someapp2"].Version.Version != "0.3" {
		t.Error("wrong version")
	}
	if elem["someapp3"].Version.Version != "0.7" {
		t.Error("wrong version")
	}
	if elem["someapp"].Version.DeploymentCount != 4 {
		t.Error("wrong deployment count")
	}
}

func TestScheduler_run_UpdateScale_MultipleHost(t *testing.T) {
	initAll()

	state_cloud.GlobalCloudLayout.Current.AddHost("hostId1", state_cloud.CloudLayoutElement{
		HostId: "hostId1",
		IpAddress: "0.0.0.0",
		HabitatVersion: "0.0",
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state_cloud.GlobalCloudLayout.Current.AddApp("hostId1", "someapp", "0.2", 1)
	state_cloud.GlobalCloudLayout.Current.AddApp("hostId1", "someapp2", "0.4", 5)
	state_cloud.GlobalCloudLayout.Current.AddApp("hostId1", "someapp3", "0.6", 5)

	state_cloud.GlobalCloudLayout.Desired.AddHost("hostId1", state_cloud.CloudLayoutElement{
		HostId: "hostId1",
		IpAddress: "0.0.0.0",
		HabitatVersion: "0.0",
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state_cloud.GlobalCloudLayout.Desired.AddApp("hostId1", "someapp", "0.2", 4)
	state_cloud.GlobalCloudLayout.Desired.AddApp("hostId1", "someapp2", "0.3", 5)
	state_cloud.GlobalCloudLayout.Desired.AddApp("hostId1", "someapp3", "0.7", 5)


	state_cloud.GlobalCloudLayout.Current.AddHost("hostId2", state_cloud.CloudLayoutElement{
		HostId: "hostId2",
		IpAddress: "0.0.0.0",
		HabitatVersion: "0.0",
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state_cloud.GlobalCloudLayout.Current.AddApp("hostId2", "someappR", "0.2R", 10)
	state_cloud.GlobalCloudLayout.Current.AddApp("hostId2", "someappR2", "0.4R", 50)

	state_cloud.GlobalCloudLayout.Desired.AddHost("hostId2", state_cloud.CloudLayoutElement{
		HostId: "hostId2",
		IpAddress: "0.0.0.0",
		HabitatVersion: "0.0",
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state_cloud.GlobalCloudLayout.Desired.AddApp("hostId2", "someappR", "0.2R", 40)
	state_cloud.GlobalCloudLayout.Desired.AddApp("hostId2", "someappR3", "0.7R", 50)

	run()

	if planner.Queue.Empty("hostId1") {
		t.Error("queue is empty")
	}
	if !planner.Queue.Empty("unkown") {
		t.Error("queue should be empty")
	}

	elem, _ := planner.Queue.Get("hostId1")

	if elem["someapp2"].Version.Version != "0.3" {
		t.Error("wrong version")
	}
	if elem["someapp3"].Version.Version != "0.7" {
		t.Error("wrong version")
	}
	if elem["someapp"].Version.DeploymentCount != 4 {
		t.Error("wrong deployment count")
	}

	elem2, _ := planner.Queue.Get("hostId2")

	if elem2["someappR2"].Version.Version != "0.4R" {
		t.Error("wrong version")
	}
	if elem2["someappR2"].Version.DeploymentCount != 0 {
		t.Error("ap remove not working")
	}
	if elem2["someappR3"].Version.Version != "0.7R" {
		t.Error("wrong version")
	}
	if elem2["someappR"].Version.DeploymentCount != 40 {
		t.Error("wrong deployment count")
	}


}