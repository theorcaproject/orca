package docker

import (
	"testing"
	"gatoor/orca/base"
	"gatoor/orca/client/types"
)

func TestClient_InstallApp(t *testing.T) {
	cli := Client{}
	cli.Init()
	appConf := base.AppConfiguration{}
	appConf.Name = "nginx"
	appConf.Version = "latest"
	appConf.DockerConfig.Reference = "docker.io"
	appConf.DockerConfig.Repository = "nginx"
	appsState := types.AppsState{}
	//retry := types.RetryCounter{}
	conf := types.Configuration{}
	cli.InstallApp(appConf, &appsState, &conf)
}


func TestClient_RunStopApp(t *testing.T) {
	cli := Client{}
	cli.Init()

	appConf := base.AppConfiguration{}
	appConf.Name = "nginx"
	appConf.Version = "latest"
	appConf.DockerConfig.Reference = "docker.io"
	appConf.DockerConfig.Repository = "nginx"
	appsState := types.AppsState{}
	conf := types.Configuration{}

	if !cli.RunApp("superapp1", appConf, &appsState, &conf) {
		t.Error()
	}

	if !cli.RunApp("superapp2", appConf, &appsState, &conf) {
		t.Error()
	}

	if !cli.StopApp("superapp1", appConf, &appsState, &conf) {
		t.Error()
	}

	if !cli.StopApp("superapp2", appConf, &appsState, &conf) {
		t.Error()
	}
}

func TestClient_QueryApp(t *testing.T) {
	cli := Client{}
	cli.Init()

	appConf := base.AppConfiguration{}
	appConf.Name = "nginx"
	appConf.Version = "latest"
	appConf.DockerConfig.Reference = "docker.io"
	appConf.DockerConfig.Repository = "nginx"
	appsState := types.AppsState{}
	conf := types.Configuration{}

	if !cli.RunApp("superapp1", appConf, &appsState, &conf) {
		t.Error()
	}

	if !cli.QueryApp("superapp1", appConf, &appsState, &conf) {
		t.Error()
	}

	dockerCli.StopContainer("superapp1", 0)

	if cli.QueryApp("superapp1", appConf, &appsState, &conf) {
		t.Error()
	}

	cli.StopApp("superapp1", appConf, &appsState, &conf)
}


func TestClient_AppMetrics(t *testing.T) {
	cli := Client{}
	cli.Init()

	appConf := base.AppConfiguration{}
	appConf.Name = "nginx"
	appConf.Version = "latest"
	appConf.DockerConfig.Reference = "docker.io"
	appConf.DockerConfig.Repository = "nginx"
	appsState := types.AppsState{}
	conf := types.Configuration{}
	metrics := types.AppsMetricsById{}

	if !cli.RunApp("superapp1", appConf, &appsState, &conf) {
		t.Error()
	}

	if !cli.AppMetrics("superapp1", appConf, &appsState, &conf, &metrics) {
		t.Error()
	}

	if !cli.StopApp("superapp1", appConf, &appsState, &conf) {
		t.Error()
	}
}


func TestClient_DeleteApp(t *testing.T) {
	cli := Client{}
	cli.Init()

	appConf := base.AppConfiguration{}
	appConf.Name = "nginx"
	appConf.Version = "latest"
	appConf.DockerConfig.Reference = "docker.io"
	appConf.DockerConfig.Repository = "nginx"
	appsState := types.AppsState{}
	conf := types.Configuration{}

	if !cli.DeleteApp(appConf, &appsState, &conf) {
		t.Error()
	}

	if cli.DeleteApp(appConf, &appsState, &conf) {
		t.Error()
	}
}



