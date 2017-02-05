package state

import (
	"errors"
	"time"
)

type ChangeApplication struct {
	Id     string
	Type   string
	HostId string
	Time   string
}

type ChangeServer struct {
	Id   string
	Type string
	Time string
}

type HostResources struct {

}

type Application struct {
	Name     string
	State    string
	Version  int
	ChangeId string
	Metrics  string
}

type ApplicationStateFromHost struct {
	Name           string
	Application    Application
	ChangesApplied map[string]bool
}

type Host struct {
	Id        string
	LastSeen  string
	FirstSeen string
	State     string
	Apps      []Application
	Changes   []ChangeApplication
	Resources HostResources
}

type StateStore struct {
	hosts map[string]*Host;
}

func (store *StateStore) Init() {
	store.hosts = make(map[string]*Host);
}

func (store *StateStore) GetConfiguration(hostId string) (*Host, error) {
	if app, ok := store.hosts[hostId]; ok {
		return app, nil;
	}
	return nil, errors.New("Could not ");
}

func (store *StateStore) GetAllHosts() map[string]*Host {
	return store.hosts
}

func (store *StateStore) GetApplication(hostId string, applicationName string) (Application, error) {
	host, _ := store.GetConfiguration(hostId)
	for _, application := range host.Apps {
		if application.Name == applicationName {
			return application, nil
		}
	}
	return Application{}, errors.New("Could not find application")
}

func (store *StateStore) HostCheckin(hostId string, apps []ApplicationStateFromHost) (*Host, error) {
	host, err := store.GetConfiguration(hostId)
	if err != nil {
		host = &Host{
			Id: hostId, LastSeen: "", FirstSeen: time.Now().Format(time.RFC3339Nano), State: "running", Apps: []Application{}, Changes: []ChangeApplication{}, Resources: HostResources{},
		}
		store.hosts[hostId] = host
	}

	for _, application := range apps {
		for change, contains:= range application.ChangesApplied {
			if contains {
				store.RemoveChange(host.Id, change)
			}
		}
	}

	host.LastSeen = time.Now().Format(time.RFC3339Nano)
	host.Apps = make([]Application, 0)
	for _, appStateFromHost := range apps {
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
	newChanges := make([]ChangeApplication, 0)
	for _, change := range host.Changes {
		if change.Id != changeId {
			newChanges = append(newChanges, change)
		}
	}
	host.Changes = newChanges
}