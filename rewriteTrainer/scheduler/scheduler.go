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

package scheduler


import (
	Logger "gatoor/orca/rewriteTrainer/log"
	"time"
	"gatoor/orca/rewriteTrainer/state/cloud"
	"gatoor/orca/rewriteTrainer/planner"
	"gatoor/orca/rewriteTrainer/state/configuration"
	"gatoor/orca/rewriteTrainer/db"
)

var SchedulerLogger = Logger.LoggerWithField(Logger.Logger, "module", "scheduler")

var ticker *time.Ticker
var trackerTicker *time.Ticker


func Start() {
	SchedulerLogger.Infof("Scheduler starting")
	ticker := time.NewTicker(time.Second * 60)
	go func () {
		for {
			<- ticker.C
			run()
		}
	}()
}

func TriggerRun() {
	run()
}

func run() {
	SchedulerLogger.Info("Starting run()")
	state_cloud.GlobalCloudLayout.CheckCheckinTimeout()
	planner.Plan()

	/* Collect some statistics on whats running and whats desired */
	for appName, app := range state_configuration.GlobalConfigurationState.AllAppsLatest() {
		count_snapshot := db.ApplicationCountStatistic{}
		current_state := state_cloud.GlobalCloudLayout.Snapshot()

		running_count, _ :=current_state.Current.DeploymentCount(appName, app.Version)
		desired_count, _ :=current_state.Current.DeploymentCount(appName, app.Version)

		count_snapshot.AppName = appName
		count_snapshot.Running = running_count
		count_snapshot.Desired = desired_count
		count_snapshot.Timestamp = time.Now()

		db.Audit.Insert__ApplicationCountStatistic(count_snapshot)
	}

	/* Collect performance metrics on each application. Sadly this must be done this way otherwise we end up with no data
	gaps, which is not strictly correct */
	for targetAppName, _ := range state_configuration.GlobalConfigurationState.AllAppsLatest() {
		statistic := db.ApplicationUtilisationStatistic{}
		statistic.AppName = targetAppName
		statistic.Timestamp = time.Now()

		current_state := state_cloud.GlobalCloudLayout.Snapshot()
		for _, layoutElem := range current_state.Current.Layout {
			for appName, application := range layoutElem.Apps {
				if appName == targetAppName {
					minute_threshold := time.Now().Unix() - 120
					if application.StatisticPointTimestamp.Unix() > minute_threshold {
						statistic.Cpu += application.StatisticPoint.CpuUsage
						statistic.Mbytes += application.StatisticPoint.MemoryUsage
						statistic.Network += application.StatisticPoint.NetworkUsage
					}
				}
			}
		}

		db.Audit.Insert__ApplicationUtilisationStatistic(statistic)
	}

}
