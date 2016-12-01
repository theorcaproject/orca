package main

import (
	"gatoor/orca/rewriteTrainer/state/needs"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/rewriteTrainer/state/configuration"
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/rewriteTrainer/config"
	"gatoor/orca/rewriteTrainer/api"
	"gatoor/orca/rewriteTrainer/cloud"
	"gatoor/orca/rewriteTrainer/db"
	"gatoor/orca/rewriteTrainer/scheduler"
	"gatoor/orca/rewriteTrainer/planner"
	"time"
)


const CHECKIN_WAIT_TIME = 5

var (
	TRAINER_CONFIGURATION_FILE = "/orca/config/trainer/trainer.json"
	APPS_CONFIGURATION_FILE = "/orca/config/trainer/apps.json"
	AVAILABLE_INSTANCES_CONFIGURATION_FILE = "/orca/config/trainer/available_instances.json"
	CLOUD_PROVIDER_CONFIGURATION_FILE = "/orca/config/trainer/cloud_provider.json"
)

func main() {
	Logger.InitLogger.Info("Starting trainer...")
	initState()
	initConfig()
	cloud.Init()
	db.Init("")
	initApi()
	waitForCheckin()
	scheduler.Start()
	planner.InitialPlan()
	Logger.InitLogger.Info("Trainer started")
	ticker := time.NewTicker(time.Second * 60)
	for {
		<- ticker.C
	}
}

func waitForCheckin() {
	Logger.InitLogger.Infof("Waiting %ds for existsing clients to check in", CHECKIN_WAIT_TIME)
	time.Sleep(time.Duration(CHECKIN_WAIT_TIME * time.Second))
	Logger.InitLogger.Info("Done waiting")
}

func initState() {
	state_configuration.GlobalConfigurationState.Init()
	state_cloud.GlobalCloudLayout.Init()
	state_needs.GlobalAppsNeedState = state_needs.AppsNeedState{}
}

func initConfig() {
	var baseConfiguration config.JsonConfiguration
	baseConfiguration.Load(TRAINER_CONFIGURATION_FILE, APPS_CONFIGURATION_FILE, AVAILABLE_INSTANCES_CONFIGURATION_FILE, CLOUD_PROVIDER_CONFIGURATION_FILE)
	baseConfiguration.Check()
	baseConfiguration.ApplyToState()
}


func initApi() {
	var a api.Api
	a.Init()
}


//func main() {
//	//go http.ListenAndServe(":8080", http.DefaultServeMux)
//	Logger.SetLogLevel(logrus.ErrorLevel)
//	//state_configuration.GlobalConfigurationState.Init()
//	//state_cloud.GlobalCloudLayout.Init()
//	//state_needs.GlobalAppsNeedState = state_needs.AppsNeedState{}
//	//applySampleConfig()
//	//initCloudProvider()
//
//	example.AwsJsonConfig()
//	//appCount := 0
//	//
//	//for i:= 1; i <= 50000; i++ {
//	//	state_cloud.GlobalAvailableInstances.Update(base.HostId("filllerHost_" + fmt.Sprint(i)), state_cloud.InstanceResources{
//	//		TotalCpuResource: 50,
//	//		TotalMemoryResource: 10,
//	//		TotalNetworkResource: 10,
//	//	})
//	//}
//	//
//	//for i:= 1; i <= 10; i++ {
//	//	c := 5000//rand.Intn(5000)
//	//
//	//	appCount += c
//	//	state_configuration.GlobalConfigurationState.ConfigureApp(state_configuration.AppConfiguration{
//	//		Name: base.AppName("fillerHttp_" + fmt.Sprint(i)),
//	//		Type: base.APP_HTTP,
//	//		Version: "1.0",
//	//		MinDeploymentCount: base.DeploymentCount(c),
//	//	})
//	//	state_needs.GlobalAppsNeedState.UpdateNeeds(base.AppName("fillerHttp_" + fmt.Sprint(i)), "1.0", needs.AppNeeds{
//	//		CpuNeeds: needs.CpuNeeds(rand.Intn(10) + 1), MemoryNeeds: needs.MemoryNeeds(rand.Intn(10) + 1), NetworkNeeds: needs.NetworkNeeds(rand.Intn(10) + 1),
//	//	})
//	//}
//	//for i:= 1; i <= 0; i++ {
//	//	d := rand.Intn(10000)
//	//	appCount += d
//	//	state_configuration.GlobalConfigurationState.ConfigureApp(state_configuration.AppConfiguration{
//	//		Name: base.AppName("fillerWorker_" + fmt.Sprint(i)),
//	//		Type: base.APP_WORKER,
//	//		Version: "1.0",
//	//		MinDeploymentCount: base.DeploymentCount(d),
//	//	})
//	//	state_needs.GlobalAppsNeedState.UpdateNeeds(base.AppName("fillerWorker_" + fmt.Sprint(i)), "1.0", needs.AppNeeds{
//	//		CpuNeeds: needs.CpuNeeds(rand.Intn(10) + 1), MemoryNeeds: needs.MemoryNeeds(rand.Intn(10) + 1), NetworkNeeds: needs.NetworkNeeds(rand.Intn(10) + 1),
//	//	})
//	//}
//	//
//	//targetCpu := 0
//	//targetMem := 0
//	//targetNet := 0
//	//for appname, app := range state_needs.GlobalAppsNeedState {
//	//	for vername, ver := range app{
//	//		appConf, _ := state_configuration.GlobalConfigurationState.GetApp(appname, vername)
//	//		targetCpu += int(appConf.MinDeploymentCount) * int(ver.CpuNeeds)
//	//		targetMem += int(appConf.MinDeploymentCount) * int(ver.MemoryNeeds)
//	//		targetNet += int(appConf.MinDeploymentCount) * int(ver.NetworkNeeds)
//	//	}
//	//}
//	//
//	//start := time.Now()
//	//doStuff()
//	//elapsed := time.Since(start)
//	////time.Sleep(time.Second * 10)
//	//fmt.Printf("Planning took %s for %d apps totalIter %d", elapsed, appCount, planner.TotalIter)
//	////fmt.Println("")
//	////fmt.Println(state_cloud.TotalTime)
//	////fmt.Println("")
//	////fmt.Println(state_cloud.TotalTimeCalc)
//	//fmt.Println("")
//	//fmt.Printf("%+v", state_cloud.GlobalAvailableInstances.GlobalResourceConsumption())
//	//fmt.Println("")
//	//fmt.Printf("%d %d %d", targetCpu, targetMem, targetNet)
//	//fmt.Println("")
//	////fmt.Println(s)
//	////fmt.Println(planner.HttpOccupiedCache)
//	//fmt.Println("fail ")
//	//fmt.Println(len(planner.FailedAssigned))
//	//fmt.Println(planner.FailedAssigned)
//	//fmt.Println("missing")
//	//fmt.Println(planner.MissingAssigned)
//	//fmt.Println("")
//
//	//fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>")
//	//fmt.Println(state_cloud.GlobalCloudLayout.Desired.Layout)
//	//fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>")
//	//fmt.Println(state_cloud.GlobalAvailableInstances)
//	//fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>")
//}




func applySampleConfig() {
	conf := config.JsonConfiguration{}

	conf.Trainer.Port = 5000

	//conf.Habitats = []config.HabitatJsonConfiguration{
	//	{
	//		Name: "habitat1",
	//		Version: "0.1",
	//		InstallCommands: []base.OsCommand{
	//			{
	//				Type: base.EXEC_COMMAND,
	//				Command: base.Command{"ls", "/home"},
	//			},
	//			{
	//				Type: base.FILE_COMMAND,
	//				Command: base.Command{"/etc/orca.conf", "somefilecontent as a string"},
	//			},
	//		},
	//	},
	//	{
	//		Name: "habitat2",
	//		Version: "0.1",
	//		InstallCommands: []base.OsCommand{
	//			{
	//				Type: base.EXEC_COMMAND,
	//				Command: base.Command{"ps", "aux"},
	//			},
	//			{
	//				Type: base.FILE_COMMAND,
	//				Command: base.Command{"/etc/orca.conf", "different config"},
	//			},
	//		},
	//	},
	//}

	//httpApp1 := config.AppJsonConfiguration{
	//	Name: "httpApp_1",
	//	Version: "http_1.0",
	//	Type: base.APP_HTTP,
	//	MinDeploymentCount: 3,
	//	MaxDeploymentCount: 10,
	//	InstallCommands: []base.OsCommand{
	//		{
	//			Type: base.EXEC_COMMAND,
	//			Command: base.Command{"ls", "/home"},
	//		},
	//		{
	//			Type: base.FILE_COMMAND,
	//			Command: base.Command{"/server/app1/app1.conf", "somefilecontent as a string"},
	//		},
	//	},
	//	QueryStateCommand: base.OsCommand{
	//		Type: base.EXEC_COMMAND,
	//		Command: base.Command{"wget", "http://localhost:1234/check"},
	//	},
	//	RemoveCommand: base.OsCommand{
	//		Type: base.EXEC_COMMAND,
	//		Command: base.Command{"rm", "-rf /server/app1"},
	//	},
	//	Needs: needs.AppNeeds{
	//		MemoryNeeds: needs.MemoryNeeds(1),
	//		CpuNeeds: needs.CpuNeeds(1),
	//		NetworkNeeds: needs.NetworkNeeds(1),
	//	},
	//}
	//
	//httpApp1_v2 := config.AppJsonConfiguration{
	//	Name: "httpApp_1",
	//	Version: "http_1.1",
	//	Type: base.APP_HTTP,
	//	MinDeploymentCount: 2,
	//	MaxDeploymentCount: 10,
	//	InstallCommands: []base.OsCommand{
	//		{
	//			Type: base.EXEC_COMMAND,
	//			Command: base.Command{"ls", "/home"},
	//		},
	//		{
	//			Type: base.FILE_COMMAND,
	//			Command: base.Command{"/server/app1/app1.conf", "somefilecontent as a string"},
	//		},
	//	},
	//	QueryStateCommand: base.OsCommand{
	//		Type: base.EXEC_COMMAND,
	//		Command: base.Command{"wget", "http://localhost:1234/check"},
	//	},
	//	RemoveCommand: base.OsCommand{
	//		Type: base.EXEC_COMMAND,
	//		Command: base.Command{"rm", "-rf /server/app1"},
	//	},
	//	Needs: needs.AppNeeds{
	//		MemoryNeeds: needs.MemoryNeeds(2),
	//		CpuNeeds: needs.CpuNeeds(2),
	//		NetworkNeeds: needs.NetworkNeeds(5),
	//	},
	//}
	//
	//httpApp2 := config.AppJsonConfiguration{
	//	Name: "httpApp_2",
	//	Version: "http_2.0",
	//	Type: base.APP_HTTP,
	//	MinDeploymentCount: 4,
	//	MaxDeploymentCount: 10,
	//	InstallCommands: []base.OsCommand{
	//		{
	//			Type: base.EXEC_COMMAND,
	//			Command: base.Command{"ls", "/home"},
	//		},
	//		{
	//			Type: base.FILE_COMMAND,
	//			Command: base.Command{"/server/app1/app1.conf", "somefilecontent as a string"},
	//		},
	//	},
	//	QueryStateCommand: base.OsCommand{
	//		Type: base.EXEC_COMMAND,
	//		Command: base.Command{"wget", "http://localhost:1234/check"},
	//	},
	//	RemoveCommand: base.OsCommand{
	//		Type: base.EXEC_COMMAND,
	//		Command: base.Command{"rm", "-rf /server/app1"},
	//	},
	//	Needs: needs.AppNeeds{
	//		MemoryNeeds: needs.MemoryNeeds(1),
	//		CpuNeeds: needs.CpuNeeds(1),
	//		NetworkNeeds: needs.NetworkNeeds(1),
	//	},
	//}
	//
	//workerApp1 := config.AppJsonConfiguration{
	//	Name: "workerApp_1",
	//	Version: "worker_1.0",
	//	Type: base.APP_WORKER,
	//	MinDeploymentCount: 1,
	//	MaxDeploymentCount: 1,
	//	InstallCommands: []base.OsCommand{
	//		{
	//			Type: base.EXEC_COMMAND,
	//			Command: base.Command{"ls", "/home"},
	//		},
	//		{
	//			Type: base.FILE_COMMAND,
	//			Command: base.Command{"/server/app1/app1.conf", "somefilecontent as a string"},
	//		},
	//	},
	//	QueryStateCommand: base.OsCommand{
	//		Type: base.EXEC_COMMAND,
	//		Command: base.Command{"wget", "http://localhost:1234/check"},
	//	},
	//	RemoveCommand: base.OsCommand{
	//		Type: base.EXEC_COMMAND,
	//		Command: base.Command{"rm", "-rf /server/app1"},
	//	},
	//	Needs: needs.AppNeeds{
	//		CpuNeeds: needs.CpuNeeds(50),
	//		MemoryNeeds: needs.MemoryNeeds(10),
	//		NetworkNeeds: needs.NetworkNeeds(10),
	//	},
	//}
	//
	//workerApp1_v2 := config.AppJsonConfiguration{
	//	Name: "workerApp_1",
	//	Version: "worker_1.1",
	//	Type: base.APP_WORKER,
	//	MinDeploymentCount: 1,
	//	MaxDeploymentCount: 1,
	//	InstallCommands: []base.OsCommand{
	//		{
	//			Type: base.EXEC_COMMAND,
	//			Command: base.Command{"ls", "/home"},
	//		},
	//		{
	//			Type: base.FILE_COMMAND,
	//			Command: base.Command{"/server/app1/app1.conf", "somefilecontent as a string"},
	//		},
	//	},
	//	QueryStateCommand: base.OsCommand{
	//		Type: base.EXEC_COMMAND,
	//		Command: base.Command{"wget", "http://localhost:1234/check"},
	//	},
	//	RemoveCommand: base.OsCommand{
	//		Type: base.EXEC_COMMAND,
	//		Command: base.Command{"rm", "-rf /server/app1"},
	//	},
	//	Needs: needs.AppNeeds{
	//		CpuNeeds: needs.CpuNeeds(70),
	//		MemoryNeeds: needs.MemoryNeeds(40),
	//		NetworkNeeds: needs.NetworkNeeds(30),
	//	},
	//}
	//
	//workerApp2 := config.AppJsonConfiguration{
	//	Name: "workerApp_2",
	//	Version: "worker_2.0",
	//	Type: base.APP_WORKER,
	//	MinDeploymentCount: 5,
	//	MaxDeploymentCount: 10,
	//	InstallCommands: []base.OsCommand{
	//		{
	//			Type: base.EXEC_COMMAND,
	//			Command: base.Command{"ls", "/home"},
	//		},
	//		{
	//			Type: base.FILE_COMMAND,
	//			Command: base.Command{"/server/app1/app1.conf", "somefilecontent as a string"},
	//		},
	//	},
	//	QueryStateCommand: base.OsCommand{
	//		Type: base.EXEC_COMMAND,
	//		Command: base.Command{"wget", "http://localhost:1234/check"},
	//	},
	//	RemoveCommand: base.OsCommand{
	//		Type: base.EXEC_COMMAND,
	//		Command: base.Command{"rm", "-rf /server/app1"},
	//	},
	//	Needs: needs.AppNeeds{
	//		CpuNeeds: needs.CpuNeeds(23),
	//		MemoryNeeds: needs.MemoryNeeds(23),
	//		NetworkNeeds: needs.NetworkNeeds(23),
	//	},
	//}
	//
	//workerApp3 := config.AppJsonConfiguration{
	//	Name: "workerApp_3",
	//	Version: "worker_3.0",
	//	Type: base.APP_WORKER,
	//	MinDeploymentCount: 100,
	//	MaxDeploymentCount: 200,
	//	InstallCommands: []base.OsCommand{
	//		{
	//			Type: base.EXEC_COMMAND,
	//			Command: base.Command{"ls", "/home"},
	//		},
	//		{
	//			Type: base.FILE_COMMAND,
	//			Command: base.Command{"/server/app1/app1.conf", "somefilecontent as a string"},
	//		},
	//	},
	//	QueryStateCommand: base.OsCommand{
	//		Type: base.EXEC_COMMAND,
	//		Command: base.Command{"wget", "http://localhost:1234/check"},
	//	},
	//	RemoveCommand: base.OsCommand{
	//		Type: base.EXEC_COMMAND,
	//		Command: base.Command{"rm", "-rf /server/app1"},
	//	},
	//	Needs: needs.AppNeeds{
	//		CpuNeeds: needs.CpuNeeds(7),
	//		MemoryNeeds: needs.MemoryNeeds(2),
	//		NetworkNeeds: needs.NetworkNeeds(1),
	//	},
	//}
	//


	//conf.Apps = []config.AppJsonConfiguration{
	//	httpApp1, httpApp1_v2, httpApp2,
	//	workerApp1, workerApp1_v2, workerApp2, workerApp3,
	//}

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