package scheduler


import (
	Logger "gatoor/orca/rewriteTrainer/log"
	"time"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/rewriteTrainer/planner"
	"gatoor/orca/rewriteTrainer/tracker"
	"fmt"
	"gatoor/orca/rewriteTrainer/state/needs"
	"gatoor/orca/rewriteTrainer/state/configuration"
)

var SchedulerLogger = Logger.LoggerWithField(Logger.Logger, "module", "scheduler")

var ticker *time.Ticker
var trackerTicker *time.Ticker


func Start() {
	SchedulerLogger.Infof("Scheduler starting")
	ticker := time.NewTicker(time.Second * 60)
	trackerTicker = time.NewTicker(time.Second * 10)
	go func () {
		for {
			<- ticker.C
			run()
		}
	}()
	go func () {
	       for {
		       <- trackerTicker.C
		       tracker.GlobalHostTracker.CheckCheckinTimeout()
	       }
	}()
}

func Stop() {
	ticker.Stop()
	trackerTicker.Stop()
	SchedulerLogger.Infof("Scheduler stopped")
}

func TriggerRun() {
	run()
}

func run() {
	SchedulerLogger.Info("Starting run()")
	//if planner.checkFailures() {
	//	// we have a failure handle
	//	//planner.handleFailures()
	//	return
	//}
	//
	//if diff := planner.Diff(nil , nil) {
	//	//we have diff, get cloud into nice state first
	//	planner.Queue.Apply(diff)
	//	return
	//}
	//planner.\

	fmt.Println(".........")
	fmt.Printf("Layout: %+v", state_cloud.GlobalCloudLayout)
	fmt.Println("")
	fmt.Println("")
	fmt.Printf("AvailableInstances: %+v", state_cloud.GlobalAvailableInstances)
	fmt.Println("")
	fmt.Println("")
	fmt.Printf("Needs: %+v", state_needs.GlobalAppsNeedState)
	fmt.Println("")
	fmt.Println("")
	fmt.Printf("Config: %+v", state_configuration.GlobalConfigurationState)
	fmt.Println("")
	fmt.Println("")
	//fmt.Printf("CloudProvider: %+v", cloud.CurrentProviderConfig)
	//fmt.Println("")
	//fmt.Println("")

	diff := planner.Diff(state_cloud.GlobalCloudLayout.Desired, state_cloud.GlobalCloudLayout.Current)
	fmt.Println("")
	fmt.Println("")
	fmt.Printf("Diff: %+v", diff)
	fmt.Println("")
	fmt.Println("")

	planner.Queue.Apply(diff)

	fmt.Println("")
	fmt.Println("")
	fmt.Printf("Queue: %+v", planner.Queue.Queue)
	fmt.Println("")
	fmt.Println("")
	//
	////analyzer.DoStuff
	//
	planner.Plan()

	fmt.Println("")
	fmt.Println("")
	fmt.Printf("Planned Layout: %+v", state_cloud.GlobalCloudLayout)
	fmt.Println("")
	fmt.Println("")
	fmt.Printf("Tracker: %+v", tracker.GlobalHostTracker)
	fmt.Println("")

	fmt.Println(".........")
	SchedulerLogger.Info("Finished run()")
}
