package docker

import (
	"testing"
	"gatoor/orca/base"
	"gatoor/orca/hostRewrite/types"
)

func TestClient_InstallApp_NewApp(t *testing.T) {
	cli := Client{}
	appConf := base.AppConfiguration{}
	appConf.Name = "nginx"
	appConf.Version = "1.0"
	appConf.DockerConfig.Reference = "nginx"
	appConf.DockerConfig.Repository = "docker.io"
	appsState := types.AppsState{}
	retry := types.RetryCounter{}
	conf := types.Configuration{}
	cli.InstallApp(appConf, &appsState, &retry, &conf)
}


