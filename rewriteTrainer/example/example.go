package example

import (
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/rewriteTrainer/config"
	"gatoor/orca/base"
	"gatoor/orca/rewriteTrainer/needs"
)

func ExampleCloudState() {
	state := state_cloud.CloudLayoutAll{}
	state.Init()
	state.Current.AddEmptyHost("host1")
	state.Current.AddEmptyHost("host2")
	state.Current.AddEmptyHost("host3")
	state.Current.AddApp("host1", "app1", 1, 1)
	state.Current.AddApp("host1", "app11", 1, 2)
	state.Current.AddApp("host2", "app2", 2, 10)

	state.Desired.AddEmptyHost("host1")
	state.Desired.AddEmptyHost("host2")
	state.Desired.AddEmptyHost("host3")
	state.Desired.AddApp("host1", "app1", 1, 1)
	state.Desired.AddApp("host1", "app11", 1, 2)
	state.Desired.AddApp("host2", "app2", 2, 5)
	state.Desired.AddApp("host3", "app2", 2, 5)
}

func ExampleJsonConfig() config.JsonConfiguration {
	conf := config.JsonConfiguration{}

	conf.Trainer.Port = 5000
	conf.Trainer.Policies.TRY_TO_REMOVE_HOSTS = true

	conf.Apps = []base.AppConfiguration{
		{
			Name: "http1",
			Version: 1,
			Type: base.APP_HTTP,
			MinDeploymentCount: 2,
			TargetDeploymentCount: 2,
			Needs: needs.AppNeeds{
				MemoryNeeds: needs.MemoryNeeds(5),
				CpuNeeds: needs.CpuNeeds(5),
				NetworkNeeds: needs.NetworkNeeds(5),
			},
		},{
			Name: "app1",
			Version: 1,
			Type: base.APP_WORKER,
			MinDeploymentCount: 2,
			TargetDeploymentCount: 2,
			Needs: needs.AppNeeds{
				MemoryNeeds: needs.MemoryNeeds(5),
				CpuNeeds: needs.CpuNeeds(5),
				NetworkNeeds: needs.NetworkNeeds(5),
			},
		},
		{
			Name: "app11",
			Version: 1,
			Type: base.APP_WORKER,
			MinDeploymentCount: 2,
			TargetDeploymentCount: 2,
			Needs: needs.AppNeeds{
				MemoryNeeds: needs.MemoryNeeds(5),
				CpuNeeds: needs.CpuNeeds(5),
				NetworkNeeds: needs.NetworkNeeds(5),
			},
		},
		{
			Name: "app2",
			Version: 2,
			Type: base.APP_WORKER,
			MinDeploymentCount: 2,
			TargetDeploymentCount: 2,
			Needs: needs.AppNeeds{
				MemoryNeeds: needs.MemoryNeeds(5),
				CpuNeeds: needs.CpuNeeds(5),
				NetworkNeeds: needs.NetworkNeeds(5),
			},
		},
	}
	return conf
}

