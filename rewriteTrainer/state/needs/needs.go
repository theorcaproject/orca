package state_needs

import (
	"gatoor/orca/rewriteTrainer/base"
	"sync"
	"errors"
)

var needsStateMutex = &sync.Mutex{}
var GlobalAppsNeedState AppsNeedState

type AppsNeedState map[base.AppName]AppNeedVersion

type AppNeedVersion map[base.Version]AppNeeds


type Needs float32

func (m Needs) Get() (Needs){
	return m
}

func (m Needs) Set(n float32) {
	if n < 0 {
		m = 0.0
	} else if n > 1 {
		m = 1.00
	} else {
		m = Needs(n)
	}
}


type MemoryNeeds Needs
type CpuNeeds Needs
type NetworkNeeds Needs

type AppNeeds struct {
	MemoryNeeds MemoryNeeds
	CpuNeeds CpuNeeds
	NetworkNeeds NetworkNeeds
}

func (a AppsNeedState) GetAll(app base.AppName) (AppNeedVersion, error) {
	needsStateMutex.Lock()
	defer needsStateMutex.Unlock()
	if _, exists := a[app]; !exists {
		return AppNeedVersion{}, errors.New("No such App")
	}
	res := a[app]
	return res, nil
}

func (a AppsNeedState) Get(app base.AppName, version base.Version) (AppNeeds, error) {
	needsStateMutex.Lock()
	defer needsStateMutex.Unlock()
	if _, exists := a[app]; !exists {
		return AppNeeds{}, errors.New("No such App")
	}
	if _, exists := a[app][version]; !exists {
		return AppNeeds{}, errors.New("No such Version")
	}
	res := a[app][version]
	return res, nil
}

func (a AppsNeedState) UpdateNeeds(app base.AppName, version base.Version, needs AppNeeds) {
	needsStateMutex.Lock()
	defer needsStateMutex.Unlock()
	if _, exists := a[app]; !exists {
		a[app] = make(map[base.Version]AppNeeds)
	}
	a[app][version] = needs
}

func (a AppsNeedState) Snapshot() AppsNeedState {
	needsStateMutex.Lock()
	defer needsStateMutex.Unlock()
	res := a
	return res
}




