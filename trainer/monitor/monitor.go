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

func (monitor *Monitor) DataQueue(cloudProvider *cloud.CloudProvider, queue model.DataQueue) {
	numMsgs := cloudProvider.MonitorQueue(queue.Name)
	alertThreshold, err := strconv.Atoi(queue.AlertThreshold)
	if err != nil {
		fmt.Println(err)
		return
	}

	if state, ok := monitor.monitorStates[queue.Name]; !ok {
		state = &MonitorState{}
		state.Alarm = false
		state.Name = queue.Name
		state.CountValue = 0
		monitor.monitorStates[queue.Name] = state
	}

	alarmState := numMsgs >= alertThreshold
	if alarmState != monitor.monitorStates[queue.Name].Alarm {
		monitor.monitorStates[queue.Name].Alarm = alarmState
		monitor.monitorStates[queue.Name].Alarm = alarmState
		monitor.Alert(monitor.monitorStates[queue.Name])
	}

}
