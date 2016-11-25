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
	"gatoor/orca/rewriteTrainer/responder"
	"gatoor/orca/rewriteTrainer/db"
	"gatoor/orca/rewriteTrainer/tracker"
	"gatoor/orca/base"
)

type Api struct{}
var ApiLogger = Logger.LoggerWithField(Logger.Logger, "module", "api")

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
	hostInfo, stats, err := metrics.ParsePush(r)
	ApiLogger.Infof("Got metrics from host '%s'", hostInfo.HostId)
	if err != nil {
		json.NewEncoder(w).Encode(nil)
		return
	}

	doHandlePush(hostInfo, stats)

	config, err := responder.GetConfigForHost(hostInfo.HostId)

	if err != nil {
		ApiLogger.Infof("Sending empty response to host '%s'", hostInfo.HostId)
		returnJson(w, nil)
		return
	}
	ApiLogger.Infof("Sending new config to host '%s': '%+v'", hostInfo.HostId, config)
	returnJson(w, config)
}

func doHandlePush(hostInfo base.HostInfo, stats base.MetricsWrapper) {
	timeString, time := db.GetNow()

	metrics.RecordStats(hostInfo.HostId, stats, timeString)
	metrics.RecordHostInfo(hostInfo, timeString)

	state_cloud.UpdateCurrent(hostInfo, timeString)
	responder.CheckAppState(hostInfo)
	tracker.GlobalHostTracker.Update(hostInfo.HostId, time)
}

func getStateConfiguration(w http.ResponseWriter, r *http.Request) {
	ApiLogger.Infof("Query to getStateConfiguration")
	returnJson(w, state_configuration.GlobalConfigurationState.Snapshot())
}

func getStateCloud(w http.ResponseWriter, r *http.Request) {
	ApiLogger.Infof("Query to getStateCloud")
	returnJson(w, state_cloud.GlobalCloudLayout.Snapshot())
}

func getStateNeeds(w http.ResponseWriter, r *http.Request) {
	ApiLogger.Infof("Query to getStateNeeds")
	returnJson(w, state_needs.GlobalAppsNeedState.Snapshot())
}

func doHandleCloudEvent() {

}