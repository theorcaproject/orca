package state_cloud

import (
	"sync"
	"errors"
	"gatoor/orca/rewriteTrainer/base"
)

var GlobalCloudLayout CloudLayoutAll
var cloudLayoutMutex = &sync.Mutex{}
type CloudLayout map[base.HostId]CloudLayoutElement

type CloudLayoutAll struct {
	Current CloudLayout
	Desired CloudLayout
	Vision CloudLayout
}

type CloudLayoutElement struct {
	HostId base.HostId
	IpAddress base.IpAddr
	HabitatVersion base.Version
	Apps map[base.AppName]AppsVersion
}

type AppsVersion struct {
	Version base.Version
	DeploymentCount int
}

func (c *CloudLayoutAll) Init() {
	c.Current = CloudLayout{}
	c.Desired = CloudLayout{}
	c.Vision = CloudLayout{}
}

func (c *CloudLayoutAll) Snapshot() CloudLayoutAll {
	cloudLayoutMutex.Lock()
	defer cloudLayoutMutex.Unlock()
	res := *c
	return res
}

func (c *CloudLayout) GetHost(host base.HostId) (CloudLayoutElement, error) {
	if _, exists :=(*c)[host]; !exists {
		return CloudLayoutElement{}, errors.New("No such host")
	}
	return (*c)[host], nil
}

func (c *CloudLayout) AddHost(host base.HostId, elem CloudLayoutElement) {
	cloudLayoutMutex.Lock()
	defer cloudLayoutMutex.Unlock()
	(*c)[host] = elem
}

func (c *CloudLayout) AddEmptyHost(host base.HostId) {
	cloudLayoutMutex.Lock()
	defer cloudLayoutMutex.Unlock()
	(*c)[host] = CloudLayoutElement{
		HostId: host,
		IpAddress: "",
		HabitatVersion: "",
		Apps: make(map[base.AppName]AppsVersion),
	}
}

func (c *CloudLayout) RemoveHost(host base.HostId) {
	cloudLayoutMutex.Lock()
	defer cloudLayoutMutex.Unlock()
	delete(*c, host)
}
func (c *CloudLayout) AddApp(host base.HostId, app base.AppName, version base.Version, count int) {
	cloudLayoutMutex.Lock()
	defer cloudLayoutMutex.Unlock()
	if val, exists := (*c)[host]; exists {
		if _, ex := val.Apps[app]; !ex {
			val.Apps[app] = AppsVersion{version, count}
		}
	}
}
func (c *CloudLayout) RemoveApp(host base.HostId, app base.AppName) {
	cloudLayoutMutex.Lock()
	defer cloudLayoutMutex.Unlock()
	if val, exists := (*c)[host]; exists {
		delete(val.Apps, app)
	}
}
