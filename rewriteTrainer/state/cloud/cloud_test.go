package state_cloud_test

import (
	"gatoor/orca/rewriteTrainer/base"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"testing"
	"gatoor/orca/rewriteTrainer/metrics"
)

func prepareLayoutState() state_cloud.CloudLayoutAll {
	var GlobalCloudLayout state_cloud.CloudLayoutAll
	GlobalCloudLayout.Init()
	return GlobalCloudLayout
}

func TestLayoutAddHost(t *testing.T) {
	state := prepareLayoutState()
	state.Current.AddHost("hostId1", state_cloud.CloudLayoutElement{
		HostId: "hostId1",
		IpAddress: "0.0.0.0",
		HabitatVersion: "0.0",
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})

	if _, err := state.Current.GetHost("unknown"); err == nil {
		t.Error()
	}
	if val, err := state.Current.GetHost("hostId1"); err == nil {
		if val.HostId != "hostId1" {
			t.Error()
		}
	} else {
		t.Error(err)
	}
}

func TestLayoutRemoveHost(t *testing.T) {
	state := prepareLayoutState()
	state.Current.AddHost("hostId1", state_cloud.CloudLayoutElement{
		HostId: "hostId1",
		IpAddress: "0.0.0.0",
		HabitatVersion: "0.0",
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state.Current.AddHost("hostId2", state_cloud.CloudLayoutElement{
		HostId: "hostId2",
		IpAddress: "0.0.0.0",
		HabitatVersion: "0.0",
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	if val, err := state.Current.GetHost("hostId1"); err == nil {
		if val.HostId != "hostId1" {
			t.Error()
		}
	} else {
		t.Error(err)
	}
	state.Current.RemoveHost("hostId1")

	if _, err := state.Current.GetHost("hostId1"); err == nil {
		t.Error()
	}

	if val, err := state.Current.GetHost("hostId2"); err == nil {
		if val.HostId != "hostId2" {
			t.Error()
		}
	} else {
		t.Error(err)
	}
}

func TestLayoutAddApp(t *testing.T) {
	state := prepareLayoutState()
	state.Current.AddHost("hostId1", state_cloud.CloudLayoutElement{
		HostId: "hostId1",
		IpAddress: "0.0.0.0",
		HabitatVersion: "0.0",
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state.Current.AddApp("hostId1", "someapp", "0.2", 1)

	if val, err := state.Current.GetHost("hostId1"); err == nil {
		if val.HostId != "hostId1" {
			t.Error()
		}
		if val.Apps["someapp"].Version != "0.2" {
			t.Error()
		}
	} else {
		t.Error(err)
	}
}

func TestLayoutRemoveApp(t *testing.T) {
	state := prepareLayoutState()
	state.Current.AddHost("hostId1", state_cloud.CloudLayoutElement{
		HostId: "hostId1",
		IpAddress: "0.0.0.0",
		HabitatVersion: "0.0",
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state.Current.AddApp("hostId1", "someapp", "0.2", 1)
	state.Current.AddApp("hostId1", "someapp2", "0.4", 5)

	if val, err := state.Current.GetHost("hostId1"); err == nil {
		if val.HostId != "hostId1" {
			t.Error()
		}
		if val.Apps["someapp"].Version!= "0.2" {
			t.Error()
		}
	} else {
		t.Error(err)
	}

	state.Current.RemoveApp("hostId1", "someapp")

	if val, err := state.Current.GetHost("hostId1"); err == nil {
		if val.HostId != "hostId1" {
			t.Error()
		}
		if val.Apps["someapp"].Version == "0.2" {
			t.Error()
		}
	} else {
		t.Error(err)
	}

	if val, err := state.Current.GetHost("hostId1"); err == nil {
		if val.HostId != "hostId1" {
			t.Error()
		}
		if val.Apps["someapp2"].Version != "0.4" {
			t.Error()
		}
	} else {
		t.Error(err)
	}
}

func TestLayoutIllegalAccess(t *testing.T) {
	state := prepareLayoutState()
	state.Current.RemoveHost("somehost")
	state.Current.RemoveApp("somehost", "someapp")
}

func TestLayout_FindHostsWithApp(t *testing.T) {
	state := prepareLayoutState()
	state.Current.AddHost("hostId1", state_cloud.CloudLayoutElement{
		HostId: "hostId1",
		IpAddress: "0.0.0.0",
		HabitatVersion: "0.0",
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state.Current.AddHost("hostId2", state_cloud.CloudLayoutElement{
		HostId: "hostId2",
		IpAddress: "0.0.0.0",
		HabitatVersion: "0.0",
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state.Current.AddHost("hostId3", state_cloud.CloudLayoutElement{
		HostId: "hostId3",
		IpAddress: "0.0.0.0",
		HabitatVersion: "0.0",
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state.Current.AddApp("hostId1", "app1", "10.0", 10)
	state.Current.AddApp("hostId1", "app2", "10.0", 10)
	state.Current.AddApp("hostId2", "app1", "10.0", 10)
	state.Current.AddApp("hostId2", "app2", "10.0", 10)
	state.Current.AddApp("hostId3", "app1", "10.0", 10)

	res := state.Current.FindHostsWithApp("app1")
	if !res["hostId1"] || !res["hostId2"] || !res["hostId3"] {
		t.Error(res)
	}
	res1 := state.Current.FindHostsWithApp("unnkown")
	if res1["hostId2"] {
		t.Error(res)
	}
	res2 := state.Current.FindHostsWithApp("app2")
	if  !res2["hostId1"] || !res2["hostId2"] || res2["hostId3"]{
		t.Error(res2)
	}
}

func TestLAvailableInstances(t *testing.T) {
	instances := state_cloud.AvailableInstances{}

	if len(instances) != 0 {
		t.Error("should be empty")
	}

	instances.Update("host1", state_cloud.InstanceResources{
		TotalCpuResource: 10.0,
		TotalMemoryResource: 20.0,
		TotalNetworkResource: 30.0,
	})

	instances.Update("host2", state_cloud.InstanceResources{
		TotalCpuResource: 100.0,
		TotalMemoryResource: 200.0,
		TotalNetworkResource: 300.0,
	})

	if len(instances) != 2 {
		t.Error("len should be 2")
	}

	_, err := instances.GetResources("unknown")

	if err == nil {
		t.Error("should err")
	}

	elem, _ := instances.GetResources("host1")
	if elem.TotalMemoryResource != 20.0 {
		t.Error("wrong memory resource")
	}
	elem2, _ := instances.GetResources("host2")
	if elem2.TotalCpuResource != 100.0 {
		t.Error("wrong cpu resource")
	}

	instances.Update("host1", state_cloud.InstanceResources{
		TotalCpuResource: 15.0,
		TotalMemoryResource: 25.0,
		TotalNetworkResource: 35.0,

	})

	if len(instances) != 2 {
		t.Error("len should be 2")
	}
	elem3, _ := instances.GetResources("host1")
	if elem3.TotalMemoryResource != 25.0 {
		t.Error("wrong memory resource")
	}

	instances.Remove("host4")
	if len(instances) != 2 {
		t.Error("len should be 2")
	}

	instances.Remove("host2")
	if len(instances) != 1 {
		t.Error("len should be 1")
	}
}

func TestUpdateHost(t *testing.T) {
	state := prepareLayoutState()
	apps :=  make(map[base.AppName]state_cloud.AppsVersion)
	apps["app1"] = state_cloud.AppsVersion{
		"0.1", 5,
	}
	apps["app2"] = state_cloud.AppsVersion{
		"0.2", 20,
	}

	state.Current.AddHost("hostId1", state_cloud.CloudLayoutElement{
		HostId: "hostId1",
		IpAddress: "0.0.0.0",
		HabitatVersion: "0.0",
		Apps: apps,
	})

	if _, err := state.Current.GetHost("unknown"); err == nil {
		t.Error()
	}
	if val, err := state.Current.GetHost("hostId1"); err == nil {
		if val.HostId != "hostId1" {
			t.Error()
		}
		if len(val.Apps) != 2 {
			t.Error(val.Apps)
		}
	} else {
		t.Error(err)
	}


	appInfo1 := metrics.AppInfo{
		Type: base.APP_HTTP,
		Name: "app1",
		Version: "1.0",
		Status: base.STATUS_DEPLOYING,
	}

	appInfo2 := metrics.AppInfo{
		Type: base.APP_HTTP,
		Name: "app1",
		Version: "1.0",
		Status: base.STATUS_RUNNING,
	}
	appInfo3 := metrics.AppInfo{
		Type: base.APP_HTTP,
		Name: "app2",
		Version: "2.0",
		Status: base.STATUS_RUNNING,
	}
	appInfo4 := metrics.AppInfo{
		Type: base.APP_HTTP,
		Name: "app2",
		Version: "2.0",
		Status: base.STATUS_RUNNING,
	}

	hostInfo := metrics.HostInfo{
		HostId: "hostId1",
		IpAddr: "1.1.1.1",
		OsInfo: metrics.OsInfo{},
		HabitatInfo: metrics.HabitatInfo{},
		Apps: []metrics.AppInfo{appInfo1, appInfo2, appInfo3, appInfo4},
	}

	state.Current.UpdateHost(hostInfo)

	if _, err := state.Current.GetHost("unknown"); err == nil {
		t.Error()
	}
	if val2, err := state.Current.GetHost("hostId1"); err == nil {
		if val2.HostId != "hostId1" {
			t.Error()
		}
		if val2.IpAddress != "1.1.1.1" {
			t.Error(val2.IpAddress)
		}
		if len(val2.Apps) != 2 {
			t.Error(val2.Apps)
		}
		if val2.Apps["app2"].Version != "2.0" {
			t.Error(val2.Apps["app2"].Version)
		}
		if val2.Apps["app2"].DeploymentCount != 2 {
			t.Error(val2.Apps["app2"].DeploymentCount)
		}
	} else {
		t.Error(err)
	}
}


