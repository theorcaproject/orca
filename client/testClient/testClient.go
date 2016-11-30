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
	return !strings.Contains(string(appConf.Version), "installfail")
}

func (c *Client) RunApp(appId base.AppId, appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	return !strings.Contains(string(appConf.Version), "runfail")
}

func (c *Client) QueryApp(appId base.AppId, appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	return !strings.Contains(string(appConf.Version), "queryfail")
}

func (c *Client) StopApp(appId base.AppId, appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	return !strings.Contains(string(appConf.Version), "stopfail")
}

func (c *Client) DeleteApp(appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	return !strings.Contains(string(appConf.Version), "deletefail")
}

func (c *Client) AppMetrics(appId base.AppId, appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration, metrics *types.AppsMetricsById) bool {
	if !strings.Contains(string(appConf.Version), "metrics=") {
		return false
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	arr := strings.Split(strings.Split(string(appConf.Version), "metrics=")[1], "_")
	cpu, _ := strconv.Atoi(arr[0])
	mem, _ := strconv.Atoi(arr[1])
	net, _ := strconv.Atoi(arr[2])
	resp, _ := strconv.Atoi(arr[3])
	metrics.Add(appId, now, base.AppStats{
		CpuUsage: base.Usage(cpu), MemoryUsage: base.Usage(mem), NetworkUsage: base.Usage(net), ResponsePerformance: uint64(resp),
	})
	return true
}


