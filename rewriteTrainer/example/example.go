package example

import (
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/rewriteTrainer/config"
	"gatoor/orca/base"
	"gatoor/orca/rewriteTrainer/needs"
)

func ExampleCloudState() {
	state := state_cloud.CloudLayoutAll{}
	state.Init()
	state.Current.AddEmptyHost("host1")
	state.Current.AddEmptyHost("host2")
	state.Current.AddEmptyHost("host3")
	state.Current.AddApp("host1", "app1", "0.1", 1)
	state.Current.AddApp("host1", "app11", "0.1", 2)
	state.Current.AddApp("host2", "app2", "0.2", 10)

	state.Desired.AddEmptyHost("host1")
	state.Desired.AddEmptyHost("host2")
	state.Desired.AddEmptyHost("host3")
	state.Desired.AddApp("host1", "app1", "0.1", 1)
	state.Desired.AddApp("host1", "app11", "0.1", 2)
	state.Desired.AddApp("host2", "app2", "0.2", 5)
	state.Desired.AddApp("host3", "app2", "0.2", 5)
}

func ExampleJsonConfig() config.JsonConfiguration {
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

	conf.Apps = []config.AppJsonConfiguration{
		{
			Name: "http1",
			Version: "0.1",
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
					Command: base.Command{"/server/http1/app1.conf", "somefilecontent as a string"},
				},
			},
			QueryStateCommand: base.OsCommand{
				Type: base.EXEC_COMMAND,
				Command: base.Command{"wget", "http://localhost:1234/check"},
			},
			RemoveCommand: base.OsCommand{
				Type: base.EXEC_COMMAND,
				Command: base.Command{"rm", "-rf /server/http1"},
			},
			Needs: needs.AppNeeds{
				MemoryNeeds: needs.MemoryNeeds(5),
				CpuNeeds: needs.CpuNeeds(5),
				NetworkNeeds: needs.NetworkNeeds(5),
			},
		},{
			Name: "app1",
			Version: "0.1",
			Type: base.APP_WORKER,
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
				MemoryNeeds: needs.MemoryNeeds(5),
				CpuNeeds: needs.CpuNeeds(5),
				NetworkNeeds: needs.NetworkNeeds(5),
			},
		},
		{
			Name: "app11",
			Version: "0.1",
			Type: base.APP_WORKER,
			MinDeploymentCount: 2,
			MaxDeploymentCount: 10,
			InstallCommands: []base.OsCommand{
				{
					Type: base.EXEC_COMMAND,
					Command: base.Command{"ls", "/home"},
				},
				{
					Type: base.FILE_COMMAND,
					Command: base.Command{"/server/app11/app11.conf", "somefilecontent as a string"},
				},
			},
			QueryStateCommand: base.OsCommand{
				Type: base.EXEC_COMMAND,
				Command: base.Command{"wget", "http://localhost:1235/check"},
			},
			RemoveCommand: base.OsCommand{
				Type: base.EXEC_COMMAND,
				Command: base.Command{"rm", "-rf /server/app11"},
			},
			Needs: needs.AppNeeds{
				MemoryNeeds: needs.MemoryNeeds(5),
				CpuNeeds: needs.CpuNeeds(5),
				NetworkNeeds: needs.NetworkNeeds(5),
			},
		},
		{
			Name: "app2",
			Version: "0.2",
			Type: base.APP_WORKER,
			MinDeploymentCount: 2,
			MaxDeploymentCount: 10,
			InstallCommands: []base.OsCommand{
				{
					Type: base.EXEC_COMMAND,
					Command: base.Command{"ls", "/home"},
				},
				{
					Type: base.FILE_COMMAND,
					Command: base.Command{"/server/app2/app2.conf", "somefilecontent as a string"},
				},
			},
			QueryStateCommand: base.OsCommand{
				Type: base.EXEC_COMMAND,
				Command: base.Command{"wget", "http://localhost:1236/check"},
			},
			RemoveCommand: base.OsCommand{
				Type: base.EXEC_COMMAND,
				Command: base.Command{"rm", "-rf /server/app2"},
			},
			Needs: needs.AppNeeds{
				MemoryNeeds: needs.MemoryNeeds(5),
				CpuNeeds: needs.CpuNeeds(5),
				NetworkNeeds: needs.NetworkNeeds(5),
			},
		},
	}
	return conf
}

