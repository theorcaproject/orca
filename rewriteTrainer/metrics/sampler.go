package metrics

import (
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/rewriteTrainer/base"
	"net/http"
	"encoding/json"
	"errors"
	"gatoor/orca/rewriteTrainer/db"
)

var MetricsLogger = Logger.LoggerWithField(Logger.Logger, "module", "metrics")



type Usage float32

type HostStats struct {
	MemoryUsage Usage
	CpuUsage Usage
	NetworkUsage Usage
}

type AppStats struct {
	MemoryUsage Usage
	CpuUsage Usage
	NetworkUsage Usage
}

type HostInfo struct {
	HostId base.HostId
	IpAddr base.IpAddr
	OsInfo OsInfo
	HabitatInfo HabitatInfo
	Apps []AppInfo
}

type OsInfo struct {
	Os string
	Arch string
}

type HabitatInfo struct {
	Version base.Version
	Name base.HabitatName
	Status base.Status
}

type AppInfo struct {
	Type base.AppType
	Name base.AppName
	Version base.Version
	Status base.Status
}

type StatsWrapper struct {
	Host HostStats
	Apps map[base.AppName]AppStats
}

type PushWrapper struct {
	HostInfo HostInfo
	Stats StatsWrapper
}

func ParsePush(r *http.Request) (HostInfo, StatsWrapper, error) {
	decoder := json.NewDecoder(r.Body)
	var wrapper PushWrapper
	err := decoder.Decode(&wrapper)
	if err != nil {
		MetricsLogger.Errorf("StatsWrapper parsing failed - %s", err)
		return HostInfo{}, StatsWrapper{}, errors.New("Parsing failed")
	}
	return wrapper.HostInfo, wrapper.Stats, nil
}

func RecordStats(host base.HostId, stats StatsWrapper, time string) {
	MetricsLogger.WithField("host", host).Infof("Recording stats for host '%s'", host)
	MetricsLogger.WithField("host", host).Infof("Stats: %+v", stats)
	db.Audit.Add(db.BUCKET_AUDIT_RECEIVED_STATS, string(host) + "_" + time, stats)
}

func RecordHostInfo(info HostInfo, time string) {
	MetricsLogger.WithField("host", info.HostId).Infof("Recording info for host %s", info.HostId)
	MetricsLogger.WithField("host", info.HostId).Infof("Info: %+v", info)
	db.Audit.Add(db.BUCKET_AUDIT_RECEIVED_HOST_INFO, string(info.HostId) + "_" + time, info)
}
