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

package main

//import (
//	"testing"
//	"time"
//	"gatoor/orca/rewriteTrainer/installer"
//	"gatoor/orca/rewriteTrainer/state/cloud"
//	"gatoor/orca/rewriteTrainer/scheduler"
//	"gatoor/orca/rewriteTrainer/tracker"
//	"gatoor/orca/rewriteTrainer/planner"
//)


//func before(t *testing.T) {
//	go main()
//	time.Sleep(500 * time.Millisecond)
//}
//
//func TestInstall(t *testing.T) {
//	before(t)
//	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 0 {
//		t.Error(state_cloud.GlobalCloudLayout.Current)
//	}
//	if len(state_cloud.GlobalAvailableInstances) != 0 {
//		t.Error(state_cloud.GlobalAvailableInstances)
//	}
//
//	res := installer.InstallNewInstance(installer.TestClientConfig("orca_client_1_metrics=20_20_10"), "172.16.147.189")
//	if !res {
//		t.Error("install failed")
//	}
//	res = installer.InstallNewInstance(installer.TestClientConfig("orca_client_2_metrics=100_100_100"), "172.16.147.190")
//	if !res {
//		t.Error("install failed")
//	}
//
//	time.Sleep(15 * time.Second)
//
//	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 2 {
//		t.Error(state_cloud.GlobalCloudLayout.Current)
//	}
//	if len(state_cloud.GlobalAvailableInstances) != 2 {
//		t.Error(state_cloud.GlobalAvailableInstances)
//	}
//	resour, _ := state_cloud.GlobalAvailableInstances.GetResources("orca_client_1_metrics=20_20_10")
//	if resour.TotalCpuResource != 20 || resour.TotalNetworkResource != 10 {
//		t.Error(resour)
//	}
//
//	resour2, _ := state_cloud.GlobalAvailableInstances.GetResources("orca_client_2_metrics=100_100_100")
//	if resour2.TotalCpuResource != 100 || resour2.TotalMemoryResource != 100 {
//		t.Error(resour2)
//	}
//
//	scheduler.TriggerRun()
//
//	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 2 {
//		t.Error(state_cloud.GlobalCloudLayout.Current)
//	}
//	if len(state_cloud.GlobalCloudLayout.Desired.Layout) != 2 {
//		t.Error(state_cloud.GlobalCloudLayout.Desired)
//	}
//}
//
//
//func TestRunningOk(t *testing.T) {
//	before(t)
//
//	time.Sleep(15 * time.Second)
//
//	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 2 {
//		t.Error(state_cloud.GlobalCloudLayout.Current)
//	}
//
//	if state_cloud.GlobalAvailableInstances.GlobalResourceConsumption().UsedCpuResource != 0 || state_cloud.GlobalAvailableInstances.GlobalResourceConsumption().TotalCpuResource != 120 {
//		t.Error(state_cloud.GlobalAvailableInstances)
//	}
//
//	scheduler.TriggerRun()
//
//	if len(state_cloud.GlobalCloudLayout.Desired.Layout) != 2 {
//		t.Error(state_cloud.GlobalCloudLayout.Desired)
//	}
//
//	client1, _ := state_cloud.GlobalCloudLayout.Desired.GetHost("orca_client_1_metrics=20_20_10")
//	if client1.Apps["workerApp3"].DeploymentCount != 0 || client1.Apps["workerApp2"].DeploymentCount != 1 || client1.Apps["httpApp1"].DeploymentCount != 1 {
//		t.Error(client1)
//	}
//
//	client2, _ := state_cloud.GlobalCloudLayout.Desired.GetHost("orca_client_2_metrics=100_100_100")
//	if client2.Apps["workerApp3"].DeploymentCount != 10 || client2.Apps["workerApp2"].DeploymentCount != 1 || client2.Apps["httpApp1"].DeploymentCount != 0 {
//		t.Error(client2)
//	}
//
//	if state_cloud.GlobalAvailableInstances.GlobalResourceConsumption().UsedCpuResource != 75 || state_cloud.GlobalAvailableInstances.GlobalResourceConsumption().TotalCpuResource != 120 {
//		t.Error(state_cloud.GlobalAvailableInstances)
//	}
//
//	time.Sleep(15 * time.Second)
//
//	current1,_ := state_cloud.GlobalCloudLayout.Desired.GetHost("orca_client_1_metrics=20_20_10")
//	if current1.Apps["workerApp3"].DeploymentCount != 0 || current1.Apps["workerApp2"].DeploymentCount != 1 || current1.Apps["httpApp1"].DeploymentCount != 1 {
//		t.Error(client1)
//	}
//
//	current2,_ := state_cloud.GlobalCloudLayout.Desired.GetHost("orca_client_2_metrics=100_100_100")
//	if current2.Apps["workerApp3"].DeploymentCount != 10 || current2.Apps["workerApp2"].DeploymentCount != 1 || current2.Apps["httpApp1"].DeploymentCount != 0 {
//		t.Error(current2)
//	}
//
//	worker2Rating, _ := tracker.GlobalAppsStatusTracker.GetRating("workerApp2", "2.0")
//	worker3Rating, _ := tracker.GlobalAppsStatusTracker.GetRating("workerApp3", "3.0")
//	httpRating, _ := tracker.GlobalAppsStatusTracker.GetRating("httpApp1", "1.0")
//	if httpRating != tracker.RATING_STABLE || worker2Rating != tracker.RATING_STABLE || worker3Rating != tracker.RATING_STABLE {
//		t.Error(tracker.GlobalAppsStatusTracker)
//	}
//
//	t.Error(state_cloud.GlobalCloudLayout.Current)
//	t.Error(state_cloud.GlobalCloudLayout.Desired)
//	t.Error(planner.Diff(state_cloud.GlobalCloudLayout.Desired, state_cloud.GlobalCloudLayout.Current))
//	t.Error(planner.Queue)
//}