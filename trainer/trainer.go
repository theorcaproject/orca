package main

import (
	"encoding/json"
	log "gatoor/orca/base/log"
	"fmt"
	"os"
	"gatoor/orca/util"
	"gatoor/orca/modules/cloud"
	"gatoor/orca/modules/cloud/aws"
)

var Logger = log.Logger

type TrainerConfiguration struct {
	Port int
}

const (
	TRAINER_CONFIG_FILE = "/etc/orca/trainer/trainer.conf"
)

var conf TrainerConfiguration
var cloudProvider cloud.CloudProvider
var cloudLayoutCurrent cloud.CloudLayout
var cloudLayoutDesired cloud.CloudLayout

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
	if err := decoder.Decode(&conf); err != nil {
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
	conf.Port = 5000
	initLayout()
	initCloud()
	initApi()
	Logger.Info("Trainer initialized.")
}

func initCloud() {
	Logger.Info(fmt.Sprintf("Initializing Cloud Provider %s", "aws"))
	cloudProvider = aws.AwsProvider{}
}

func initLayout() {
	cloudLayoutCurrent = make(map[string]cloud.CloudLayoutElement)
	cloudLayoutDesired = make(map[string]cloud.CloudLayoutElement)
}




