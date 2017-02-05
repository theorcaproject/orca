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

package api

import (
	"github.com/gorilla/mux"
	"net/http"
	"fmt"
	"encoding/json"
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/trainer/configuration"
	"gatoor/orca/trainer/state"
)

type Api struct{
	configurationStore *configuration.ConfigurationStore
	state *state.StateStore
}

var ApiLogger = Logger.LoggerWithField(Logger.Logger, "module", "api")

func (api *Api) Init(port int, configurationStore *configuration.ConfigurationStore, state *state.StateStore) {
	api.configurationStore = configurationStore
	api.state = state
	ApiLogger.Infof("Initializing Api on Port %d", port)

	r := mux.NewRouter()

	/* Routes for the client */
	r.HandleFunc("/config", api.getAllConfiguration)
	r.HandleFunc("/state", api.getAllRunningState)
	r.HandleFunc("/checkin", api.hostCheckin)

	http.Handle("/", r)

	func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
		if err != nil {
			ApiLogger.Fatalf("Api failed to start - %s", err)
		}
	}()
}

func returnJson(w http.ResponseWriter, obj interface{}) {
	fmt.Printf("%+v", obj)

	j, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		ApiLogger.Errorf("Json serialization failed - %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func (api *Api) getAllConfiguration(w http.ResponseWriter, r *http.Request) {
	returnJson(w, api.configurationStore.GetAllConfiguration())
}

func (api *Api) getAllRunningState(w http.ResponseWriter, r *http.Request) {
	returnJson(w, api.state.GetAllHosts())
}


func (api *Api) hostCheckin(w http.ResponseWriter, r *http.Request) {
	var apps []state.ApplicationStateFromHost
	hostId := r.URL.Query().Get("host")

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&apps); err != nil {
		ApiLogger.Infof("An error occurred while reading the application information")
	}

	result, err := api.state.HostCheckin(hostId, apps)
	if err == nil {
		returnJson(w, result)
		return
	} else {
		returnJson(w, nil)
	}
}
