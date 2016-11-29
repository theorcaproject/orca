package main


import (
	"testing"
	"gatoor/orca/hostRewrite/types"
	"gatoor/orca/base"
	"gatoor/orca/hostRewrite/client"
)


func TestPrepareData(t *testing.T) {
	state := types.AppsState{}
	client.Configuration.HostId = "somehost"
	state.Add("app1_1", base.AppInfo{Name: "app1", Version: "1.0", Status:base.STATUS_RUNNING, Id: "app1_1", Type: base.APP_HTTP})
	metrics := base.AppMetrics{}
	metrics.Add("app1", "1.0", "sometimestring", base.AppStats{CpuUsage: 100, MemoryUsage: 20, NetworkUsage: 10, ResponseTime: 3})
	metrics.Add("app1", "1.0", "sometimestring2", base.AppStats{CpuUsage: 100, MemoryUsage: 20, NetworkUsage: 10, ResponseTime: 3})
	res := prepareData(state, metrics)
	if res.HostInfo.HostId != "somehost" || len(res.Stats.AppMetrics["app1"]["1.0"]) != 2 {
		t.Error(res)
	}
}