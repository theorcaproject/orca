package main

import (
	"net/http"
	"encoding/json"
	"gatoor/orca/base"
	"fmt"
)

func recordStats(r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var wrapper base.StatsWrapper
	err := decoder.Decode(&wrapper)
	if err != nil {
		Logger.Error(err)
	} else {
		updateCurrentCloudLayout(wrapper.HostInfo)
		Logger.Info(fmt.Printf("%+v", wrapper))
	}
}
