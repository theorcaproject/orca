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

package scheduler

import (
	"testing"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/rewriteTrainer/state/configuration"
	"gatoor/orca/rewriteTrainer/planner"
	"gatoor/orca/rewriteTrainer/cloud"
	"gatoor/orca/base"
)


func initAll() {
	state_cloud.GlobalCloudLayout.Init()
	state_configuration.GlobalConfigurationState.Init()
	cloud.Init()
}



func TestScheduler_run_NoChanges(t *testing.T) {
	initAll()

	state_cloud.GlobalCloudLayout.Current.AddHost("hostId1", state_cloud.CloudLayoutElement{
		HostId: "hostId1",
		IpAddress: "0.0.0.0",
		HabitatVersion: 0,
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state_cloud.GlobalCloudLayout.Current.AddApp("hostId1", "someapp", 2, 1)
	state_cloud.GlobalCloudLayout.Current.AddApp("hostId1", "someapp2", 4, 5)

	state_cloud.GlobalCloudLayout.Desired.AddHost("hostId1", state_cloud.CloudLayoutElement{
		HostId: "hostId1",
		IpAddress: "0.0.0.0",
		HabitatVersion: 0,
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state_cloud.GlobalCloudLayout.Desired.AddApp("hostId1", "someapp", 2, 1)
	state_cloud.GlobalCloudLayout.Desired.AddApp("hostId1", "someapp2", 4, 5)

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
		HabitatVersion: 0,
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state_cloud.GlobalCloudLayout.Current.AddApp("hostId1", "someapp", 2, 1)
	state_cloud.GlobalCloudLayout.Current.AddApp("hostId1", "someapp2", 4, 5)
	state_cloud.GlobalCloudLayout.Current.AddApp("hostId1", "someapp3", 6, 5)

	state_cloud.GlobalCloudLayout.Desired.AddHost("hostId1", state_cloud.CloudLayoutElement{
		HostId: "hostId1",
		IpAddress: "0.0.0.0",
		HabitatVersion: 0,
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state_cloud.GlobalCloudLayout.Desired.AddApp("hostId1", "someapp", 2, 4)
	state_cloud.GlobalCloudLayout.Desired.AddApp("hostId1", "someapp2", 3, 5)
	state_cloud.GlobalCloudLayout.Desired.AddApp("hostId1", "someapp3", 7, 5)

	run()

	if planner.Queue.Empty("hostId1") {
		t.Error("queue is empty")
	}
	if !planner.Queue.Empty("unkown") {
		t.Error("queue should be empty")
	}

	elem, _ := planner.Queue.Get("hostId1")

	if elem["someapp2"].Version.Version != 3 {
		t.Error("wrong version")
	}
	if elem["someapp3"].Version.Version != 7 {
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
		HabitatVersion: 0,
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state_cloud.GlobalCloudLayout.Current.AddApp("hostId1", "someapp", 2, 1)
	state_cloud.GlobalCloudLayout.Current.AddApp("hostId1", "someapp2", 4, 5)
	state_cloud.GlobalCloudLayout.Current.AddApp("hostId1", "someapp3", 6, 5)

	state_cloud.GlobalCloudLayout.Desired.AddHost("hostId1", state_cloud.CloudLayoutElement{
		HostId: "hostId1",
		IpAddress: "0.0.0.0",
		HabitatVersion: 0,
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state_cloud.GlobalCloudLayout.Desired.AddApp("hostId1", "someapp", 2, 4)
	state_cloud.GlobalCloudLayout.Desired.AddApp("hostId1", "someapp2", 3, 5)
	state_cloud.GlobalCloudLayout.Desired.AddApp("hostId1", "someapp3", 7, 5)


	state_cloud.GlobalCloudLayout.Current.AddHost("hostId2", state_cloud.CloudLayoutElement{
		HostId: "hostId2",
		IpAddress: "0.0.0.0",
		HabitatVersion: 0,
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state_cloud.GlobalCloudLayout.Current.AddApp("hostId2", "someappR", 8, 10)
	state_cloud.GlobalCloudLayout.Current.AddApp("hostId2", "someappR2", 9, 50)

	state_cloud.GlobalCloudLayout.Desired.AddHost("hostId2", state_cloud.CloudLayoutElement{
		HostId: "hostId2",
		IpAddress: "0.0.0.0",
		HabitatVersion: 0,
		Apps: make(map[base.AppName]state_cloud.AppsVersion),
	})
	state_cloud.GlobalCloudLayout.Desired.AddApp("hostId2", "someappR", 88, 40)
	state_cloud.GlobalCloudLayout.Desired.AddApp("hostId2", "someappR3", 99, 50)

	run()

	if planner.Queue.Empty("hostId1") {
		t.Error("queue is empty")
	}
	if !planner.Queue.Empty("unkown") {
		t.Error("queue should be empty")
	}

	elem, _ := planner.Queue.Get("hostId1")

	if elem["someapp2"].Version.Version != 3 {
		t.Error("wrong version")
	}
	if elem["someapp3"].Version.Version != 7 {
		t.Error("wrong version")
	}
	if elem["someapp"].Version.DeploymentCount != 4 {
		t.Error("wrong deployment count")
	}

	elem2, _ := planner.Queue.Get("hostId2")

	if elem2["someappR2"].Version.Version != 9 {
		t.Error(elem2["someappR2"].Version.Version)
	}
	if elem2["someappR2"].Version.DeploymentCount != 0 {
		t.Error("ap remove not working")
	}
	if elem2["someappR3"].Version.Version != 99 {
		t.Error(elem2["someappR3"].Version.Version)
	}
	if elem2["someappR"].Version.DeploymentCount != 40 {
		t.Error("wrong deployment count")
	}


}