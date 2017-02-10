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
	SpawnInstanceSync(InstanceType, string, []model.SecurityGroup) *model.Host
	SpawnSpotInstanceSync(InstanceType, string, []model.SecurityGroup) *model.Host
	GetInstanceType(HostId) InstanceType
	TerminateInstance(HostId) bool
	GetHostInfo(HostId) (string, string, []model.SecurityGroup)

	GetIp(hostId string) string

	GetPem() string
	RegisterWithLb(hostId string, elb string)
	DeRegisterWithLb(hostId string, elb string)
	//GetIp(HostId) base.IpAddr
	//UpdateLoadBalancers(hostId HostId, app base.AppName, version base.Version, event string)
}

