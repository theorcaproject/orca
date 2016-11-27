package needs

import (
	"time"
)

type Needs int

type MemoryNeeds Needs
type CpuNeeds Needs
type NetworkNeeds Needs

type AppNeeds struct {
	MemoryNeeds MemoryNeeds
	CpuNeeds CpuNeeds
	NetworkNeeds NetworkNeeds
}

const (
	NEEDS_DELTA = 15
	NEEDS_CAUTION_INTERVAL = 2
)


type Minutes int

type TimeBasedNeeds map[Minutes]AppNeeds

type WeeklyNeeds map[time.Weekday]TimeBasedNeeds


func (t TimeBasedNeeds) Get(minutes Minutes) AppNeeds {
	return t[minutes]
}

func (w WeeklyNeeds) Get(day time.Weekday, minutes Minutes) AppNeeds {
	var maxNeeds AppNeeds
	for i := (minutes - NEEDS_CAUTION_INTERVAL * NEEDS_DELTA); i <= minutes + NEEDS_CAUTION_INTERVAL * NEEDS_DELTA; i += NEEDS_DELTA {
		if i >= 0 && i < 1440 {
			current := w[day][Minutes(i)]
			if current.CpuNeeds > maxNeeds.CpuNeeds {
				maxNeeds.CpuNeeds = current.CpuNeeds
			}
			if current.MemoryNeeds > maxNeeds.MemoryNeeds {
				maxNeeds.MemoryNeeds = current.MemoryNeeds
			}
			if current.NetworkNeeds > maxNeeds.NetworkNeeds {
				maxNeeds.NetworkNeeds = current.NetworkNeeds
			}
		} else if i < 0 {
			var dayBefore time.Weekday
			if day == time.Sunday {
				dayBefore = time.Saturday
			} else {
				dayBefore = day - 1
			}
			currentBefore := w[dayBefore][Minutes(1440 + i)]
			if currentBefore.CpuNeeds > maxNeeds.CpuNeeds {
				maxNeeds.CpuNeeds = currentBefore.CpuNeeds
			}
			if currentBefore.MemoryNeeds > maxNeeds.MemoryNeeds {
				maxNeeds.MemoryNeeds = currentBefore.MemoryNeeds
			}
			if currentBefore.NetworkNeeds > maxNeeds.NetworkNeeds {
				maxNeeds.NetworkNeeds = currentBefore.NetworkNeeds
			}
		} else if i >= 1440 {
			var dayAfter time.Weekday
			if day == time.Saturday {
				dayAfter = time.Sunday
			} else {
				dayAfter = day + 1
			}
			currentAfter := w[dayAfter][Minutes(i - 1440)]
			if currentAfter.CpuNeeds > maxNeeds.CpuNeeds {
				maxNeeds.CpuNeeds = currentAfter.CpuNeeds
			}
			if currentAfter.MemoryNeeds > maxNeeds.MemoryNeeds {
				maxNeeds.MemoryNeeds = currentAfter.MemoryNeeds
			}
			if currentAfter.NetworkNeeds > maxNeeds.NetworkNeeds {
				maxNeeds.NetworkNeeds = currentAfter.NetworkNeeds
			}
		}
	}
	return maxNeeds
}

// initialized by JSONconfig
//added by metrics.analyzer
func (w WeeklyNeeds) Set(day time.Weekday, minutes Minutes, ns AppNeeds) {
	if _, exists := w[day]; !exists {
		w[day] = make(map[Minutes]AppNeeds)
	}
	w[day][minutes] = ns
}

// helper method for Apps that have the same needs 24/7
func (w WeeklyNeeds) SetFlat(ns AppNeeds) {
	 for i := 0; i < 7; i++ {
		 w[time.Weekday(i)] = make(map[Minutes]AppNeeds)
		 for m := 0; m < (24 * 60); m += NEEDS_DELTA {
			 w[time.Weekday(i)][Minutes(m)] = ns
		 }
	 }
}

//get weekday and minutes in NEEDS_DELTA increments. always rounded down
func timeToWeekdayMinutes(t time.Time) (time.Weekday, Minutes) {
	utcTime := t.UTC()
	w := utcTime.Weekday()
	m := utcTime.Hour() * 60 + utcTime.Minute()
	if m % NEEDS_DELTA != 0 {
		m = int(m / NEEDS_DELTA) * NEEDS_DELTA
	}
	return w, Minutes(m)
}

func CurrentTimeForNeeds() (time.Weekday, Minutes) {
	return timeToWeekdayMinutes(time.Now())
}
