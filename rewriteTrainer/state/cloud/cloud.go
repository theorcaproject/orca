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

package state_cloud

import (
	"github.com/satori/go.uuid"
	"errors"
	"gatoor/orca/base"
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/rewriteTrainer/db"
	"gatoor/orca/rewriteTrainer/state/needs"
	"sync"
	"gatoor/orca/rewriteTrainer/needs"
	"fmt"
	"sort"
	"gatoor/orca/rewriteTrainer/cloud"
	"gatoor/orca/rewriteTrainer/state/configuration"
	"time"
)

var StateCloudLogger = Logger.LoggerWithField(Logger.Logger, "module", "state_cloud")

var GlobalCloudLayout CloudLayoutAll
var cloudLayoutMutex = &sync.Mutex{}

type CloudLayout struct {
	Layout map[base.HostId]CloudLayoutElement
}

type PlannedCloudLayout struct {
	Layout map[base.HostId]PlannedCloudLayoutElement
}

type ResourceObjList []ResourceObj

type ResourceObj struct {
	HostId                     base.HostId
	CombinedAvailableResources int
}

type CloudLayoutAll struct {
	Current CloudLayout
	Changes []base.ChangeRequest
}

const (
	HOST_NORMAL="HOST_NORMAL"
	HOST_VANISHED="HOST_VANISHED"
	HOST_PLANNING_TERMINATING="HOST_PLANNING_TERMINATING"
	HOST_PLANNING_TERMINATED="HOST_PLANNING_TERMINATED"
)


//TODO: Why is this object shared between desired and current. Current has state information.
type CloudLayoutElement struct {
	HostId         base.HostId
	IpAddress      base.IpAddr
	InstanceType   base.InstanceType
	SafeInstance   base.SafeInstance
	HabitatVersion base.Version
	Apps           map[base.AppName]AppsVersion

	LastSeen       time.Time
	FirstSeen      time.Time
	HostState      string

	AvailableResources base.InstanceResources
}

type PlannedCloudLayoutElement struct {
	HostId         base.HostId
	Apps           map[base.AppName]PlannedAppsVersion
}

type AppsVersion struct {
	Version         base.Version
	DeploymentCount base.DeploymentCount
	AppState      string

	StatisticPoint  base.AppStats
	StatisticPointTimestamp time.Time
}

type PlannedAppsVersion struct {
	Version         base.Version
	DeploymentCount base.DeploymentCount
}

func (c *CloudLayoutAll) Init() {
	c.Current = CloudLayout{
		Layout: make(map[base.HostId]CloudLayoutElement),
	}

	c.Changes = make([]base.ChangeRequest, 0)
}

func (object *CloudLayoutAll) AddChange(change base.ChangeRequest){
	change.Id = uuid.NewV4().String()

	if change.CreatedTime.Unix() == 0 {
		change.CreatedTime = time.Now()
	}
	object.Changes = append(object.Changes, change)
}

func (object *CloudLayoutAll) GetChanges(host base.HostId) []base.ChangeRequest{
	changes := make([]base.ChangeRequest, 0)
	for _, change := range object.Changes {
		if change.Host == host {
			changes = append(changes, change)
		}
	}

	return changes
}

func (object *CloudLayoutAll) DeleteChange(id string){
	changesMinusHost := make([]base.ChangeRequest, 0)
	for _, change := range object.Changes {
		if change.Id != id {
			changesMinusHost= append(changesMinusHost, change)
		}
	}

	object.Changes = changesMinusHost
}

func (object *CloudLayoutAll) PopChanges(host base.HostId) []base.ChangeRequest{
	ret := make([]base.ChangeRequest, 0)
	changesMinusHost := make([]base.ChangeRequest, 0)
	for _, change := range object.Changes {

		if change.Host == host {
			ret = append(ret, change)
		}else{
			changesMinusHost = append(ret, change)
		}
	}

	object.Changes = changesMinusHost
	return ret
}

func (c *CloudLayoutAll) InitBaseInstances(){
	if len(GlobalCloudLayout.Current.Layout) < int(state_configuration.GlobalConfigurationState.CloudProvider.MinInstances) {
		for i := len(GlobalCloudLayout.Current.Layout); i < int(state_configuration.GlobalConfigurationState.CloudProvider.MinInstances); i++ {
			cloud.CurrentProvider.SpawnInstanceSync(state_configuration.GlobalConfigurationState.CloudProvider.BaseInstanceType)
		}
	}
}

func (c *CloudLayoutAll) CheckCheckinTimeout(){
	for _, host := range c.Current.Layout{
		if host.LastSeen.Before(time.Now().UTC().Add(-600)) && host.HostState == HOST_NORMAL{
			/* We seem to have lost a host for some reason, in the next planning stage we will need to fix this */
			host.HostState = HOST_VANISHED
		}
	}
}

func (c *CloudLayoutAll) Snapshot() CloudLayoutAll {
	cloudLayoutMutex.Lock()
	defer cloudLayoutMutex.Unlock()
	res := *c
	return res
}

func (c *CloudLayout) GetHost(host base.HostId) (CloudLayoutElement, error) {
	cloudLayoutMutex.Lock()
	defer cloudLayoutMutex.Unlock()
	if _, exists := (*c).Layout[host]; !exists {
		return CloudLayoutElement{}, errors.New("No such host")
	}
	return (*c).Layout[host], nil
}

func (c *CloudLayoutElement) HostHasApp(appName base.AppName) bool {
	cloudLayoutMutex.Lock()
	defer cloudLayoutMutex.Unlock()
	_, exists := c.Apps[appName]
	return exists
}


func (c *CloudLayout) FindHostsWithApp(appName base.AppName) map[base.HostId]bool {
	cloudLayoutMutex.Lock()
	defer cloudLayoutMutex.Unlock()
	hosts := make(map[base.HostId]bool)
	//sorting to make it testable
	var hostsSort []string
	for k := range (*c).Layout {
		hostsSort = append(hostsSort, string(k))
	}
	sort.Strings(hostsSort)

	for _, hostId := range hostsSort {
		for app := range (*c).Layout[base.HostId(hostId)].Apps {
			if app == appName {
				hosts[base.HostId(hostId)] = true
			}
		}
	}

	StateCloudLogger.Debugf("Found Hosts %+v with App '%s'", hosts, appName)
	return hosts
}

func (c *CloudLayout) AddHost(host base.HostId, elem CloudLayoutElement) {
	StateCloudLogger.Infof("Adding host '%s': '%+v'", host, elem)
	cloudLayoutMutex.Lock()
	defer cloudLayoutMutex.Unlock()

	if _, ok := (*c).Layout[host]; !ok {
		db.Audit.Insert__AuditEvent(db.AuditEvent{Details:map[string]string{
			"message": fmt.Sprintf("New host discovered, '%s': '%+v'", host, elem),
			"subsystem": "cloud",
			"level": "info",
		}})
	}

	if (elem.InstanceType == "") {
		elem.InstanceType = cloud.CurrentProvider.GetInstanceType(host)
	}

	if (elem.IpAddress == "") {
		elem.IpAddress = cloud.CurrentProvider.GetIp(host)
	}

	(*c).Layout[host] = elem
}

func (c *PlannedCloudLayout) AddHost(host base.HostId, elem PlannedCloudLayoutElement) {
	StateCloudLogger.Infof("Adding host '%s': '%+v'", host, elem)
	cloudLayoutMutex.Lock()
	defer cloudLayoutMutex.Unlock()
	(*c).Layout[host] = elem
}

func (c *CloudLayout) DeploymentCount(app base.AppName, version base.Version) (base.DeploymentCount, base.DeploymentCount) {
	cloudLayoutMutex.Lock()
	defer cloudLayoutMutex.Unlock()
	var versionCount base.DeploymentCount = 0
	var total base.DeploymentCount = 0
	for _, host := range (*c).Layout {
		if app, exists := host.Apps[app]; exists {
			if app.Version == version {
				versionCount += app.DeploymentCount
			}
			total += app.DeploymentCount
		}
	}
	return versionCount, total
}

func (c *CloudLayout) AddEmptyHost(host base.HostId) {
	StateCloudLogger.Infof("Adding empty host '%s'", host)
	entity := CloudLayoutElement{
		HostId: host,
		HabitatVersion: 0,
		Apps: make(map[base.AppName]AppsVersion),
	}
	c.AddHost(host, entity)
}


func (c *PlannedCloudLayout) AddEmptyHost(host base.HostId) {
	StateCloudLogger.Infof("Adding empty host '%s'", host)
	entity := PlannedCloudLayoutElement{
		HostId: host,
		Apps: make(map[base.AppName]PlannedAppsVersion),
	}
	c.AddHost(host, entity)
}

func (c *PlannedCloudLayout) Wipe() {
	StateCloudLogger.Infof("Wipe")
	cloudLayoutMutex.Lock()
	defer cloudLayoutMutex.Unlock()
	(*c).Layout = make(map[base.HostId]PlannedCloudLayoutElement)
}

func (c *CloudLayout) RemoveHost(host base.HostId) {
	StateCloudLogger.Infof("Removing host '%s'", host)
	cloudLayoutMutex.Lock()
	defer cloudLayoutMutex.Unlock()
	delete(c.Layout, host)
}

func (c *CloudLayout) AddApp(host base.HostId, app base.AppName, version base.Version, count base.DeploymentCount) {
	StateCloudLogger.Infof("Adding App %s:%d to host '%s' %d times", app, version, host, count)
	cloudLayoutMutex.Lock()
	defer cloudLayoutMutex.Unlock()
	if val, exists := (*c).Layout[host]; exists {
		if val.Apps == nil {
			val.Apps = make(map[base.AppName]AppsVersion)
		}

		if _, ex := val.Apps[app]; !ex {
			val.Apps[app] = AppsVersion{
				Version:version,
				DeploymentCount:count,
			}
		}

		(*c).Layout[host] = val
	}
}

func (c *CloudLayout) RemoveApp(host base.HostId, app base.AppName) {
	StateCloudLogger.Infof("Removing app '%s' from host '%s'", app, host)
	cloudLayoutMutex.Lock()
	defer cloudLayoutMutex.Unlock()
	if val, exists := (*c).Layout[host]; exists {
		delete(val.Apps, app)
	}
}

func (c *CloudLayout) Needs(app base.AppName) needs.AppNeeds {
	cloudLayoutMutex.Lock()
	defer cloudLayoutMutex.Unlock()
	ns := needs.AppNeeds{}
	for _, elem := range (*c).Layout {
		if appObj, exists := elem.Apps[app]; exists {
			currentNeeds, err := state_needs.GlobalAppsNeedState.Get(app, appObj.Version)
			if err == nil {
				ns.CpuNeeds += needs.CpuNeeds(int(appObj.DeploymentCount) * int(currentNeeds.CpuNeeds))
				ns.MemoryNeeds += needs.MemoryNeeds(int(appObj.DeploymentCount) * int(currentNeeds.MemoryNeeds))
				ns.NetworkNeeds += needs.NetworkNeeds(int(appObj.DeploymentCount) * int(currentNeeds.NetworkNeeds))
			}
		}
	}
	return ns
}

func (c *CloudLayout) AllNeeds() needs.AppNeeds {
	cloudLayoutMutex.Lock()
	defer cloudLayoutMutex.Unlock()
	ns := needs.AppNeeds{}
	for _, elem := range (*c).Layout {
		for appName, appObj := range elem.Apps {
			currentNeeds, err := state_needs.GlobalAppsNeedState.Get(appName, appObj.Version)
			if err == nil {
				ns.CpuNeeds += needs.CpuNeeds(int(appObj.DeploymentCount) * int(currentNeeds.CpuNeeds))
				ns.MemoryNeeds += needs.MemoryNeeds(int(appObj.DeploymentCount) * int(currentNeeds.MemoryNeeds))
				ns.NetworkNeeds += needs.NetworkNeeds(int(appObj.DeploymentCount) * int(currentNeeds.NetworkNeeds))
			}
		}
	}
	return ns
}

func (c *CloudLayout) UpdateHost(hostInfo base.HostInfo, stats base.MetricsWrapper) {
	apps := make(map[base.AppName]AppsVersion)

	for _, val := range hostInfo.Apps {
		newApp := AppsVersion{
			DeploymentCount:1,
			Version:val.Version,
		}

		for _, stat := range stats.AppMetrics[val.Name][newApp.Version]{
			newApp.StatisticPoint = stat
			newApp.StatisticPointTimestamp = time.Now()
			break
		}

		if newApp.AppState != string(val.Status) {
			//TODO: Create an audit event here
		}

		newApp.AppState = string(val.Status)
		apps[val.Name] = newApp
	}

	existingHostObject, err := c.GetHost(hostInfo.HostId)
	if err != nil {
		/* Most likely could not find the host, we need to create a new object */
		existingHostObject = CloudLayoutElement{
			HostId: hostInfo.HostId,
		}

		/* Figure out the instance type by host Id */
		instance_type := cloud.CurrentProvider.GetInstanceType(hostInfo.HostId)
		existingHostObject.AvailableResources = cloud.CurrentProvider.GetResources(instance_type)
		existingHostObject.IpAddress = cloud.CurrentProvider.GetIp(hostInfo.HostId)
		existingHostObject.FirstSeen = time.Now()
		existingHostObject.HostState = HOST_NORMAL
	}

	existingHostObject.Apps = apps
	existingHostObject.LastSeen = time.Now()

	StateCloudLogger.Infof("UpdateHost for host '%s': %+v", hostInfo.HostId, hostInfo)
	c.AddHost(hostInfo.HostId, existingHostObject)
}

func (a *CloudLayoutElement) HostHasResourcesForApp(ns needs.AppNeeds) bool {
	/* If there are changes that have not been picked up we need to check them first */
	if int(a.AvailableResources.TotalCpuResource - a.AvailableResources.UsedCpuResource) >= int(ns.CpuNeeds) &&
		int(a.AvailableResources.TotalMemoryResource - a.AvailableResources.UsedMemoryResource) >= int(ns.MemoryNeeds) &&
		int(a.AvailableResources.TotalNetworkResource - a.AvailableResources.UsedNetworkResource) >= int(ns.NetworkNeeds) {
		return true
	}
	return false
}

func (p ResourceObjList) Len() int {
	return len(p)
}
func (p ResourceObjList) Less(i, j int) bool {
	return p[i].CombinedAvailableResources < p[j].CombinedAvailableResources
}
func (p ResourceObjList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
