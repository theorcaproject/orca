package planner


import (
	"testing"
	"gatoor/orca/rewriteTrainer/base"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"fmt"
)

func TestPlannerQueue_AllEmpty(t *testing.T) {
	queue := NewPlannerQueue()

	if queue.AllEmpty() == false {
		t.Error("should be all empty")
	}

	queue.Append("host1", state_cloud.AppsVersion{"1.0", 2})

	if queue.AllEmpty() == true {
		t.Error("should have elements")
	}
}

func TestPlannerQueue_Empty(t *testing.T) {
	queue := NewPlannerQueue()

	if queue.Empty("somehost") == false {
		t.Error("should be empty")
	}

	queue.Append("host1", state_cloud.AppsVersion{"1.0", 2})

	if queue.Empty("somehost") == false {
		t.Error("should be empty")
	}

	if queue.Empty("host1") == true {
		t.Error("should have elements")
	}
}

func TestPlannerQueue_PopSuccessFailState(t *testing.T) {
	queue := NewPlannerQueue()

	queue.Append("host1", state_cloud.AppsVersion{"1.0", 2})
	queue.Append("host2", state_cloud.AppsVersion{"10.0", 2})
	queue.Append("host1", state_cloud.AppsVersion{"2.0", 2})
	queue.Append("host1", state_cloud.AppsVersion{"2.0", 3})

	if queue.Empty("host1") == true || queue.Empty("host2") == true{
		t.Error("should have elements")
	}

	queue.SetState("host2", STATE_SUCCESS)

	elem, err := queue.Pop("host2")
	if err != nil {
		t.Error("unexpected error")
	}
	if elem.Version != "10.0" {
		t.Error("wrong version")
	}
	if queue.Empty("host2") == false {
		t.Error("pop didn't remove element")
	}

	queue.SetState("host1", STATE_FAIL)

	elem1, err1 := queue.Pop("host1")
	if err1 != nil {
		t.Error("unexpected error")
	}
	if elem1.Version != "1.0" {
		t.Error("wrong version")
	}
}


func TestPlannerQueue_PopQueuedApplyingState(t *testing.T) {
	queue := NewPlannerQueue()

	queue.Append("host1", state_cloud.AppsVersion{"1.0", 2})
	queue.Append("host2", state_cloud.AppsVersion{"10.0", 2})
	queue.Append("host1", state_cloud.AppsVersion{"2.0", 2})
	queue.Append("host1", state_cloud.AppsVersion{"2.0", 3})

	elem1, err1 := queue.Pop("host1")
	if err1 != nil {
		t.Error("unexpected error")
	}
	if elem1.Version != "1.0" {
		t.Error("wrong version")
	}



	elem2, err2 := queue.Pop("host1")
	if err2 != nil {
		t.Error("unexpected error")
	}

	if elem2.Version != "1.0" {
		t.Error("wrong version")
	}
	queue.SetState("host1", STATE_APPLYING)

	elem4, err4 := queue.Pop("host1")
	if err4 != nil {
		t.Error("unexpected error")
	}

	if elem4.Version != "1.0" {
		t.Error("wrong version")
	}
	queue.SetState("host1", STATE_FAIL)
	elem3, err3 := queue.Pop("host1")
	if err3 != nil {
		t.Error("unexpected error")
	}
	if elem3.DeploymentCount != 2 {
		t.Error("wrong deployment count")
	}

	if queue.Empty("host1") == true {
		t.Error("pop removed too many elements")
	}
}


func initAppsDiff() (map[base.AppName]state_cloud.AppsVersion, map[base.AppName]state_cloud.AppsVersion) {
	return make(map[base.AppName]state_cloud.AppsVersion), make(map[base.AppName]state_cloud.AppsVersion)
}

func TestPlanner_appsDiff_Equal(t *testing.T) {
	master, slave := initAppsDiff()

	master["app1"] = state_cloud.AppsVersion{"1.0", 1}
	master["app2"] = state_cloud.AppsVersion{"1.1", 2}

	slave["app1"] = state_cloud.AppsVersion{"1.0", 1}
	slave["app2"] = state_cloud.AppsVersion{"1.1", 2}

	diff := appsDiff(master, slave)

	if len(diff) != 0 {
		t.Error("found diff in equal apps")
	}
}


func TestPlanner_appsDiff_Update(t *testing.T) {
	master, slave := initAppsDiff()

	master["app1"] = state_cloud.AppsVersion{"2.0", 1}
	master["app2"] = state_cloud.AppsVersion{"2.1", 2}

	slave["app1"] = state_cloud.AppsVersion{"1.0", 1}
	slave["app2"] = state_cloud.AppsVersion{"1.1", 2}

	diff := appsDiff(master, slave)

	if len(diff) != 2 {
		t.Error("found no diff")
	}

	if diff["app1"].Version != "2.0" {
		t.Error("wrong version")
	}
	if diff["app2"].Version != "2.1" {
		t.Error("wrong version")
	}
}


func TestPlanner_appsDiff_ScaleUp(t *testing.T) {
	master, slave := initAppsDiff()

	master["app1"] = state_cloud.AppsVersion{"1.0", 2}
	master["app2"] = state_cloud.AppsVersion{"1.1", 2}

	slave["app1"] = state_cloud.AppsVersion{"1.0", 1}
	slave["app2"] = state_cloud.AppsVersion{"1.1", 2}

	diff := appsDiff(master, slave)

	if len(diff) != 1 {
		t.Error("found no diff")
	}

	if diff["app1"].Version != "1.0" {
		t.Error("wrong version")
	}
	if diff["app1"].DeploymentCount != 2 {
		t.Error("wrong count")
	}
}


func TestPlanner_appsDiff_DeployNew(t *testing.T) {
	master, slave := initAppsDiff()

	master["app1"] = state_cloud.AppsVersion{"1.0", 1}
	master["app2"] = state_cloud.AppsVersion{"1.1", 2}

	slave["app1"] = state_cloud.AppsVersion{"1.0", 1}

	diff := appsDiff(master, slave)

	if len(diff) != 1 {
		t.Error("found no diff")
	}

	if diff["app2"].Version != "1.1" {
		t.Error("wrong version")
	}
}


func TestPlanner_appsDiff_RemoveApp(t *testing.T) {
	master, slave := initAppsDiff()

	master["app1"] = state_cloud.AppsVersion{"1.0", 1}

	slave["app1"] = state_cloud.AppsVersion{"1.0", 1}
	slave["app2"] = state_cloud.AppsVersion{"1.1", 2}

	diff := appsDiff(master, slave)
	if len(diff) != 1 {
		t.Error("found no diff")
	}

	if diff["app2"].Version != "1.1" {
		t.Error("wrong version")
	}

	if diff["app2"].DeploymentCount != 0 {
		t.Error("wrong version")
	}
}


func TestPlanner_appsDiff_RollbackVersion(t *testing.T) {
	master, slave := initAppsDiff()

	master["app1"] = state_cloud.AppsVersion{"1.0", 1}
	master["app2"] = state_cloud.AppsVersion{"1.1", 2}

	slave["app1"] = state_cloud.AppsVersion{"1.0", 1}
	slave["app2"] = state_cloud.AppsVersion{"2.0", 2}

	diff := appsDiff(master, slave)

	if len(diff) != 1 {
		t.Error("found no diff")
	}

	if diff["app2"].Version != "1.1" {
		t.Error("wrong version")
	}
}


func TestPlanner_appsDiff_ScaleDown(t *testing.T) {
	master, slave := initAppsDiff()

	master["app1"] = state_cloud.AppsVersion{"1.0", 1}
	master["app2"] = state_cloud.AppsVersion{"1.1", 1}

	slave["app1"] = state_cloud.AppsVersion{"1.0", 1}
	slave["app2"] = state_cloud.AppsVersion{"1.1", 2}

	diff := appsDiff(master, slave)

	if len(diff) != 1 {
		t.Error("found no diff")
	}

	if diff["app2"].Version != "1.1" {
		t.Error("wrong version")
	}
	if diff["app2"].DeploymentCount != 1 {
		t.Error("wrong deployment count")
	}
}

func TestPlanner_appsDiff_Combination(t *testing.T) {
	master, slave := initAppsDiff()

	//do nothing
	master["app1"] = state_cloud.AppsVersion{"1.0", 1}
	slave["app1"] = state_cloud.AppsVersion{"1.0", 1}

	// update
	master["app2"] = state_cloud.AppsVersion{"2.0", 1}
	slave["app2"] = state_cloud.AppsVersion{"1.0", 1}

	//rollback
	master["app3"] = state_cloud.AppsVersion{"1.0", 1}
	slave["app3"] = state_cloud.AppsVersion{"2.0", 1}

	//deploy new
	master["app4"] = state_cloud.AppsVersion{"2.0", 1}

	//remove
	slave["app5"] = state_cloud.AppsVersion{"1.0", 1}

	//rollback scale up
	master["app6"] = state_cloud.AppsVersion{"1.0", 2}
	slave["app6"] = state_cloud.AppsVersion{"2.0", 1}

	//scale down
	master["app7"] = state_cloud.AppsVersion{"1.0", 1}
	slave["app7"] = state_cloud.AppsVersion{"1.0", 5}


	diff := appsDiff(master, slave)

	if len(diff) != 6 {
		t.Error("found no diff")
	}

	if diff["app2"].Version != "2.0" {
		t.Error("wrong version")
	}
	if diff["app2"].DeploymentCount != 1 {
		t.Error("wrong deployment count")
	}

	if diff["app3"].Version != "1.0" {
		t.Error("wrong version")
	}
	if diff["app3"].DeploymentCount != 1 {
		t.Error("wrong deployment count")
	}

	if diff["app4"].Version != "2.0" {
		t.Error("wrong version")
	}
	if diff["app4"].DeploymentCount != 1 {
		t.Error("wrong deployment count")
	}

	if diff["app5"].Version != "1.0" {
		t.Error("wrong version")
	}
	if diff["app5"].DeploymentCount != 0 {
		t.Error("wrong deployment count")
	}

	if diff["app6"].Version != "1.0" {
		t.Error("wrong version")
	}
	if diff["app6"].DeploymentCount != 2 {
		t.Error("wrong deployment count")
	}

	if diff["app7"].Version != "1.0" {
		t.Error("wrong version")
	}
	if diff["app7"].DeploymentCount != 1 {
		t.Error("wrong deployment count")
	}
}

func TestPlannerQueue_Snapshot(t *testing.T) {
	queue := NewPlannerQueue()

	queue.Append("host1", state_cloud.AppsVersion{"1.0", 2})
	queue.Append("host2", state_cloud.AppsVersion{"10.0", 2})

	snapshot := queue.Snapshot()

	queue.RemoveHost("host2")

	if queue.Empty("host2") == false {
		t.Error("RemoveHost didn't remove element")
	}

	//sanity check of empty check below
	if _, exists := queue.Queue["host2"]; exists {
		t.Error("RemoveHost didn't remove elem")
	}

	if _, exists := snapshot["host2"]; !exists {
		fmt.Println(queue.Queue)
		fmt.Println(snapshot)
		t.Error("RemoveHost removed snapshotted elem")
	}
}
