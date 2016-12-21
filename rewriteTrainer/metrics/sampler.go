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

package metrics

import (
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/base"
	"net/http"
	"encoding/json"
	"errors"
	"gatoor/orca/rewriteTrainer/db"
	"io/ioutil"
	"fmt"
	"os"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

var MetricsLogger = Logger.LoggerWithField(Logger.Logger, "module", "metrics")

func ParsePush(r *http.Request) (base.HostInfo, base.MetricsWrapper, error) {
	decoder := json.NewDecoder(r.Body)
	var wrapper base.TrainerPushWrapper
	err := decoder.Decode(&wrapper)
	if err != nil {
		MetricsLogger.Errorf("TrainerPushWrapper parsing failed - %s", err)
		htmlData, err0 := ioutil.ReadAll(r.Body)
		if err0 != nil {
			fmt.Println(err0)
			os.Exit(1)
		}
		fmt.Println(">>>")
		fmt.Println(os.Stdout, string(htmlData))
		fmt.Println(">>>")

		return base.HostInfo{}, base.MetricsWrapper{}, errors.New("Parsing failed")
	}
	return wrapper.HostInfo, wrapper.Stats, nil
}

type UtilisationStatistic struct {
	Timestamp  time.Time
	Cpu        base.Usage
	Mbytes     base.Usage
	Network    base.Usage

	AppName    base.AppName
	AppVersion base.Version
	Host       base.HostId
}


//TODO: MongoDB version so that we are more flexible
//++ Memory/CPU/Network is pushed from the hosts every minute and stored in mongodb
//++ Queries are then grouped on the minute so that an aggregate can be computed
//++ Statistical queries are minute orientated and averages could be taken over larger time periods

func RecordStats(host base.HostId, stats base.MetricsWrapper, _ string) {
	//TODO handle string version
	session, err := mgo.Dial("ec2-52-78-246-226.ap-northeast-2.compute.amazonaws.com")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	c := session.DB("orca").C("app_metrics")
	for appName, obj := range stats.AppMetrics {
		for appVersion, appMetrics := range obj {
			for _, metric := range appMetrics {
				entity := UtilisationStatistic{
					AppName:appName,
					AppVersion:appVersion,
					Cpu:metric.CpuUsage,
					Mbytes:metric.MemoryUsage,
					Network:metric.NetworkUsage,
					Host:host,
					Timestamp:time.Now(),
				}

				err = c.Insert(&entity)
				if err != nil {
				}
			}
		}
	}

	MetricsLogger.WithField("host", host).Infof("Recording stats for host '%s'", host)
	MetricsLogger.WithField("host", host).Infof("Stats: %+v", stats)
}

func RecordHostInfo(info base.HostInfo, time string) {
	MetricsLogger.WithField("host", info.HostId).Infof("Recording info for host %s", info.HostId)
	MetricsLogger.WithField("host", info.HostId).Infof("Info: %+v", info)
	db.Audit.Add(db.BUCKET_AUDIT_RECEIVED_HOST_INFO, string(info.HostId) + "_" + time, info)
}

func QueryStats__ApplicationPerformance(application base.AppName) []UtilisationStatistic {
	session, err := mgo.Dial("ec2-52-78-246-226.ap-northeast-2.compute.amazonaws.com")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	var result []UtilisationStatistic
	c := session.DB("orca").C("app_metrics")
	err = c.Find(bson.M{"appname": application}).All(&result)
	if err != nil {
		panic(err)
	}
	return result
}

func QueryStats__ApplicationPerformance__ByMinute(application base.AppName) []UtilisationStatistic {
	session, err := mgo.Dial("ec2-52-78-246-226.ap-northeast-2.compute.amazonaws.com")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	c := session.DB("orca").C("app_metrics")

	var result = make([]UtilisationStatistic, 0)
	operations := []bson.M{
		bson.M{
			"$match" :bson.M{"appname":application},
		},
		bson.M{
			"$group": bson.M{
				"_id": bson.M{
					"utc_year": bson.M{"$year": "$timestamp" },
					"utc_dayOfYMonth": bson.M{"$dayOfMonth": "$timestamp" },
					"utc_month": bson.M{"$month": "$timestamp" },
					"utc_hour": bson.M{"$hour": "$timestamp" },
					"utc_minute": bson.M{"$minute": "$timestamp" },
				},
				"cpu": bson.M{
					"$sum": "$cpu",
				},
				"mbytes": bson.M{
					"$sum": "$mbytes",
				},
				"network": bson.M{
					"$sum": "$network",
				},
			},
		}}

	pipe := c.Pipe(operations)
	results := []bson.M{}
	err1 := pipe.All(&results)

	if err1 != nil {
		fmt.Printf("ERROR : %s\n", err1.Error())
	}

	for _, item := range results {
		timestamp := time.Date(
			item["_id"].(bson.M)["utc_year"].(int),
			time.Month(item["_id"].(bson.M)["utc_month"].(int)),
			item["_id"].(bson.M)["utc_dayOfYMonth"].(int),
			item["_id"].(bson.M)["utc_hour"].(int),
			item["_id"].(bson.M)["utc_minute"].(int), 0, 0,
			time.UTC,
		)

		entry := UtilisationStatistic{
			Timestamp:timestamp,
			Mbytes: base.Usage(item["mbytes"].(int64)),
			Cpu: base.Usage(item["cpu"].(int64)),
			Network: base.Usage(item["network"].(int64)),
		}
		result = append(result, entry)
	}

	return result
}
