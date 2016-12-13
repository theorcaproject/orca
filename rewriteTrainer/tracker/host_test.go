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

package tracker


import (
	"testing"
	"time"
	"gatoor/orca/rewriteTrainer/cloud"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/base"
)



func TestHostTracker_Get(t *testing.T) {
	cloud.Init()
	state_cloud.GlobalCloudLayout.Init()
	_, err := GlobalHostTracker.Get("host1")
	if err == nil {
		t.Error("should throw error")
	}
	ti := time.Now()

	GlobalHostTracker.Update("host1", ti)

	res, _ := GlobalHostTracker.Get("host1")

	if res.LastCheckin != ti {
		t.Error(res)
	}
}


func TestHostTracker_CheckCheckinTimeout(t *testing.T) {
	cloud.Init()
	ti := time.Now().Add(-time.Duration(time.Minute * 1))
	state_cloud.GlobalCloudLayout.Init()
	state_cloud.GlobalCloudLayout.Current.AddEmptyHost("host1")

	GlobalHostTracker.Update("host1", ti)
	cloud.Init()

	res, _ := GlobalHostTracker.Get("host1")

	if res.LastCheckin != ti {
		t.Error(res)
	}

	if len(state_cloud.GlobalAvailableInstances) != 1 || state_cloud.GlobalAvailableInstances["host1"].TotalCpuResource != 10 {
		t.Error(state_cloud.GlobalAvailableInstances)
	}

	GlobalHostTracker.CheckCheckinTimeout()

	if len(GlobalHostCrashHandler) != 0 {
		t.Error(GlobalHostCrashHandler)
	}

	GlobalHostTracker.Update("host1", time.Now().Add(-time.Duration(time.Minute * 1000)))
	GlobalHostTracker.CheckCheckinTimeout()

	if GlobalHostCrashHandler["host1"].OldHostId != "host1" {
		t.Error(res)
	}
	_, err := state_cloud.GlobalCloudLayout.Current.GetHost("host1")
	if err == nil {
		t.Error(state_cloud.GlobalCloudLayout.Current)
	}
}


func TestHostTracker_CheckCloudProvider(t *testing.T) {
	GlobalHostCrashHandler = HostCrashHandler{}
	GlobalHostTracker = HostTracker{}
	cloud.Init()
	if len(GlobalHostCrashHandler) != 0 {
		t.Error(GlobalHostCrashHandler)
	}

	GlobalHostTracker.CheckCloudProvider()

	if len(GlobalHostCrashHandler) != 0 {
		t.Error(GlobalHostCrashHandler)
	}

	GlobalHostTracker.Update("healthy", time.Now().UTC())

	GlobalHostTracker.CheckCloudProvider()

	if len(GlobalHostCrashHandler) != 0 {
		t.Error(GlobalHostCrashHandler)
	}
	GlobalHostTracker.Update("dead", time.Now().UTC())

	GlobalHostTracker.CheckCloudProvider()

	res := GlobalHostCrashHandler["dead"]
	if res.OldHostId != "dead" || res.Status != HOST_STATUS_SPAWN_TRIGGERED || res.NewHostId != "new_dead" {
		t.Error(res)
	}
}


func TestHostTracker_HandleCloudProviderEvent(t *testing.T) {
	GlobalHostCrashHandler = HostCrashHandler{}
	GlobalHostTracker = HostTracker{}

	if len(GlobalHostCrashHandler) != 0 {
		t.Error(GlobalHostCrashHandler)
	}

	GlobalHostTracker.HandleCloudProviderEvent(cloud.ProviderEvent{"somehost", cloud.PROVIDER_EVENT_KILLED})

	res := GlobalHostCrashHandler["somehost"]
	if res.OldHostId != "somehost" || res.Status != HOST_STATUS_SPAWN_TRIGGERED || res.NewHostId != "new_somehost" {
		t.Error(res)
	}

	GlobalHostTracker.HandleCloudProviderEvent(cloud.ProviderEvent{"new_somehost", cloud.PROVIDER_EVENT_READY})

	res1, err := GlobalHostCrashHandler.Get("somehost")
	if err == nil {
		t.Error(res1)
	}
}


func TestHostCrashHandler_checkinHost(t *testing.T) {
	GlobalHostCrashHandler = HostCrashHandler{}
	cloud.Init()
	if len(GlobalHostCrashHandler) != 0 {
		t.Error(GlobalHostCrashHandler)
	}

	GlobalHostTracker.HandleCloudProviderEvent(cloud.ProviderEvent{"somehost", cloud.PROVIDER_EVENT_KILLED})

	res := GlobalHostCrashHandler["somehost"]
	if res.OldHostId != "somehost" || res.Status != HOST_STATUS_SPAWN_TRIGGERED || res.NewHostId != "new_somehost" {
		t.Error(res)
	}

	GlobalHostCrashHandler.checkinHost("new_somehost")
	res1, err := GlobalHostCrashHandler.Get("somehost")
	if err == nil {
		t.Error(res1)
	}
}

func Test_removeHostFromState(t *testing.T) {
	state_cloud.GlobalCloudLayout.Init()
        state_cloud.GlobalCloudLayout.Current.AddEmptyHost("lamehost")
	state_cloud.GlobalAvailableInstances.Update("lamehost", base.InstanceResources{})
	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 1 {
		t.Error(state_cloud.GlobalCloudLayout.Current.Layout)
	}
	if len(state_cloud.GlobalAvailableInstances) != 1 {
		t.Error(state_cloud.GlobalAvailableInstances)
	}

	removeHostFromState("otherhost")

	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 1 {
		t.Error(state_cloud.GlobalCloudLayout.Current.Layout)
	}
	if len(state_cloud.GlobalAvailableInstances) != 1 {
		t.Error(state_cloud.GlobalAvailableInstances)
	}

	removeHostFromState("lamehost")

	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 0 {
		t.Error(state_cloud.GlobalCloudLayout.Current.Layout)
	}
	if len(state_cloud.GlobalAvailableInstances) != 0 {
		t.Error(state_cloud.GlobalAvailableInstances)
	}
}