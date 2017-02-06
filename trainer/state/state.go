package state

import (
	"errors"
	"time"
	"gatoor/orca/trainer/model"
	"fmt"
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

	fmt.Println("Metrics were ", checkin.Metrics)
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