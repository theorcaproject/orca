package schedule

import (
	"time"
)

const (
	CAUTION_INTERVAL = 0
	MINUTES_DELTA = 60
)


type WeekdaySchedule struct {
	Schedule WeekdayBased
}

type Minutes int

type WeekdayBased map[time.Weekday]map[Minutes]int

func (w WeekdayBased) get(t time.Time) int {
	day, minutes := timeToWeekdayMinutes(t)
	var max int
	for i := (minutes - CAUTION_INTERVAL * MINUTES_DELTA); i <= minutes + CAUTION_INTERVAL * MINUTES_DELTA; i += MINUTES_DELTA {
		if i >= 0 && i < 1440 {
			current := w[day][Minutes(i)]
			if current > max {
				max = current
			}
		} else if i < 0 {
			var dayBefore time.Weekday
			if day == time.Sunday {
				dayBefore = time.Saturday
			} else {
				dayBefore = day - 1
			}
			currentBefore := w[dayBefore][Minutes(1440 + i)]
			if currentBefore > max {
				max = currentBefore
			}
		} else if i >= 1440 {
			var dayAfter time.Weekday
			if day == time.Saturday {
				dayAfter = time.Sunday
			} else {
				dayAfter = day + 1
			}
			currentAfter := w[dayAfter][Minutes(i - 1440)]
			if currentAfter > max {
				max =currentAfter
			}
		}
	}
	return max
}

func (w WeekdayBased) set(day time.Weekday, minutes Minutes, val int) {
	if _, exists := w[day]; !exists {
		w[day] = make(map[Minutes]int)
	}
	w[day][minutes] = val
}

func (w WeekdayBased) setAll(val int) {
	for i := 0; i < 7; i++ {
		w[time.Weekday(i)] = make(map[Minutes]int)
		for m := 0; m < (24 * 60); m += MINUTES_DELTA {
			w[time.Weekday(i)][Minutes(m)] = val
		}
	}
}

func (w WeekdayBased) isEmpty() bool {
	if len(w) != 7 {
		return true
	}
	for i := 0; i < 7; i++ {
		for m := 0; m < (24 * 60); m += MINUTES_DELTA {
			if w[time.Weekday(i)][Minutes(m)] != 0 {
				return false
			}
		}
	}
	return true
}

//get weekday and minutes in MINUTES_DELTA increments. always rounded down
func timeToWeekdayMinutes(t time.Time) (time.Weekday, Minutes) {
	utcTime := t.UTC()
	w := utcTime.Weekday()
	m := utcTime.Hour() * 60 + utcTime.Minute()
	if m % MINUTES_DELTA != 0 {
		m = int(m / MINUTES_DELTA) * MINUTES_DELTA
	}
	return w, Minutes(m)
}

