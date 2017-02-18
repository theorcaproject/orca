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
	"gopkg.in/olivere/elastic.v5"
	"golang.org/x/net/context"
	"orca/trainer/logs"
	"fmt"
	"reflect"
)

type OrcaDb struct {
	enabled bool
	client  *elastic.Client
	ctx context.Context
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
	AUDIT__ERROR = AuditSeverity("error")
	AUDIT__INFO = AuditSeverity("info")
	AUDIT__DEBUG = AuditSeverity("debug")
	LOG__STDOUT = "stdout"
	LOG__STDERR = "stderr"

	LOG_MAPPING = `{
                        "mappings" : {
                            "log" : {
                                "properties" : {
                                    "Timestamp" : { "type" : "date" },
                                    "LogLevel" : { "type" : "string", "index" : "not_analyzed" },
                                    "HostId" : { "type" : "string", "index" : "not_analyzed" },
                                    "AppId" : { "type" : "string", "index" : "not_analyzed" },
                                    "Message" : { "type" : "string"}
                                }
                            }
                        }
                    }`

	AUDIT_MAPPING = `{
                        "mappings" : {
                            "event" : {
                                "properties" : {
                                    "Timestamp" : { "type" : "date" },
                                    "HostId" : { "type" : "string", "index" : "not_analyzed" },
                                    "AppId" : { "type" : "string", "index" : "not_analyzed" },
                                    "Message" : { "type" : "string"},
                                    "Severity" : { "type" : "string"}
                                }
                            }
                        }
                    }`
)

var Audit OrcaDb

//"http://127.0.0.1:9200"
func (a *OrcaDb) Init(hostname string) {
	if hostname == "" {
		a.enabled = false
		return
	}
	a.enabled = true
	ctx := context.Background()
	a.ctx = ctx
	cli, err := elastic.NewClient(elastic.SetURL(hostname))
	if err != nil {
		fmt.Println("Cloud not connect to elasticsearch")
		return
	}
	a.client = cli
	exists, _ := a.client.IndexExists("audit").Do(ctx)
	if !exists {
		a.client.CreateIndex("audit").Body(AUDIT_MAPPING).Do(ctx)
	}

	existsLogs, _ := a.client.IndexExists("logs").Do(ctx)
	if !existsLogs {
		a.client.CreateIndex("logs").Body(LOG_MAPPING).Do(ctx)
	}

}

func (a *OrcaDb) Close() {
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

	if db.client == nil {
		return
	}

	event.Timestamp = time.Now()
	_, err := db.client.Index().
		Index("audit").
		Type("event").
		BodyJson(event).
		Do(db.ctx)
	if err != nil {
		fmt.Println(err)
	}
}

func (db *OrcaDb) Query__AuditEvents() []AuditEvent {
	if !db.enabled {
		return []AuditEvent{}
	}
	events, err := db.client.Search().Index("audit").Query(elastic.NewMatchAllQuery()).Sort("Timestamp", false).Size(10000).Do(db.ctx)
	var eventType AuditEvent
	var results []AuditEvent
	if err != nil {
		return results
	}
	for _, item := range events.Each(reflect.TypeOf(eventType)) {
		if t, ok := item.(AuditEvent); ok {
			results = append(results, t)
		}
	}
	return results
}

func (db *OrcaDb) Query__AuditEventsHost(host string) []AuditEvent {
	if !db.enabled {
		return []AuditEvent{}
	}
	auditRes, err := db.client.Search().Index("audit").Query(elastic.NewTermQuery("HostId", host)).Sort("Timestamp", false).Size(10000).Do(db.ctx)
	var eventType AuditEvent
	var results []AuditEvent
	if err != nil {
		return results
	}
	for _, item := range auditRes.Each(reflect.TypeOf(eventType)) {
		if t, ok := item.(AuditEvent); ok {
			results = append(results, t)
		}
	}
	return results
}

func (db *OrcaDb) Query__AuditEventsApplication(application string) []AuditEvent {
	if !db.enabled {
		return []AuditEvent{}
	}
	auditRes, err := db.client.Search().Index("audit").Query(elastic.NewTermQuery("AppId", application)).Sort("Timestamp", false).Size(10000).Do(db.ctx)
	var eventType AuditEvent
	var results []AuditEvent
	if err != nil {
		return results
	}
	for _, item := range auditRes.Each(reflect.TypeOf(eventType)) {
		if t, ok := item.(AuditEvent); ok {
			results = append(results, t)
		}
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

	if db.client == nil {
		return
	}

	log.Timestamp = time.Now()
	_, err := db.client.Index().
		Index("logs").
		Type("log").
		BodyJson(log).
		Do(db.ctx)
	if err != nil {
		fmt.Println(err)
	}
}

func (db *OrcaDb) Query__HostLog(host string) []LogEvent {
	if !db.enabled {
		return []LogEvent{}
	}
	logsRes, err := db.client.Search().Index("logs").Query(elastic.NewTermQuery("HostId", host)).Sort("Timestamp", false).Size(10000).Do(db.ctx)
	var logType LogEvent
	var results []LogEvent
	if err != nil {
		fmt.Println(err)
		return results
	}
	for _, item := range logsRes.Each(reflect.TypeOf(logType)) {
		if t, ok := item.(LogEvent); ok {
			results = append(results, t)
		}
	}
	return results
}

func (db *OrcaDb) Query__AppLog(app string) []LogEvent {
	if !db.enabled {
		return []LogEvent{}
	}
	logsRes, err := db.client.Search().Index("logs").Query(elastic.NewTermQuery("AppId", app)).Sort("Timestamp", false).Size(10000).Do(db.ctx)
	var logType LogEvent
	var results []LogEvent
	if err != nil {
		fmt.Println(err)
		return results
	}
	for _, item := range logsRes.Each(reflect.TypeOf(logType)) {
		if t, ok := item.(LogEvent); ok {
			results = append(results, t)
		}
	}
	return results
}