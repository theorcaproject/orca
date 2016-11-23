package api

import (
	"github.com/gorilla/mux"
	"net/http"
	"gatoor/orca/rewriteTrainer/state/configuration"
	"encoding/json"
	"gatoor/orca/rewriteTrainer/state/needs"
	"gatoor/orca/rewriteTrainer/state/cloud"
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/rewriteTrainer/metrics"
	"gatoor/orca/rewriteTrainer/planner"
	"fmt"
	"gatoor/orca/rewriteTrainer/config"
)

type Api struct{}
var ApiLogger = Logger.LoggerWithField(Logger.Logger, "module", "api")
var Sampler metrics.Sampler

func (api Api) Init (port int) {
	ApiLogger.Infof("Initializing Api on Port %d", port)

	r := mux.NewRouter()

	r.HandleFunc("/push", pushHandler)

	r.HandleFunc("/state/config", getStateConfiguration)
	r.HandleFunc("/state/cloud", getStateCloud)
	r.HandleFunc("/state/needs", getStateNeeds)

	r.HandleFunc("/state/config/applications", getConfigApps)

	http.Handle("/", r)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		ApiLogger.Fatalf("Api failed to start - %s", err)
	}
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

func pushHandler(w http.ResponseWriter, r *http.Request) {
	hostInfo, stats, err := Sampler.ParsePush(r)
	if err != nil {
		json.NewEncoder(w).Encode(nil)
		return
	}
	Sampler.RecordStats(hostInfo.HostId, stats)
	Sampler.RecordHostInfo(hostInfo)
	returnJson(w, planner.Config(hostInfo.HostId))
}

func getStateConfiguration(w http.ResponseWriter, r *http.Request) {
	returnJson(w, state_configuration.GlobalConfigurationState.Snapshot())
}

func getStateCloud(w http.ResponseWriter, r *http.Request) {
	returnJson(w, state_cloud.GlobalCloudLayout.Snapshot())
}

func getConfigApps(w http.ResponseWriter, r *http.Request) {
	if(r.Method == "POST"){
		/* We need to create a new configuration object */
		var object config.AppJsonConfiguration
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&object); err != nil {
			ApiLogger.Infof("An error occurred while reading the application information")
		}

		ApiLogger.Infof("Read new configuration for application %s", object.Name)
		ApiLogger.Infof("version %s", object.Version)

		var new_version = object.Version + 1
		state_configuration.GlobalConfigurationState.ConfigureApp(state_configuration.AppConfiguration{
			Name: object.Name,
			Type: object.Type,
			Version: new_version,
			InstallFiles: object.InstallFiles,
			InstallCommands: object.InstallCommands,
			QueryStateCommand: object.QueryStateCommand,
			RemoveCommand: object.RemoveCommand,

			Min: object.Min,
			Desired: object.Desired,
			Max: object.Max,
		})
	}
	returnJson(w, state_configuration.GlobalConfigurationState.GetAllApps())
}

func getStateNeeds(w http.ResponseWriter, r *http.Request) {
	returnJson(w, state_needs.GlobalAppsNeedState.Snapshot())
}