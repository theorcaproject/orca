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

package planner

import (
	"orca/trainer/configuration"
	"orca/trainer/state"
	"orca/trainer/model"
)

type PlanningChange struct {
	Id string
	Type string /* Create Server, Add/Remove Application */

	/* Creation or removal of application */
	HostId string
	ApplicationName string

	/* Creation or removal of server*/
	InstanceId string
	InstanceNeeds string
	RequiresReliableInstance bool
	Network string
	SecurityGroups []model.SecurityGroup
}

type Planner interface {
	Init()

	Plan (configurationStore configuration.ConfigurationStore, currentState state.StateStore) ([]PlanningChange)
}
