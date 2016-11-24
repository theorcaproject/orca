package api

import (
	"testing"
	"gatoor/orca/rewriteTrainer/metrics"
	"gatoor/orca/rewriteTrainer/base"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/rewriteTrainer/db"
	"gatoor/orca/rewriteTrainer/state/configuration"
	"gatoor/orca/rewriteTrainer/config"
	"gatoor/orca/rewriteTrainer/state/needs"
	"gatoor/orca/rewriteTrainer/tracker"
	"time"
	"gatoor/orca/rewriteTrainer/responder"
	"gatoor/orca/rewriteTrainer/planner"
)


func applySampleConfig() {
	conf := config.JsonConfiguration{}

	conf.Trainer.Port = 5000

	conf.Habitats = []config.HabitatJsonConfiguration{
		{
			Name: "habitat1",
			Version: "0.1",
			InstallCommands: []base.OsCommand{
				{
					Type: base.EXEC_COMMAND,
					Command: base.Command{"ls", "/home"},
				},
				{
					Type: base.FILE_COMMAND,
					Command: base.Command{"/etc/orca.conf", "somefilecontent as a string"},
				},
			},
		},
		{
			Name: "habitat2",
			Version: "0.1",
			InstallCommands: []base.OsCommand{
				{
					Type: base.EXEC_COMMAND,
					Command: base.Command{"ps", "aux"},
				},
				{
					Type: base.FILE_COMMAND,
					Command: base.Command{"/etc/orca.conf", "different config"},
				},
			},
		},
	}

	httpApp1 := config.AppJsonConfiguration{
		Name: "httpApp_1",
		Version: "http_1.0",
		Type: base.APP_HTTP,
		MinDeploymentCount: 3,
		MaxDeploymentCount: 10,
		InstallCommands: []base.OsCommand{
			{
				Type: base.EXEC_COMMAND,
				Command: base.Command{"ls", "/home"},
			},
			{
				Type: base.FILE_COMMAND,
				Command: base.Command{"/server/app1/app1.conf", "somefilecontent as a string"},
			},
		},
		QueryStateCommand: base.OsCommand{
			Type: base.EXEC_COMMAND,
			Command: base.Command{"wget", "http://localhost:1234/check"},
		},
		RemoveCommand: base.OsCommand{
			Type: base.EXEC_COMMAND,
			Command: base.Command{"rm", "-rf /server/app1"},
		},
		Needs: state_needs.AppNeeds{
			MemoryNeeds: state_needs.MemoryNeeds(1),
			CpuNeeds: state_needs.CpuNeeds(1),
			NetworkNeeds: state_needs.NetworkNeeds(1),
		},
	}

	httpApp1_v2 := config.AppJsonConfiguration{
		Name: "httpApp_1",
		Version: "http_1.1",
		Type: base.APP_HTTP,
		MinDeploymentCount: 2,
		MaxDeploymentCount: 10,
		InstallCommands: []base.OsCommand{
			{
				Type: base.EXEC_COMMAND,
				Command: base.Command{"ls", "/home"},
			},
			{
				Type: base.FILE_COMMAND,
				Command: base.Command{"/server/app1/app1.conf", "somefilecontent as a string"},
			},
		},
		QueryStateCommand: base.OsCommand{
			Type: base.EXEC_COMMAND,
			Command: base.Command{"wget", "http://localhost:1234/check"},
		},
		RemoveCommand: base.OsCommand{
			Type: base.EXEC_COMMAND,
			Command: base.Command{"rm", "-rf /server/app1"},
		},
		Needs: state_needs.AppNeeds{
			MemoryNeeds: state_needs.MemoryNeeds(2),
			CpuNeeds: state_needs.CpuNeeds(2),
			NetworkNeeds: state_needs.NetworkNeeds(5),
		},
	}

	httpApp2 := config.AppJsonConfiguration{
		Name: "httpApp_2",
		Version: "http_2.0",
		Type: base.APP_HTTP,
		MinDeploymentCount: 4,
		MaxDeploymentCount: 10,
		InstallCommands: []base.OsCommand{
			{
				Type: base.EXEC_COMMAND,
				Command: base.Command{"ls", "/home"},
			},
			{
				Type: base.FILE_COMMAND,
				Command: base.Command{"/server/app1/app1.conf", "somefilecontent as a string"},
			},
		},
		QueryStateCommand: base.OsCommand{
			Type: base.EXEC_COMMAND,
			Command: base.Command{"wget", "http://localhost:1234/check"},
		},
		RemoveCommand: base.OsCommand{
			Type: base.EXEC_COMMAND,
			Command: base.Command{"rm", "-rf /server/app1"},
		},
		Needs: state_needs.AppNeeds{
			MemoryNeeds: state_needs.MemoryNeeds(1),
			CpuNeeds: state_needs.CpuNeeds(1),
			NetworkNeeds: state_needs.NetworkNeeds(1),
		},
	}

	workerApp1 := config.AppJsonConfiguration{
		Name: "workerApp_1",
		Version: "worker_1.0",
		Type: base.APP_WORKER,
		MinDeploymentCount: 1,
		MaxDeploymentCount: 1,
		InstallCommands: []base.OsCommand{
			{
				Type: base.EXEC_COMMAND,
				Command: base.Command{"ls", "/home"},
			},
			{
				Type: base.FILE_COMMAND,
				Command: base.Command{"/server/app1/app1.conf", "somefilecontent as a string"},
			},
		},
		QueryStateCommand: base.OsCommand{
			Type: base.EXEC_COMMAND,
			Command: base.Command{"wget", "http://localhost:1234/check"},
		},
		RemoveCommand: base.OsCommand{
			Type: base.EXEC_COMMAND,
			Command: base.Command{"rm", "-rf /server/app1"},
		},
		Needs: state_needs.AppNeeds{
			CpuNeeds: state_needs.CpuNeeds(50),
			MemoryNeeds: state_needs.MemoryNeeds(10),
			NetworkNeeds: state_needs.NetworkNeeds(10),
		},
	}

	workerApp1_v2 := config.AppJsonConfiguration{
		Name: "workerApp_1",
		Version: "worker_1.1",
		Type: base.APP_WORKER,
		MinDeploymentCount: 1,
		MaxDeploymentCount: 1,
		InstallCommands: []base.OsCommand{
			{
				Type: base.EXEC_COMMAND,
				Command: base.Command{"ls", "/home"},
			},
			{
				Type: base.FILE_COMMAND,
				Command: base.Command{"/server/app1/app1.conf", "somefilecontent as a string"},
			},
		},
		QueryStateCommand: base.OsCommand{
			Type: base.EXEC_COMMAND,
			Command: base.Command{"wget", "http://localhost:1234/check"},
		},
		RemoveCommand: base.OsCommand{
			Type: base.EXEC_COMMAND,
			Command: base.Command{"rm", "-rf /server/app1"},
		},
		Needs: state_needs.AppNeeds{
			CpuNeeds: state_needs.CpuNeeds(70),
			MemoryNeeds: state_needs.MemoryNeeds(40),
			NetworkNeeds: state_needs.NetworkNeeds(30),
		},
	}

	workerApp2 := config.AppJsonConfiguration{
		Name: "workerApp_2",
		Version: "worker_2.0",
		Type: base.APP_WORKER,
		MinDeploymentCount: 5,
		MaxDeploymentCount: 10,
		InstallCommands: []base.OsCommand{
			{
				Type: base.EXEC_COMMAND,
				Command: base.Command{"ls", "/home"},
			},
			{
				Type: base.FILE_COMMAND,
				Command: base.Command{"/server/app1/app1.conf", "somefilecontent as a string"},
			},
		},
		QueryStateCommand: base.OsCommand{
			Type: base.EXEC_COMMAND,
			Command: base.Command{"wget", "http://localhost:1234/check"},
		},
		RemoveCommand: base.OsCommand{
			Type: base.EXEC_COMMAND,
			Command: base.Command{"rm", "-rf /server/app1"},
		},
		Needs: state_needs.AppNeeds{
			CpuNeeds: state_needs.CpuNeeds(23),
			MemoryNeeds: state_needs.MemoryNeeds(23),
			NetworkNeeds: state_needs.NetworkNeeds(23),
		},
	}

	workerApp3 := config.AppJsonConfiguration{
		Name: "workerApp_3",
		Version: "worker_3.0",
		Type: base.APP_WORKER,
		MinDeploymentCount: 100,
		MaxDeploymentCount: 200,
		InstallCommands: []base.OsCommand{
			{
				Type: base.EXEC_COMMAND,
				Command: base.Command{"ls", "/home"},
			},
			{
				Type: base.FILE_COMMAND,
				Command: base.Command{"/server/app1/app1.conf", "somefilecontent as a string"},
			},
		},
		QueryStateCommand: base.OsCommand{
			Type: base.EXEC_COMMAND,
			Command: base.Command{"wget", "http://localhost:1234/check"},
		},
		RemoveCommand: base.OsCommand{
			Type: base.EXEC_COMMAND,
			Command: base.Command{"rm", "-rf /server/app1"},
		},
		Needs: state_needs.AppNeeds{
			CpuNeeds: state_needs.CpuNeeds(7),
			MemoryNeeds: state_needs.MemoryNeeds(2),
			NetworkNeeds: state_needs.NetworkNeeds(1),
		},
	}



	conf.Apps = []config.AppJsonConfiguration{
		httpApp1, httpApp1_v2, httpApp2,
		workerApp1, workerApp1_v2, workerApp2, workerApp3,
	}

	conf.ApplyToState()
}



func prepare() {
	db.Init("api_integration_test_")
	state_configuration.GlobalConfigurationState.Init()
	state_cloud.GlobalCloudLayout.Init()
	tracker.GlobalHostTracker = tracker.HostTracker{}
	tracker.GlobalAppsStatusTracker = tracker.AppsStatusTracker{}
	applySampleConfig()
}

func after() {
	db.Close()
}

func TestApi_doHandlePush_NoChanges(t *testing.T) {
	prepare()
	defer after()
	app1 := metrics.AppInfo{
		Type: base.APP_HTTP,
		Name: "httpApp_1",
		Version: "http_1.0",
		Status: base.STATUS_RUNNING,
	}
	app2 := metrics.AppInfo{
		Type: base.APP_WORKER,
		Name: "workerApp_2",
		Version: "worker_2.0",
		Status: base.STATUS_RUNNING,
	}

	info := metrics.HostInfo{
	      HostId: "host1", Apps: []metrics.AppInfo{app1, app2, app2},
	}

	stats := metrics.StatsWrapper{}

	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 0 {
		t.Error(state_cloud.GlobalCloudLayout.Current)
	}

	doHandlePush(info, stats)
	//nothing changed, so the config should be empty
	_, errC := responder.GetConfigForHost("host1")

	elem, _ := state_cloud.GlobalCloudLayout.Current.GetHost("host1")
	if elem.HostId != "host1" || elem.Apps["httpApp_1"].DeploymentCount != 1 || elem.Apps["workerApp_2"].DeploymentCount != 2 {
		t.Error(elem)
	}
	if errC == nil {
		t.Error(errC)
	}

	info2 := metrics.HostInfo{
		HostId: "host1", Apps: []metrics.AppInfo{app1, app2, app2, app2},
	}
	doHandlePush(info2, stats)

	elem2, _ := state_cloud.GlobalCloudLayout.Current.GetHost("host1")
	if elem2.HostId != "host1" || elem2.Apps["httpApp_1"].DeploymentCount != 1 || elem2.Apps["workerApp_2"].DeploymentCount != 3 {
		t.Error(elem2)
	}

	track, _ := tracker.GlobalHostTracker.Get("host1")
	if !track.LastCheckin.After(time.Now().UTC().Add(-time.Duration(time.Second * 2))) {
		t.Error(track)
	}

	appTrack := tracker.GlobalAppsStatusTracker["httpApp_1"]["http_1.0"]
	if appTrack.Rating != tracker.RATING_STABLE || appTrack.RunningCount != 2 || len(appTrack.CrashDetails) != 0{
		t.Errorf("%+v", appTrack)
	}
	_, errC2 := responder.GetConfigForHost("host1")

	if errC2 == nil {
		t.Error(errC2)
	}
}


func TestApi_doHandlePush_AppShouldBeUpdated(t *testing.T) {
	//check that the config contains the updated version
	prepare()
	defer after()
	app1 := metrics.AppInfo{
		Type: base.APP_HTTP,
		Name: "httpApp_1",
		Version: "http_1.0",
		Status: base.STATUS_RUNNING,
	}
	app2 := metrics.AppInfo{
		Type: base.APP_WORKER,
		Name: "workerApp_1",
		Version: "worker_1.0",
		Status: base.STATUS_RUNNING,
	}

	info := metrics.HostInfo{
		HostId: "host1", Apps: []metrics.AppInfo{app1, app2, app2},
	}

	stats := metrics.StatsWrapper{}

	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 0 {
		t.Error(state_cloud.GlobalCloudLayout.Current)
	}

	doHandlePush(info, stats)
	//nothing changed, so the config should be empty
	_, errC := responder.GetConfigForHost("host1")

	elem, _ := state_cloud.GlobalCloudLayout.Current.GetHost("host1")
	if elem.HostId != "host1" || elem.Apps["httpApp_1"].DeploymentCount != 1 || elem.Apps["workerApp_1"].DeploymentCount != 2 {
		t.Error(elem)
	}
	if errC == nil {
		t.Error(errC)
	}


	info2 := metrics.HostInfo{
		HostId: "host1", Apps: []metrics.AppInfo{app1, app2, app2, app2},
	}

	//set an update:
	planner.Queue.Add("host1", "httpApp_1", state_cloud.AppsVersion{Version: "http_1.1", DeploymentCount: 1})
	doHandlePush(info2, stats)

	elem2, _ := state_cloud.GlobalCloudLayout.Current.GetHost("host1")
	if elem2.HostId != "host1" || elem2.Apps["httpApp_1"].DeploymentCount != 1 || elem2.Apps["workerApp_1"].DeploymentCount != 3 {
		t.Error(elem2)
	}

	track, _ := tracker.GlobalHostTracker.Get("host1")
	if !track.LastCheckin.After(time.Now().UTC().Add(-time.Duration(time.Second * 2))) {
		t.Error(track)
	}

	appTrack := tracker.GlobalAppsStatusTracker["httpApp_1"]["http_1.0"]
	if appTrack.Rating != tracker.RATING_STABLE || appTrack.RunningCount != 2 || len(appTrack.CrashDetails) != 0{
		t.Errorf("%+v", appTrack)
	}
	conf, _ := responder.GetConfigForHost("host1")

	if conf.DeploymentCount != 1 || conf.AppConfiguration.Name != "httpApp_1" || conf.AppConfiguration.Type != base.APP_HTTP || conf.AppConfiguration.Version != "http_1.1" {
		t.Error(conf)
	}

	//same with worker app

	planner.Queue.RemoveHost("host1")
	planner.Queue.Add("host1", "workerApp_1", state_cloud.AppsVersion{Version: "worker_1.1", DeploymentCount: 10})
	doHandlePush(info2, stats)

	elem3, _ := state_cloud.GlobalCloudLayout.Current.GetHost("host1")
	if elem3.HostId != "host1" || elem3.Apps["httpApp_1"].DeploymentCount != 1 || elem3.Apps["workerApp_1"].DeploymentCount != 3 {
		t.Error(elem3)
	}

	track1, _ := tracker.GlobalHostTracker.Get("host1")
	if !track.LastCheckin.After(time.Now().UTC().Add(-time.Duration(time.Second * 2))) {
		t.Error(track1)
	}

	appTrack1 := tracker.GlobalAppsStatusTracker["httpApp_1"]["http_1.0"]
	if appTrack1.Rating != tracker.RATING_STABLE || appTrack1.RunningCount != 3 || len(appTrack1.CrashDetails) != 0{
		t.Errorf("%+v", appTrack1)
	}
	conf1, _ := responder.GetConfigForHost("host1")

	if conf1.DeploymentCount != 10 || conf1.AppConfiguration.Name != "workerApp_1" || conf1.AppConfiguration.Type != base.APP_WORKER || conf1.AppConfiguration.Version != "worker_1.1" {
		t.Error(conf)
	}
}


func TestApi_doHandlePush_AppStillUpdating(t *testing.T) {
	// An app update is not done jet.
	// check that the CurrentLayout is NOT updated
	// check that the app tracker is NOT updated
	prepare()
	defer after()
	app1 := metrics.AppInfo{
		Type: base.APP_HTTP,
		Name: "httpApp_1",
		Version: "http_1.1",
		Status: base.STATUS_DEPLOYING,
	}
	app2 := metrics.AppInfo{
		Type: base.APP_WORKER,
		Name: "workerApp_1",
		Version: "worker_1.0",
		Status: base.STATUS_RUNNING,
	}

	info := metrics.HostInfo{
		HostId: "host1", Apps: []metrics.AppInfo{app1, app2, app2},
	}

	stats := metrics.StatsWrapper{}

	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 0 {
		t.Error(state_cloud.GlobalCloudLayout.Current)
	}

	planner.Queue.Add("host1", "httpApp_1", state_cloud.AppsVersion{"http_1.1", 1})
	planner.Queue.Add("host2", "httpApp_1", state_cloud.AppsVersion{"http_1.1", 1})
	planner.Queue.SetState("host1", "httpApp_1", planner.STATE_APPLYING)
	s, _ := planner.Queue.GetState("host1", "httpApp_1")
	if s != planner.STATE_APPLYING {
		t.Error(s)
	}

	doHandlePush(info, stats)

	conf, err := responder.GetConfigForHost("host1")
	if err != nil {
		t.Error(conf)
	}

	//app is not rated yet:
	tA := tracker.GlobalAppsStatusTracker["httpApp_1"]["http_1.1"]
	if tA.RunningCount != 0 || tA.Rating != "" {
		t.Error(tA)
	}

	layout := state_cloud.GlobalCloudLayout.Current.Layout["host1"].Apps
	//current layout for host1 should show no httpApp_1, it is not running!
	if layout["workerApp_1"].Version != "worker_1.0" || layout["httpApp_1"].Version != "" {
		t.Error(layout)
	}
	//check that other queued updates still exist:
	s2, err2 := planner.Queue.Get("host2")
	if err2 != nil {
		t.Error(s2)
	}
}


func TestApi_doHandlePush_AppUpdate(t *testing.T) {
	// An app updated successfully.
	// check that the CurrentLayout is updated
	// check that the app tracker is updated


	prepare()
	defer after()
	app1 := metrics.AppInfo{
		Type: base.APP_HTTP,
		Name: "httpApp_1",
		Version: "http_1.1",
		Status: base.STATUS_RUNNING,
	}
	app2 := metrics.AppInfo{
		Type: base.APP_WORKER,
		Name: "workerApp_1",
		Version: "worker_1.0",
		Status: base.STATUS_RUNNING,
	}

	info := metrics.HostInfo{
		HostId: "host1", Apps: []metrics.AppInfo{app1, app2, app2},
	}

	stats := metrics.StatsWrapper{}

	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 0 {
		t.Error(state_cloud.GlobalCloudLayout.Current)
	}

	planner.Queue.Add("host1", "httpApp_1", state_cloud.AppsVersion{"http_1.1", 1})
	planner.Queue.Add("host2", "httpApp_1", state_cloud.AppsVersion{"http_1.1", 1})
	planner.Queue.SetState("host1", "httpApp_1", planner.STATE_APPLYING)
	s, _ := planner.Queue.GetState("host1", "httpApp_1")
	if s != planner.STATE_APPLYING {
		t.Error(s)
	}

	doHandlePush(info, stats)

	conf, err := responder.GetConfigForHost("host1")
	if err != nil {
		t.Error(conf)
	}

	tA := tracker.GlobalAppsStatusTracker["httpApp_1"]["http_1.1"]
	if tA.RunningCount != 1 || tA.Rating != tracker.RATING_STABLE {
		t.Error(tA)
	}

	layout := state_cloud.GlobalCloudLayout.Current.Layout["host1"].Apps
	if layout["workerApp_1"].Version != "worker_1.0" || layout["httpApp_1"].Version != "http_1.1" {
		t.Error(layout)
	}

	//check that other queued updates still exist:
	s2, err2 := planner.Queue.Get("host2")
	if err2 != nil {
		t.Error(s2)
	}
}


func TestApi_doHandlePush_AppRollback(t *testing.T) {
	// An app rolled back.
	// check that the CurrentLayout is NOT updated
	// check that the app tracker is updated
	// check that the planner queue is updated (all updates of the app should be removed)


	prepare()
	defer after()
	app1 := metrics.AppInfo{
		Type: base.APP_HTTP,
		Name: "httpApp_1",
		Version: "http_1.0",
		Status: base.STATUS_RUNNING,
	}
	app2 := metrics.AppInfo{
		Type: base.APP_WORKER,
		Name: "workerApp_1",
		Version: "worker_1.0",
		Status: base.STATUS_RUNNING,
	}

	info := metrics.HostInfo{
		HostId: "host1", Apps: []metrics.AppInfo{app1, app2, app2},
	}

	stats := metrics.StatsWrapper{}

	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 0 {
		t.Error(state_cloud.GlobalCloudLayout.Current)
	}

	planner.Queue.Add("host1", "httpApp_1", state_cloud.AppsVersion{"http_1.1", 1})
	planner.Queue.Add("host2", "httpApp_1", state_cloud.AppsVersion{"http_1.1", 1})
	planner.Queue.SetState("host1", "httpApp_1", planner.STATE_APPLYING)
	s, _ := planner.Queue.GetState("host1", "httpApp_1")
	if s != planner.STATE_APPLYING {
		t.Error(s)
	}

	doHandlePush(info, stats)

	conf, err := responder.GetConfigForHost("host1")
	if err == nil {
		t.Error(conf)
	}

	//new app version never checked in, handle as crash:
	tA := tracker.GlobalAppsStatusTracker["httpApp_1"]["http_1.1"]
	if tA.RunningCount != 0 || tA.Rating != tracker.RATING_CRASHED {
		t.Error(tA)
	}

	layout := state_cloud.GlobalCloudLayout.Current.Layout["host1"].Apps
	if layout["workerApp_1"].Version != "worker_1.0" || layout["httpApp_1"].Version != "http_1.0" {
		t.Error(layout)
	}

	//check that other queued updates were removed:
	s2, err2 := planner.Queue.Get("host2")
	if err2 == nil {
		t.Error(s2)
	}
}


func TestApi_doHandlePush_AppShouldBeRemovedFromHost(t *testing.T) {
	//check that the config contains the new deploymentCount 0

	prepare()
	defer after()
	app1 := metrics.AppInfo{
		Type: base.APP_HTTP,
		Name: "httpApp_1",
		Version: "http_1.1",
		Status: base.STATUS_RUNNING,
	}
	app2 := metrics.AppInfo{
		Type: base.APP_WORKER,
		Name: "workerApp_1",
		Version: "worker_1.0",
		Status: base.STATUS_RUNNING,
	}

	info := metrics.HostInfo{
		HostId: "host1", Apps: []metrics.AppInfo{app1, app2, app2},
	}

	stats := metrics.StatsWrapper{}

	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 0 {
		t.Error(state_cloud.GlobalCloudLayout.Current)
	}

	planner.Queue.Add("host1", "httpApp_1", state_cloud.AppsVersion{"http_1.1", 0})
	planner.Queue.Add("host2", "httpApp_1", state_cloud.AppsVersion{"http_1.1", 1})
	planner.Queue.SetState("host1", "httpApp_1", planner.STATE_APPLYING)
	s, _ := planner.Queue.GetState("host1", "httpApp_1")
	if s != planner.STATE_APPLYING {
		t.Error(s)
	}

	doHandlePush(info, stats)

	conf, err := responder.GetConfigForHost("host1")
	if err != nil {
		t.Error(conf)
	}

	if conf.DeploymentCount != 0 {
		t.Error(conf)
	}

	tA := tracker.GlobalAppsStatusTracker["httpApp_1"]["http_1.1"]
	if tA.RunningCount != 1 || tA.Rating != tracker.RATING_STABLE {
		t.Error(tA)
	}

	layout := state_cloud.GlobalCloudLayout.Current.Layout["host1"].Apps
	if layout["workerApp_1"].Version != "worker_1.0" || layout["httpApp_1"].Version != "http_1.1" {
		t.Error(layout)
	}

	//check that other queued updates still exist:
	s2, err2 := planner.Queue.Get("host2")
	if err2 != nil {
		t.Error(s2)
	}
}

func TestApi_doHandlePush_AppShouldBeScaled(t *testing.T) {
	//check that the config contains the new deploymentCount
	prepare()
	defer after()
	app1 := metrics.AppInfo{
		Type: base.APP_HTTP,
		Name: "httpApp_1",
		Version: "http_1.1",
		Status: base.STATUS_RUNNING,
	}
	app2 := metrics.AppInfo{
		Type: base.APP_WORKER,
		Name: "workerApp_1",
		Version: "worker_1.0",
		Status: base.STATUS_RUNNING,
	}

	info := metrics.HostInfo{
		HostId: "host1", Apps: []metrics.AppInfo{app1, app2, app2},
	}

	stats := metrics.StatsWrapper{}

	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 0 {
		t.Error(state_cloud.GlobalCloudLayout.Current)
	}

	planner.Queue.Add("host1", "httpApp_1", state_cloud.AppsVersion{"http_1.1", 5})
	planner.Queue.Add("host2", "httpApp_1", state_cloud.AppsVersion{"http_1.1", 1})
	planner.Queue.SetState("host1", "httpApp_1", planner.STATE_APPLYING)
	s, _ := planner.Queue.GetState("host1", "httpApp_1")
	if s != planner.STATE_APPLYING {
		t.Error(s)
	}

	doHandlePush(info, stats)

	conf, err := responder.GetConfigForHost("host1")
	if err != nil {
		t.Error(conf)
	}

	if conf.DeploymentCount != 5 {
		t.Error(conf)
	}

	tA := tracker.GlobalAppsStatusTracker["httpApp_1"]["http_1.1"]
	if tA.RunningCount != 1 || tA.Rating != tracker.RATING_STABLE {
		t.Error(tA)
	}

	layout := state_cloud.GlobalCloudLayout.Current.Layout["host1"].Apps
	if layout["workerApp_1"].Version != "worker_1.0" || layout["httpApp_1"].Version != "http_1.1" {
		t.Error(layout)
	}

	//check that other queued updates still exist:
	s2, err2 := planner.Queue.Get("host2")
	if err2 != nil {
		t.Error(s2)
	}
}


func TestApi_doHandlePush_NewInstance(t *testing.T) {
	// This is the first time the new host checks in.
	// Check that the CurrentLayout is updated.
	// Check that the HostTracker is updated.

	prepare()
	defer after()
	state_cloud.GlobalCloudLayout.Current.AddEmptyHost("somehost")
	info := metrics.HostInfo{
		HostId: "new_host", Apps: []metrics.AppInfo{},
	}

	stats := metrics.StatsWrapper{}

	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 1 {
		t.Error(state_cloud.GlobalCloudLayout.Current)
	}

	doHandlePush(info, stats)

	conf, err := responder.GetConfigForHost("host1")
	if err == nil {
		t.Error(conf)
	}

	layout := state_cloud.GlobalCloudLayout.Current.Layout["new_host"].Apps
	if len(layout) != 0 {
		t.Error(layout)
	}
	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 2 {
		t.Error(state_cloud.GlobalCloudLayout.Current.Layout)
	}
}


func TestApi_doHandlePush_AppDied(t *testing.T) {
	// An app died.
	// check that the CurrentLayout is updated
	// check that the app tracker is updated


	prepare()
	defer after()
	app1 := metrics.AppInfo{
		Type: base.APP_HTTP,
		Name: "httpApp_1",
		Version: "http_1.0",
		Status: base.STATUS_RUNNING,
	}
	app2 := metrics.AppInfo{
		Type: base.APP_WORKER,
		Name: "workerApp_2",
		Version: "worker_2.0",
		Status: base.STATUS_DEAD,
	}

	info := metrics.HostInfo{
		HostId: "host1", Apps: []metrics.AppInfo{app1, app2, app2},
	}

	stats := metrics.StatsWrapper{}

	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 0 {
		t.Error(state_cloud.GlobalCloudLayout.Current)
	}

	doHandlePush(info, stats)

	_, errC := responder.GetConfigForHost("host1")

	elem, _ := state_cloud.GlobalCloudLayout.Current.GetHost("host1")

	if elem.HostId != "host1" || elem.Apps["httpApp_1"].DeploymentCount != 1 {
		t.Error(elem)
	}

	if _, exists := elem.Apps["workerApp_2"]; exists {
		t.Error(elem.Apps["workerApp_2"] )
	}

	if errC == nil {
		t.Error(errC)
	}

	info2 := metrics.HostInfo{
		HostId: "host1", Apps: []metrics.AppInfo{app1, app2, app2, app2},
	}
	doHandlePush(info2, stats)

	track, _ := tracker.GlobalHostTracker.Get("host1")
	if !track.LastCheckin.After(time.Now().UTC().Add(-time.Duration(time.Second * 2))) {
		t.Error(track)
	}

	appTrack := tracker.GlobalAppsStatusTracker["httpApp_1"]["http_1.0"]
	if appTrack.Rating != tracker.RATING_STABLE || appTrack.RunningCount != 2 || len(appTrack.CrashDetails) != 0{
		t.Errorf("%+v", appTrack)
	}

	appTrackWorker := tracker.GlobalAppsStatusTracker["workerApp_2"]["worker_2.0"]
	if appTrackWorker.Rating != tracker.RATING_CRASHED || appTrackWorker.CrashDetails[0].Cause != "APP_EVENT_CRASH" {
		t.Errorf("%+v", appTrackWorker)
	}

	_, errC2 := responder.GetConfigForHost("host1")

	if errC2 == nil {
		t.Error(errC2)
	}
}

func TestApi_doHandleCloudEvent_HostSpawned(t *testing.T) {
	// Host spawned.   TODO think about other stuff to check
	// check that the CurrentLayout is updated
	// check that the app tracker is updated
}

func TestApi_doHandleCloudEvent_HostDied(t *testing.T) {
	// Host died.   TODO think about other stuff to check
	// check that the CurrentLayout is updated
	// check that the app tracker is updated
}
