package state

import (
	"errors"
	"time"
	"fmt"
)

type Change struct {
	Id string

}

type HostResources struct {

}


type Application struct {
	Name string
	State string
	Version int
	ChangeId string
	Metrics string
}


type Host struct {
	Id string
	LastSeen string
	FirstSeen string
	State string
	Apps []Application
	Changes []Change
	Resources HostResources
}

type StateStore struct {
	hosts map[string]*Host;
}

func (store *StateStore) Init(){
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


func (store *StateStore) HostCheckin(hostId string, apps []Application) (*Host, error) {
	host, err := store.GetConfiguration(hostId)
	if err != nil {
		host = &Host{
			Id: hostId, LastSeen: "", FirstSeen: time.Now().Format(time.RFC3339Nano), State: "running", Apps: apps, Changes: []Change{}, Resources: HostResources{},
		}
		store.hosts[hostId] = host
	}
	fmt.Printf("%v", store.hosts)
	host.LastSeen = time.Now().Format(time.RFC3339Nano)
	host.Apps = apps
	return store.GetConfiguration(hostId)
}