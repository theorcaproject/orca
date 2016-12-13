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

	if GlobalAppsStatusTracker.LastStable("app1") != 1 {
		t.Error(GlobalAppsStatusTracker.LastStable("app1"))
	}

	if GlobalAppsStatusTracker.LastStable("app2") != 2 {
		t.Error(GlobalAppsStatusTracker.LastStable("app2"))
	}
}

