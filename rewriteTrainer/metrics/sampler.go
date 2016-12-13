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

func RecordStats(host base.HostId, stats base.MetricsWrapper, time string) {//TODO handle string version
	MetricsLogger.WithField("host", host).Infof("Recording stats for host '%s'", host)
	MetricsLogger.WithField("host", host).Infof("Stats: %+v", stats)
	db.Audit.Add(db.BUCKET_AUDIT_RECEIVED_STATS, string(host) + "_" + time, stats)
}

func RecordHostInfo(info base.HostInfo, time string) {
	MetricsLogger.WithField("host", info.HostId).Infof("Recording info for host %s", info.HostId)
	MetricsLogger.WithField("host", info.HostId).Infof("Info: %+v", info)
	db.Audit.Add(db.BUCKET_AUDIT_RECEIVED_HOST_INFO, string(info.HostId) + "_" + time, info)
}
