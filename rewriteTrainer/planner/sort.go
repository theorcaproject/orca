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

package planner

import (
	"gatoor/orca/base"
	"gatoor/orca/rewriteTrainer/needs"
)


/*
	Sort the available hosts of the cluster...
	TRY_TO_REMOVE_HOSTS=true: Get the host with the least amount of available resources - this allows us to shut an instance down.
	TRY_TO_REMOVE_HOSTS=false: Get the host with a lot of spare resources - balance the load on the cluster
*/

type HostPair struct {
	Key base.HostId
	Val base.InstanceResources
}

type HostList []HostPair
type HostLess func(p HostList, i, j int) bool


func (p HostList) Len() int { return len(p) }
func (p HostList) Less(i, j int) bool { return (
	int(p[i].Val.TotalCpuResource - p[i].Val.UsedCpuResource) +
		int(p[i].Val.TotalMemoryResource - p[i].Val.UsedMemoryResource) +
		int(p[i].Val.TotalNetworkResource - p[i].Val.UsedNetworkResource)) < (
	int(p[j].Val.TotalCpuResource - p[j].Val.UsedCpuResource) +
		int(p[j].Val.TotalMemoryResource - p[j].Val.UsedMemoryResource) +
		int(p[j].Val.TotalNetworkResource - p[j].Val.UsedNetworkResource))}
func (p HostList) Swap(i, j int){ p[i], p[j] = p[j], p[i] }

/*
	Sort the apps by Needs: High needs => should be planned first
 */


type ConfPair struct {
	Key base.AppName
	Val needs.AppNeeds
}

type ConfList []ConfPair

func (p ConfList) Len() int { return len(p) }
func (p ConfList) Less(i, j int) bool { return (int(p[i].Val.CpuNeeds) + int(p[i].Val.MemoryNeeds) + int(p[i].Val.NetworkNeeds)) > (int(p[j].Val.CpuNeeds) + int(p[j].Val.MemoryNeeds) + int(p[j].Val.NetworkNeeds))}
func (p ConfList) Swap(i, j int){ p[i], p[j] = p[j], p[i] }

