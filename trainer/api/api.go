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

package api

import (
	"github.com/gorilla/mux"
	"net/http"
	"fmt"
	"encoding/json"
	"orca/trainer/configuration"
	"orca/trainer/model"
	"orca/trainer/state"
	log "orca/util/log"
	"orca/trainer/cloud"
	"time"
)

type Api struct {
	configurationStore *configuration.ConfigurationStore
	state              *state.StateStore
	cloudProvider  		*cloud.CloudProvider
	globalSettings		*configuration.GlobalSettings
}

var ApiLogger = log.LoggerWithField(log.Logger, "module", "api")

func (api *Api) Init(port int, configurationStore *configuration.ConfigurationStore, state *state.StateStore, cloudProvider *cloud.CloudProvider, globalSettings *configuration.GlobalSettings) {
	api.configurationStore = configurationStore
	api.state = state
	api.cloudProvider = cloudProvider
	api.globalSettings = globalSettings

	ApiLogger.Infof("Initializing Api on Port %d", port)
	r := mux.NewRouter()

	/* Routes for the client */
	r.HandleFunc("/settings", api.getSettings)
	r.HandleFunc("/config", api.getAllConfiguration)
	r.HandleFunc("/config/applications", api.getAllConfigurationApplications)
	r.HandleFunc("/config/applications/configuration/latest", api.getAllConfigurationApplications_Configurations_Latest)
	r.HandleFunc("/state", api.getAllRunningState)
	r.HandleFunc("/checkin", api.hostCheckin)
	r.HandleFunc("/state/cloud/application/performance", api.getAppPerformance)

	r.HandleFunc("/audit", api.getAudit)
	r.HandleFunc("/audit/application", api.getAuditApplication)

	http.Handle("/", r)

	func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
		if err != nil {
			ApiLogger.Fatalf("Api failed to start - %s", err)
		}
	}()
}

func returnJson(w http.ResponseWriter, obj interface{}) {
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

func (api *Api) getAllConfigurationApplications(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		applicationName := r.URL.Query().Get("application")

		var object model.ApplicationConfiguration
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&object); err == nil {
			application, err := api.configurationStore.GetConfiguration(applicationName)
			if err != nil {
				object.Config = make(map[string]model.VersionConfig)
				application = api.configurationStore.Add(applicationName, &object)
			}

			state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
				Message:"Modified application " + applicationName + " in pool",
				Details:map[string]string{
				"application": applicationName,
			}})

			application.MinDeployment = object.MinDeployment
			application.DesiredDeployment = object.DesiredDeployment
			application.Enabled = object.Enabled
			application.DisableSchedule = object.DisableSchedule
			api.configurationStore.Save()
		}

	}

	listOfApplications := []*model.ApplicationConfiguration{}
	for _, application := range api.configurationStore.GetAllConfiguration() {
		listOfApplications = append(listOfApplications, application)
	}
	returnJson(w, listOfApplications)
}

func (api *Api) getAllConfigurationApplications_Configurations_Latest(w http.ResponseWriter, r *http.Request) {
	applicationName := r.URL.Query().Get("application")
	application, err := api.configurationStore.GetConfiguration(applicationName)
	if err == nil {
		if r.Method == "POST" {
			var object model.VersionConfig
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&object); err == nil {
				newVersion := application.GetNextVersion()
				object.Version = newVersion
				application.Config[newVersion] = object

				state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
					Message:"API: Modified application " + applicationName + ", created new configuration",
					Details:map[string]string{
					"application": applicationName,
				}})

				api.configurationStore.Save()
			}
		}

		returnJson(w, application.GetLatestConfiguration())
		return
	}

	returnJson(w, nil)
}

func (api *Api) getAllRunningState(w http.ResponseWriter, r *http.Request) {
	returnJson(w, api.state.GetAllHosts())
}

func (api *Api) hostCheckin(w http.ResponseWriter, r *http.Request) {
	var apps model.HostCheckinDataPackage
	hostId := r.URL.Query().Get("host")


	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&apps); err != nil {
		ApiLogger.Infof("An error occurred while reading the application information")
	}

	_, err := api.state.GetConfiguration(hostId)

	if err != nil {
		host := &model.Host{
			Id: hostId, LastSeen: "", FirstSeen: time.Now().Format(time.RFC3339Nano), State: "running", Apps: []model.Application{}, Changes: []model.ChangeApplication{}, Resources: model.HostResources{},
		}
		ip, subnet, secGrps := api.cloudProvider.Engine.GetHostInfo(cloud.HostId(hostId))
		host.Ip = ip
		host.Network = subnet
		host.SecurityGroups = secGrps
		api.state.Add(hostId, host)

		state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
			Message: "Discovered new host " + hostId,
			Details:map[string]string{
			"host": hostId,
		}})
	}

	result, err := api.state.HostCheckin(hostId, apps)
	if err == nil {
		/* Lets tell the cloud provider that this host has checked in */
		api.cloudProvider.NotifyHostCheckIn(result)
		returnJson(w, result.Changes)
		return
	} else {
		returnJson(w, nil)
	}
}

func (api *Api) getAudit(w http.ResponseWriter, r *http.Request) {
	returnJson(w, state.Audit.Query__AuditEvents(""))
}

func (api *Api) getAuditApplication(w http.ResponseWriter, r *http.Request) {
	applicationName := r.URL.Query().Get("application")
	returnJson(w, state.Audit.Query__AuditEvents(applicationName))
}

func (api *Api) getAppPerformance(w http.ResponseWriter, r *http.Request) {
	ApiLogger.Infof("Query to getAppPerformance")
	application := r.URL.Query().Get("application")
	returnJson(w, state.Stats.Query__ApplicationUtilisationStatistic(application))
}

func (api *Api) getSettings(w http.ResponseWriter, r *http.Request) {
	ApiLogger.Infof("Query to getSettings")
	returnJson(w, api.globalSettings)
}
