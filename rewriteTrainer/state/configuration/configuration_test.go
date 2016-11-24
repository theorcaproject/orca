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
		"appname", base.APP_HTTP, "0.1", 1, 2,
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
	GlobalConfig.ConfigureHabitat(base.HabitatConfiguration{"habname", "0.1", []base.OsCommand{
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

func TestAppConfigurationVersions_LatestVersion(t *testing.T) {
	dict := state_configuration.AppConfigurationVersions{}

	dict["0.1"] = base.AppConfiguration{}
	dict["0.12"] = base.AppConfiguration{}

	latest := dict.LatestVersion()

	if latest != "0.12" {
		t.Error("wrong version")
	}

	dict["0.2"] = base.AppConfiguration{}

	latest2 := dict.LatestVersion()

	if latest2 != "0.2" {
		t.Error("wrong version")
	}
}

func TestAppConfigurationVersions_AllAppsLatestVersion(t *testing.T) {
	GlobalConfig := prepareConfigState()
	GlobalConfig.ConfigureApp(base.AppConfiguration{
		"appname", base.APP_HTTP, "0.1", 1, 2,
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
	GlobalConfig.ConfigureApp(base.AppConfiguration{
		"appname", base.APP_HTTP, "1.1", 1, 2,
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
	GlobalConfig.ConfigureApp(base.AppConfiguration{
		"appname2", base.APP_HTTP, "2.0", 1, 2,
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

	latest := GlobalConfig.AllAppsLatest()

	if latest["appname"].Version != "1.1" {
		t.Error("wrong version")
	}
	if latest["appname2"].Version != "2.0" {
		t.Error("wrong version")
	}
}

