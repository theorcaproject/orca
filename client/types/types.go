package types

import (
	"gatoor/orca/base"
	"fmt"
	"math/rand"
	"sync"
)

type ClientType string

const (
	DOCKER_CLIENT = ClientType("DOCKER_CLIENT")
	RAW_CLIENT = ClientType("RAW_CLIENT")
	TEST_CLIENT = ClientType("TEST_CLIENT")
)


type Configuration struct {
	Type ClientType
	TrainerPollInterval int
	AppStatusPollInterval int
	MetricsPollInterval int
	TrainerUrl string
	Port int
	HostId base.HostId
}

//type RetryCounter struct {
//	Install int
//	Run int
//	Query int
//	Stop int
//	Delete int
//}

type AppsState map[base.AppId]base.AppInfo
var appsStateMutex = &sync.Mutex{}
type AppsConfiguration map[base.AppId]base.AppConfiguration
var appsConfMutex = &sync.Mutex{}
//type AppsRetryCounter map[base.AppId]RetryCounter
//var appsRetryMutex = &sync.Mutex{}

type AppsMetricsById map[base.AppId]map[string]base.AppStats
var appsMetricsIdMutex = &sync.Mutex{}

func GenerateId(app base.AppName) base.AppId {
	return base.AppId(fmt.Sprintf("%s_%d", app, rand.Int31()))
}

func (a *AppsMetricsById) Add(id base.AppId, time string, stats base.AppStats) {
	appsMetricsIdMutex.Lock()
	defer appsMetricsIdMutex.Unlock()
	if _, exists := (*a)[id]; !exists {
		(*a)[id] = make(map[string]base.AppStats)
	}
	(*a)[id][time] = stats
}

func (a *AppsMetricsById) Clear() {
	appsMetricsIdMutex.Lock()
	defer appsMetricsIdMutex.Unlock()
	(*a) = make(map[base.AppId]map[string]base.AppStats)
}

func (a *AppsMetricsById) All() map[base.AppId]map[string]base.AppStats {
	appsMetricsIdMutex.Lock()
	defer appsMetricsIdMutex.Unlock()
	res := (*a)
	return res
}

func (a *AppsConfiguration) Add(id base.AppId, conf base.AppConfiguration) {
	appsConfMutex.Lock()
	defer appsConfMutex.Unlock()
	(*a)[id] = conf
}

func (a *AppsConfiguration) Remove(id base.AppId) {
	appsConfMutex.Lock()
	defer appsConfMutex.Unlock()
	delete((*a), id)
}

func (a *AppsConfiguration) Get(id base.AppId) base.AppConfiguration {
	appsConfMutex.Lock()
	defer appsConfMutex.Unlock()
	conf := (*a)[id]
	return conf
}

func (a *AppsState) Add(id base.AppId, info base.AppInfo) {
	appsStateMutex.Lock()
	defer appsStateMutex.Unlock()
	(*a)[id] = info
}

func (a *AppsState) Remove(id base.AppId) {
	appsStateMutex.Lock()
	defer appsStateMutex.Unlock()
	delete((*a), id)
}

func (a *AppsState) Get(id base.AppId) base.AppInfo {
	appsStateMutex.Lock()
	defer appsStateMutex.Unlock()
	info := (*a)[id]
	return info
}

func (a *AppsState) GetAll(name base.AppName) []base.AppInfo {
	appsStateMutex.Lock()
	defer appsStateMutex.Unlock()
	var infos []base.AppInfo
	for _, info := range (*a) {
		if info.Name == name {
			infos = append(infos, info)
		}
	}
	return infos
}

func (a *AppsState) All() []base.AppInfo {
	appsStateMutex.Lock()
	defer appsStateMutex.Unlock()
	var res []base.AppInfo
	for _, info := range (*a) {
		res = append(res, info)
	}
	return res
}

func (a *AppsState) GetAllWithVersion(name base.AppName, version base.Version) []base.AppInfo {
	appsStateMutex.Lock()
	defer appsStateMutex.Unlock()
	var infos []base.AppInfo
	for _, info := range (*a) {
		if info.Name == name && info.Version == version {
			infos = append(infos, info)
		}
	}
	return infos
}

func (a *AppsState) Set(id base.AppId, status base.Status) {
	appsStateMutex.Lock()
	defer appsStateMutex.Unlock()
	info := (*a)[id]
	info.Status = status
	(*a)[id] = info
}