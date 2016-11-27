package main

import (
      "testing"
	"gatoor/orca/base"
	"os"
)


func before() {
	hostInfo = base.HostInfo{}
	AppConfigCache = AppConfig{}
	StableAppVersionsCache = StableAppVersions{}
}

func pollFalse(conf base.AppConfiguration) bool {
	return false
}

func pollTrue(conf base.AppConfiguration) bool {
	return true
}

func Test_pollMetrics(t *testing.T) {

}

func Test_pollApp_deployingApp (t *testing.T) {
	appInfo := base.AppInfo{base.APP_HTTP, "app1", "1.0", base.STATUS_DEPLOYING, "app1_1"}
	hostInfo.Apps = []base.AppInfo{
		appInfo,
	}

	before := getAllAppStatus("app1", "1.0")
	if before != base.STATUS_DEPLOYING {
		t.Error(before)
	}

	appConf := base.AppConfiguration{Name: "app1", Version: "1.0"}
	AppConfigCache.Set(appConf)
	pollAppStatus(appInfo, pollFalse)
	after := getAllAppStatus("app1", "1.0")
	if after != base.STATUS_DEPLOYING {
		t.Error(before)
	}

	if StableAppVersionsCache.IsStable("unknown", "1.0") {
		t.Error(StableAppVersionsCache)
	}
	if StableAppVersionsCache.IsStable("app1", "1.0") {
		t.Error(StableAppVersionsCache)
	}
}

func Test_pollApp_runningAppFromDeploying (t *testing.T) {
	appInfo := base.AppInfo{base.APP_HTTP, "app1", "1.0", base.STATUS_DEPLOYING, "app1_1"}
	hostInfo.Apps = []base.AppInfo{
		appInfo,
	}

	before := getAllAppStatus("app1", "1.0")
	if before != base.STATUS_DEPLOYING {
		t.Error(before)
	}

	appConf := base.AppConfiguration{Name: "app1", Version: "1.0"}
	AppConfigCache.Set(appConf)
	pollAppStatus(appInfo, pollTrue)
	after := getAllAppStatus("app1", "1.0")
	if after != base.STATUS_RUNNING {
		t.Error(before)
	}

	if !StableAppVersionsCache.IsStable("app1", "1.0") {
		t.Error(StableAppVersionsCache)
	}
}

func Test_pollApp_runningApp (t *testing.T) {
	appInfo := base.AppInfo{base.APP_HTTP, "app1", "1.0", base.STATUS_DEAD, "app1_1"}
	StableAppVersionsCache.Set("app1", "0.1", true)
	hostInfo.Apps = []base.AppInfo{
		appInfo,
	}

	before := getAllAppStatus("app1", "1.0")
	if before != base.STATUS_DEAD {
		t.Error(before)
	}

	appConf := base.AppConfiguration{Name: "app1", Version: "1.0"}
	AppConfigCache.Set(appConf)
	pollAppStatus(appInfo, pollTrue)
	after := getAllAppStatus("app1", "1.0")
	if after != base.STATUS_RUNNING {
		t.Error(before)
	}

	if !StableAppVersionsCache.IsStable("app1", "1.0") {
		t.Error(StableAppVersionsCache)
	}
	if  StableAppVersionsCache.GetLatestStable("app1") != "1.0" {
		t.Error(StableAppVersionsCache)
	}
}

func Test_pollApp_deadApp (t *testing.T) {
	StableAppVersionsCache.Set("app1", "0.1", true)
	appInfo := base.AppInfo{base.APP_HTTP, "app1", "1.0", base.STATUS_RUNNING, "app1_1"}
	hostInfo.Apps = []base.AppInfo{
		appInfo,
	}

	before := getAllAppStatus("app1", "1.0")
	if before != base.STATUS_RUNNING {
		t.Error(before)
	}

	appConf := base.AppConfiguration{Name: "app1", Version: "1.0"}
	AppConfigCache.Set(appConf)
	pollAppStatus(appInfo, pollFalse)
	after := getAllAppStatus("app1", "1.0")
	if after != base.STATUS_DEAD {
		t.Error(before)
	}
	if StableAppVersionsCache.IsStable("app1", "1.0") {
		t.Error(StableAppVersionsCache)
	}

	if  StableAppVersionsCache.GetLatestStable("app1") != "0.1" {
		t.Error(StableAppVersionsCache)
	}
}


func Test_installApp_newApp(t *testing.T) {
	before()
	conf := base.PushConfiguration{
		DeploymentCount: 2,
		AppConfiguration: base.AppConfiguration{
			Name: "app1",
			Type: base.APP_WORKER,
			Version: "0.0.1",
			MinDeploymentCount: 1,
			MaxDeploymentCount: 5,
			InstallCommands: []base.OsCommand{{base.EXEC_COMMAND,base.Command{"ls", "/"}},},
			RunCommand: base.OsCommand{base.EXEC_COMMAND,base.Command{"ls", "/"}},
			QueryStateCommand: base.OsCommand{base.EXEC_COMMAND,base.Command{"ls", "/"}},
			RemoveCommand: base.OsCommand{base.EXEC_COMMAND,base.Command{"ls", "/"}},
		},
	}
	installApp(conf.AppConfiguration, 1)
	if len(hostInfo.Apps) != 1 {
		t.Error(hostInfo)
	}
	if len(AppConfigCache) != 1 {
		t.Error(AppConfigCache)
	}
	if AppConfigCache.Get("app1", "0.0.1").Type != base.APP_WORKER {
		t.Error(AppConfigCache)
	}
}

func Test_installApp_appIsDeployingOrRunning(t *testing.T) {
	before()
	hostInfo.Apps = []base.AppInfo{
		base.AppInfo{
			Name: "app1",
			Type: base.APP_WORKER,
			Version: "0.0.1",
			Status: base.STATUS_DEPLOYING},
	}
	conf := base.PushConfiguration{
		DeploymentCount: 2,
		AppConfiguration: base.AppConfiguration{
			Name: "app1",
			Type: base.APP_WORKER,
			Version: "0.0.1",
			MinDeploymentCount: 1,
			MaxDeploymentCount: 5,
			InstallCommands: []base.OsCommand{{base.EXEC_COMMAND,base.Command{"ls", "/"}},},
			RunCommand: base.OsCommand{base.EXEC_COMMAND,base.Command{"ls", "/"}},
			QueryStateCommand: base.OsCommand{base.EXEC_COMMAND,base.Command{"ls", "/"}},
			RemoveCommand: base.OsCommand{base.EXEC_COMMAND,base.Command{"ls", "/"}},
		},
	}
	installApp(conf.AppConfiguration, 1)
	if len(hostInfo.Apps) != 1 {
		t.Error(hostInfo)
	}
	if len(AppConfigCache) != 0 {
		t.Error(AppConfigCache)
	}
}

func Test_installApp_updatedApp(t *testing.T) {
	before()
	conf := base.PushConfiguration{
		DeploymentCount: 2,
		AppConfiguration: base.AppConfiguration{
			Name: "app1",
			Type: base.APP_WORKER,
			Version: "0.0.1",
			MinDeploymentCount: 1,
			MaxDeploymentCount: 5,
			InstallCommands: []base.OsCommand{{base.EXEC_COMMAND,base.Command{"ls", "/"}},},
			RunCommand: base.OsCommand{base.EXEC_COMMAND,base.Command{"ls", "/"}},
			QueryStateCommand: base.OsCommand{base.EXEC_COMMAND,base.Command{"ls", "/"}},
			RemoveCommand: base.OsCommand{base.EXEC_COMMAND,base.Command{"ls", "/"}},
		},
	}
	installApp(conf.AppConfiguration, 1)
	if len(hostInfo.Apps) != 1 {
		t.Error(hostInfo)
	}
	if len(AppConfigCache) != 1 {
		t.Error(AppConfigCache)
	}
	if AppConfigCache.Get("app1", "0.0.1").Type != base.APP_WORKER {
		t.Error(AppConfigCache)
	}
	conf2 := base.PushConfiguration{
		DeploymentCount: 2,
		AppConfiguration: base.AppConfiguration{
			Name: "app1",
			Type: base.APP_WORKER,
			Version: "0.0.2",
			MinDeploymentCount: 4,
			MaxDeploymentCount: 5,
			InstallCommands: []base.OsCommand{{base.EXEC_COMMAND,base.Command{"ls", "/"}},},
			RunCommand: base.OsCommand{base.EXEC_COMMAND,base.Command{"ls", "/"}},
			QueryStateCommand: base.OsCommand{base.EXEC_COMMAND,base.Command{"ls", "/"}},
			RemoveCommand: base.OsCommand{base.EXEC_COMMAND,base.Command{"ls", "/"}},
		},
	}
	installApp(conf2.AppConfiguration, 1)
	if len(hostInfo.Apps) != 1 {
		t.Error(hostInfo)
	}
	if len(AppConfigCache) != 1 {
		t.Error(AppConfigCache)
	}
	if len(AppConfigCache["app1"]) != 2 {
		t.Error(AppConfigCache)
	}
	if AppConfigCache["app1"]["0.0.2"].MinDeploymentCount != 4 {
		t.Error(AppConfigCache)
	}
}

func Test_installApp_rollback(t *testing.T) {

}


func Test_installApp_integration_allOk(t *testing.T) {
	before()
	conf := base.PushConfiguration{
		DeploymentCount: 2,
		AppConfiguration: base.AppConfiguration{
			Name: "app1",
			Type: base.APP_WORKER,
			Version: "0.0.1",
			MinDeploymentCount: 1,
			MaxDeploymentCount: 5,
			InstallCommands: []base.OsCommand{{base.FILE_COMMAND, base.Command{"/tmp/orca_test_install", "some stuff install"}},},
			RunCommand: base.OsCommand{base.FILE_COMMAND, base.Command{"/tmp/orca_test_run", "some stuff run"}},
			QueryStateCommand: base.OsCommand{base.EXEC_COMMAND, base.Command{"test", "/tmp/orca_test_install && echo true || echo false "}},
			RemoveCommand: base.OsCommand{base.EXEC_COMMAND,base.Command{"ls", "/"}},
		},
	}
	installApp(conf.AppConfiguration, 1)

	if hostInfo.Apps[0].Version != "0.0.1" || hostInfo.Apps[0].Status != base.STATUS_RUNNING {
		t.Error(hostInfo)
	}
	if StableAppVersionsCache.GetLatestStable("app1") != "0.0.1" {
		t.Error(StableAppVersionsCache)
	}
}

func Test_installApp_integration_installFailNewApp(t *testing.T) {
	before()
	conf := base.PushConfiguration{
		DeploymentCount: 2,
		AppConfiguration: base.AppConfiguration{
			Name: "app1",
			Type: base.APP_WORKER,
			Version: "0.0.1",
			MinDeploymentCount: 1,
			MaxDeploymentCount: 5,
			InstallCommands: []base.OsCommand{{base.EXEC_COMMAND, base.Command{"/tmp/orca_test_install", "some stuff install"}},},
			RunCommand: base.OsCommand{base.FILE_COMMAND, base.Command{"/tmp/orca_test_run", "some stuff run"}},
			QueryStateCommand: base.OsCommand{base.EXEC_COMMAND, base.Command{"test", "/tmp/orca_test_install && echo true || echo false "}},
			RemoveCommand: base.OsCommand{base.EXEC_COMMAND,base.Command{"ls", "/"}},
		},
	}
	installApp(conf.AppConfiguration, 1)

	if len(hostInfo.Apps) != 0 {
		t.Errorf("%+v", hostInfo)
	}

	if StableAppVersionsCache.GetLatestStable("app1") == "0.0.1" {
		t.Error(StableAppVersionsCache)
	}
}

func Test_installApp_integration_Rollback(t *testing.T) {
	before()

	conf := base.PushConfiguration{
		DeploymentCount: 2,
		AppConfiguration: base.AppConfiguration{
			Name: "app1",
			Type: base.APP_WORKER,
			Version: "0.0.1",
			MinDeploymentCount: 1,
			MaxDeploymentCount: 5,
			InstallCommands: []base.OsCommand{{base.FILE_COMMAND, base.Command{"/tmp/orca_test_install", "some stuff install"}},},
			RunCommand: base.OsCommand{base.FILE_COMMAND, base.Command{"/tmp/orca_test_run", "some stuff run"}},
			QueryStateCommand: base.OsCommand{base.EXEC_COMMAND, base.Command{"test", "/tmp/orca_test_install && echo true || echo false "}},
			RemoveCommand: base.OsCommand{base.EXEC_COMMAND,base.Command{"ls", "/"}},
		},
	}
	installApp(conf.AppConfiguration, 1)

	if hostInfo.Apps[0].Version != "0.0.1" || hostInfo.Apps[0].Status != base.STATUS_RUNNING {
		t.Error(hostInfo)
	}
	if StableAppVersionsCache.GetLatestStable("app1") != "0.0.1" {
		t.Error(StableAppVersionsCache)
	}


	confFail := base.PushConfiguration{
		DeploymentCount: 2,
		AppConfiguration: base.AppConfiguration{
			Name: "app1",
			Type: base.APP_WORKER,
			Version: "2.0.1",
			MinDeploymentCount: 1,
			MaxDeploymentCount: 5,
			InstallCommands: []base.OsCommand{{base.EXEC_COMMAND, base.Command{"/tmp/orca_test_install", "some stuff install"}},},
			RunCommand: base.OsCommand{base.FILE_COMMAND, base.Command{"/tmp/orca_test_run", "some stuff run"}},
			QueryStateCommand: base.OsCommand{base.EXEC_COMMAND, base.Command{"test", "/tmp/orca_test_install && echo true || echo false "}},
			RemoveCommand: base.OsCommand{base.EXEC_COMMAND,base.Command{"ls", "/"}},
		},
	}
	installApp(confFail.AppConfiguration, 1)

	if len(hostInfo.Apps) != 1 {
		t.Errorf("%+v", hostInfo)
	}

	if StableAppVersionsCache.GetLatestStable("app1") != "0.0.1" {
		t.Error(StableAppVersionsCache)
	}
}


func Test_installApp_integration_update(t *testing.T) {
	before()

	conf := base.PushConfiguration{
		DeploymentCount: 2,
		AppConfiguration: base.AppConfiguration{
			Name: "app1",
			Type: base.APP_WORKER,
			Version: "0.0.1",
			MinDeploymentCount: 1,
			MaxDeploymentCount: 5,
			InstallCommands: []base.OsCommand{{base.FILE_COMMAND, base.Command{"/tmp/orca_test_install", "some stuff install"}},},
			RunCommand: base.OsCommand{base.FILE_COMMAND, base.Command{"/tmp/orca_test_run", "some stuff run"}},
			QueryStateCommand: base.OsCommand{base.EXEC_COMMAND, base.Command{"test", "/tmp/orca_test_install && echo true || echo false "}},
			RemoveCommand: base.OsCommand{base.EXEC_COMMAND, base.Command{"rm", "/tmp/orca_test_install"}},
		},
	}
	installApp(conf.AppConfiguration, 1)

	if hostInfo.Apps[0].Version != "0.0.1" || hostInfo.Apps[0].Status != base.STATUS_RUNNING {
		t.Error(hostInfo)
	}
	if StableAppVersionsCache.GetLatestStable("app1") != "0.0.1" {
		t.Error(StableAppVersionsCache)
	}


	conf2 := base.PushConfiguration{
		DeploymentCount: 2,
		AppConfiguration: base.AppConfiguration{
			Name: "app1",
			Type: base.APP_WORKER,
			Version: "2.0.1",
			MinDeploymentCount: 1,
			MaxDeploymentCount: 5,
			InstallCommands: []base.OsCommand{{base.FILE_COMMAND, base.Command{"/tmp/orca_test_install2", "some stuff install"}},},
			RunCommand: base.OsCommand{base.FILE_COMMAND, base.Command{"/tmp/orca_test_run", "some stuff run"}},
			QueryStateCommand: base.OsCommand{base.EXEC_COMMAND, base.Command{"test", "/tmp/orca_test_install2 && echo true || echo false "}},
			RemoveCommand: base.OsCommand{},
		},
	}

	if _, err := os.Stat("/tmp/orca_test_install"); err != nil {
		t.Error("file should exist")
	}

	installApp(conf2.AppConfiguration, 2)

	if hostInfo.Apps[0].Version != "2.0.1" || hostInfo.Apps[0].Status != base.STATUS_RUNNING {
		t.Errorf("%+v", hostInfo)
		t.Errorf("%+v", AppConfigCache)
		t.Errorf("%+v", StableAppVersionsCache)
	}

	if StableAppVersionsCache.GetLatestStable("app1") != "2.0.1" {
		t.Error(StableAppVersionsCache)
	}

	if _, err := os.Stat("/tmp/orca_test_install"); err == nil {
		t.Error("file should be deleted by app uninstall")
	}
}


func Test_installApp_integration_scaleUp(t *testing.T) {
	before()

	conf := base.PushConfiguration{
		DeploymentCount: 2,
		AppConfiguration: base.AppConfiguration{
			Name: "app1",
			Type: base.APP_WORKER,
			Version: "0.0.1",
			MinDeploymentCount: 1,
			MaxDeploymentCount: 5,
			InstallCommands: []base.OsCommand{{base.FILE_COMMAND, base.Command{"/tmp/orca_test_install", "some stuff install"}},},
			RunCommand: base.OsCommand{base.FILE_COMMAND, base.Command{"/tmp/orca_test_run", "some stuff run"}},
			QueryStateCommand: base.OsCommand{base.EXEC_COMMAND, base.Command{"test", "/tmp/orca_test_install && echo true || echo false "}},
			RemoveCommand: base.OsCommand{base.EXEC_COMMAND, base.Command{"rm", "/tmp/orca_test_install"}},
		},
	}
	installApp(conf.AppConfiguration, 1)

	if hostInfo.Apps[0].Version != "0.0.1" || hostInfo.Apps[0].Status != base.STATUS_RUNNING {
		t.Error(hostInfo)
	}
	if StableAppVersionsCache.GetLatestStable("app1") != "0.0.1" {
		t.Error(StableAppVersionsCache)
	}

	if _, err := os.Stat("/tmp/orca_test_install"); err != nil {
		t.Error("file should exist")
	}

	installApp(conf.AppConfiguration, 5)

	if StableAppVersionsCache.GetLatestStable("app1") != "0.0.1" {
		t.Error(StableAppVersionsCache)
	}
	if len(hostInfo.Apps) != 5 {
		t.Error(hostInfo)
	}

}


func Test_saveState(t *testing.T) {
	hostInfo = base.HostInfo{
		HostId: "somehost",
		IpAddr: "1.2.3.4",
		OsInfo: base.OsInfo{},
		Apps: []base.AppInfo{{
			Type: base.APP_HTTP,
			Name: "app1",
			Version: "1.0",
			Status: base.STATUS_RUNNING,
			Id: "app1_123"}},
	}

	AppConfigCache.Set(base.AppConfiguration{Name: "app1", Type: base.APP_HTTP, Version: "1.0"})
	saveState()
	hostInfo = base.HostInfo{}
	AppConfigCache = AppConfig{}
	if hostInfo.HostId != "" {
		t.Error(hostInfo)
	}
	if AppConfigCache.Get("app1", "1.0").Name != "" {
		t.Error(AppConfigCache)
	}
	loadLastState()
	if hostInfo.HostId != "somehost" {
		t.Error(hostInfo)
	}
	if AppConfigCache.Get("app1", "1.0").Name != "app1" {
		t.Error(AppConfigCache)
	}
}