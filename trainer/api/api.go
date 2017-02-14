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
	"github.com/twinj/uuid"
)

type Api struct {
	configurationStore *configuration.ConfigurationStore
	state              *state.StateStore
	cloudProvider      *cloud.CloudProvider

	sessions           map[string]bool
}

type Logs struct {
	StdOut string
	StdErr string
}

var ApiLogger = log.LoggerWithField(log.Logger, "module", "api")

func (api *Api) Init(port int, configurationStore *configuration.ConfigurationStore, state *state.StateStore, cloudProvider *cloud.CloudProvider) {
	api.configurationStore = configurationStore
	api.state = state
	api.cloudProvider = cloudProvider
	api.sessions = make(map[string]bool)

	ApiLogger.Infof("Initializing Api on Port %d", port)
	r := mux.NewRouter()

	/* Routes for the client */
	r.HandleFunc("/authenticate", api.authenticate)
	r.HandleFunc("/settings", api.getSettings)
	r.HandleFunc("/config", api.getAllConfiguration)
	r.HandleFunc("/config/applications", api.getAllConfigurationApplications)
	r.HandleFunc("/config/applications/configuration/latest", api.getAllConfigurationApplications_Configurations_Latest)
	r.HandleFunc("/state", api.getAllRunningState)
	r.HandleFunc("/checkin", api.hostCheckin)

	r.HandleFunc("/state/cloud/host/performance", api.getHostPerformance)
	r.HandleFunc("/state/cloud/application/performance", api.getAppPerformance)

	r.HandleFunc("/state/cloud/audit", api.getAudit)
	r.HandleFunc("/state/cloud/host/audit", api.getHostAudit)
	r.HandleFunc("/state/cloud/application/audit", api.getApplicationAudit)

	r.HandleFunc("/state/cloud/logs", api.getAllLogs)
	r.HandleFunc("/state/cloud/host/logs", api.getHostLogs)
	r.HandleFunc("/state/cloud/application/logs", api.getApplicationLogs)

	r.HandleFunc("/log", api.getLogs)
	r.HandleFunc("/log/apps", api.pushLogs)

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
	if (api.authenticate_user(w, r)) {
		returnJson(w, api.configurationStore.GetAllConfiguration())
	}
}

func (api *Api) getAllConfigurationApplications(w http.ResponseWriter, r *http.Request) {
	if (api.authenticate_user(w, r)) {
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
				application.DeploymentSchedule = object.DeploymentSchedule
				api.configurationStore.Save()
			}

		}

		listOfApplications := []*model.ApplicationConfiguration{}
		for _, application := range api.configurationStore.GetAllConfiguration() {
			listOfApplications = append(listOfApplications, application)
		}
		returnJson(w, listOfApplications)
	}
}

func (api *Api) getAllConfigurationApplications_Configurations_Latest(w http.ResponseWriter, r *http.Request) {
	if (api.authenticate_user(w, r)) {
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
}

func (api *Api) getAllRunningState(w http.ResponseWriter, r *http.Request) {
	if (api.authenticate_user(w, r)) {
		returnJson(w, api.state.GetAllHosts())
	}
}

func (api *Api) hostCheckin(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if (api.configurationStore.GlobalSettings.HostToken == token){
		var apps model.HostCheckinDataPackage
		hostId := r.URL.Query().Get("host")

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&apps); err != nil {
			ApiLogger.Infof("An error occurred while reading the application information")
		}

		_, err := api.state.GetConfiguration(hostId)

		if err != nil {
			host := &model.Host{
				Id: hostId,
				LastSeen: "",
				FirstSeen: time.Now().Format(time.RFC3339Nano),
				State: "running",
				Apps: []model.Application{},
				Changes: []model.ChangeApplication{},
				Resources: model.HostResources{},
			}
			ip, subnet, secGrps, isSpot := api.cloudProvider.Engine.GetHostInfo(cloud.HostId(hostId))
			host.Ip = ip
			host.Network = subnet
			host.SecurityGroups = secGrps
			host.SpotInstance = isSpot

			api.state.Add(hostId, host)

			state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
				Message: fmt.Sprintf("Discovered new server %s, ip: %s, subnet: %s spot: %t", hostId, ip, subnet, isSpot),
				Details:map[string]string{
					"host": hostId,
				}})
		}

		result, err := api.state.HostCheckin(hostId, apps)
		if err == nil {
			/* Lets tell the cloud provider that this host has checked in */
			api.cloudProvider.NotifyHostCheckIn(result)

			/* Lets save some stats */
			state.Stats.Insert__HostUtilisationStatistic(state.HostUtilisationStatistic{
				Cpu: apps.HostMetrics.CpuUsage,
				Mbytes: apps.HostMetrics.MemoryUsage,
				Network: apps.HostMetrics.NetworkUsage,
				HardDiskUsage: apps.HostMetrics.HardDiskUsage,
				HardDiskUsagePercent: apps.HostMetrics.HardDiskUsagePercent,
				Host: hostId,
				Timestamp:time.Now(),
			})

			returnJson(w, result.Changes)
			return
		} else {
			returnJson(w, nil)
		}
	}else{
		http.Error(w, "Token invalid", 403)
	}
}

func (api *Api) getLogs(w http.ResponseWriter, r *http.Request) {
	if (api.authenticate_user(w, r)) {
		returnJson(w, state.Audit.Query__HostLog(""))
	}
}


func (api *Api) pushLogs(w http.ResponseWriter, r *http.Request) {
	var logs map[string]Logs
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&logs); err == nil {
		host := r.URL.Query().Get("host")
		fmt.Println(fmt.Sprintf("Got logs from %s", host))
		for app, appLogs := range logs {
			state.Audit.Insert__Log(state.LogEvent{
				HostId: host, AppId: app, Message: appLogs.StdOut, LogLevel: "stdout",
			})
			state.Audit.Insert__Log(state.LogEvent{
				HostId: host, AppId: app, Message: appLogs.StdErr, LogLevel: "stderr",
			})
		}
	} else {
		fmt.Println(fmt.Sprintf("Log parsing error: %s", err))
	}
}

func (api *Api) getApplicationLogs(w http.ResponseWriter, r *http.Request) {
	if (api.authenticate_user(w, r)) {
		application := r.URL.Query().Get("application")
		returnJson(w, state.Audit.Query__AppLog(application))
	}
}

func (api *Api) getAllLogs(w http.ResponseWriter, r *http.Request) {
	if (api.authenticate_user(w, r)) {
		//returnJson(w, state.Audit.Query__AuditEvents())
	}
}

func (api *Api) getHostLogs(w http.ResponseWriter, r *http.Request) {
	if (api.authenticate_user(w, r)) {
		hostAudit := r.URL.Query().Get("host")
		returnJson(w, state.Audit.Query__HostLog(hostAudit))
	}
}

func (api *Api) getAudit(w http.ResponseWriter, r *http.Request) {
	if (api.authenticate_user(w, r)) {
		returnJson(w, state.Audit.Query__AuditEvents())
	}
}

func (api *Api) getHostAudit(w http.ResponseWriter, r *http.Request) {
	if (api.authenticate_user(w, r)) {
		hostAudit := r.URL.Query().Get("host")
		returnJson(w, state.Audit.Query__AuditEventsHost(hostAudit))
	}
}


func (api *Api) getApplicationAudit(w http.ResponseWriter, r *http.Request) {
	if (api.authenticate_user(w, r)) {
		application := r.URL.Query().Get("application")
		returnJson(w, state.Audit.Query__AuditEventsApplication(application))
	}
}

func (api *Api) getAppPerformance(w http.ResponseWriter, r *http.Request) {
	if (api.authenticate_user(w, r)) {
		application := r.URL.Query().Get("application")
		returnJson(w, state.Stats.Query__ApplicationUtilisationStatistic(application))
	}
}

func (api *Api) getHostPerformance(w http.ResponseWriter, r *http.Request) {
	if (api.authenticate_user(w, r)) {
		host := r.URL.Query().Get("host")
		returnJson(w, state.Stats.Query__HostUtilisationStatistic(host))
	}
}

func (api *Api) authenticate_user(w http.ResponseWriter, r *http.Request) bool {
	token := r.URL.Query().Get("token")
	if !api.sessions[token] {
		http.Error(w, "access denied", 403)
		return false
	}

	return true
}

func (api *Api) getSettings(w http.ResponseWriter, r *http.Request) {
	if (api.authenticate_user(w, r)) {
		returnJson(w, api.configurationStore.GlobalSettings)
	}
}

type AuthenticationResponse struct {
	Token string
}

func (api *Api) authenticate(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	password := r.URL.Query().Get("password")

	for name, userObject := range api.configurationStore.GlobalSettings.Users {
		if username == name && password == userObject.Password {
			token := uuid.NewV4().String()
			api.sessions[token] = true

			ar := AuthenticationResponse{
				Token:token,
			}
			returnJson(w, ar)
			return
		}
	}

	http.Error(w, "Authentication error", 403)
}
