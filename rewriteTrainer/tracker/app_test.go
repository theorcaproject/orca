package tracker

import (
	"testing"
)

func TestAppsStatusTracker_Update(t *testing.T) {
	if len(GlobalAppsStatusTracker) != 0 {
		t.Error(GlobalAppsStatusTracker)
	}

	GlobalAppsStatusTracker.Update("somehost", "app1", 1, APP_EVENT_CRASH)

	if len(GlobalAppsStatusTracker) != 1 {
		t.Error(GlobalAppsStatusTracker)
	}

	e, _ := GlobalAppsStatusTracker.GetRating("app1", 1)
	if e != RATING_CRASHED || len(GlobalAppsStatusTracker["app1"][1].CrashDetails) != 1 {
		t.Error(e)
	}

	GlobalAppsStatusTracker.Update("somehost", "app1", 1, APP_EVENT_ROLLBACK)
	e2, _ := GlobalAppsStatusTracker.GetRating("app1", 1)
	if e2 != RATING_CRASHED || len(GlobalAppsStatusTracker["app1"][1].CrashDetails) != 2 {
		t.Error(e2)
	}

	GlobalAppsStatusTracker.Update("somehost", "app2", 2, APP_EVENT_SUCCESSFUL_UPDATE)
	GlobalAppsStatusTracker.Update("somehost", "app2", 2, APP_EVENT_CHECKIN)

	if len(GlobalAppsStatusTracker) != 2 {
		t.Error(GlobalAppsStatusTracker)
	}
	if GlobalAppsStatusTracker["app2"][2].RunningCount != 2 {
		t.Error(GlobalAppsStatusTracker["app2"][2])
	}
}
