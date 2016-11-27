package responder

import (
	"gatoor/orca/rewriteTrainer/state/cloud"
	"testing"
	"gatoor/orca/rewriteTrainer/planner"
	"gatoor/orca/rewriteTrainer/state/configuration"
	"gatoor/orca/rewriteTrainer/state/needs"
	"gatoor/orca/base"
	"gatoor/orca/rewriteTrainer/cloud"
	"gatoor/orca/rewriteTrainer/config"
	"gatoor/orca/rewriteTrainer/needs"
)

func initTrainer() {
	state_configuration.GlobalConfigurationState.Init()
	state_cloud.GlobalCloudLayout.Init()
	state_needs.GlobalAppsNeedState = state_needs.AppsNeedState{}
	applySampleConfig()
	initCloudProvider()
	planner.Queue = *planner.NewPlannerQueue()
	cloud.Init()
}

func applySampleConfig() {
	conf := config.JsonConfiguration{}

	conf.Trainer.Port = 5000


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
		Needs: needs.AppNeeds{
			MemoryNeeds: needs.MemoryNeeds(1),
			CpuNeeds: needs.CpuNeeds(1),
			NetworkNeeds: needs.NetworkNeeds(1),
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
		Needs: needs.AppNeeds{
			MemoryNeeds: needs.MemoryNeeds(2),
			CpuNeeds: needs.CpuNeeds(2),
			NetworkNeeds: needs.NetworkNeeds(5),
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
		Needs: needs.AppNeeds{
			MemoryNeeds: needs.MemoryNeeds(1),
			CpuNeeds: needs.CpuNeeds(1),
			NetworkNeeds: needs.NetworkNeeds(1),
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
		Needs: needs.AppNeeds{
			CpuNeeds: needs.CpuNeeds(50),
			MemoryNeeds: needs.MemoryNeeds(10),
			NetworkNeeds: needs.NetworkNeeds(10),
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
		Needs: needs.AppNeeds{
			CpuNeeds: needs.CpuNeeds(70),
			MemoryNeeds: needs.MemoryNeeds(40),
			NetworkNeeds: needs.NetworkNeeds(30),
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
		Needs: needs.AppNeeds{
			CpuNeeds: needs.CpuNeeds(23),
			MemoryNeeds: needs.MemoryNeeds(23),
			NetworkNeeds: needs.NetworkNeeds(23),
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
		Needs: needs.AppNeeds{
			CpuNeeds: needs.CpuNeeds(7),
			MemoryNeeds: needs.MemoryNeeds(2),
			NetworkNeeds: needs.NetworkNeeds(1),
		},
	}



	conf.Apps = []config.AppJsonConfiguration{
		httpApp1, httpApp1_v2, httpApp2,
		workerApp1, workerApp1_v2, workerApp2, workerApp3,
	}

	conf.ApplyToState()
}

func initCloudProvider() {
	state_cloud.GlobalAvailableInstances.Update("cpuHost_1", state_cloud.InstanceResources{
		TotalCpuResource: 500,
		TotalMemoryResource: 100,
		TotalNetworkResource: 100,
	})
	state_cloud.GlobalAvailableInstances.Update("cpuHost_2", state_cloud.InstanceResources{
		TotalCpuResource: 501,
		TotalMemoryResource: 101,
		TotalNetworkResource: 101,
	})
	state_cloud.GlobalAvailableInstances.Update("memoryHost_1", state_cloud.InstanceResources{
		TotalCpuResource: 200,
		TotalMemoryResource: 300,
		TotalNetworkResource: 100,
	})
	state_cloud.GlobalAvailableInstances.Update("generalHost_1", state_cloud.InstanceResources{
		TotalCpuResource: 101,
		TotalMemoryResource: 101,
		TotalNetworkResource: 101,
	})
	state_cloud.GlobalAvailableInstances.Update("generalHost_2", state_cloud.InstanceResources{
		TotalCpuResource: 102,
		TotalMemoryResource: 102,
		TotalNetworkResource: 102,
	})
	state_cloud.GlobalAvailableInstances.Update("generalHost_3", state_cloud.InstanceResources{
		TotalCpuResource: 103,
		TotalMemoryResource: 103,
		TotalNetworkResource: 103,
	})
	state_cloud.GlobalAvailableInstances.Update("emptyHost", state_cloud.InstanceResources{
		TotalCpuResource: 1000,
		TotalMemoryResource: 1000,
		TotalNetworkResource: 1000,
	})
}

func testInitConfig(t *testing.T) {
	if len(state_configuration.GlobalConfigurationState.Apps) != 5 {
		t.Error("init state_config apps wrong len")
	}
	if len(state_configuration.GlobalConfigurationState.Apps["httpApp_1"]) != 2 {
		t.Error("init state_config apps wrong len")
	}
	if len(state_configuration.GlobalConfigurationState.Apps["httpApp_2"]) != 1 {
		t.Error("init state_config apps wrong len")
	}
	if len(state_configuration.GlobalConfigurationState.Apps["workerApp_1"]) != 2 {
		t.Error("init state_config apps wrong len")
	}
	if len(state_configuration.GlobalConfigurationState.Apps["workerApp_3"]) != 1 {
		t.Error("init state_config apps wrong len")
	}
	if len(state_configuration.GlobalConfigurationState.Habitats) != 2 {
		t.Error("init state_config habitats wrong len")
	}

	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 0 {
		t.Error("init state_cloud current should be empty")
	}
	if len(state_cloud.GlobalCloudLayout.Desired.Layout) != 0 {
		t.Error("init state_cloud desired should be empty")
	}

	if len(state_needs.GlobalAppsNeedState) != 5 {
		t.Error("init state_needs wrong len")
	}
	elem , _ := state_needs.GlobalAppsNeedState.Get("httpApp_1", "http_1.1")
	if elem.CpuNeeds != 2 {
		t.Error("wrong needs")
	}
	elem2 , _ := state_needs.GlobalAppsNeedState.Get("workerApp_3", "worker_3.0")
	if elem2.MemoryNeeds != 2 {
		t.Error("wrong needs")
	}
}

func testLayout(t *testing.T) {
	if len(state_cloud.GlobalAvailableInstances) != 7 {
		t.Error(state_cloud.GlobalAvailableInstances)
	}
}

func testInit(t * testing.T) {
	testInitConfig(t)
	testLayout(t)
}













func TestResponder_GetConfigForHost_Integration(t *testing.T) {
	initTrainer()
	testInit(t)

	state_cloud.GlobalCloudLayout.Current.AddEmptyHost("cpuHost_1")
	state_cloud.GlobalCloudLayout.Current.AddEmptyHost("cpuHost_2")
	state_cloud.GlobalCloudLayout.Current.AddEmptyHost("memoryHost_1")
	state_cloud.GlobalCloudLayout.Current.AddEmptyHost("generalHost_1")
	state_cloud.GlobalCloudLayout.Current.AddEmptyHost("generalHost_2")
	state_cloud.GlobalCloudLayout.Current.AddEmptyHost("generalHost_3")
	state_cloud.GlobalCloudLayout.Current.AddEmptyHost("emptyHost")

	//httpApp_1 already deployed MinDeploymentCount
	state_cloud.GlobalCloudLayout.Current.AddApp("cpuHost_1", "httpApp_1", "http_1.0", 1)
	state_cloud.GlobalCloudLayout.Current.AddApp("cpuHost_2", "httpApp_1", "http_1.1", 1)

	//httpApp_2 missing 1 app for MinDeploymentCount
	state_cloud.GlobalCloudLayout.Current.AddApp("generalHost_1", "httpApp_2", "http_2.0", 1)
	state_cloud.GlobalCloudLayout.Current.AddApp("generalHost_2", "httpApp_2", "http_2.0", 1)
	state_cloud.GlobalCloudLayout.Current.AddApp("generalHost_3", "httpApp_2", "http_2.0", 1)

	//workerApp_1 missing

	//workerApp_2 already deployed MinDeploymentCount all on one host
	state_cloud.GlobalCloudLayout.Current.AddApp("cpuHost_1", "workerApp_2", "worker_2.0", 5)

	//workerApp_3 missing 90 for MinDeploymentCount
	state_cloud.GlobalCloudLayout.Current.AddApp("generalHost_2", "workerApp_3", "worker_3.0", 5)
	state_cloud.GlobalCloudLayout.Current.AddApp("generalHost_3", "workerApp_3", "worker_3.0", 5)


	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 7 {
		t.Error(state_cloud.GlobalCloudLayout.Current)
	}
	if len(state_cloud.GlobalCloudLayout.Desired.Layout) != 0 {
		t.Error(state_cloud.GlobalCloudLayout.Desired)
	}

	planner.Plan()

	if len(state_cloud.GlobalCloudLayout.Current.Layout) == 0 {
		t.Error(state_cloud.GlobalCloudLayout.Current)
	}
	if len(state_cloud.GlobalCloudLayout.Desired.Layout) != 7 {
		t.Error(state_cloud.GlobalCloudLayout.Desired)
	}

	//check httpApp_1 update
	host, _ := state_cloud.GlobalCloudLayout.Desired.GetHost("cpuHost_1")
	if  host.Apps["httpApp_1"].Version != "http_1.1" ||  host.Apps["httpApp_1"].DeploymentCount != 1 {
		t.Error(state_cloud.GlobalCloudLayout.Desired)
	}

	diff := planner.Diff(state_cloud.GlobalCloudLayout.Desired, state_cloud.GlobalCloudLayout.Current)

	if diff["cpuHost_1"]["httpApp_1"].Version != "http_1.1" || diff["cpuHost_1"]["httpApp_1"].DeploymentCount != 1 {
		t.Error(diff["cpuHost_1"])
	}
	if len(diff["emptyHost"]) != 0{
		t.Error(diff["emptyHost"])
	}
	if _, exists := diff["generalHost_1"]["httpApp_2"]; exists {
		t.Error(diff["gernerHost_1"])
	}
	if diff["generalHost_3"]["workerApp_3"].DeploymentCount != 14 {
		t.Error(diff["generalHost_3"]["workerApp_3"])
	}

	instances := state_cloud.GlobalAvailableInstances

	cpuHost_1, _ := instances.GetResources("cpuHost_1")
	if cpuHost_1.TotalCpuResource != 500 || cpuHost_1.UsedCpuResource != 94 {
		t.Error(cpuHost_1)
	}
	cpuHost_2, _ := instances.GetResources("cpuHost_2")
	if cpuHost_2.TotalCpuResource != 501 || cpuHost_2.UsedCpuResource != 310 || cpuHost_2.UsedMemoryResource != 90 {
		t.Error(cpuHost_2)
	}
	memoryHost_1, _ := instances.GetResources("memoryHost_1")
	if memoryHost_1.TotalCpuResource != 200 || memoryHost_1.UsedCpuResource != 197 || memoryHost_1.UsedMemoryResource != 57 {
		t.Error(memoryHost_1)
	}
	generalHost_1, _ := instances.GetResources("generalHost_1")
	if generalHost_1.TotalCpuResource != 101 || generalHost_1.UsedCpuResource != 94 || generalHost_1.UsedMemoryResource != 64 {
		t.Error(generalHost_1)
	}
	generalHost_2, _ := instances.GetResources("generalHost_2")
	if generalHost_2.TotalCpuResource != 102 || generalHost_2.UsedCpuResource != 99 || generalHost_2.UsedMemoryResource != 29 {
		t.Error(generalHost_2)
	}
	emptyHost, _ := instances.GetResources("emptyHost")
	if emptyHost.TotalCpuResource != 1000 || emptyHost.UsedCpuResource != 0 || emptyHost.UsedMemoryResource != 0 || emptyHost.UsedNetworkResource != 0 {
		t.Error(emptyHost)
	}


	// planning worked, now test the responder
	planner.Queue.Apply(diff)


	cpuHost_1Config, _ := GetConfigForHost("cpuHost_1")
	if cpuHost_1Config.DeploymentCount != 1 {
		t.Errorf("%+v", cpuHost_1Config)
	}

}