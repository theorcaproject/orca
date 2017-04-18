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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"orca/trainer/configuration"
	"orca/trainer/logs"
	"reflect"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"
)

type OrcaDb struct {
	enabled            bool
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
	AUDIT__ERROR = AuditSeverity("error")
	AUDIT__INFO  = AuditSeverity("info")
	AUDIT__DEBUG = AuditSeverity("debug")
	LOG__STDOUT  = "stdout"
	LOG__STDERR  = "stderr"

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

//"http://127.0.0.1:9200"
func (a *OrcaDb) Init(configurationStore *configuration.ConfigurationStore) {
	a.configurationStore = configurationStore
	if configurationStore.GlobalSettings.AuditDatabaseUri == "" {
		a.enabled = false
		return
	}
	a.enabled = true
	ctx := context.Background()
	a.ctx = ctx
	cli, err := elastic.NewClient(elastic.SetURL(configurationStore.GlobalSettings.AuditDatabaseUri))
	if err != nil {
		fmt.Println("Cloud not connect to elasticsearch")
		panic(err)
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

	if db.client == nil {
		return
	}

	event.Timestamp = time.Now()
	_, err := db.client.Index().
		Index("audit").
		Type("event").
		BodyJson(event).
		TTL("24h").
		Do(db.ctx)
	if err != nil {
		fmt.Println(err)
	}
}

func (db *OrcaDb) Query__AuditEvents(limit string, search string, lasttime string) []AuditEvent {
	if !db.enabled {
		return []AuditEvent{}
	}
	limitInteger, _ := strconv.Atoi(limit)
	q := elastic.NewBoolQuery()
	if len(search) > 0 {
		q.Must(elastic.NewWildcardQuery("Message", search))
	}
	if len(lasttime) > 0 {
		q.Must(elastic.NewRangeQuery("Timestamp").Gt(lasttime))
	}
	events, err := db.client.Search().Index("audit").Query(q).Sort("Timestamp", false).Size(limitInteger).Do(db.ctx)
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

func (db *OrcaDb) Query__AuditEventsHost(host string, limit string, search string, lasttime string) []AuditEvent {
	if !db.enabled {
		return []AuditEvent{}
	}
	limitInteger, _ := strconv.Atoi(limit)
	q := elastic.NewBoolQuery()
	q.Must(elastic.NewTermQuery("HostId", host))
	if len(search) > 0 {
		q.Must(elastic.NewWildcardQuery("Message", search))
	}
	if len(lasttime) > 0 {
		q.Must(elastic.NewRangeQuery("Timestamp").Gt(lasttime))
	}

	auditRes, err := db.client.Search().Index("audit").Query(q).Sort("Timestamp", false).Size(limitInteger).Do(db.ctx)
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

func (db *OrcaDb) Query__AuditEventsApplication(application string, limit string, search string, lasttime string) []AuditEvent {
	if !db.enabled {
		return []AuditEvent{}
	}

	limitInteger, _ := strconv.Atoi(limit)
	q := elastic.NewBoolQuery()
	q.Must(elastic.NewTermQuery("AppId", application))
	if len(search) > 0 {
		q.Must(elastic.NewWildcardQuery("Message", search))
	}
	if len(lasttime) > 0 {
		q.Must(elastic.NewRangeQuery("Timestamp").Gt(lasttime))
	}
	auditRes, err := db.client.Search().Index("audit").Query(q).Sort("Timestamp", false).Size(limitInteger).Do(db.ctx)
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

	/* Run hooks */
	for _, hook := range db.configurationStore.GlobalSettings.LoggingWebHooks {
		j, err := json.MarshalIndent(log, "", "  ")
		b := new(bytes.Buffer)
		b.Write(j)
		if err != nil {

			req, _ := http.NewRequest("PUT", hook.Uri, b)
			req.Header.Set("Content-Type", "application/json")
			if len(hook.User) > 0 && len(hook.Password) > 0 {
				req.SetBasicAuth(hook.User, hook.Password)
			}

			var transport = &http.Transport{}

			if len(hook.Certificate) > 0 {
				certPool := x509.NewCertPool()
				certPool.AppendCertsFromPEM([]byte(hook.Certificate))

				transport.TLSClientConfig = &tls.Config{RootCAs: certPool, InsecureSkipVerify: false}
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
	_, err := db.client.Index().
		Index("logs").
		Type("log").
		BodyJson(log).
		TTL("24h").
		Do(db.ctx)
	if err != nil {
		fmt.Println(err)
	}
}

func (db *OrcaDb) Query__HostLog(host string, limit string, search string, lasttime string) []LogEvent {
	if !db.enabled {
		return []LogEvent{}
	}

	limitInteger, _ := strconv.Atoi(limit)
	q := elastic.NewBoolQuery()
	q.Must(elastic.NewTermQuery("HostId", host))

	if len(search) > 0 {
		q.Must(elastic.NewWildcardQuery("Message", search))
	}
	if len(lasttime) > 0 {
		q.Must(elastic.NewRangeQuery("Timestamp").Gt(lasttime))
	}
	logsRes, err := db.client.Search().Index("logs").Query(q).Sort("Timestamp", false).Size(limitInteger).Do(db.ctx)
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

func (db *OrcaDb) Query__AppLog(app string, limit string, search string, lasttime string) []LogEvent {
	if !db.enabled {
		return []LogEvent{}
	}

	limitInteger, _ := strconv.Atoi(limit)
	q := elastic.NewBoolQuery()
	q.Must(elastic.NewTermQuery("AppId", app))
	if len(search) > 0 {
		q.Must(elastic.NewWildcardQuery("Message", search))
	}
	if len(lasttime) > 0 {
		q.Must(elastic.NewRangeQuery("Timestamp").Gt(lasttime))
	}
	logsRes, err := db.client.Search().Index("logs").Query(q).Sort("Timestamp", false).Size(limitInteger).Do(db.ctx)
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
