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
)

type StatisticsDb struct {
	session *mgo.Session
	db      *mgo.Database
}

var Stats StatisticsDb

func (a *StatisticsDb) Init(hostname string) {
	session, err := mgo.Dial(hostname)
	if err != nil {
		panic(err)
	}

	a.session = session
	a.db = session.DB("orca")
}

func (a *StatisticsDb) Close() {
	a.session.Close()
}

type ApplicationUtilisationStatistic struct {
	Cpu       int64
	Mbytes    int64
	Network   int64
	InstanceCount int64
	AppName   string
	Timestamp time.Time
}

type HostUtilisationStatistic struct {
	Cpu       int64
	Mbytes    int64
	Network   int64
	HardDiskUsage        int64
	HardDiskUsagePercent int64

	Host   string
	Timestamp time.Time
}

func (db *StatisticsDb) Insert__ApplicationUtilisationStatistic(event ApplicationUtilisationStatistic) {
	if db.session == nil {
		return
	}

	event.Timestamp = time.Now()
	c := db.db.C("app_utilisation")
	err := c.Insert(&event)
	if err != nil {
		return
	}
}

func (db *StatisticsDb) Insert__HostUtilisationStatistic(event HostUtilisationStatistic) {
	if db.session == nil {
		return
	}

	event.Timestamp = time.Now()
	c := db.db.C("host_utilisation")
	err := c.Insert(&event)
	if err != nil {
		return
	}
}

func (db *StatisticsDb) Query__ApplicationUtilisationStatistic(application string) []ApplicationUtilisationStatistic {
	c := db.db.C("app_utilisation")
	var results []ApplicationUtilisationStatistic
	err := c.Find(bson.M{"appname": application}).Sort("-Timestamp").All(&results)
	if err != nil {
		panic("error querying db")
	}

	return results
}

func (db *StatisticsDb) Query__HostUtilisationStatistic(host string) []HostUtilisationStatistic {
	c := db.db.C("host_utilisation")
	var results []HostUtilisationStatistic
	err := c.Find(bson.M{"host": host}).Sort("-Timestamp").All(&results)
	if err != nil {
		panic("error querying db")
	}

	return results
}


