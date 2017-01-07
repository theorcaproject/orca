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
	"gatoor/orca/rewriteTrainer/state/configuration"
	"encoding/json"
	"gatoor/orca/rewriteTrainer/state/needs"
	"gatoor/orca/rewriteTrainer/state/cloud"
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/rewriteTrainer/db"
	"gatoor/orca/base"
	"gatoor/orca/rewriteTrainer/config"
)

type Api struct{
	ConfigManager *config.JsonConfiguration
}

var ApiLogger = Logger.LoggerWithField(Logger.Logger, "module", "api")
var apiInstance Api

func (api Api) Init() {
	apiInstance = api
	ApiLogger.Infof("Initializing Api on Port %d", state_configuration.GlobalConfigurationState.Trainer.Port)

	r := mux.NewRouter()

	/* Routes for the client */
	r.HandleFunc("/client/changes", getChanges) //This route pushes state and already pulls back []ChangeRequests
	r.HandleFunc("/client/push/events", pushEventsHandler) //TODO: This route is to push errors/failures
	r.HandleFunc("/client/push/logs", pushLogsHandler) //TODO: This route is to push logs from stdout/stderr
	r.HandleFunc("/client/push/state", pushStateHandler)

	r.HandleFunc("/state/config", getStateConfiguration)
	r.HandleFunc("/state/config/applications", getStateConfigurationApplications)
	r.HandleFunc("/state/config/applications/configuration", getStateConfigurationApplicationsConfigurationSet)
	r.HandleFunc("/state/config/cloud", getStateConfigurationCloudProviders)
	r.HandleFunc("/state/cloud", getStateCloud)
	r.HandleFunc("/state/cloud/application/performance", getAppPerformance)
	r.HandleFunc("/state/cloud/application/count", getAppCount)
	r.HandleFunc("/state/needs", getStateNeeds)
	r.HandleFunc("/audit", getAuditEvents)

	http.Handle("/", r)

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", state_configuration.GlobalConfigurationState.Trainer.Port), nil)
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

func getChanges(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Query().Get("host")
	changes := state_cloud.GlobalCloudLayout.PopChanges(base.HostId(host))
	returnJson(w, changes)
}

func pushStateHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var wrapper base.TrainerPushWrapper
	err := decoder.Decode(&wrapper)

	ApiLogger.Infof("Got metrics from host '%s'", wrapper.HostInfo.HostId)
	if err != nil {
		json.NewEncoder(w).Encode(nil)
		return
	}

	state_cloud.GlobalCloudLayout.Current.UpdateHost(wrapper.HostInfo,wrapper.Stats)
	returnJson(w, nil)
}

func pushEventsHandler(w http.ResponseWriter, r *http.Request) {
}

func pushLogsHandler(w http.ResponseWriter, r *http.Request) {
}

func getStateConfiguration(w http.ResponseWriter, r *http.Request) {
	ApiLogger.Infof("Query to getStateConfiguration")
	returnJson(w, state_configuration.GlobalConfigurationState.Snapshot())
}

func getStateConfigurationCloudProviders(w http.ResponseWriter, r *http.Request) {
	ApiLogger.Infof("Query to getStateConfigurationCloudProviders")
	if (r.Method == "POST") {
		db.Audit.Insert__AuditEvent(db.AuditEvent{Details:map[string]string{
			"message": "Cloud configuration was changed via the api",
		}})

		var object base.ProviderConfiguration
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&object); err != nil {
			ApiLogger.Infof("An error occurred while reading the application information")
		}

		state_configuration.GlobalConfigurationState.CloudProvider.MinInstances = object.MinInstances
		state_configuration.GlobalConfigurationState.CloudProvider.MaxInstances = object.MaxInstances
		state_configuration.GlobalConfigurationState.CloudProvider.Type = object.Type
		state_configuration.GlobalConfigurationState.CloudProvider.SSHKey = object.SSHKey
		state_configuration.GlobalConfigurationState.CloudProvider.SSHUser = object.SSHUser

		if (object.Type == "AWS") {
			state_configuration.GlobalConfigurationState.CloudProvider.AWSConfiguration.AMI = object.AWSConfiguration.AMI
			state_configuration.GlobalConfigurationState.CloudProvider.AWSConfiguration.Key = object.AWSConfiguration.Key
			state_configuration.GlobalConfigurationState.CloudProvider.AWSConfiguration.Secret = object.AWSConfiguration.Secret
			state_configuration.GlobalConfigurationState.CloudProvider.AWSConfiguration.Region = object.AWSConfiguration.Region
			state_configuration.GlobalConfigurationState.CloudProvider.AWSConfiguration.SecurityGroupId = object.AWSConfiguration.SecurityGroupId
		}

		apiInstance.ConfigManager.Save()
	}

	returnJson(w, state_configuration.GlobalConfigurationState.CloudProvider)
}

func getStateConfigurationApplications(w http.ResponseWriter, r *http.Request) {
	ApiLogger.Infof("Query to getStateConfigurationApplications")
	if (r.Method == "POST") {
		/* We need to create a new configuration object */

		var object base.AppConfiguration
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&object); err != nil {
			ApiLogger.Infof("An error occurred while reading the application information")
		}

		db.Audit.Insert__AuditEvent(db.AuditEvent{Details:map[string]string{
			"message": "Scaling configuration details were changed for application "+ string(object.Name) +" via the api",
			"application":string(object.Name),
		}})

		ApiLogger.Infof("Read new configuration for application %s", object.Name)
		existingConfiguration, err := state_configuration.GlobalConfigurationState.GetApp(object.Name)
		configurationSets := make([]base.AppConfigurationSet, 0)
		if err == nil {
			configurationSets = existingConfiguration.ConfigurationSets
		}

		state_configuration.GlobalConfigurationState.ConfigureApp(base.AppConfiguration{
			Name: object.Name,
			Type: object.Type,
			TargetDeploymentCount: object.TargetDeploymentCount,
			MinDeploymentCount: object.MinDeploymentCount,
			Needs: object.Needs,
			ConfigurationSets: configurationSets,
		})

		apiInstance.ConfigManager.Save()
	}

	returnJson(w, state_configuration.GlobalConfigurationState.AllAppsLatest())
}

func getStateConfigurationApplicationsConfigurationSet(w http.ResponseWriter, r *http.Request) {
	ApiLogger.Infof("Query to setStateConfigurationApplicationsConfigurationSet")
	applicationName := r.URL.Query().Get("application")

	if (r.Method == "POST") {
		/* We need to create a new configuration object */
		var object base.AppConfigurationSet
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&object); err != nil {
			ApiLogger.Infof("An error occurred while reading the application information")
		}

		db.Audit.Insert__AuditEvent(db.AuditEvent{Details:map[string]string{
			"message": "Application configuration details were changed for application "+  applicationName +" via the api",
			"application":applicationName,
		}})

		ApiLogger.Infof("Read new configuration for application %s", applicationName)
		targetApplication, _ := state_configuration.GlobalConfigurationState.GetApp(base.AppName(applicationName))
		configurationSets := targetApplication.ConfigurationSets

		//Add new configuration version
		object.Version = targetApplication.FindNewConfigurationSetId()
		targetApplication.ConfigurationSets = append(configurationSets, object)
		state_configuration.GlobalConfigurationState.ConfigureApp(targetApplication)
		apiInstance.ConfigManager.Save()
	}

	latestAppConfiguration, _ := state_configuration.GlobalConfigurationState.GetApp(base.AppName(applicationName))
	returnJson(w, latestAppConfiguration.LatestConfiguration())
}

func getStateCloud(w http.ResponseWriter, r *http.Request) {
	ApiLogger.Infof("Query to getStateCloud")
	returnJson(w, state_cloud.GlobalCloudLayout.Snapshot())
}

func getStateNeeds(w http.ResponseWriter, r *http.Request) {
	ApiLogger.Infof("Query to getStateNeeds")
	returnJson(w, state_needs.GlobalAppsNeedState.Snapshot())
}

func getAuditEvents(w http.ResponseWriter, r *http.Request) {
	ApiLogger.Infof("Query to getAuditEvents")

	application := r.URL.Query().Get("application")
	returnJson(w, db.Audit.Query__AuditEvents(base.AppName(application)))
}

func getAppPerformance(w http.ResponseWriter, r *http.Request) {
	ApiLogger.Infof("Query to getAppPerformance")
	application := r.URL.Query().Get("application")
	hostid := r.URL.Query().Get("hostid")
	if hostid == ""{
		returnJson(w, db.Audit.Query__ApplicationUtilisationStatistic(base.AppName(application)))
	}else{
		returnJson(w, db.Audit.Query__AppMetrics_Performance__ByMinute_SingleHost(base.AppName(application), base.HostId(hostid)))
	}
}

func getAppCount(w http.ResponseWriter, r *http.Request) {
	ApiLogger.Infof("Query to getAppCount")
	application := r.URL.Query().Get("application")
	returnJson(w, db.Audit.Query__ApplicationCountStatistic(base.AppName(application)))
}
