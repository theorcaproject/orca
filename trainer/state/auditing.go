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

package state

import (
	"time"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"fmt"
)

type OrcaDb struct {
	enabled bool
	session *mgo.Session
	db      *mgo.Database
}

type AuditEvent struct {
	Timestamp time.Time
	Details   map[string]string
}

var Audit OrcaDb

func (a *OrcaDb) Init(hostname string) {
	if hostname == "" {
		a.enabled = false
		return
	}
	a.enabled = true
	session, err := mgo.Dial(hostname)
	if err != nil {
		panic(err)
	}

	a.session = session
	a.db = session.DB("orca")
}

func (a *OrcaDb) Close() {
	a.session.Close()
}

func (db *OrcaDb) Insert__AuditEvent(event AuditEvent) {
	if !db.enabled {
		return
	}
	fmt.Printf("AUDIT: %s\n", event.Details["message"])

	if db.session == nil {
		return
	}

	event.Timestamp = time.Now()
	c := db.db.C("audit")
	err := c.Insert(&event)
	if err != nil {
		return
	}
}

func (db *OrcaDb) Query__AuditEvents(application string) []AuditEvent {
	if !db.enabled {
		return []AuditEvent{}
	}
	c := db.db.C("audit")
	var results []AuditEvent
	if application != "" {
		err := c.Find(bson.M{"details.application": application}).Sort("-Timestamp").All(&results)
		if err != nil {
			panic("error querying db")
		}

	} else {
		err := c.Find(nil).Sort("-Timestamp").All(&results)
		if err != nil {
			panic("error querying db")
		}
	}

	return results
}

