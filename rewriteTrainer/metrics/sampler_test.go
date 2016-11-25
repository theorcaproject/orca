package metrics

import (
	"testing"
	"gatoor/orca/rewriteTrainer/db"
	"gatoor/orca/base"
)


func TestSampler_RecordStats(t *testing.T) {
	db.Init("_test")
	time := "sometimestamp"
	stats :=  base.StatsWrapper{
		base.HostStats{1.0, 2.0, 3.5}, make(map[base.AppName]base.AppStats),
	}
	RecordStats("host1", stats, time)

	res := db.Audit.Get(db.BUCKET_AUDIT_RECEIVED_STATS, "host1_sometimestamp")

	if res != "{\"Host\":{\"MemoryUsage\":1,\"CpuUsage\":2,\"NetworkUsage\":3.5},\"Apps\":{}}" {
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
