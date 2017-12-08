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
	"orca/trainer/configuration"
	"orca/trainer/logs"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type StatisticsDb struct {
	session            *mgo.Session
	db                 *mgo.Database
	configurationStore *configuration.ConfigurationStore
}

var Stats StatisticsDb

func (db *StatisticsDb) Init(configurationStore *configuration.ConfigurationStore) {
	db.configurationStore = configurationStore

	session, err := mgo.Dial(db.configurationStore.GlobalSettings.StatsDatabaseUri)
	if err != nil {
		panic(err)
	}

	db.session = session
	s := db.session.Copy()
	defer s.Close()

	indexAppUtilisationTTL := mgo.Index{
		Key:         []string{"timestamp"},
		Unique:      false,
		DropDups:    false,
		Background:  true,
		ExpireAfter: time.Hour * 24 * 7}

	if err := s.DB("orca").C("app_utilisation").EnsureIndex(indexAppUtilisationTTL); err != nil {
		panic(err)
	}

	indexAppUtilisation := mgo.Index{
		Key:        []string{"timestamp", "appname"},
		Unique:     false,
		DropDups:   false,
		Background: true}

	if err := s.DB("orca").C("app_utilisation").EnsureIndex(indexAppUtilisation); err != nil {
		panic(err)
	}

	indexAppHostUtilisationTTL := mgo.Index{
		Key:         []string{"timestamp"},
		Unique:      false,
		DropDups:    false,
		Background:  true,
		ExpireAfter: time.Hour * 24 * 7}

	if err := s.DB("orca").C("app_host_utilisation").EnsureIndex(indexAppHostUtilisationTTL); err != nil {
		panic(err)
	}

	indexAppHostUtilisation := mgo.Index{
		Key:        []string{"timestamp", "appname", "host"},
		Unique:     false,
		DropDups:   false,
		Background: true}

	if err := s.DB("orca").C("app_host_utilisation").EnsureIndex(indexAppHostUtilisation); err != nil {
		panic(err)
	}

	indexHostUtilisationTTL := mgo.Index{
		Key:         []string{"timestamp"},
		Unique:      false,
		DropDups:    false,
		Background:  true,
		ExpireAfter: time.Hour * 24 * 7}

	if err := s.DB("orca").C("host_utilisation").EnsureIndex(indexHostUtilisationTTL); err != nil {
		panic(err)
	}

	indexHostUtilisation := mgo.Index{
		Key:         []string{"timestamp", "host"},
		Unique:      false,
		DropDups:    false,
		Background:  true,
		ExpireAfter: time.Hour * 24 * 7}

	if err := s.DB("orca").C("host_utilisation").EnsureIndex(indexHostUtilisation); err != nil {
		panic(err)
	}
}

func (db *StatisticsDb) Close() {
	db.session.Close()
}

type ApplicationUtilisationStatistic struct {
	Cpu     int64
	Mbytes  int64
	Network int64

	InstanceCount        int64
	DesiredInstanceCount int64

	AppName   string
	Timestamp time.Time
}

type ApplicationHostUtilisationStatistic struct {
	Cpu     int64
	Mbytes  int64
	Network int64

	AppName string
	Host    string

	Timestamp time.Time
}

type HostUtilisationStatistic struct {
	Cpu                  int64
	Mbytes               int64
	Network              int64
	HardDiskUsage        int64
	HardDiskUsagePercent int64

	Host      string
	Timestamp time.Time
}

func (db *StatisticsDb) Insert__ApplicationUtilisationStatistic(event ApplicationUtilisationStatistic) {
	s := db.session.Copy()
	defer s.Close()

	event.Timestamp = time.Now()
	c := s.DB("orca").C("app_utilisation")
	err := c.Insert(&event)
	if err != nil {
		return
	}
}

func (db *StatisticsDb) Insert__ApplicationHostUtilisationStatistic(event ApplicationHostUtilisationStatistic) {
	s := db.session.Copy()
	defer s.Close()

	event.Timestamp = time.Now()
	c := s.DB("orca").C("app_host_utilisation")
	err := c.Insert(&event)
	if err != nil {
		return
	}
}

func (db *StatisticsDb) Insert__HostUtilisationStatistic(event HostUtilisationStatistic) {
	s := db.session.Copy()
	defer s.Close()

	event.Timestamp = time.Now()
	c := s.DB("orca").C("host_utilisation")
	err := c.Insert(&event)
	if err != nil {
		return
	}
}

func (db *StatisticsDb) Query__ApplicationUtilisationStatistic(application string) []ApplicationUtilisationStatistic {
	s := db.session.Copy()
	defer s.Close()

	c := s.DB("orca").C("app_utilisation")
	var results []ApplicationUtilisationStatistic
	err := c.Find(bson.M{"appname": application}).Sort("-timestamp").All(&results)
	if err != nil {
		panic("error querying db")
	}

	return results
}

func (db *StatisticsDb) Query__HostUtilisationStatistic(host string) []HostUtilisationStatistic {
	s := db.session.Copy()
	defer s.Close()

	c := s.DB("orca").C("host_utilisation")
	var results []HostUtilisationStatistic
	err := c.Find(bson.M{"host": host}).Sort("-timestamp").All(&results)
	if err != nil {
		panic("error querying db")
	}

	return results
}

func (db *StatisticsDb) Query__LatestHostUtilisationStatistic(host string) HostUtilisationStatistic {
	s := db.session.Copy()
	defer s.Close()

	c := s.DB("orca").C("host_utilisation")
	var results HostUtilisationStatistic
	err := c.Find(bson.M{"host": host}).Sort("-timestamp").One(&results)
	if err != nil {
		logs.AuditLogger.Errorln(err)
		return HostUtilisationStatistic{}
	}

	return results
}

func (db *StatisticsDb) Query__ApplicationHostUtilisationStatistic(application string, host string) []ApplicationHostUtilisationStatistic {
	s := db.session.Copy()
	defer s.Close()

	c := s.DB("orca").C("app_host_utilisation")
	var results []ApplicationHostUtilisationStatistic
	err := c.Find(bson.M{"host": host, "appname": application}).Sort("-timestamp").All(&results)
	if err != nil {
		panic("error querying db")
	}

	return results
}
