/*
Copyright Alex Mack
This file is part of Orca.

Orca is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Orca is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Orca.  If not, see <http://www.gnu.org/licenses/>.
*/

package docker

import (
	"testing"
	"gatoor/orca/base"
	"gatoor/orca/client/types"
	"time"
)

func TestClient_InstallApp(t *testing.T) {
	cli := Client{}
	cli.Init()
	appConf := base.AppConfiguration{}
	appConf.Name = "nginx"
	appConf.Version = 1
	appConf.DockerConfig.Reference = "docker.io"
	appConf.DockerConfig.Repository = "nginx"
	appConf.DockerConfig.Tag = "latest"
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
	appConf.Version = 1
	appConf.DockerConfig.Reference = "docker.io"
	appConf.DockerConfig.Repository = "nginx"
	appConf.DockerConfig.Tag = "latest"
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
	appConf.Version = 1
	appConf.DockerConfig.Reference = "docker.io"
	appConf.DockerConfig.Repository = "nginx"
	appConf.DockerConfig.Tag = "latest"
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
	appConf.Version = 1
	appConf.DockerConfig.Reference = "docker.io"
	appConf.DockerConfig.Repository = "nginx"
	appConf.DockerConfig.Tag = "latest"
	appsState := types.AppsState{}
	conf := types.Configuration{}
	metrics := types.AppsMetricsById{}

	if !cli.RunApp("superapp1", appConf, &appsState, &conf) {
		t.Error()
	}

	if !cli.AppMetrics("superapp1", appConf, &appsState, &conf, &metrics) {
		t.Error()
	}
	time.Sleep(50 * time.Millisecond)
	if !cli.AppMetrics("superapp1", appConf, &appsState, &conf, &metrics) {
		t.Error()
	time.Sleep(50 * time.Millisecond)
	}
	if !cli.AppMetrics("superapp1", appConf, &appsState, &conf, &metrics) {
		t.Error()
	}

	if len(metrics.All()) != 3 {
		t.Error(metrics.All())
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
	appConf.Version = 1
	appConf.DockerConfig.Reference = "docker.io"
	appConf.DockerConfig.Repository = "nginx"
	appConf.DockerConfig.Tag = "latest"
	appsState := types.AppsState{}
	conf := types.Configuration{}

	if !cli.DeleteApp(appConf, &appsState, &conf) {
		t.Error()
	}

	if cli.DeleteApp(appConf, &appsState, &conf) {
		t.Error()
	}
}



