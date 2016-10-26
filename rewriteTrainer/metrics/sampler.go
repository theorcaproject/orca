package metrics

import (
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/rewriteTrainer/base"
	"net/http"
	"encoding/json"
	"errors"
)

var MetricsLogger = Logger.LoggerWithField(Logger.Logger, "module", "metrics")

type Sampler struct {}

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

func (s Sampler) ParsePush(r *http.Request) (HostInfo, StatsWrapper, error) {
	decoder := json.NewDecoder(r.Body)
	var wrapper PushWrapper
	err := decoder.Decode(&wrapper)
	if err != nil {
		MetricsLogger.Errorf("StatsWrapper parsing failed - %s", err)
		return HostInfo{}, StatsWrapper{}, errors.New("Parsing failed")
	}
	return wrapper.HostInfo, wrapper.Stats, nil
}

func (s Sampler) RecordStats(host base.HostId, stats StatsWrapper) {
	MetricsLogger.WithField("host", host).Info("Got stats")
	// save host and apps usage
}

func (s Sampler) RecordHostInfo(info HostInfo) {
	MetricsLogger.WithField("host", info.HostId).Info("Got info")
	//save to db
}
