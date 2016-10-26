package state_needs_test

import (
	"testing"
	"gatoor/orca/rewriteTrainer/state/needs"
)


func prepareNeedsState() state_needs.AppsNeedState {
	return state_needs.AppsNeedState{}
}


func TestAppsNeedState_GetNeeds(t *testing.T) {
	needs := prepareNeedsState()

	needs.UpdateNeeds("app1", "0.1", state_needs.AppNeeds{
		CpuNeeds: state_needs.CpuNeeds(0.3),
		MemoryNeeds: state_needs.MemoryNeeds(1.0),
		NetworkNeeds: state_needs.NetworkNeeds(0.1),
	})

	_, err0 := needs.Get("unknown", "aa")
	if err0 == nil {
		t.Error("found an app that's not there")
	}
	_, err1 := needs.Get("app1", "aa")
	if err1 == nil {
		t.Error("found a version that's not there")
	}
	val, err2 := needs.Get("app1", "0.1")
	if err2 != nil {
		t.Error("did not find app/version")
	}
	if val.MemoryNeeds != 1.0 {
		t.Error("got wrong needs value")
	}
}


func TestAppsNeedState_GetAall(t *testing.T) {
	needs := prepareNeedsState()

	needs.UpdateNeeds("app1", "0.1", state_needs.AppNeeds{
		CpuNeeds: state_needs.CpuNeeds(0.1),
		MemoryNeeds: state_needs.MemoryNeeds(0.1),
		NetworkNeeds: state_needs.NetworkNeeds(0.1),
	})
	needs.UpdateNeeds("app1", "0.2", state_needs.AppNeeds{
		CpuNeeds: state_needs.CpuNeeds(0.2),
		MemoryNeeds: state_needs.MemoryNeeds(0.2),
		NetworkNeeds: state_needs.NetworkNeeds(0.2),
	})

	_, err0 := needs.GetAll("unknown")
	if err0 == nil {
		t.Error("found an app that's not there")
	}

	val, err2 := needs.GetAll("app1")
	if err2 != nil {
		t.Error("did not find app")
	}
	if len(val) != 2 {
		t.Error("didn't get all versions")
	}
	if val["0.1"].MemoryNeeds != 0.1 {
		t.Error("got wrong needs value")
	}
}



