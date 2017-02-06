/*
Copyright Alex Mack and Michael Lawson (michael@sphinix.com)
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
)

type StateStore struct {
	hosts map[string]*model.Host;
}

func (store *StateStore) Init() {
	store.hosts = make(map[string]*model.Host);
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

func (store *StateStore) HostCheckin(hostId string, checkin model.HostCheckinDataPackage) (*model.Host, error) {
	host, err := store.GetConfiguration(hostId)
	if err != nil {
		host = &model.Host{
			Id: hostId, LastSeen: "", FirstSeen: time.Now().Format(time.RFC3339Nano), State: "running", Apps: []model.Application{}, Changes: []model.ChangeApplication{}, Resources: model.HostResources{},
		}
		store.hosts[hostId] = host

		Audit.Insert__AuditEvent(AuditEvent{Details:map[string]string{
			"message": "Discovered new host " + hostId,
			"host": hostId,
		}})
	}

	for change, contains := range checkin.ChangesApplied {
		if contains {
			store.RemoveChange(host.Id, change)
		}
	}

	host.LastSeen = time.Now().Format(time.RFC3339Nano)
	host.Apps = make([]model.Application, 0)
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