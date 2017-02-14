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
	"orca/trainer/logs"
)

type OrcaDb struct {
	enabled bool
	session *mgo.Session
	db      *mgo.Database
}

type AuditEvent struct {
	Timestamp time.Time
	Details   map[string]string
	Severity  AuditSeverity
	Message   string
}

type LogEvent struct {
	Timestamp time.Time
	LogLevel  string
	HostId    string
	AppId     string
	Message   string
}

type AuditSeverity string
type AuditMessage string

const (
	AUDIT__ERROR = AuditSeverity("error")
	AUDIT__INFO = AuditSeverity("info")
	AUDIT__DEBUG = AuditSeverity("debug")
	LOG__STDOUT = "stdout"
	LOG__STDERR = "stderr"
)

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

	if (event.Severity == AUDIT__ERROR) {
		logs.AuditLogger.Errorln(event.Message)
	} else if event.Severity == AUDIT__INFO {
		logs.AuditLogger.Infoln(event.Message)
	} else if event.Severity == AUDIT__DEBUG {
		logs.AuditLogger.Debugln(event.Message)
	}

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

func (db *OrcaDb) Query__AuditEvents() []AuditEvent {
	if !db.enabled {
		return []AuditEvent{}
	}
	c := db.db.C("audit")
	var results []AuditEvent
	err := c.Find(nil).Sort("-Timestamp").All(&results)
	if err != nil {
		panic("error querying db")
	}

	return results
}

func (db *OrcaDb) Query__AuditEventsHost(host string) []AuditEvent {
	if !db.enabled {
		return []AuditEvent{}
	}
	c := db.db.C("audit")
	var results []AuditEvent
	err := c.Find(bson.M{"host": host}).Sort("-Timestamp").All(&results)
	if err != nil {
		panic("error querying db")
	}

	return results
}

func (db *OrcaDb) Query__AuditEventsApplication(application string) []AuditEvent {
	if !db.enabled {
		return []AuditEvent{}
	}
	c := db.db.C("audit")
	var results []AuditEvent
	err := c.Find(bson.M{"application": application}).Sort("-Timestamp").All(&results)
	if err != nil {
		panic("error querying db")
	}

	return results
}

func (db *OrcaDb) Insert__Log(log LogEvent) {
	if !db.enabled {
		return
	}
	if log.Message == "" {
		return
	}
	if (log.LogLevel == LOG__STDERR) {
		logs.AuditLogger.Errorln(log.Message)
	} else if log.LogLevel == LOG__STDOUT {
		logs.AuditLogger.Infoln(log.Message)
	}

	if db.session == nil {
		return
	}

	log.Timestamp = time.Now()
	c := db.db.C("logs")
	err := c.Insert(&log)
	if err != nil {
		return
	}
}

func (db *OrcaDb) Query__HostLog(host string) []LogEvent {
	if !db.enabled {
		return []LogEvent{}
	}
	c := db.db.C("logs")
	var results []LogEvent
	if host != "" {
		err := c.Find(bson.M{"host": host}).Sort("-Timestamp").All(&results)
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

func (db *OrcaDb) Query__AppLog(app string) []LogEvent {
	if !db.enabled {
		return []LogEvent{}
	}
	c := db.db.C("logs")
	var results []LogEvent
	if app != "" {
		err := c.Find(bson.M{"app": app}).Sort("-Timestamp").All(&results)
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