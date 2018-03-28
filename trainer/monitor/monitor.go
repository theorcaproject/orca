/*
Copyright Alex Mack (al9mack@gmail.com) and Michael Lawson (michael@sphinix.com)
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
package monitor

import (
	"fmt"
	"orca/trainer/cloud"
	"orca/trainer/model"
	"orca/trainer/state"
	"strconv"
)

type MonitorState struct {
	Alarm       bool
	Name        string
	CountValue  uint64
	StringValue string
}

type Monitor struct {
	monitorStates map[string]*MonitorState
}

var Monit Monitor

func (monitor *Monitor) Init() {
	monitor.monitorStates = make(map[string]*MonitorState)
}

func (monitor *Monitor) Alert(monitorState *MonitorState) {
	alarmString := "OK"
	if monitorState.Alarm {
		alarmString = "ALARM"
	}
	state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__MONITOR,
		Message: fmt.Sprintf(" %s alert state is %s", monitorState.Name, alarmString),
	})
}

func (monitor *Monitor) monitDataQueue(name string, alertThreshold int, numMsgs int) {
	if state, ok := monitor.monitorStates[name]; !ok {
		state = &MonitorState{}
		state.Alarm = false
		state.Name = name
		state.CountValue = 0
		monitor.monitorStates[name] = state
	}
	alarmState := numMsgs >= alertThreshold
	if alarmState != monitor.monitorStates[name].Alarm {
		monitor.monitorStates[name].Alarm = alarmState
		monitor.Alert(monitor.monitorStates[name])
	}

}

func (monitor *Monitor) DataQueue(cloudProvider *cloud.CloudProvider, queue model.DataQueue) {
	// monitor main queue
	if len(queue.AlertThreshold) > 0 {
		alertThreshold, err := strconv.Atoi(queue.AlertThreshold)
		if err != nil {
			fmt.Println(err)
			return
		}
		numMsgs := cloudProvider.MonitorQueue(queue.Name)
		if numMsgs >= 0 {
			monitor.monitDataQueue(queue.Name, alertThreshold, numMsgs)

			// monitor corresponding rogue queue
			if len(queue.RogueAlertThreshold) > 0 {
				rogueAlertThreshold, err := strconv.Atoi(queue.RogueAlertThreshold)
				if err != nil {
					fmt.Println(err)
					return
				}
				rogueNumMsgs := cloudProvider.MonitorQueue(queue.RogueName)
				monitor.monitDataQueue(queue.RogueName, rogueAlertThreshold, rogueNumMsgs)
			}
		}
	}
}
