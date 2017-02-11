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

package schedule

import (
	"testing"
	"time"
	"encoding/json"
	"fmt"
)

func TestWeekdayBasedDeploymentCount(t *testing.T) {
	schedule := DeploymentSchedule{}
	schedule.SetAll(100)

	res := schedule.Get(time.Now())
	if res != 100 {
		t.Error(res)
	}

	t1, _ := time.Parse(time.RFC3339Nano, "2016-11-26T00:15:44+00:00")
	w1, m1 := timeToWeekdayMinutes(t1)
	schedule.Set(w1, m1, 150)

	if schedule.Get(t1) != 150 {
		t.Error(schedule.Get(t1))
	}
	// 15 min after
	t2, _ := time.Parse(time.RFC3339Nano, "2016-11-26T00:30:44+00:00")
	if schedule.Get(t2) != 150 {
		t.Error(schedule.Get(t2))
	}
	// 15 min before
	t3, _ := time.Parse(time.RFC3339Nano, "2016-11-26T00:00:44+00:00")
	if schedule.Get(t3) != 150 {
		t.Error(schedule.Get(t3))
	}
	// 45 min after
	t4, _ := time.Parse(time.RFC3339Nano, "2016-11-26T01:00:44+00:00")
	if schedule.Get(t4) != 100 {
		t.Error(schedule.Get(t4))
	}
	// 45 min before
	t5, _ := time.Parse(time.RFC3339Nano, "2016-11-25T23:30:44+00:00")
	if schedule.Get(t5) != 100 {
		t.Error(schedule.Get(t5))
	}
}

func TestGenerateSchedule (t *testing.T){
	schedule := DeploymentSchedule{}
	schedule.SetAll(100)

	res, _ := json.MarshalIndent(schedule, "", "  ")
	var result = string(res)
	fmt.Printf("%s", result)

}
