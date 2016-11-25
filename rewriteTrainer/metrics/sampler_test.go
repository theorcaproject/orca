package metrics

import (
	"testing"
	"gatoor/orca/rewriteTrainer/db"
	"gatoor/orca/base"
	"time"
)


func TestSampler_RecordStats(t *testing.T) {
	db.Init("_test")
	tim := "sometimestamp"
	stats :=  base.MetricsWrapper{}
	stats.HostMetrics = make(map[time.Time]base.HostStats)
	stats.AppMetrics = make(map[base.AppName]map[time.Time]base.AppStats)
	stats.HostMetrics[time.Unix(0,0)] = base.HostStats{3, 2, 1}
	RecordStats("host1", stats, tim)

	res := db.Audit.Get(db.BUCKET_AUDIT_RECEIVED_STATS, "host1_sometimestamp")

	if res != "{\"HostMetrics\":{\"1970-01-01T12:00:00+12:00\":{\"MemoryUsage\":3,\"CpuUsage\":2,\"NetworkUsage\":1}},\"AppMetrics\":{}}" {
		t.Error(res)
	}
	db.Close()
}


func TestSampler_RecordHostInfo(t *testing.T) {
	db.Init("_test")
	time := "sometimestamp"
	info :=  base.HostInfo{
		HostId: "host1",
		IpAddr: "0.0.0.0",
		OsInfo: base.OsInfo{},
		Apps: []base.AppInfo{},
	}
	RecordHostInfo(info, time)

	res := db.Audit.Get(db.BUCKET_AUDIT_RECEIVED_HOST_INFO, "host1_sometimestamp")

	if res != "{\"HostId\":\"host1\",\"IpAddr\":\"0.0.0.0\",\"OsInfo\":{\"Os\":\"\",\"Arch\":\"\"},\"Apps\":[]}" {
		t.Error(res)
	}
	db.Close()
}
