package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"fmt"
	log "gatoor/orca/base/log"
	"encoding/json"
)


var ApiLogger = log.LoggerWithField(log.Logger, "Type", "Api")

func initApi() {
	ApiLogger.Info("Init Api...")
	r := mux.NewRouter()
	r.HandleFunc("/stats", statsHandler)
	r.HandleFunc("/maintenance/instance/new", maintenanceInstanceNew)
	r.HandleFunc("/status", status)
	http.Handle("/", r)
	ApiLogger.Info(fmt.Sprintf("Api running at port %d", conf.Port))
	http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), nil)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	recordStats(r)
	sendConfig(w)
}

func maintenanceInstanceNew(w http.ResponseWriter, r *http.Request) {
	ApiLogger.Info("Got manual Instance provisioning request.")
	createInstance()
}

func status(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(orcaCloud)
}