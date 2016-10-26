package state_cloud_test

import (
	"gatoor/orca/rewriteTrainer/base"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"testing"
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