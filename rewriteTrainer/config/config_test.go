package config

import (
	"testing"
	"gatoor/orca/rewriteTrainer/state/configuration"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/rewriteTrainer/state/needs"
	"gatoor/orca/rewriteTrainer/base"
)


func TestConfig_ApplyToState(t *testing.T) {
	state_configuration.GlobalConfigurationState.Init()
	state_cloud.GlobalCloudLayout.Init()

	if len(state_configuration.GlobalConfigurationState.Apps) != 0 {
		t.Error("init state_config apps should be empty")
	}
	if len(state_configuration.GlobalConfigurationState.Habitats) != 0 {
		t.Error("init state_config habitats should be empty")
	}

	if len(state_cloud.GlobalCloudLayout.Current.Layout) != 0 {
		t.Error("init state_cloud current should be empty")
	}
	if len(state_cloud.GlobalCloudLayout.Desired.Layout) != 0 {
		t.Error("init state_cloud desired should be empty")
	}

	if len(state_needs.GlobalAppsNeedState) != 0 {
		t.Error("init state_needs should be empty")
	}

	conf := JsonConfiguration{}

	conf.Trainer.Port = 5000

	conf.Habitats = []HabitatJsonConfiguration{
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

	conf.Apps = []AppJsonConfiguration{
		{
			Name: "app1",
			Version: "0.1",
			Type: base.APP_WORKER,
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
				MemoryNeeds: state_needs.MemoryNeeds(5),
				CpuNeeds: state_needs.CpuNeeds(5),
				NetworkNeeds: state_needs.NetworkNeeds(5),
			},
		},
		{
			Name: "app11",
			Version: "0.1",
			Type: base.APP_WORKER,
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
			Needs: state_needs.AppNeeds{
				MemoryNeeds: state_needs.MemoryNeeds(5),
				CpuNeeds: state_needs.CpuNeeds(5),
				NetworkNeeds: state_needs.NetworkNeeds(5),
			},
		},
		{
			Name: "app2",
			Version: "0.2",
			Type: base.APP_WORKER,
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
			Needs: state_needs.AppNeeds{
				MemoryNeeds: state_needs.MemoryNeeds(5),
				CpuNeeds: state_needs.CpuNeeds(5),
				NetworkNeeds: state_needs.NetworkNeeds(5),
			},
		},
	}

	conf.ApplyToState()


	if len(state_configuration.GlobalConfigurationState.Apps) != 3 {
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

	if len(state_needs.GlobalAppsNeedState) != 3 {
		t.Error("init state_needs wrong len")
	}

}