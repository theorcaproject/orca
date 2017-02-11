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

import "time"

type DeploymentSchedule WeekdaySchedule


func (w DeploymentSchedule) IsEmpty() bool {
	return w.Schedule.isEmpty()
}

func (w DeploymentSchedule) Get(time time.Time) int {
	return w.Schedule.get(time)
}

func (w *DeploymentSchedule) Set(day time.Weekday, minutes Minutes, ns int) {
	if len(w.Schedule) == 0 {
		w.Schedule = make(map[time.Weekday]map[Minutes]int)
	}
	w.Schedule.set(day, minutes, ns)
}

func (w *DeploymentSchedule) SetAll(ns int) {
	if len(w.Schedule) == 0 {
		w.Schedule = make(map[time.Weekday]map[Minutes]int)
	}
	w.Schedule.setAll(ns)
}
