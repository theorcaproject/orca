package cloud

import "gatoor/orca/base"

type OrcaCloud struct {
	Current CloudLayout
	Desired CloudLayout
}

type CloudLayout struct {
	Layout map[string]CloudLayoutElement
}

type CloudLayoutElement struct {
	HabitatVersion string
	AppsVersion map[string]string
}

func (oc OrcaCloud) UpdateDesired(update base.TrainerUpdate) {
	if _, exists := oc.Desired.Layout[update.TargetHostId]; !exists {
		oc.Desired.Layout[update.TargetHostId] = CloudLayoutElement{"", make(map[string]string),}
	}
	elem := oc.Desired.Layout[update.TargetHostId]
	elem.HabitatVersion = update.HabitatConfiguration.Version
	for _, appConf := range update.AppsConfiguration {
		elem.AppsVersion[appConf.AppName] = appConf.Version
	}
	oc.Desired.Layout[update.TargetHostId] = elem
}

func (oc OrcaCloud) UpdateCurrent(hostInfo base.HostInfo) {
	if _, exists := oc.Current.Layout[hostInfo.Id]; !exists {
		oc.Current.Layout[hostInfo.Id] = CloudLayoutElement{"", make(map[string]string),}
	}
	elem := oc.Current.Layout[hostInfo.Id]
	elem.HabitatVersion = hostInfo.HabitatInfo.Version
	for _, app := range hostInfo.Apps {
		elem.AppsVersion[app.Name] = app.CurrentVersion
	}
	oc.Current.Layout[hostInfo.Id] = elem
}

func (c CloudLayout) DeleteHost(id string) {
	delete(c.Layout, id)
}

func (c CloudLayout) AddHost(id string) {
	elem := CloudLayoutElement{}
	elem.HabitatVersion = "0"
	elem.AppsVersion = make(map[string]string)
	c.Layout[id] = elem
}

func (c CloudLayout) DeleteApp(hostId string, appName string) {
	delete(c.Layout[hostId].AppsVersion, appName)
}

type CloudProvider interface {
	NewInstance() string
}
