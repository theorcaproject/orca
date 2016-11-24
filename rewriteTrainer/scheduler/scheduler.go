package scheduler


import (
	//"gatoor/orca/rewriteTrainer/planner"
	//"gatoor/orca/rewriteTrainer/state/cloud"
	Logger "gatoor/orca/rewriteTrainer/log"
	"time"
)

var SchedulerLogger = Logger.LoggerWithField(Logger.Logger, "module", "scheduler")

var ticker *time.Ticker


func Start() {
	SchedulerLogger.Infof("Scheduler starting")
	ticker = time.NewTicker(time.Second * 10)
	for t := range ticker.C {
		SchedulerLogger.Infof("Scheduler tick %s, starting goroutine", t)
		go run()
	}
}

func Stop() {
	ticker.Stop()
	SchedulerLogger.Infof("Scheduler stopped")
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

	//diff := planner.Diff(state_cloud.GlobalCloudLayout.Desired, state_cloud.GlobalCloudLayout.Current)
	//planner.Queue.Apply(diff)

	//analyzer.DoStuff

	//planner.Plan()

	SchedulerLogger.Info("Finished run()")
}
