package main

import (
	"encoding/json"
	log "gatoor/orca/base/log"
	"fmt"
	"os"
	"gatoor/orca/util"
	"gatoor/orca/modules/cloud"
	"gatoor/orca/modules/cloud/aws"
	"gatoor/orca/base"
)

var Logger = log.Logger

const (
	TRAINER_CONFIG_FILE = "/etc/orca/trainer/trainer.conf"
)

var jsonConf JsonConfiguration
var cloudProvider cloud.CloudProvider
var orcaCloud cloud.OrcaCloud

func main() {
	initTrainer()
}

func loadConfig () {
	Logger.Info("Starting Trainer...")
	file, err := os.Open(TRAINER_CONFIG_FILE)
	if err != nil {
		Logger.Fatal(err)
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&jsonConf); err != nil {
		extra := ""
		if serr, ok := err.(*json.SyntaxError); ok {
			line, col, highlight := util.HighlightBytePosition(file, serr.Offset)
			extra = fmt.Sprintf(":\nError at line %d, column %d (file offset %d):\n%s",
				line, col, serr.Offset, highlight)
		}
		Logger.Fatal("error parsing JSON object in config file %s%s\n%v",
			file.Name(), extra, err)
	}
}


func initTrainer() {
	jsonConf.Trainer.Port = 5000
	//initLayout()
	//initCloud()
	//initApi()
	Logger.Info("Trainer initialized.")
}

func initCloud() {
	Logger.Info(fmt.Sprintf("Initializing Cloud Provider %s", "aws"))
	cloudProvider = aws.AwsProvider{}
}

func initLayout() {
	orcaCloud.Current.Layout = make(map[base.HostId]cloud.CloudLayoutElement)
	orcaCloud.Desired.Layout = make(map[base.HostId]cloud.CloudLayoutElement)
}




