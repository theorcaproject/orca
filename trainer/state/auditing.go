package state

import (
	"time"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"fmt"
)

type OrcaDb struct {
	session *mgo.Session
	db      *mgo.Database
}

type AuditEvent struct {
	Timestamp time.Time
	Details   map[string]string
}

var Audit OrcaDb

func (a *OrcaDb) Init(hostname string) {
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

