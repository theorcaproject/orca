package tracker


import (
	"testing"
	"time"
	"gatoor/orca/rewriteTrainer/cloud"
	"gatoor/orca/rewriteTrainer/state/cloud"
)



func TestHostTracker_Get(t *testing.T) {
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
	ti := time.Now().Add(-time.Duration(time.Minute * 1))
	state_cloud.GlobalCloudLayout.Init()
	state_cloud.GlobalCloudLayout.Current.AddEmptyHost("host1")
	GlobalHostTracker.Update("host1", ti)

	res, _ := GlobalHostTracker.Get("host1")

	if res.LastCheckin != ti {
		t.Error(res)
	}

	GlobalHostTracker.CheckCheckinTimeout()

	if len(GlobalHostCrashHandler) != 0 {
		t.Error(GlobalHostCrashHandler)
	}

	GlobalHostTracker.Update("host1", time.Now().Add(-time.Duration(time.Minute * 10)))
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