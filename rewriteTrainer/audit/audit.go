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

package audit

import (
	"time"
	"gatoor/orca/base/log"
)

type Event struct {
	Timestamp int64
	Details map[string]string
}

type AuditEngine struct {
	Events []Event
}

func (p *AuditEngine) Init() {
	p.Events = make([]Event, 0)
}

func (p *AuditEngine) AddEvent(event map[string]string) {
	log.AuditLogger.Info(event["message"])
	p.Events = append(p.Events, Event{Timestamp:time.Now().Unix(), Details:event})
}


//TODO: Make this method immutable and apply the filter
func (p *AuditEngine) ListEvents(filter map[string]string) []Event{
	res := p.Events
	return res
}

var Audit AuditEngine




