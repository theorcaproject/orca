package cloud

import "gatoor/orca/base"

type CloudSnapshot struct {

}

type OrcaCloud struct {
	Current CloudLayout
	Desired CloudLayout
}

type CloudLayout struct {
	Layout map[base.HostId]CloudLayoutElement
}

type CloudLayoutElement struct {
	HabitatVersion base.Version
	IpAddress base.IpAddr
	AppsVersion map[base.AppName]base.Version
}

func (oc OrcaCloud) UpdateDesired(update base.TrainerUpdate) {
	if _, exists := oc.Desired.Layout[update.TargetHostId]; !exists {
		oc.Desired.Layout[update.TargetHostId] = CloudLayoutElement{
			HabitatVersion: "0",
			IpAddress: "unknown",
			AppsVersion: make(map[base.AppName]base.Version),
		}
	}
	elem := oc.Desired.Layout[update.TargetHostId]
	elem.HabitatVersion = update.HabitatConfiguration.Version
	for _, appConf := range update.AppsConfiguration {
		elem.AppsVersion[appConf.Name] = appConf.Version
	}
	oc.Desired.Layout[update.TargetHostId] = elem
}

func (oc OrcaCloud) UpdateCurrent(hostInfo base.HostInfo) {
	if _, exists := oc.Current.Layout[hostInfo.HostId]; !exists {
		oc.Current.Layout[hostInfo.HostId] = CloudLayoutElement{
			HabitatVersion: "0",
			IpAddress: hostInfo.IpAddr,
			AppsVersion: make(map[base.AppName]base.Version),
		}
	}
	elem := oc.Current.Layout[hostInfo.HostId]
	elem.HabitatVersion = hostInfo.HabitatInfo.Version
	for _, app := range hostInfo.Apps {
		elem.AppsVersion[app.Name] = app.Version
	}
	oc.Current.Layout[hostInfo.HostId] = elem
}

func (c CloudLayout) DeleteHost(id base.HostId) {
	delete(c.Layout, id)
}

func (c CloudLayout) AddHost(id base.HostId, ipAddr base.IpAddr) {
	c.Layout[id] = CloudLayoutElement{
		HabitatVersion: "0",
		IpAddress: ipAddr,
		AppsVersion: make(map[base.AppName]base.Version),
	}
}

func (c CloudLayout) DeleteApp(hostId base.HostId, appName base.AppName) {
	delete(c.Layout[hostId].AppsVersion, appName)
}

type CloudProvider interface {
	NewInstance() (base.HostId, base.IpAddr)
}
