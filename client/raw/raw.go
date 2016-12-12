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