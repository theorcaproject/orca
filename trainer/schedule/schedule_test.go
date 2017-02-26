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

func TestMidnightWorks(t *testing.T) {
	schedule := DeploymentSchedule{}
	schedule.SetAll(100)

	res := schedule.Get(time.Now())
	if res != 100 {
		t.Error(res)
	}

	t1, _ := time.Parse(time.RFC3339Nano, "2017-02-22T00:15:44+00:00")
	schedule.Set(time.Wednesday, Minutes(00), 150)
	if schedule.Get(t1) != 150 {
		t.Error(schedule.Get(t1))
	}
}

func TestBeforeMidnightWorks(t *testing.T) {
	schedule := DeploymentSchedule{}
	schedule.SetAll(100)

	res := schedule.Get(time.Now())
	if res != 100 {
		t.Error(res)
	}

	t1, _ := time.Parse(time.RFC3339Nano, "2017-02-22T23:15:44+00:00")
	schedule.Set(time.Wednesday, Minutes(1380), 150)
	if schedule.Get(t1) != 150 {
		t.Error(schedule.Get(t1))
	}
}

/*
            "1020": 5,
            "1080": 5,
            "1140": 5,
            "120": 5,
            "1200": 5,
            "1260": 5,
            "1320": 5,
            "1380": 5,
            "1440": 0,
            "180": 5,
            "240": 1,
            "300": 1,
            "360": 1,
            "420": 1,
            "480": 1,
            "540": 1,
            "600": 1,
            "660": 1,
            "720": 1,
            "780": 1,
            "840": 1,
            "900": 1,
            "960": 1
*/
func TestComplex(t *testing.T) {
	schedule := DeploymentSchedule{}
	schedule.SetAll(100)

	res := schedule.Get(time.Now())
	if res != 100 {
		t.Error(res)
	}

	schedule.Set(time.Wednesday, Minutes(0), 5)
	schedule.Set(time.Wednesday, Minutes(60), 5)
	schedule.Set(time.Wednesday, Minutes(120), 5)
	schedule.Set(time.Wednesday, Minutes(180), 5)
	schedule.Set(time.Wednesday, Minutes(240), 1)
	schedule.Set(time.Wednesday, Minutes(300), 1)
	schedule.Set(time.Wednesday, Minutes(360), 1)
	schedule.Set(time.Wednesday, Minutes(420), 1)
	schedule.Set(time.Wednesday, Minutes(480), 1)
	schedule.Set(time.Wednesday, Minutes(540), 1)
	schedule.Set(time.Wednesday, Minutes(600), 1)
	schedule.Set(time.Wednesday, Minutes(660), 1)
	schedule.Set(time.Wednesday, Minutes(720), 1)
	schedule.Set(time.Wednesday, Minutes(780), 1)
	schedule.Set(time.Wednesday, Minutes(840), 1)
	schedule.Set(time.Wednesday, Minutes(900), 1)
	schedule.Set(time.Wednesday, Minutes(960), 1)
	schedule.Set(time.Wednesday, Minutes(1020), 5)
	schedule.Set(time.Wednesday, Minutes(1080), 5)
	schedule.Set(time.Wednesday, Minutes(1140), 5)
	schedule.Set(time.Wednesday, Minutes(1200), 5)
	schedule.Set(time.Wednesday, Minutes(1260), 5)
	schedule.Set(time.Wednesday, Minutes(1320), 5)
	schedule.Set(time.Wednesday, Minutes(1380), 5)

	schedule.Set(time.Wednesday, Minutes(1440), 0)

	t1, _ := time.Parse(time.RFC3339Nano, "2017-02-22T23:15:44+00:00")
	if schedule.Get(t1) != 5 {
		t.Error(schedule.Get(t1))
	}

	t2, _ := time.Parse(time.RFC3339Nano, "2017-02-22T00:15:44+00:00")
	if schedule.Get(t2) != 5 {
		t.Error(schedule.Get(t2))
	}
}


func TestGenerateSchedule (t *testing.T){
	schedule := DeploymentSchedule{}
	schedule.SetAll(100)

	res, _ := json.MarshalIndent(schedule, "", "  ")
	var result = string(res)
	fmt.Printf("%s", result)

}
