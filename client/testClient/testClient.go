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

package testClient

import (
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/client/types"
	"gatoor/orca/base"
	"strings"
	"time"
	"strconv"
)

var TestLogger = Logger.LoggerWithField(Logger.Logger, "module", "testClient")

type Client struct {

}

func (c *Client) Init() {

}

func (c *Client) Type() types.ClientType {
	return types.TEST_CLIENT
}

func (c *Client) InstallApp(appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	return !strings.Contains(string(appConf.DockerConfig.Tag), "installfail")
}

func (c *Client) RunApp(appId base.AppId, appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	return !strings.Contains(string(appConf.DockerConfig.Tag), "runfail")
}

func (c *Client) QueryApp(appId base.AppId, appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	return !strings.Contains(string(appConf.DockerConfig.Tag), "queryfail")
}

func (c *Client) StopApp(appId base.AppId, appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	return !strings.Contains(string(appConf.DockerConfig.Tag), "stopfail")
}

func (c *Client) DeleteApp(appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	return !strings.Contains(string(appConf.DockerConfig.Tag), "deletefail")
}

func (c *Client) AppMetrics(appId base.AppId, appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration, metrics *types.AppsMetricsById) bool {
	if !strings.Contains(string(appConf.DockerConfig.Tag), "metrics=") {
		return false
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	arr := strings.Split(strings.Split(string(appConf.DockerConfig.Tag), "metrics=")[1], "_")
	cpu, _ := strconv.Atoi(arr[0])
	mem, _ := strconv.Atoi(arr[1])
	net, _ := strconv.Atoi(arr[2])
	resp, _ := strconv.Atoi(arr[3])
	metrics.Add(appId, now, base.AppStats{
		CpuUsage: base.Usage(cpu), MemoryUsage: base.Usage(mem), NetworkUsage: base.Usage(net), ResponsePerformance: uint64(resp),
	})
	return true
}


