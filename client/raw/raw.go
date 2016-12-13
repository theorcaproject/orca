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

package raw

import (
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/client/types"
	"gatoor/orca/base"
	"strings"
)

var RawLogger = Logger.LoggerWithField(Logger.Logger, "module", "raw")

type Client struct {

}


func (c *Client) Init() {

}

func (c *Client) Type() types.ClientType {
	return types.RAW_CLIENT
}

func (c *Client) InstallApp(appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	return false
}

func (c *Client) RunApp(appId base.AppId, appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	return strings.Contains(string(appConf.DockerConfig.Tag), "fail")
}


func (c *Client) QueryApp(appId base.AppId, appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	return !strings.Contains(string(appConf.DockerConfig.Tag), "queryfail")
}

func (c *Client) StopApp(appId base.AppId, appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	return !strings.Contains(string(appConf.DockerConfig.Tag), "stopfail")
}

func (c *Client) DeleteApp(appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	return false
}

func (c *Client) AppMetrics(appId base.AppId, appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration, metrics *types.AppsMetricsById) bool {
	return false
}