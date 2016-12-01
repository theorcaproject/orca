package main

import (
	"testing"
	"time"
	"gatoor/orca/rewriteTrainer/installer"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/rewriteTrainer/scheduler"
	"gatoor/orca/rewriteTrainer/tracker"
	"gatoor/orca/rewriteTrainer/planner"
)


func before(t *testing.T) {
	TRAINER_CONFIGURATION_FILE = "/orca/config/trainer/aws_test_trainer.json"
	APPS_CONFIGURATION_FILE = "/orca/config/trainer/aws_test_apps.json"
	AVAILABLE_INSTANCES_CONFIGURATION_FILE = "/orca/config/trainer/aws_test_available_instances.json"
	CLOUD_PROVIDER_CONFIGURATION_FILE = "/orca/config/trainer/aws_test_cloud_provider.json"
	go main()
	time.Sleep(500 * time.Millisecond)
}

func Test(t *testing.T) {
	before(t)
	//start with no instances
	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 0 {
		t.Error(state_cloud.GlobalCloudLayout.Current)
	}
	if len(state_cloud.GlobalAvailableInstances) != 0 {
		t.Error(state_cloud.GlobalAvailableInstances)
	}

	time.Sleep(120 * time.Second)

	//CloudProvider.MinInstanceCount should make us spawn 2 instances
	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 2 {
		t.Error(state_cloud.GlobalCloudLayout.Current)
	}
	if len(state_cloud.GlobalAvailableInstances) != 2 {
		t.Error(state_cloud.GlobalAvailableInstances)
	}

	scheduler.TriggerRun()

	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 2 {
		t.Error(state_cloud.GlobalCloudLayout.Current)
	}
	if len(state_cloud.GlobalCloudLayout.Desired.Layout) != 2 {
		t.Error(state_cloud.GlobalCloudLayout.Desired)
	}

	time.Sleep(120 * time.Second)

	//Nginx was installed on both instances, ubuntuWorker as well

	nginxHosts := state_cloud.GlobalCloudLayout.Current.FindHostsWithApp("nginx")
	if len(nginxHosts) != 2 {
		t.Error(state_cloud.GlobalCloudLayout.Current)
	}

	ubuntuHosts := state_cloud.GlobalCloudLayout.Current.FindHostsWithApp("ubuntuWorker")
	if len(ubuntuHosts) != 2 {
		t.Error(state_cloud.GlobalCloudLayout.Current)
	}

	for _, host := range state_cloud.GlobalCloudLayout.Current.Layout {
		if len(host.Apps) != 3 {     //one nginx and 2 ubuntu
			t.Error(host)
		}
	}
}
