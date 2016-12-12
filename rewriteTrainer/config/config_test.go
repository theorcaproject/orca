package config

import (
	"testing"
	"gatoor/orca/rewriteTrainer/state/configuration"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/rewriteTrainer/state/needs"
	"gatoor/orca/base"
	"gatoor/orca/rewriteTrainer/needs"
	"os"
	"gatoor/orca/rewriteTrainer/cloud"
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


	conf.Apps = []base.AppConfiguration{
		{
			Name: "app1",
			Version: 1,
			Type: base.APP_WORKER,
			//InstallCommands: []base.OsCommand{
			//	{
			//		Type: base.EXEC_COMMAND,
			//		Command: base.Command{"ls", "/home"},
			//	},
			//	{
			//		Type: base.FILE_COMMAND,
			//		Command: base.Command{"/server/app1/app1.conf", "somefilecontent as a string"},
			//	},
			//},
			//QueryStateCommand: base.OsCommand{
			//	Type: base.EXEC_COMMAND,
			//	Command: base.Command{"wget", "http://localhost:1234/check"},
			//},
			//RemoveCommand: base.OsCommand{
			//	Type: base.EXEC_COMMAND,
			//	Command: base.Command{"rm", "-rf /server/app1"},
			//},
			Needs: needs.AppNeeds{
				MemoryNeeds: needs.MemoryNeeds(5),
				CpuNeeds: needs.CpuNeeds(5),
				NetworkNeeds: needs.NetworkNeeds(5),
			},
		},
		{
			Name: "app11",
			Version: 1,
			Type: base.APP_WORKER,
			//InstallCommands: []base.OsCommand{
			//	{
			//		Type: base.EXEC_COMMAND,
			//		Command: base.Command{"ls", "/home"},
			//	},
			//	{
			//		Type: base.FILE_COMMAND,
			//		Command: base.Command{"/server/app11/app11.conf", "somefilecontent as a string"},
			//	},
			//},
			//QueryStateCommand: base.OsCommand{
			//	Type: base.EXEC_COMMAND,
			//	Command: base.Command{"wget", "http://localhost:1235/check"},
			//},
			//RemoveCommand: base.OsCommand{
			//	Type: base.EXEC_COMMAND,
			//	Command: base.Command{"rm", "-rf /server/app11"},
			//},
			Needs: needs.AppNeeds{
				MemoryNeeds: needs.MemoryNeeds(5),
				CpuNeeds: needs.CpuNeeds(5),
				NetworkNeeds: needs.NetworkNeeds(5),
			},
		},
		{
			Name: "app2",
			Version: 2,
			Type: base.APP_WORKER,
			//InstallCommands: []base.OsCommand{
			//	{
			//		Type: base.EXEC_COMMAND,
			//		Command: base.Command{"ls", "/home"},
			//	},
			//	{
			//		Type: base.FILE_COMMAND,
			//		Command: base.Command{"/server/app2/app2.conf", "somefilecontent as a string"},
			//	},
			//},
			//QueryStateCommand: base.OsCommand{
			//	Type: base.EXEC_COMMAND,
			//	Command: base.Command{"wget", "http://localhost:1236/check"},
			//},
			//RemoveCommand: base.OsCommand{
			//	Type: base.EXEC_COMMAND,
			//	Command: base.Command{"rm", "-rf /server/app2"},
			//},
			Needs: needs.AppNeeds{
				MemoryNeeds: needs.MemoryNeeds(5),
				CpuNeeds: needs.CpuNeeds(5),
				NetworkNeeds: needs.NetworkNeeds(5),
			},
		},
	}

	conf.ApplyToState()


	if len(state_configuration.GlobalConfigurationState.Apps) != 3 {
		t.Error("init state_config apps wrong len")
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


func Test_applyAwsConfiguration(t *testing.T) {
	file, err := os.Open("/orca/config/trainer/aws_cloud_provider.json")
	if err != nil {
		t.Fatal(err)
	}
	j := JsonConfiguration{}
	loadConfigFromFile(file, &j.CloudProvider)
	file.Close()
	applyCloudProviderConfiguration(j.CloudProvider)
	cloud.Init()

	if state_configuration.GlobalConfigurationState.CloudProvider.Type != "AWS" {
		t.Error(state_configuration.GlobalConfigurationState.CloudProvider)
	}

	awsProvider := cloud.CurrentProvider.(*cloud.AWSProvider)
	if awsProvider.Type != "AWS" {
		t.Error(cloud.CurrentProvider)
	}

	if state_configuration.GlobalConfigurationState.CloudProvider.AWSConfiguration.InstanceCost["t2.nano"] != 65 {
		t.Error(state_configuration.GlobalConfigurationState.CloudProvider)
	}
}