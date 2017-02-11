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

package state

import (
	"errors"
	"time"
	"orca/trainer/model"
	"fmt"
)

type StateStore struct {
	hosts map[string]*model.Host;
}

func (store *StateStore) Init() {
	store.hosts = make(map[string]*model.Host);
}

func (store *StateStore) Add(hostId string, host *model.Host) {
	store.hosts[hostId] = host
}

func (store *StateStore) GetConfiguration(hostId string) (*model.Host, error) {
	if app, ok := store.hosts[hostId]; ok {
		return app, nil;
	}
	return nil, errors.New("Could not ");
}

func (store *StateStore) GetAllHosts() map[string]*model.Host {
	return store.hosts
}

func (store *StateStore) GetApplication(hostId string, applicationName string) (model.Application, error) {
	host, _ := store.GetConfiguration(hostId)
	for _, application := range host.Apps {
		if application.Name == applicationName {
			return application, nil
		}
	}
	return model.Application{}, errors.New("Could not find application")
}

func (store *StateStore) HostInit(host *model.Host) {
	host.State = "initializing"
	store.hosts[host.Id] = host
}

func (store *StateStore) HostCheckin(hostId string, checkin model.HostCheckinDataPackage) (*model.Host, error) {
	host, _ := store.GetConfiguration(hostId)

	for change, contains := range checkin.ChangesApplied {
		if contains {
			changeObject := host.GetChange(change)
			if changeObject != nil {
				if changeObject.Type == "add_application" {
					Audit.Insert__AuditEvent(AuditEvent{Severity: AUDIT__INFO,
						Message: fmt.Sprintf("Application %s was installed on host %s", changeObject.Name, changeObject.HostId),
						Details:map[string]string{
							"host": hostId,
							"application": changeObject.Name,
						}})

				}else if changeObject.Type == "remove_application" {
					Audit.Insert__AuditEvent(AuditEvent{Severity: AUDIT__INFO,
						Message: fmt.Sprintf("Application %s was uninstalled from host %s", changeObject.Name, changeObject.HostId),
						Details:map[string]string{
							"host": hostId,
							"application": changeObject.Name,
						}})
				}

				host.NumberOfChangeFailuresInRow = 0
			}

			store.RemoveChange(host.Id, change)
		}
	}
	host.LastSeen = time.Now().Format(time.RFC3339Nano)
	host.Apps = make([]model.Application, 0)

	if host.State != "running" {
		Audit.Insert__AuditEvent(AuditEvent{Severity: AUDIT__INFO,
			Message: fmt.Sprintf("Server %s state changed to running", hostId),
			Details:map[string]string{
				"host": hostId,
			}})
	}


	host.State = "running"
	for _, appStateFromHost := range checkin.State {
		host.Apps = append(host.Apps, appStateFromHost.Application)
	}
	return store.GetConfiguration(hostId)
}

func (store *StateStore) HasChanges() bool {
	for _, host := range store.hosts {
		if len(host.Changes) > 0 {
			return true;
		}
	}

	return false;
}

func (store *StateStore) RemoveChange(hostId string, changeId string) {
	host, _ := store.GetConfiguration(hostId)
	newChanges := make([]model.ChangeApplication, 0)
	for _, change := range host.Changes {
		if change.Id != changeId {
			newChanges = append(newChanges, change)
		}
	}
	host.Changes = newChanges
}

func (store *StateStore) RemoveHost(hostId string) {
	delete(store.hosts, hostId)
}

func (store *StateStore) ListOfHosts() []*model.Host {
	hosts := make([]*model.Host, 0)
	for _, host := range store.GetAllHosts() {
		hosts = append(hosts, host)
	}

	return hosts
}