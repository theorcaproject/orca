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
	"gatoor/orca/rewriteTrainer/metrics"
	"gatoor/orca/rewriteTrainer/planner"
)

type Api struct{}
var ApiLogger = Logger.LoggerWithField(Logger.Logger, "module", "api")
var Sampler metrics.Sampler

func (api Api) Init () {
	ApiLogger.Infof("Initializing Api on Port %d", state_configuration.GlobalConfigurationState.Trainer.Port)

	r := mux.NewRouter()

	r.HandleFunc("/push", pushHandler)

	r.HandleFunc("/state/config", getStateConfiguration)
	r.HandleFunc("/state/cloud", getStateCloud)
	r.HandleFunc("/state/needs", getStateNeeds)

	http.Handle("/", r)

	err := http.ListenAndServe(fmt.Sprintf(":%d", state_configuration.GlobalConfigurationState.Trainer.Port), nil)
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

func getStateNeeds(w http.ResponseWriter, r *http.Request) {
	returnJson(w, state_needs.GlobalAppsNeedState.Snapshot())
}