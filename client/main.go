package main


import (
	Logger "gatoor/orca/rewriteTrainer/log"
	"os"
	"encoding/json"
	"gatoor/orca/util"
	"fmt"
	"gatoor/orca/client/types"
	"gatoor/orca/client/client"
	"time"
	"io/ioutil"
	"gatoor/orca/base"
	"bytes"
	"net/http"
)

const (
	ORCA_VERSION = "0.1"
	CLIENT_CONFIG_FILE_PATH = "/orca/config/client/client.conf"
	APP_STATUS_FILE_PATH = "/orca/data/client/app_status.json"
	APP_CONFIG_FILE_PATH = "/orca/data/client/app_config.json"
)



var MainLogger = Logger.LoggerWithField(Logger.Logger, "module", "main")

func main() {
	loadConfig()
	client.Init()
	loadLastStateAndConfig()
	startScheduledTasks()
}

func startScheduledTasks() {
	pollTicker := time.NewTicker(time.Duration(client.Configuration.AppStatusPollInterval) * time.Second)
	metricsTicker := time.NewTicker(time.Duration(client.Configuration.MetricsPollInterval) * time.Second)
	trainerTicker := time.NewTicker(time.Duration(client.Configuration.TrainerPollInterval) * time.Second)
	go func () {
		for {
			<- metricsTicker.C
			client.PollMetrics()
		}
	}()
	go func () {
		for {
			<- pollTicker.C
			client.PollAppsState()
		}
	}()
	func () {
		for {
			<- trainerTicker.C
			CallTrainer()
		}
	}()
}

func CallTrainer() {
	MainLogger.Infof("Calling Trainer...")
	metrics := client.GetAppMetrics()
	state := client.AppsState
	config := client.AppsConfiguration
	saveStateAndConfig(state, config)

	wrapper := prepareData(state, metrics)

	MainLogger.Infof("Sending data to trainer: %+v", wrapper)
	b := new(bytes.Buffer)
	jsonErr := json.NewEncoder(b).Encode(wrapper)
	if jsonErr != nil {
		MainLogger.Errorf("Could not encode Metrics: %+v", jsonErr)
	}
	res, err := http.Post(client.Configuration.TrainerUrl, "application/json; charset=utf-8", b)
	if err != nil {
		MainLogger.Errorf("Could not send data to trainer: %+v", err)
	} else {
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			MainLogger.Errorf("Could not read reponse from trainer: %+v", err)
		} else {
			handleTrainerResponse(body)
		}
	}

	MainLogger.Infof("Metrics: %+v", metrics)
	MainLogger.Infof("State: %+v", state)
}

func prepareData(state types.AppsState, metrics base.AppMetrics) base.TrainerPushWrapper {
	var apps []base.AppInfo
	for _, app := range state {
		apps = append(apps, app)
	}
	hostInfo := base.HostInfo{HostId: client.Configuration.HostId, Apps: apps}
	return base.TrainerPushWrapper{hostInfo, base.MetricsWrapper{AppMetrics: metrics}}
}

func handleTrainerResponse(body []byte) {
	var config base.PushConfiguration
	if err := json.Unmarshal(body, &config); err != nil {
		MainLogger.Errorf("Failed to parse response - %s HTTP_BODY: %s", err, string(body))
	} else {
		MainLogger.Infof("Got Config with OrcaVersion %s: %+v", config.OrcaVersion, config)
		client.Handle(config)
	}
}

func saveStateAndConfig(state types.AppsState, conf types.AppsConfiguration) {
	var stateJson, err = json.Marshal(state)
	if err != nil {
		MainLogger.Errorf("AppsState JSON serialization failed: %+v", err)
		return
	}
	err = ioutil.WriteFile(APP_STATUS_FILE_PATH, stateJson, 0644)
	if err != nil {
		MainLogger.Errorf("Could not save file %s: %s", APP_STATUS_FILE_PATH, err)
	}
	var confJson, errConf = json.Marshal(conf)
	if errConf != nil {
		MainLogger.Errorf("AppsConfiguration JSON serialization failed: %+v", errConf)
		return
	}
	err = ioutil.WriteFile(APP_CONFIG_FILE_PATH, confJson, 0644)
	if err != nil {
		MainLogger.Errorf("Could not save file %s: %s", APP_CONFIG_FILE_PATH, err)
	}
}


func loadConfig() {
	file, err := os.Open(CLIENT_CONFIG_FILE_PATH)
	if err != nil {
		MainLogger.Fatalf("Could not open client config file at %s: %s", CLIENT_CONFIG_FILE_PATH, err)
	}
	loadJsonFile(file, &client.Configuration)
	file.Close()
}

func loadLastStateAndConfig() {
	hostFile, err := os.Open(APP_STATUS_FILE_PATH)
	if err != nil {
		MainLogger.Warnf("Failed to load AppStatus from file : %v", err)
	} else {
		loadJsonFile(hostFile, &client.AppsState)
	}
	hostFile.Close()
	appFile, err := os.Open(APP_CONFIG_FILE_PATH)
	if err != nil {
		MainLogger.Warnf("Failed to load AppConfig from file : %v", err)
	} else {
		loadJsonFile(appFile, &client.AppsConfiguration)
	}
	appFile.Close()
}

func loadJsonFile(file *os.File, t interface{}) {
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(t); err != nil {
		extra := ""
		if serr, ok := err.(*json.SyntaxError); ok {
			line, col, highlight := util.HighlightBytePosition(file, serr.Offset)
			extra = fmt.Sprintf(":\nError at line %d, column %d (file offset %d):\n%s",
				line, col, serr.Offset, highlight)
		}
		MainLogger.Errorf("error parsing JSON object in config file %s %s %v",
			file.Name(), extra, err)
	}
}