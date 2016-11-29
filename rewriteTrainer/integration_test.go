package main

import (
	"testing"
	"time"
	"gatoor/orca/rewriteTrainer/installer"
	"gatoor/orca/rewriteTrainer/state/cloud"
)


func before(t *testing.T) {
	go main()
	time.Sleep(500 * time.Millisecond)
}

func TestInstall(t *testing.T) {
	before(t)
	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 0 {
		t.Error(state_cloud.GlobalCloudLayout.Current)
	}
	if len(state_cloud.GlobalAvailableInstances) != 0 {
		t.Error(state_cloud.GlobalAvailableInstances)
	}

	res := installer.InstallNewInstance(installer.TestClientConfig("orca_client_1"), "172.16.147.189")
	if !res {
		t.Error("install failed")
	}
	res = installer.InstallNewInstance(installer.TestClientConfig("orca_client_2"), "172.16.147.190")
	if !res {
		t.Error("install failed")
	}

	time.Sleep(15 * time.Second)

	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 2 {
		t.Error(state_cloud.GlobalCloudLayout.Current)
	}
	if len(state_cloud.GlobalAvailableInstances) != 2 {
		t.Error(state_cloud.GlobalAvailableInstances)
	}
}
