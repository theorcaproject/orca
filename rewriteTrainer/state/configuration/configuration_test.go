package state_configuration_test

import (
	"gatoor/orca/rewriteTrainer/state/configuration"
	"gatoor/orca/base"
	"testing"
)

func prepareConfigState() state_configuration.ConfigurationState {
	var GlobalConfig state_configuration.ConfigurationState
	GlobalConfig.Init()
	return GlobalConfig
}

func TestConfigureApp(t *testing.T) {
	GlobalConfig := prepareConfigState()
	GlobalConfig.ConfigureApp(base.AppConfiguration{
		"appname", base.APP_HTTP, 1, 1, 2,
		base.DockerConfig{}, base.RawConfig{}, "", "",
	})

	if _, err := GlobalConfig.GetApp("unknown", 2); err == nil {
		t.Error()
	}
	if _, err := GlobalConfig.GetApp("appname", 3); err == nil {
		t.Error()
	}
	if val, err :=  GlobalConfig.GetApp("appname", 1); err == nil {
		if val.Name != "appname" {
			t.Error()
		}
	} else {
		t.Error(err)
	}
}


func TestConfigureHabitat(t *testing.T) {
	GlobalConfig := prepareConfigState()
	GlobalConfig.ConfigureHabitat(base.HabitatConfiguration{"habname", 1, []base.OsCommand{
		{
			base.EXEC_COMMAND,
			base.Command{
					":aa", "mmm",
				},
			},
		},
	})

	if _, err := GlobalConfig.GetHabitat("unknown", 2); err == nil {
		t.Error()
	}
	if _, err := GlobalConfig.GetHabitat("habname", 2); err == nil {
		t.Error()
	}
	if val, err :=  GlobalConfig.GetHabitat("habname", 1); err == nil {
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

func TestAppConfigurationVersions_LatestVersion(t *testing.T) {
	dict := state_configuration.AppConfigurationVersions{}

	dict[1] = base.AppConfiguration{}
	dict[2] = base.AppConfiguration{}

	latest := dict.LatestVersion()

	if latest != 2 {
		t.Error("wrong version")
	}

	dict[2] = base.AppConfiguration{}

	latest2 := dict.LatestVersion()

	if latest2 != 2 {
		t.Error("wrong version")
	}
}

func TestAppConfigurationVersions_AllAppsLatestVersion(t *testing.T) {
	GlobalConfig := prepareConfigState()
	GlobalConfig.ConfigureApp(base.AppConfiguration{
		"appname", base.APP_HTTP, 1, 1, 2,
		base.DockerConfig{}, base.RawConfig{}, "", "",
	})
	GlobalConfig.ConfigureApp(base.AppConfiguration{
		"appname", base.APP_HTTP, 2, 1, 2,
		base.DockerConfig{},base.RawConfig{}, "", "",
	})
	GlobalConfig.ConfigureApp(base.AppConfiguration{
		"appname2", base.APP_HTTP, 3, 1, 2,
		base.DockerConfig{},base.RawConfig{}, "", "",
	})

	latest := GlobalConfig.AllAppsLatest()

	if latest["appname"].Version != 2 {
		t.Error("wrong version")
	}
	if latest["appname2"].Version != 3 {
		t.Error("wrong version")
	}
}

