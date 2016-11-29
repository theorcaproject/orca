package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"fmt"
	log "gatoor/orca/base/log"
	"encoding/json"
	"gatoor/orca/base"
)


var ApiLogger = log.LoggerWithField(log.Logger, "Type", "Api")

func initApi() {
	ApiLogger.Info("Init Api...")
	r := mux.NewRouter()
	r.HandleFunc("/stats", statsHandler)
	r.HandleFunc("/maintenance/instance/new", maintenanceInstanceNew)
	r.HandleFunc("/status", status)
	http.Handle("/", r)
	ApiLogger.Info(fmt.Sprintf("Api running at port %d", jsonConf.Trainer.Port))
	http.ListenAndServe(fmt.Sprintf(":%d", jsonConf.Trainer.Port), nil)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	recordStats(r)
	sendConfig(w, determineTrainerUpdate())
}


func determineTrainerUpdate() base.TrainerUpdate {
	var trainerUpdate base.TrainerUpdate
	trainerUpdate.TargetHostId = "172.16.147.189"
	trainerUpdate.HabitatConfiguration = buildHabitatConfiguration()
	trainerUpdate.AppsConfiguration = make(map[base.HostId]base.AppConfiguration)
	trainerUpdate.AppsConfiguration["ngin"] = buildAppConfiguration()
	return trainerUpdate
}

func maintenanceInstanceNew(w http.ResponseWriter, r *http.Request) {
	ApiLogger.Info("Got manual Instance provisioning request.")
	createInstance()
}

func status(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(orcaCloud)
}