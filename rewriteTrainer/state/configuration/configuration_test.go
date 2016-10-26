package state_configuration_test

import (
	"gatoor/orca/rewriteTrainer/state/configuration"
	"gatoor/orca/rewriteTrainer/base"
	"testing"
)

func prepareConfigState() state_configuration.ConfigurationState {
	var GlobalConfig state_configuration.ConfigurationState
	GlobalConfig.Init()
	return GlobalConfig
}

func TestConfigureApp(t *testing.T) {
	GlobalConfig := prepareConfigState()
	GlobalConfig.ConfigureApp(state_configuration.AppConfiguration{
		"appname", base.APP_HTTP, "0.1",
		[]base.OsCommand{
			{
				base.EXEC_COMMAND,
				base.Command{
					":aa", "mmm",
				},
			},
		},
		base.OsCommand{
			base.EXEC_COMMAND,
			base.Command{
				":bb", "uu",
			},
		},
		base.OsCommand{
			base.FILE_COMMAND,
			base.Command{
				":cc", "ii",
			},
		},
	})

	if _, err := GlobalConfig.GetApp("unknown", "1"); err == nil {
		t.Error()
	}
	if _, err := GlobalConfig.GetApp("appname", "0.2"); err == nil {
		t.Error()
	}
	if val, err :=  GlobalConfig.GetApp("appname", "0.1"); err == nil {
		if val.Name != "appname" {
			t.Error()
		}
		if val.InstallCommands[0].Type != base.EXEC_COMMAND {
			t.Error()
		}
		if val.RemoveCommand.Type != base.FILE_COMMAND {
			t.Error()
		}
	} else {
		t.Error(err)
	}
}


func TestConfigureHabitat(t *testing.T) {
	GlobalConfig := prepareConfigState()
	GlobalConfig.ConfigureHabitat(state_configuration.HabitatConfiguration{"habname", "0.1", []base.OsCommand{
		{
			base.EXEC_COMMAND,
			base.Command{
					":aa", "mmm",
				},
			},
		},
	})

	if _, err := GlobalConfig.GetHabitat("unknown", "1"); err == nil {
		t.Error()
	}
	if _, err := GlobalConfig.GetHabitat("habname", "0.2"); err == nil {
		t.Error()
	}
	if val, err :=  GlobalConfig.GetHabitat("habname", "0.1"); err == nil {
		if val.Name != "habname" {
			t.Error()
		}
		if val.InstallCommands[0].Type != base.EXEC_COMMAND {
			t.Error()
		}
	} else {
		t.Error(err)
	}
}