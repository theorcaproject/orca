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

package main


import (
	"testing"
	"gatoor/orca/client/types"
	"gatoor/orca/base"
	"gatoor/orca/client/client"
	"encoding/json"
	"io/ioutil"
	"os"
)


func TestPrepareData(t *testing.T) {
	state := types.AppsState{}
	client.Configuration.HostId = "somehost"
	state.Add("app1_1", base.AppInfo{Name: "app1", Version: 1, Status:base.STATUS_RUNNING, Id: "app1_1", Type: base.APP_HTTP})
	metrics := base.AppMetrics{}
	metrics.Add("app1", 1, "sometimestring", base.AppStats{CpuUsage: 100, MemoryUsage: 20, NetworkUsage: 10, ResponsePerformance: 3})
	metrics.Add("app1", 1, "sometimestring2", base.AppStats{CpuUsage: 100, MemoryUsage: 20, NetworkUsage: 10, ResponsePerformance: 3})
	res := prepareData(state, metrics)
	if res.HostInfo.HostId != "somehost" || len(res.Stats.AppMetrics["app1"]["1"]) != 2 {
		t.Error(res)
	}
}


func TestLoadConfiguration(t *testing.T) {
	var conf, err = json.Marshal(types.Configuration{HostId: "somehost", AppStatusPollInterval: 100, MetricsPollInterval: 5, TrainerUrl: "http://some/url"})
	if err != nil {
		t.Error(err)
	}
	err = ioutil.WriteFile("./testconf.json", conf, 0644)

	if err != nil {
		t.Error(err)
	}

	var loadedConf types.Configuration
	file, err := os.Open("./testconf.json")
	loadJsonFile(file, &loadedConf)

	if loadedConf.AppStatusPollInterval != 100 || loadedConf.HostId != "somehost" {
		t.Error(loadedConf)
	}
}

func TestSaveLoadStateAndConfig(t *testing.T) {
	state := types.AppsState{}
	conf := types.AppsConfiguration{}
	state.Add("app1_1", base.AppInfo{Type:base.APP_HTTP, Name:"app1", Version: 1, Id:"app1_1", Status:base.STATUS_RUNNING})
	state.Add("app1_2", base.AppInfo{Type:base.APP_HTTP, Name:"app1", Version: 1, Id:"app1_2", Status:base.STATUS_DEAD})
	state.Add("app2_1", base.AppInfo{Type:base.APP_HTTP, Name:"app2", Version: 2, Id:"app2_1", Status:base.STATUS_DEPLOYING})

	conf.Add("app1_1", base.AppConfiguration{Name:"app1", Type:base.APP_HTTP, Version: 1})
	conf.Add("app1_2", base.AppConfiguration{Name:"app1", Type:base.APP_HTTP, Version: 1})
	conf.Add("app2_1", base.AppConfiguration{Name:"app2", Type:base.APP_WORKER, Version: 2})

	saveStateAndConfig(state, conf)
	client.AppsConfiguration = types.AppsConfiguration{}
	client.AppsState = types.AppsState{}
	loadLastStateAndConfig()

	if len(client.AppsState) != 3 || client.AppsState.Get("app1_1").Status != base.STATUS_RUNNING {
		t.Error(client.AppsState)
	}
	if len(client.AppsConfiguration) != 3 || client.AppsConfiguration.Get("app2_1").Type != base.APP_WORKER {
		t.Error(client.AppsState)
	}
}