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

package cloud

import "orca/trainer/model"

type InstanceType string
type HostId string

type CloudEngine interface {
	SpawnInstanceSync(change *model.ChangeServer) *model.Host
	SpawnSpotInstanceSync(change *model.ChangeServer) *model.Host
	GetInstanceType(HostId) InstanceType
	TerminateInstance(HostId) bool
	GetHostInfo(HostId) (string, string, []model.SecurityGroup, bool, string)
	WasSpotInstanceTerminatedDueToPrice(spotRequestId string) (bool, string)
	GetIp(hostId string) string
	GetPem() string
	RegisterWithLb(hostId string, elb string)
	DeRegisterWithLb(hostId string, elb string)
	SanityCheckHosts(map[string]*model.Host)
	AddNameTag(newHostId string, appName string)
	RemoveNameTag(newHostId string, appName string)
	SetTag(newHostId string, tagKey string, tagValue string)
	GetTag(tagKey string, newHostId string) string
	BackupConfiguration(configuration string) bool
	CreateDataQueue(name string, rogueName string)
	MonitorDataQueue(name string) int
}
