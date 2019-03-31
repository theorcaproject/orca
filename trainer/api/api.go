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
	"encoding/json"
	"fmt"
	"net/http"
	"orca/trainer/cloud"
	"orca/trainer/configuration"
	"orca/trainer/model"
	"orca/trainer/state"
	log "orca/util/log"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/twinj/uuid"
)

type Api struct {
	configurationStore *configuration.ConfigurationStore
	state              *state.StateStore
	cloudProvider      *cloud.CloudProvider

	sessions map[string]bool
}

type Logs struct {
	StdOut string
	StdErr string
}

type ApplicationStatus struct {
	Name              string
	MinDeployment     int
	DesiredDeployment int
	Enabled           bool
	PendingChanges    int
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
	r.HandleFunc("/properties", api.getAllProperties)
	r.HandleFunc("/config", api.getAllConfiguration)
	r.HandleFunc("/config/applications", api.getAllConfigurationApplications)
	r.HandleFunc("/config/applications/status", api.getAllConfigurationApplications_Status)
	r.HandleFunc("/config/applications/configuration/latest", api.getAllConfigurationApplications_Configurations_Latest)
	r.HandleFunc("/config/applications/configuration/pending", api.configurationPending)
	r.HandleFunc("/config/applications/configuration/pending/approve", api.configurationApprove)
	r.HandleFunc("/state", api.getAllRunningState)
	r.HandleFunc("/checkin", api.hostCheckin)

	r.HandleFunc("/state/cloud/host/performance", api.getHostPerformance)
	r.HandleFunc("/state/cloud/host/latest/performance", api.getHostLatestPerformance)
	r.HandleFunc("/state/cloud/application/performance", api.getAppPerformance)
	r.HandleFunc("/state/cloud/application/host/performance", api.getAppHostPerformance)

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

func (api *Api) persistConfiguration() {
	api.configurationStore.Save()
	api.cloudProvider.BackupConfiguration(api.configurationStore.GetConfigAsString())
}

func (api *Api) getAllConfiguration(w http.ResponseWriter, r *http.Request) {
	if api.authenticate_user(w, r) {
		returnJson(w, api.configurationStore.GetAllConfiguration())
	}
}

func (api *Api) getAllConfigurationApplications_Status(w http.ResponseWriter, r *http.Request) {
	if api.authenticate_user(w, r) {
		var listOfApplications []*ApplicationStatus
		for _, application := range api.configurationStore.GetAllConfiguration() {
			var applicationStatus = &ApplicationStatus{}
			applicationStatus.Name = application.Name
			applicationStatus.MinDeployment = application.MinDeployment
			applicationStatus.DesiredDeployment = application.DesiredDeployment
			applicationStatus.Enabled = application.Enabled
			applicationStatus.PendingChanges = 0

			for _, publishedConfigChange := range application.PublishedConfig {
				if !publishedConfigChange.Approved {
					applicationStatus.PendingChanges += 1
				}
			}

			listOfApplications = append(listOfApplications, applicationStatus)
		}
		returnJson(w, listOfApplications)
	}
}

func (api *Api) getAllConfigurationApplications(w http.ResponseWriter, r *http.Request) {
	if api.authenticate_user(w, r) {
		if r.Method == "POST" {
			applicationName := r.URL.Query().Get("application")

			var object model.ApplicationConfiguration
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&object); err == nil {
				application, err := api.configurationStore.GetConfiguration(applicationName)
				if err != nil {
					object.Config = make(map[string]*model.VersionConfig)
					application = api.configurationStore.Add(applicationName, &object)
				}

				state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
					Message: "Modified application " + applicationName + " in pool",
					AppId:   applicationName,
				})

				application.MinDeployment = object.MinDeployment
				application.DesiredDeployment = object.DesiredDeployment
				application.Enabled = object.Enabled
				application.DisableSchedule = object.DisableSchedule
				application.DeploymentSchedule = object.DeploymentSchedule
				application.PropertyGroups = object.PropertyGroups
				application.ScheduleParts = object.ScheduleParts
				application.Depends = object.Depends
				api.persistConfiguration()
			}
		} else if r.Method == "DELETE" {
			applicationName := r.URL.Query().Get("application")
			application, err := api.configurationStore.GetConfiguration(applicationName)
			if err == nil && !application.Enabled {
				api.configurationStore.Remove(applicationName)
				api.persistConfiguration()
			} else {
				fmt.Println(err)
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
	if api.authenticate_user(w, r) {
		applicationName := r.URL.Query().Get("application")
		application, err := api.configurationStore.GetConfiguration(applicationName)
		if err == nil {
			if r.Method == "POST" {
				var object model.VersionConfig
				decoder := json.NewDecoder(r.Body)
				if err := decoder.Decode(&object); err == nil {
					newVersion := application.GetSuitableNextVersion()
					object.Version = newVersion

					application.Config[newVersion] = &object

					state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
						Message: "API: Modified application " + applicationName + ", created new configuration",
						AppId:   applicationName,
					})

					api.persistConfiguration()
				}
			}

			returnJson(w, application.GetLatestConfiguration())
			return
		}

		returnJson(w, nil)
	}
}

func (api *Api) configurationApprove(w http.ResponseWriter, r *http.Request) {
	if api.authenticate_user(w, r) {
		applicationName := r.URL.Query().Get("application")

		if applicationName == ""  {
			for _, appConfig := range api.configurationStore.GetAllConfigurationAsOrderedList() {
				for _, config := range appConfig.PublishedConfig {
					if !config.Approved {
						config.Approved = true

						state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
							Message: "API: Configuration approved,  " + applicationName + ", version" + config.Version,
							AppId:   applicationName,
						})
					}
				}
			}
		} else {
			application, err := api.configurationStore.GetConfiguration(applicationName)
			if err == nil {
				for _, config := range application.PublishedConfig {
					config.Approved = true

					state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
						Message: "API: Configuration approved,  " + applicationName + ", version" + config.Version,
						AppId:   applicationName,
					})
				}

			}
		}

		api.persistConfiguration()
		returnJson(w, nil)
		return
	}
}

func (api *Api) configurationPending(w http.ResponseWriter, r *http.Request) {
	if api.authenticate_user(w, r) {
		applicationName := r.URL.Query().Get("application")
		application, err := api.configurationStore.GetConfiguration(applicationName)
		var pending []*model.VersionConfig

		for _, config := range application.PublishedConfig {
			if config.Approved == false {
				pending = append(pending, config)
			}
		}

		if r.Method == "POST" {
			if err == nil {
			}
		}

		returnJson(w, pending)
		return
	}
}

func (api *Api) getAllRunningState(w http.ResponseWriter, r *http.Request) {
	if api.authenticate_user(w, r) {
		returnJson(w, api.state.GetAllHosts())
	}
}

func (api *Api) hostCheckin(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if api.configurationStore.GlobalSettings.HostToken == token {
		var apps model.HostCheckinDataPackage
		hostId := r.URL.Query().Get("host")

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&apps); err != nil {
			ApiLogger.Infof("An error occurred while reading the application information")
		}

		_, err := api.state.GetConfiguration(hostId)

		if err != nil {
			host := &model.Host{
				Id:        hostId,
				LastSeen:  "",
				FirstSeen: time.Now().Format(time.RFC3339Nano),
				State:     "running",
				Apps:      []model.Application{},
				Changes:   []model.ChangeApplication{},
				Resources: model.HostResources{},
			}
			ip, subnet, secGrps, isSpot, spotId := api.cloudProvider.Engine.GetHostInfo(cloud.HostId(hostId))
			host.GroupingTag = api.cloudProvider.Engine.GetTag("GroupingTag", host.Id)

			host.Ip = ip
			host.Network = subnet
			host.SecurityGroups = secGrps
			host.SpotInstance = isSpot
			host.SpotInstanceId = spotId
			api.state.Add(hostId, host)

			state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
				Message: fmt.Sprintf("Discovered new server %s, ip: %s, subnet: %s spot: %t", hostId, ip, subnet, isSpot),
				HostId:  hostId,
			})
		}

		result, err := api.state.HostCheckin(hostId, apps)
		if err == nil {
			/* Lets tell the cloud provider that this host has checked in */
			api.cloudProvider.NotifyHostCheckIn(result)

			/* Lets save some stats */
			state.Stats.Insert__HostUtilisationStatistic(state.HostUtilisationStatistic{
				Cpu:                  apps.HostMetrics.CpuUsage,
				Mbytes:               apps.HostMetrics.MemoryUsage,
				Network:              apps.HostMetrics.NetworkUsage,
				HardDiskUsage:        apps.HostMetrics.HardDiskUsage,
				HardDiskUsagePercent: apps.HostMetrics.HardDiskUsagePercent,
				Host:                 hostId,
				Timestamp:            time.Now(),
			})

			returnJson(w, result.Changes)
			return
		} else {
			returnJson(w, nil)
		}
	} else {
		http.Error(w, "Token invalid", 403)
	}
}

func (api *Api) getLogs(w http.ResponseWriter, r *http.Request) {
	if api.authenticate_user(w, r) {
		returnJson(w, state.Audit.Query__HostLog("", "100", "", ""))
	}
}

func (api *Api) pushLogs(w http.ResponseWriter, r *http.Request) {
	var logs map[string]Logs
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&logs); err == nil {
		host := r.URL.Query().Get("host")
		for app, appLogs := range logs {
			if len(appLogs.StdErr) > 0 {
				entries := strings.Split(appLogs.StdErr, "\n")

				for i := len(entries) - 1; i >= 0; i-- {
					state.Audit.Insert__Log(state.LogEvent{
						HostId: host, AppId: app, Message: entries[i], LogLevel: "stderr",
					})
				}
			}

			if len(appLogs.StdOut) > 0 {
				entries := strings.Split(appLogs.StdOut, "\n")
				for i := len(entries) - 1; i >= 0; i-- {
					state.Audit.Insert__Log(state.LogEvent{
						HostId: host, AppId: app, Message: entries[i], LogLevel: "stdout",
					})
				}
			}
		}
	} else {
		fmt.Println(fmt.Sprintf("Log parsing error: %s", err))
	}
}

func (api *Api) getApplicationLogs(w http.ResponseWriter, r *http.Request) {
	if api.authenticate_user(w, r) {
		application := r.URL.Query().Get("application")
		limit := r.URL.Query().Get("limit")
		search := r.URL.Query().Get("search")
		lasttime := r.URL.Query().Get("lasttime")

		returnJson(w, state.Audit.Query__AppLog(application, limit, search, lasttime))
	}
}

func (api *Api) getAllLogs(w http.ResponseWriter, r *http.Request) {
	if api.authenticate_user(w, r) {
		//returnJson(w, state.Audit.Query__AuditEvents())
	}
}

func (api *Api) getHostLogs(w http.ResponseWriter, r *http.Request) {
	if api.authenticate_user(w, r) {
		limit := r.URL.Query().Get("limit")
		search := r.URL.Query().Get("search")
		hostAudit := r.URL.Query().Get("host")
		lasttime := r.URL.Query().Get("lasttime")
		returnJson(w, state.Audit.Query__HostLog(hostAudit, limit, search, lasttime))
	}
}

func (api *Api) getAudit(w http.ResponseWriter, r *http.Request) {
	if api.authenticate_user(w, r) {
		limit := r.URL.Query().Get("limit")
		search := r.URL.Query().Get("search")
		lasttime := r.URL.Query().Get("lasttime")
		returnJson(w, state.Audit.Query__AuditEvents(limit, search, lasttime))
	}
}

func (api *Api) getHostAudit(w http.ResponseWriter, r *http.Request) {
	if api.authenticate_user(w, r) {
		hostAudit := r.URL.Query().Get("host")
		limit := r.URL.Query().Get("limit")
		search := r.URL.Query().Get("search")
		lasttime := r.URL.Query().Get("lasttime")
		returnJson(w, state.Audit.Query__AuditEventsHost(hostAudit, limit, search, lasttime))
	}
}

func (api *Api) getApplicationAudit(w http.ResponseWriter, r *http.Request) {
	if api.authenticate_user(w, r) {
		application := r.URL.Query().Get("application")
		limit := r.URL.Query().Get("limit")
		search := r.URL.Query().Get("search")
		lasttime := r.URL.Query().Get("lasttime")
		returnJson(w, state.Audit.Query__AuditEventsApplication(application, limit, search, lasttime))
	}
}

func (api *Api) getAppPerformance(w http.ResponseWriter, r *http.Request) {
	if api.authenticate_user(w, r) {
		application := r.URL.Query().Get("application")
		returnJson(w, state.Stats.Query__ApplicationUtilisationStatistic(application))
	}
}

func (api *Api) getAppHostPerformance(w http.ResponseWriter, r *http.Request) {
	if api.authenticate_user(w, r) {
		application := r.URL.Query().Get("application")
		host := r.URL.Query().Get("host")
		returnJson(w, state.Stats.Query__ApplicationHostUtilisationStatistic(application, host))
	}
}

func (api *Api) getHostPerformance(w http.ResponseWriter, r *http.Request) {
	if api.authenticate_user(w, r) {
		host := r.URL.Query().Get("host")
		returnJson(w, state.Stats.Query__HostUtilisationStatistic(host))
	}
}

func (api *Api) getHostLatestPerformance(w http.ResponseWriter, r *http.Request) {
	if api.authenticate_user(w, r) {
		host := r.URL.Query().Get("host")
		returnJson(w, state.Stats.Query__LatestHostUtilisationStatistic(host))
	}
}

func (api *Api) authenticate_user(w http.ResponseWriter, r *http.Request) bool {
	token := r.URL.Query().Get("token")

	for _, allowedToken := range api.configurationStore.GlobalSettings.ApiTokens {
		if token == allowedToken.Token {
			return true
		}
	}

	if api.sessions[token] {
		return true
	}

	http.Error(w, "access denied", 403)
	return false
}

func (api *Api) getSettings(w http.ResponseWriter, r *http.Request) {
	if api.authenticate_user(w, r) {
		if r.Method == "POST" {
			var object configuration.GlobalSettings
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&object); err == nil {
				state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
					Message: "Global configuration was modified",
				})
				api.configurationStore.GlobalSettings = object
				api.persistConfiguration()
			} else {
				fmt.Println(fmt.Sprintf("Config update error: %s", err))
			}
		}

		returnJson(w, api.configurationStore.GlobalSettings)
	}
}

func (api *Api) getAllProperties(w http.ResponseWriter, r *http.Request) {
	if api.authenticate_user(w, r) {
		if r.Method == "POST" {
			var object model.PropertyGroup
			decoder := json.NewDecoder(r.Body)
			if err := decoder.Decode(&object); err == nil {
				state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
					Message: "Updated property group",
				})

				object.Version = 0
				if oldVersion, ok := api.configurationStore.Properties[object.Name]; ok {
					object.Version += oldVersion.Version + 1
				}

				api.configurationStore.Properties[object.Name] = &object
				api.persistConfiguration()
			}
		}

		returnJson(w, api.configurationStore.Properties)
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
				Token: token,
			}
			returnJson(w, ar)
			return
		}
	}

	http.Error(w, "Authentication error", 403)
}
