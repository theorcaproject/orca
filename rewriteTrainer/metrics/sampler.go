package metrics

import (
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/base"
	"net/http"
	"encoding/json"
	"errors"
	"gatoor/orca/rewriteTrainer/db"
)

var MetricsLogger = Logger.LoggerWithField(Logger.Logger, "module", "metrics")

func ParsePush(r *http.Request) (base.HostInfo, base.MetricsWrapper, error) {
	decoder := json.NewDecoder(r.Body)
	var wrapper base.TrainerPushWrapper
	err := decoder.Decode(&wrapper)
	if err != nil {
		MetricsLogger.Errorf("TrainerPushWrapper parsing failed - %s", err)
		return base.HostInfo{}, base.MetricsWrapper{}, errors.New("Parsing failed")
	}
	return wrapper.HostInfo, wrapper.Stats, nil
}

func RecordStats(host base.HostId, stats base.MetricsWrapper, time string) {
	MetricsLogger.WithField("host", host).Infof("Recording stats for host '%s'", host)
	MetricsLogger.WithField("host", host).Infof("Stats: %+v", stats)
	db.Audit.Add(db.BUCKET_AUDIT_RECEIVED_STATS, string(host) + "_" + time, stats)
}

func RecordHostInfo(info base.HostInfo, time string) {
	MetricsLogger.WithField("host", info.HostId).Infof("Recording info for host %s", info.HostId)
	MetricsLogger.WithField("host", info.HostId).Infof("Info: %+v", info)
	db.Audit.Add(db.BUCKET_AUDIT_RECEIVED_HOST_INFO, string(info.HostId) + "_" + time, info)
}
