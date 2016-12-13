package responder


import (
	"testing"
	"gatoor/orca/rewriteTrainer/planner"
	"gatoor/orca/base"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/rewriteTrainer/state/configuration"
	"gatoor/orca/rewriteTrainer/tracker"
	"gatoor/orca/rewriteTrainer/cloud"
)


func initPlanner() {
	masterCloud := state_cloud.CloudLayout{Type: "master",Layout: make(map[base.HostId]state_cloud.CloudLayoutElement),}
	slaveCloud := state_cloud.CloudLayout{Type: "master",Layout: make(map[base.HostId]state_cloud.CloudLayoutElement),}
	masterCloud.AddEmptyHost("host1")
	masterCloud.AddEmptyHost("host2")
	slaveCloud.AddEmptyHost("host1")
	slaveCloud.AddEmptyHost("host2")
	masterCloud.AddApp("host1", "app1", 1, 1)
	slaveCloud.AddApp("host1", "app1", 1, 1)
	masterCloud.AddApp("host1", "app2", 2, 1)
	slaveCloud.AddApp("host1", "app2", 1, 1)
	masterCloud.AddApp("host2", "app1", 1, 10)
	slaveCloud.AddApp("host2", "app1", 1, 2)
	planner.Queue.Apply(planner.Diff(masterCloud, slaveCloud))
}

func TestResponder_getQueueElement_NoElemForHost(t *testing.T) {
	initPlanner()

	_, _, err := getQueueElement("unknown")
	if err == nil {
		t.Error("should return err")
	}
}

func TestResponder_getQueueElement_getApplyingState(t *testing.T) {
	initPlanner()

	app, obj, err := getQueueElement("host1")

	if err != nil {
		t.Error("got err")
	}
	if app != "app2" {
		t.Error("wrong app")
	}
	if obj.Version.Version != 2 {
		t.Error("wrong version")
	}
	if obj.State != planner.STATE_APPLYING {
		t.Error(obj)
	}

	app2, obj2, err2 := getQueueElement("host1")

	if err2 != nil {
		t.Error("got err")
	}
	if app2 != "app2"{
		t.Error("wrong app")
	}
	if obj2.Version.Version != 2 {
		t.Error("wrong version")
	}
	if obj2.State != planner.STATE_APPLYING {
		t.Error("wrong state")
	}
}


func TestResponder_GetConfigForHost_noChanges(t *testing.T) {
	initPlanner()

	_, _, err := getQueueElement("unknown")
	if err == nil {
		t.Error("should return err")
	}

	_, err2 := GetConfigForHost("unkown")
	if err2 == nil {
		t.Error("should return err")
	}
}


func TestResponder_GetConfigForHost_HasChanges(t *testing.T) {
	initPlanner()

	_, err0 := GetConfigForHost("host1")
	if err0 == nil {
		t.Error("expected error global config state")
	}

	state_configuration.GlobalConfigurationState.Init()
	state_configuration.GlobalConfigurationState.ConfigureApp(base.AppConfiguration{
		Name: "app1", Type: base.APP_HTTP, Version: 1,
	})
	state_configuration.GlobalConfigurationState.ConfigureApp(base.AppConfiguration{
		Name: "app2", Type: base.APP_HTTP, Version: 2,
	})

	config, err := GetConfigForHost("host1")
	if err != nil {
		t.Error("got err", err)
	}
	if config.DeploymentCount != 1 {
		t.Error("wrong deployment count")
	}
	config2, err2 := GetConfigForHost("host2")
	if err2 != nil {
		t.Error("got err", err)
	}
	if config2.DeploymentCount != 10 {
		t.Error("wrong deployment count")
	}
}


func TestResponder_simpleAppCheck(t *testing.T) {
	state_cloud.GlobalCloudLayout.Init()
	if len(tracker.GlobalAppsStatusTracker) != 0 {
		t.Error(tracker.GlobalAppsStatusTracker)
	}

	simpleAppCheck(base.AppInfo{Name: "app1", Version:1, Status:base.STATUS_RUNNING}, "host1")

	if len(tracker.GlobalAppsStatusTracker) != 1 {
		t.Error(tracker.GlobalAppsStatusTracker)
	}
	if tracker.GlobalAppsStatusTracker["app1"][1].RunningCount != 1 {
		t.Error(tracker.GlobalAppsStatusTracker)
	}

	rating0, _ := tracker.GlobalAppsStatusTracker.GetRating("app1", 1)
	if rating0 != tracker.RATING_STABLE {
		t.Error(tracker.GlobalAppsStatusTracker)
	}


	simpleAppCheck(base.AppInfo{Name: "app1", Version:1, Status:base.STATUS_DEAD}, "host1")

	if len(tracker.GlobalAppsStatusTracker) != 1 {
		t.Error(tracker.GlobalAppsStatusTracker)
	}
	rating, _ := tracker.GlobalAppsStatusTracker.GetRating("app1", 1)
	if rating != tracker.RATING_CRASHED {
		t.Error(tracker.GlobalAppsStatusTracker)
	}
}

func TestResponder_checkAppUpdate(t *testing.T) {
	state_cloud.GlobalCloudLayout.Init()
	planner.Queue = *planner.NewPlannerQueue()
	tracker.GlobalAppsStatusTracker = tracker.AppsStatusTracker{}
	if len(tracker.GlobalAppsStatusTracker) != 0 {
		t.Error(tracker.GlobalAppsStatusTracker)
	}

	if len(planner.Queue.Queue) != 0 {
		t.Error(planner.Queue.Queue)
	}

	//app is running -> do nothing

	planner.Queue.Add("host1", "app1", state_cloud.AppsVersion{Version: 1, DeploymentCount: 0})

	before, _ := planner.Queue.Get("host1")

	if before["app1"].State != planner.STATE_QUEUED {
		t.Error(before)
	}

	checkAppUpdate(base.AppInfo{Name: "app1", Version:1, Status:base.STATUS_RUNNING}, "host1", before["app1"])

	if len(tracker.GlobalAppsStatusTracker) != 0 {
		t.Error(tracker.GlobalAppsStatusTracker)
	}

	//illegal stuff, do nothing

	before1, _ := planner.Queue.Get("host1")

	if before1["app1"].State != planner.STATE_QUEUED {
		t.Error(before1)
	}

	checkAppUpdate(base.AppInfo{Name: "app1", Version:1, Status:base.STATUS_DEAD}, "host1", before1["app1"])

	if len(tracker.GlobalAppsStatusTracker) != 0 {    // Illegal state, do nothing
		t.Error(tracker.GlobalAppsStatusTracker)
	}

	//now starts the real part

	planner.Queue.SetState("host1", "app1", planner.STATE_APPLYING)
	before2, _ := planner.Queue.Get("host1")

	if before2["app1"].State != planner.STATE_APPLYING {
		t.Error(before2)
	}

	// app update was successful

	planner.Queue.SetState("host1", "app1", planner.STATE_APPLYING)
	updated := before2["app1"]
	updated.Version.Version = 11
	checkAppUpdate(base.AppInfo{Name: "app1", Version:11, Status:base.STATUS_RUNNING}, "host1", updated)

	if len(tracker.GlobalAppsStatusTracker) != 1 {
		t.Error(tracker.GlobalAppsStatusTracker)
	}
	if len(tracker.GlobalAppsStatusTracker["app1"][11].CrashDetails) != 0 {
		t.Error(tracker.GlobalAppsStatusTracker["app1"][11])
	}
	if tracker.GlobalAppsStatusTracker["app1"][11].Rating != tracker.RATING_STABLE {
		t.Error(tracker.GlobalAppsStatusTracker["app1"][11])
	}
	if !planner.Queue.Empty("host1") {
		t.Error(planner.Queue)
	}

	planner.Queue.Add("host1", "app1", state_cloud.AppsVersion{Version: 1, DeploymentCount: 1})
	planner.Queue.SetState("host1", "app1", planner.STATE_APPLYING)
	before2, _ = planner.Queue.Get("host1")

	// app is dead - the update crashed the app and no rollback

	checkAppUpdate(base.AppInfo{Name: "app1", Version:1, Status:base.STATUS_DEAD}, "host1", before2["app1"])

	if len(tracker.GlobalAppsStatusTracker) != 1 {
		t.Error(tracker.GlobalAppsStatusTracker)
	}
	if len(tracker.GlobalAppsStatusTracker["app1"][1].CrashDetails) != 1 {
		t.Error(tracker.GlobalAppsStatusTracker["app1"][1])
	}
	if tracker.GlobalAppsStatusTracker["app1"][1].Rating != tracker.RATING_CRASHED {
		t.Error(tracker.GlobalAppsStatusTracker["app1"][1])
	}

	//app will be rolled back
	tracker.GlobalAppsStatusTracker = tracker.AppsStatusTracker{}
	planner.Queue.Add("host1", "app1", state_cloud.AppsVersion{Version: 1, DeploymentCount: 1})
	planner.Queue.SetState("host1", "app1", planner.STATE_APPLYING)
	updated = before2["app1"]
	updated.Version.Version = 2
	checkAppUpdate(base.AppInfo{Name: "app1", Version:1, Status:base.STATUS_RUNNING}, "host1", updated)

	if len(tracker.GlobalAppsStatusTracker) != 1 {
		t.Error(tracker.GlobalAppsStatusTracker)
	}
	if len(tracker.GlobalAppsStatusTracker["app1"][1].CrashDetails) != 0 {
		t.Error(tracker.GlobalAppsStatusTracker["app1"][1])
	}
	if tracker.GlobalAppsStatusTracker["app1"][2].Rating != tracker.RATING_CRASHED || tracker.GlobalAppsStatusTracker["app1"][2].CrashDetails[0].Cause != tracker.APP_EVENT_ROLLBACK {
		t.Error(tracker.GlobalAppsStatusTracker["app1"][2])
	}
}

func Test_handleEmptyHost_RecentlySpawned(t *testing.T) {
	cloud.Init()
	provider := cloud.CurrentProvider.(*cloud.TestProvider)
	provider.SpawnInstance("host1")
	if len(provider.GetSpawnLog()) != 1 {
		t.Error(provider)
	}
	handleEmptyHost(base.HostInfo{"host1", "", base.OsInfo{}, []base.AppInfo{}})
	if len(provider.GetSpawnLog()) != 1 || provider.GetSpawnLog()[0] != "host1" {
		t.Errorf("%+v", cloud.CurrentProvider)
	}
	if len(provider.KillList) != 0 {
		t.Errorf("%+v", provider)
	}
}

func Test_handleEmptyHost_NotRecentlySpawned(t *testing.T) {
	cloud.Init()
	provider := cloud.CurrentProvider.(*cloud.TestProvider)
	if len(provider.GetSpawnLog()) != 0 {
		t.Error(provider)
	}
	handleEmptyHost(base.HostInfo{"host1", "", base.OsInfo{}, []base.AppInfo{}})
	//if len(provider.KillList) != 1 {
	//	t.Errorf("%+v", provider)
	//}
}
