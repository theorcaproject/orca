package schedule

import (
	"testing"
	"time"
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