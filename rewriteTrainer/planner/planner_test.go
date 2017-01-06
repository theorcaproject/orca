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

package planner


import (
	"testing"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/rewriteTrainer/state/configuration"
	"gatoor/orca/base"
	"gatoor/orca/rewriteTrainer/needs"
	"github.com/docker/docker/pkg/testutil/assert"
	"time"
	"fmt"
)

func TestPlanner_DoNothing(t *testing.T) {
	doPlanInternal()

	if len(state_cloud.GlobalCloudLayout.Changes) > 0 {
		t.Fail()
	}
}

func TestPlanner_SpawnServer(t *testing.T) {
	state_configuration.GlobalConfigurationState.Init()
	state_cloud.GlobalCloudLayout.Init()

	appConfigSets := make([]base.AppConfigurationSet, 1)
	appConfigSets[0] = base.AppConfigurationSet{
		Version:1,
	}

	state_configuration.GlobalConfigurationState.ConfigureApp(base.AppConfiguration{
		Name:"testing",
		MinDeploymentCount:1,
		Needs: needs.AppNeeds{CpuNeeds:1, MemoryNeeds:1, NetworkNeeds:1},
		ConfigurationSets:appConfigSets,
	})

	doPlanInternal()

	assert.Equal(t, len(state_cloud.GlobalCloudLayout.Changes), 1)
	assert.Equal(t, state_cloud.GlobalCloudLayout.Changes[0].ChangeType, base.CHANGE_REQUEST__SPAWN_SERVER)
}


func TestPlanner_AddApplication__MinNeedsNotSatisfied(t *testing.T) {
	state_configuration.GlobalConfigurationState.Init()
	state_cloud.GlobalCloudLayout.Init()

	appConfigSets := make([]base.AppConfigurationSet, 1)
	appConfigSets[0] = base.AppConfigurationSet{
		Version:1,
	}

	state_configuration.GlobalConfigurationState.ConfigureApp(base.AppConfiguration{
		Name:"testing",
		MinDeploymentCount:1,
		Needs: needs.AppNeeds{CpuNeeds:1, MemoryNeeds:1, NetworkNeeds:1},
		ConfigurationSets:appConfigSets,
	})

	state_cloud.GlobalCloudLayout.Init()
	state_cloud.GlobalCloudLayout.Current.AddHost("host1", state_cloud.CloudLayoutElement{
		HostId:"host1",
		InstanceType:"t2.micro",
		IpAddress:"172.16.0.1",
		AvailableResources: base.InstanceResources{
			TotalMemoryResource:100,
			TotalNetworkResource:100,
			TotalCpuResource:100,
		},
	})

	doPlanInternal()

	assert.Equal(t, len(state_cloud.GlobalCloudLayout.Changes), 1)
	assert.Equal(t, state_cloud.GlobalCloudLayout.Changes[0].ChangeType, base.UPDATE_TYPE__ADD)
	assert.Equal(t, state_cloud.GlobalCloudLayout.Changes[0].Application, base.AppName("testing"))
}


func TestPlanner_AddApplication__MinNeedsSatisfied(t *testing.T) {
	state_configuration.GlobalConfigurationState.Init()
	state_cloud.GlobalCloudLayout.Init()

	appConfigSets := make([]base.AppConfigurationSet, 1)
	appConfigSets[0] = base.AppConfigurationSet{
		Version:1,
	}

	application := base.AppConfiguration{
		Name:"testing",
		MinDeploymentCount:1,
		TargetDeploymentCount:1,
		Needs: needs.AppNeeds{CpuNeeds:1, MemoryNeeds:1, NetworkNeeds:1},
		ConfigurationSets:appConfigSets,
	}
	state_configuration.GlobalConfigurationState.ConfigureApp(application)

	state_cloud.GlobalCloudLayout.Init()
	host := state_cloud.CloudLayoutElement{
		HostId:"host1",
		InstanceType:"t2.micro",
		IpAddress:"172.16.0.1",
		AvailableResources: base.InstanceResources{
			TotalMemoryResource:100,
			TotalNetworkResource:100,
			TotalCpuResource:100,
		},
	}

	state_cloud.GlobalCloudLayout.Current.AddHost("host1", host)
	state_cloud.GlobalCloudLayout.Current.AddApp("host1", "testing", 1, 1)

	doPlanInternal()

	assert.Equal(t, len(state_cloud.GlobalCloudLayout.Changes), 0)
}


func TestPlanner_AddApplication__ToManyInstances(t *testing.T) {
	state_configuration.GlobalConfigurationState.Init()
	state_cloud.GlobalCloudLayout.Init()

	appConfigSets := make([]base.AppConfigurationSet, 1)
	appConfigSets[0] = base.AppConfigurationSet{
		Version:1,
	}

	application := base.AppConfiguration{
		Name:"testing",
		MinDeploymentCount:0,
		TargetDeploymentCount:0,
		Needs: needs.AppNeeds{CpuNeeds:1, MemoryNeeds:1, NetworkNeeds:1},
		ConfigurationSets:appConfigSets,
	}
	state_configuration.GlobalConfigurationState.ConfigureApp(application)

	state_cloud.GlobalCloudLayout.Init()
	host := state_cloud.CloudLayoutElement{
		HostId:"host1",
		InstanceType:"t2.micro",
		IpAddress:"172.16.0.1",
		AvailableResources: base.InstanceResources{
			TotalMemoryResource:100,
			TotalNetworkResource:100,
			TotalCpuResource:100,
		},
	}

	state_cloud.GlobalCloudLayout.Current.AddHost("host1", host)
	state_cloud.GlobalCloudLayout.Current.AddApp("host1", "testing", 1, 1)

	doPlanInternal()

	assert.Equal(t, len(state_cloud.GlobalCloudLayout.Changes), 1)
	assert.Equal(t, state_cloud.GlobalCloudLayout.Changes[0].ChangeType, base.UPDATE_TYPE__REMOVE)
	assert.Equal(t, state_cloud.GlobalCloudLayout.Changes[0].Application, base.AppName("testing"))
}

func TestPlanner_AddApplication__MinNeedsSatisfiedDesiredNot(t *testing.T) {
	state_configuration.GlobalConfigurationState.Init()
	state_cloud.GlobalCloudLayout.Init()

	appConfigSets := make([]base.AppConfigurationSet, 1)
	appConfigSets[0] = base.AppConfigurationSet{
		Version:1,
	}

	application := base.AppConfiguration{
		Name:"testing",
		MinDeploymentCount:1,
		TargetDeploymentCount:5,
		Needs: needs.AppNeeds{CpuNeeds:1, MemoryNeeds:1, NetworkNeeds:1},
		ConfigurationSets:appConfigSets,
	}
	state_configuration.GlobalConfigurationState.ConfigureApp(application)

	state_cloud.GlobalCloudLayout.Init()
	host := state_cloud.CloudLayoutElement{
		HostId:"host1",
		InstanceType:"t2.micro",
		IpAddress:"172.16.0.1",
		AvailableResources: base.InstanceResources{
			TotalMemoryResource:100,
			TotalNetworkResource:100,
			TotalCpuResource:100,
		},
	}

	state_cloud.GlobalCloudLayout.Current.AddHost("host1", host)
	state_cloud.GlobalCloudLayout.Current.AddApp("host1", "testing", 1, 1)

	doPlanInternal()

	assert.Equal(t, len(state_cloud.GlobalCloudLayout.Changes), 5)
	assert.Equal(t, state_cloud.GlobalCloudLayout.Changes[0].ChangeType, base.CHANGE_REQUEST__SPAWN_SERVER)
	assert.Equal(t, state_cloud.GlobalCloudLayout.Changes[0].Cost.CpuNeeds, needs.CpuNeeds(1))
	assert.Equal(t, state_cloud.GlobalCloudLayout.Changes[0].Cost.MemoryNeeds, needs.MemoryNeeds(1))
	assert.Equal(t, state_cloud.GlobalCloudLayout.Changes[0].Cost.NetworkNeeds, needs.NetworkNeeds(1))

	assert.Equal(t, state_cloud.GlobalCloudLayout.Changes[1].ChangeType, base.CHANGE_REQUEST__SPAWN_SERVER)
	assert.Equal(t, state_cloud.GlobalCloudLayout.Changes[1].Cost.CpuNeeds, needs.CpuNeeds(1))
	assert.Equal(t, state_cloud.GlobalCloudLayout.Changes[1].Cost.MemoryNeeds, needs.MemoryNeeds(1))
	assert.Equal(t, state_cloud.GlobalCloudLayout.Changes[1].Cost.NetworkNeeds, needs.NetworkNeeds(1))
}


func TestPlanner_AddApplication__MinNeedsSatisfied_KillOldVersions(t *testing.T) {
	state_configuration.GlobalConfigurationState.Init()
	state_cloud.GlobalCloudLayout.Init()

	appConfigSets := make([]base.AppConfigurationSet, 2)
	appConfigSets[0] = base.AppConfigurationSet{
		Version:1,
	}

	appConfigSets[1] = base.AppConfigurationSet{
		Version:2,
	}

	application := base.AppConfiguration{
		Name:"testing",
		MinDeploymentCount:1,
		TargetDeploymentCount:1,
		Needs: needs.AppNeeds{CpuNeeds:1, MemoryNeeds:1, NetworkNeeds:1},
		ConfigurationSets:appConfigSets,
	}
	state_configuration.GlobalConfigurationState.ConfigureApp(application)

	state_cloud.GlobalCloudLayout.Init()
	host := state_cloud.CloudLayoutElement{
		HostId:"host1",
		InstanceType:"t2.micro",
		IpAddress:"172.16.0.1",
		AvailableResources: base.InstanceResources{
			TotalMemoryResource:100,
			TotalNetworkResource:100,
			TotalCpuResource:100,
		},
	}
	host2 := state_cloud.CloudLayoutElement{
		HostId:"host1",
		InstanceType:"t2.micro",
		IpAddress:"172.16.0.1",
		AvailableResources: base.InstanceResources{
			TotalMemoryResource:100,
			TotalNetworkResource:100,
			TotalCpuResource:100,
		},
	}

	state_cloud.GlobalCloudLayout.Current.AddHost("host1", host)
	state_cloud.GlobalCloudLayout.Current.AddHost("host2", host2)
	state_cloud.GlobalCloudLayout.Current.AddApp("host1", "testing", 1, 1)
	state_cloud.GlobalCloudLayout.Current.AddApp("host2", "testing", 2, 1)

	doPlanInternal()

	assert.Equal(t, len(state_cloud.GlobalCloudLayout.Changes), 1)
	assert.Equal(t, state_cloud.GlobalCloudLayout.Changes[0].ChangeType, base.UPDATE_TYPE__REMOVE)
	assert.Equal(t, state_cloud.GlobalCloudLayout.Changes[0].Host, host.HostId)
	assert.Equal(t, state_cloud.GlobalCloudLayout.Changes[0].Application, base.AppName("testing"))
}

func TestPlanner_AddApplication__KillEmptyServers(t *testing.T) {
	state_configuration.GlobalConfigurationState.Init()
	state_cloud.GlobalCloudLayout.Init()

	appConfigSets := make([]base.AppConfigurationSet, 1)
	appConfigSets[0] = base.AppConfigurationSet{
		Version:1,
	}

	application := base.AppConfiguration{
		Name:"testing",
		MinDeploymentCount:0,
		TargetDeploymentCount:0,
		Needs: needs.AppNeeds{CpuNeeds:1, MemoryNeeds:1, NetworkNeeds:1},
		ConfigurationSets:appConfigSets,
	}
	state_configuration.GlobalConfigurationState.ConfigureApp(application)

	state_cloud.GlobalCloudLayout.Init()
	host := state_cloud.CloudLayoutElement{
		HostId:"host1",
		InstanceType:"t2.micro",
		IpAddress:"172.16.0.1",
		AvailableResources: base.InstanceResources{
			TotalMemoryResource:100,
			TotalNetworkResource:100,
			TotalCpuResource:100,
		},
	}

	state_cloud.GlobalCloudLayout.Current.AddHost("host1", host)

	doPlanInternal()

	assert.Equal(t, len(state_cloud.GlobalCloudLayout.Changes), 1)
	assert.Equal(t, state_cloud.GlobalCloudLayout.Changes[0].ChangeType, base.CHANGE_REQUEST__TERMINATE_SERVER)
	assert.Equal(t, state_cloud.GlobalCloudLayout.Changes[0].Host, host.HostId)
}

func TestPlanner__ChangesTimeOut(t *testing.T) {
	state_configuration.GlobalConfigurationState.Init()
	state_cloud.GlobalCloudLayout.Init()
	Init(base.TrainerConfigurationState{
		ChangeDefaultTimeout:100,
		ChangeSpawnTimeout:120,
	})
	/* This change will be removed */
	state_cloud.GlobalCloudLayout.AddChange(base.ChangeRequest{
		CreatedTime: time.Unix(1, 0),
		ChangeType:base.CHANGE_REQUEST__SPAWN_SERVER,
	})

	/* This change will be removed */
	state_cloud.GlobalCloudLayout.AddChange(base.ChangeRequest{
		CreatedTime: time.Unix(1, 0),
		ChangeType:base.CHANGE_REQUEST__TERMINATE_SERVER,
	})

	state_cloud.GlobalCloudLayout.AddChange(base.ChangeRequest{
		CreatedTime: time.Now(),
		ChangeType:base.CHANGE_REQUEST__TERMINATE_SERVER,
	})

	doCheckForTimedOutChanges()

	assert.Equal(t, len(state_cloud.GlobalCloudLayout.Changes), 1)
}

func TestPlanner__HostTimedOut(t *testing.T) {
	state_configuration.GlobalConfigurationState.Init()
	state_cloud.GlobalCloudLayout.Init()
	Init(base.TrainerConfigurationState{
		DeadHostTimeout:120,
	})

	/* This change will be removed */
	state_cloud.GlobalCloudLayout.Current.AddHost("host1", state_cloud.CloudLayoutElement{
		InstanceType:"TEST",
		IpAddress:"192.168.1.1",
		LastSeen:time.Unix(1,0),
		HostId:"host1",
	})

	state_cloud.GlobalCloudLayout.Current.AddHost("host2", state_cloud.CloudLayoutElement{
		InstanceType:"TEST",
		IpAddress:"192.168.1.1",
		LastSeen: time.Now(),
		HostId:"host2",
	})

	doCheckForTimeoutHosts()

	fmt.Printf("%+v", state_cloud.GlobalCloudLayout.Current.Layout)
	assert.Equal(t, len(state_cloud.GlobalCloudLayout.Current.Layout), 1)
}
