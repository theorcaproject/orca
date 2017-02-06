package schedule

import "time"

type DeploymentSchedule WeekdaySchedule


func (w DeploymentSchedule) isEmpty() bool {
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
