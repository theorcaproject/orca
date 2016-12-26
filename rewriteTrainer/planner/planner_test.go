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
	"testing"
	"gatoor/orca/base"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/rewriteTrainer/state/configuration"
	"gatoor/orca/rewriteTrainer/example"
	"gatoor/orca/rewriteTrainer/state/needs"
	"gatoor/orca/rewriteTrainer/cloud"
	"gatoor/orca/rewriteTrainer/needs"
	"gatoor/orca/rewriteTrainer/tracker"
	"time"
)

func TestPlannerQueue_AllEmpty(t *testing.T) {
	queue := NewPlannerQueue()

	if queue.AllEmpty() == false {
		t.Error("should be all empty")
	}

	queue.Add("host1", "app1", state_cloud.AppsVersion{1, 2})

	if queue.AllEmpty() == true {
		t.Error("should have elements")
	}
}
