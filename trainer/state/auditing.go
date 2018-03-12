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
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"orca/trainer/configuration"
	"orca/trainer/logs"
	"strconv"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	elastic "gopkg.in/olivere/elastic.v5"
)

type OrcaDb struct {
	enabled            bool
	session            *mgo.Session
	db                 *mgo.Database
	client             *elastic.Client
	ctx                context.Context
	configurationStore *configuration.ConfigurationStore
}

type AuditEvent struct {
	Timestamp time.Time
	HostId    string
	AppId     string
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
	AUDIT__ERROR   = AuditSeverity("error")
	AUDIT__INFO    = AuditSeverity("info")
	AUDIT__DEBUG   = AuditSeverity("debug")
	AUDIT__MONITOR = AuditSeverity("monitor")
	LOG__STDOUT    = "stdout"
	LOG__STDERR    = "stderr"

	LOG_MAPPING = `{
                        "mappings" : {
                            "log" : {
                                "properties" : {
                                    "Timestamp" : { "type" : "date" },
                                 }
                            }
                        }
                    }`

	AUDIT_MAPPING = `{
                        "mappings" : {
                            "event" : {
                                "properties" : {
                                    "Timestamp" : { "type" : "date" },
                                    "HostId" : { "type" : "string", "index" : "not_analyzed"},
                                    "AppId" : { "type" : "string", "index" : "not_analyzed"},
                                    "Message" : { "type" : "string", "index" : "not_analyzed"},
                                    "Severity" : { "type" : "string"}
                                }
                            }
                        }
                    }`
)

var Audit OrcaDb

func (db *OrcaDb) Init(configurationStore *configuration.ConfigurationStore) {
	db.configurationStore = configurationStore
	if configurationStore.GlobalSettings.AuditDatabaseUri == "" {
		db.enabled = false
		return
	}
	db.enabled = true
	session, err := mgo.Dial(db.configurationStore.GlobalSettings.StatsDatabaseUri)
	if err != nil {
		panic(err)
	}
	session.SetSocketTimeout(1 * time.Minute)

	db.session = session
	s := db.session.Copy()
	defer s.Close()

	indexAuditTTL := mgo.Index{
		Key:         []string{"timestamp"},
		Unique:      false,
		DropDups:    false,
		Background:  true,
		ExpireAfter: time.Hour * 24}

	if err := s.DB("orca").C("audit").EnsureIndex(indexAuditTTL); err != nil {
		panic(err)
	}

	indexLogTTL := mgo.Index{
		Key:         []string{"timestamp"},
		Unique:      false,
		DropDups:    false,
		Background:  true,
		ExpireAfter: time.Hour * 24}

	if err := s.DB("orca").C("logs").EnsureIndex(indexLogTTL); err != nil {
		panic(err)
	}

	indexApps := mgo.Index{
		Key:        []string{"timestamp", "appid"},
		Unique:     false,
		DropDups:   false,
		Background: true}
	if err := s.DB("orca").C("audit").EnsureIndex(indexApps); err != nil {
		panic(err)
	}

	indexHosts := mgo.Index{
		Key:        []string{"timestamp", "hostid"},
		Unique:     false,
		DropDups:   false,
		Background: true}
	if err := s.DB("orca").C("audit").EnsureIndex(indexHosts); err != nil {
		panic(err)
	}
}

func (db *OrcaDb) Insert__AuditEvent(event AuditEvent) {
	if !db.enabled {
		return
	}

	if event.Severity == AUDIT__ERROR {
		logs.AuditLogger.Errorln(event.Message)
	} else if event.Severity == AUDIT__INFO {
		logs.AuditLogger.Infoln(event.Message)
	} else if event.Severity == AUDIT__DEBUG {
		logs.AuditLogger.Debugln(event.Message)
	}

	/* Run hooks */
	for _, hook := range db.configurationStore.GlobalSettings.AuditWebhooks {
		if hook.Severity == string(event.Severity) {
			b := new(bytes.Buffer)
			b.WriteString("{\"text\":\"orca@" + db.configurationStore.GlobalSettings.EnvName + " said " + event.Message + "\"}")

			res, err := http.Post(hook.Uri, "application/json; charset=utf-8", b)
			if err != nil {
				logs.AuditLogger.Errorf("Could not send event to webhook: %+v", err)
			} else {
				defer res.Body.Close()
			}
		}
	}
	event.Timestamp = time.Now()
	s := db.session.Copy()
	defer s.Close()
	c := s.DB("orca").C("audit")
	if err := c.Insert(&event); err != nil {
		fmt.Println(err)
	}
}

func (db *OrcaDb) query(collection string, results interface{}, hostid string, appid string, limit string, search string, lasttime string) {
	limitInteger, _ := strconv.Atoi(limit)
	s := db.session.Copy()
	defer s.Close()
	c := s.DB("orca").C(collection)
	conditions := make(bson.M, 0)
	if len(hostid) > 0 {
		conditions["hostid"] = hostid
	}
	if len(appid) > 0 {
		conditions["appid"] = appid
	}
	if len(search) > 0 {
		conditions["message"] = bson.RegEx{Pattern: search, Options: "i"}
	}
	if len(lasttime) > 0 {
		lt, _ := time.Parse("2006-01-02T15:04:05.999Z", lasttime)
		conditions["timestamp"] = bson.M{"$gt": lt}
	}

	if err := c.Find(conditions).Sort("-timestamp").Limit(limitInteger).All(results); err != nil {
		panic("error querying db")
	}
}

func (db *OrcaDb) Query__AuditEvents(limit string, search string, lasttime string) []AuditEvent {
	if !db.enabled {
		return []AuditEvent{}
	}
	var results []AuditEvent
	db.query("audit", &results, "", "", limit, search, lasttime)
	return results
}

func (db *OrcaDb) Query__AuditEventsHost(host string, limit string, search string, lasttime string) []AuditEvent {
	if !db.enabled {
		return []AuditEvent{}
	}
	var results []AuditEvent
	db.query("audit", &results, host, "", limit, search, lasttime)
	return results
}

func (db *OrcaDb) Query__AuditEventsApplication(application string, limit string, search string, lasttime string) []AuditEvent {
	if !db.enabled {
		return []AuditEvent{}
	}
	var results []AuditEvent
	db.query("audit", &results, "", application, limit, search, lasttime)
	return results
}

func (db *OrcaDb) Insert__Log(log LogEvent) {
	if !db.enabled {
		return
	}

	/* Run hooks */
	for _, hook := range db.configurationStore.GlobalSettings.LoggingWebHooks {
		j, err := json.MarshalIndent(log, "", "  ")
		b := new(bytes.Buffer)
		b.Write(j)
		if err != nil {
			logs.AuditLogger.Errorf("Error while sending to LoggingWebHook %s.  %+v", hook.Uri, err)
		} else {
			req, _ := http.NewRequest("PUT", hook.Uri, b)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Region", db.configurationStore.GlobalSettings.AWSRegion)
			if len(hook.User) > 0 && len(hook.Password) > 0 {
				req.SetBasicAuth(hook.User, hook.Password)
			}

			var transport = &http.Transport{}

			if len(hook.Certificate) > 0 {
				certPool := x509.NewCertPool()
				certPool.AppendCertsFromPEM([]byte(hook.Certificate))

				transport.TLSClientConfig = &tls.Config{RootCAs: certPool, InsecureSkipVerify: true}
			}
			var client = &http.Client{
				Transport: transport,
			}

			res, err := client.Do(req)
			if err != nil {
				logs.AuditLogger.Errorf("Could not send event to webhook: %+v", err)
			} else {
				defer res.Body.Close()
			}
		}
	}
	log.Timestamp = time.Now()
	s := db.session.Copy()
	defer s.Close()
	c := s.DB("orca").C("logs")
	if err := c.Insert(&log); err != nil {
		fmt.Println(err)
	}
}

func (db *OrcaDb) Query__HostLog(host string, limit string, search string, lasttime string) []LogEvent {
	if !db.enabled {
		return []LogEvent{}
	}
	var results []LogEvent
	db.query("logs", &results, host, "", limit, search, lasttime)
	return results
}

func (db *OrcaDb) Query__AppLog(app string, limit string, search string, lasttime string) []LogEvent {
	if !db.enabled {
		return []LogEvent{}
	}
	var results []LogEvent
	db.query("logs", &results, "", app, limit, search, lasttime)
	return results
}
