/*
Copyright Alex Mack (al9mack@gmail.com) and Michael Lawson (michael@sphinix.com)
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

package model

import (
	"strconv"
	"orca/trainer/schedule"
)

type ChangeApplication struct {
	Id        string
	Type      string
	HostId    string
	Time      string
	Name      string

	/* This could be nil/empty if not needed */
	AppConfig VersionConfig
}

type ChangeServer struct {
	Id                       string
	Type                     string
	Time                     string
	NewHostId                string
	RequiresReliableInstance bool
	Network                  string
	SecurityGroups           []SecurityGroup

	// Internal Status Information
	InstanceLaunched         bool
	InstalledPackages        bool

	//Load balancer add task
	LoadBalancerName         string
	LoadBalancerAppTarget    string
	LoadBalancerAppVersion   string
}

type HostResources struct {

}

type Application struct {
	Name     string
	State    string
	Version  string
	ChangeId string

	Metrics  Metric
}

type ApplicationStateFromHost struct {
	Name        string
	Application Application
}

type Metric struct {
	CpuUsage             int64
	MemoryUsage          int64
	NetworkUsage         int64
	HardDiskUsage        int64
	HardDiskUsagePercent int64
}

type HostCheckinDataPackage struct {
	State          []ApplicationStateFromHost
	ChangesApplied map[string]bool
	HostMetrics    Metric
}

type Host struct {
	Id                          string
	CloudId                     string
	LastSeen                    string
	FirstSeen                   string
	State                       string
	Network                     string
	Ip                          string
	Apps                        []Application
	Changes                     []ChangeApplication
	Resources                   HostResources
	SpotInstance                bool
	SecurityGroups              []SecurityGroup
	NumberOfChangeFailuresInRow int64

	InstanceType string
}

func (host *Host) HasApp(name string) bool {
	for _, runningApplicationState := range host.Apps {
		if (runningApplicationState.Name == name && runningApplicationState.State == "running") {
			return true
		}
	}
	return false;
}

func (host *Host) GetChange(id string) *ChangeApplication {
	for _, change := range host.Changes {
		if (change.Id == id) {
			return &change
		}
	}
	return nil;
}

func (host *Host) HasAppWithSameVersion(name string, version string) bool {
	for _, runningApplicationState := range host.Apps {
		if (runningApplicationState.Name == name && runningApplicationState.Version == version && runningApplicationState.State == "running") {
			return true
		}
	}
	return false;
}

type DockerConfig struct {
	Username   string
	Password   string
	Email      string
	Server     string
	Tag        string
	Repository string
	Reference  string
}

type PortMapping struct {
	HostPort      string
	ContainerPort string
}

type VolumeMapping struct {
	HostPath      string
	ContainerPath string
}

type File struct {
	HostPath           string
	Base64FileContents string
}

type EnvironmentVariable struct {
	Key   string
	Value string
}

type Needs float32
type MemoryNeeds Needs
type CpuNeeds Needs
type NetworkNeeds Needs

type AppNeeds struct {
	MemoryNeeds  MemoryNeeds
	CpuNeeds     CpuNeeds
	NetworkNeeds NetworkNeeds
}

type LoadBalancerEntry struct {
	Domain string
}

type SecurityGroup struct {
	Group string
}

type ApplicationChecks struct {
	Type string /* Either HTTP or TCP */
	Goal string /* Either a port or uri */
}

type AffinityTag struct {
	Tag string
}

type VersionConfig struct {
	Version              string
	DockerConfig         DockerConfig
	Needs                AppNeeds
	LoadBalancer         []LoadBalancerEntry
	Network              string
	SecurityGroups       []SecurityGroup
	PortMappings         []PortMapping
	VolumeMappings       []VolumeMapping
	EnvironmentVariables []EnvironmentVariable
	Files                []File
	Checks               []ApplicationChecks
	Affinity             []AffinityTag
}

type ApplicationConfiguration struct {
	Name               string
	MinDeployment      int
	DesiredDeployment  int
	DisableSchedule    bool
	DeploymentSchedule schedule.DeploymentSchedule
	Config             map[string]VersionConfig

	Enabled            bool
}

func (app *ApplicationConfiguration) GetLatestVersion() string {
	version := 0
	for v, _ := range app.Config {
		iversion, _ := strconv.Atoi(v)
		if iversion > version {
			version = iversion
		}
	}

	return strconv.Itoa(version)
}

func (app *ApplicationConfiguration) GetNextVersion() string {
	ivalue, _ := strconv.Atoi(app.GetLatestVersion())
	return strconv.Itoa(ivalue + 1)
}

func (app *ApplicationConfiguration) GetLatestConfiguration() (VersionConfig) {
	last_version := app.GetLatestVersion()
	return app.Config[last_version]
}


