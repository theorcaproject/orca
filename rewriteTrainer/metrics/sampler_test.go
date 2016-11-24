package metrics

import (
	"testing"
	"gatoor/orca/rewriteTrainer/db"
	"gatoor/orca/rewriteTrainer/base"
)


func TestSampler_RecordStats(t *testing.T) {
	db.Init("_test")
	time := "sometimestamp"
	stats :=  StatsWrapper{
		HostStats{1.0, 2.0, 3.5}, make(map[base.AppName]AppStats),
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
	info :=  HostInfo{
		HostId: "host1",
		IpAddr: "0.0.0.0",
		OsInfo: OsInfo{},
		HabitatInfo: HabitatInfo{},
		Apps: []AppInfo{},
	}
	RecordHostInfo(info, time)

	res := db.Audit.Get(db.BUCKET_AUDIT_RECEIVED_HOST_INFO, "host1_sometimestamp")

	if res != "{\"HostId\":\"host1\",\"IpAddr\":\"0.0.0.0\",\"OsInfo\":{\"Os\":\"\",\"Arch\":\"\"},\"HabitatInfo\":{\"Version\":\"\",\"Name\":\"\",\"Status\":\"\"},\"Apps\":[]}" {
		t.Error(res)
	}
	db.Close()
}
