package main

import (
	"fmt"
	orcaSSh "gatoor/orca/util"
	log "gatoor/orca/base/log"
)

func ubuntu1604(trainerIp string) []string {
	return []string{
		"echo orca | sudo -S apt-get update",
		"echo orca | sudo -S apt-get install -y git golang supervisor",
		"echo orca | sudo -S mkdir -p /orca",
		"echo orca | sudo -S chmod -R 777 /orca",
		"echo orca | sudo -S mkdir -p /etc/orca",
		"echo orca | sudo -S sh -c \"echo '" + trainerIp + " orcatrainer' >> /etc/hosts\"",
		"rm -rf /orca/src/gatoor && mkdir -p /orca/src/gatoor && cd /orca/src/gatoor && git clone https://github.com/gatoor/orca.git",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca && go get github.com/c9s/goprocinfo/linux && go get github.com/Sirupsen/logrus && go get golang.org/x/crypto/ssh'",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca/base && go build'",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca/base/log && go build'",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca/util && go build'",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca/host && go install'",
		"echo orca | sudo -S sh -c \"echo '{\"PollInterval\": 10, \"TrainerUrl\": \"http://orcatrainer:5000\"}' > /etc/orca/host.conf\"",
	}
}

var InstallLogger = log.LoggerWithField(log.Logger, "Type", "install")

func installNewInstance(instanceIp string) bool {
	InstallLogger.Info(fmt.Sprintf("Starting Base Calf install at [%s]", instanceIp))
	orcaCloud.Desired.AddHost(instanceIp)
	userName := "orca"
	session, addr := orcaSSh.Connect(userName, instanceIp + ":22")
	for _, cmd := range ubuntu1604("172.16.147.1") {
		res := orcaSSh.ExecuteSshCommand(session, addr, cmd)
		if !res {
			InstallLogger.Info(fmt.Sprintf("Base Calf install failed for host [%s].", instanceIp))
			orcaCloud.Desired.DeleteHost(instanceIp)
			return false
		}
	}
	orcaCloud.Current.AddHost(instanceIp)
	InstallLogger.Info(fmt.Sprintf("Base Calf install complete for host [%s]", instanceIp))
	return true
}

func createInstance() {
	InstallLogger.Info("Starting instance provisioning.")
	instanceIp := cloudProvider.NewInstance()
	InstallLogger.Info(fmt.Sprintf("Instance provisioned at %s", instanceIp))
	installNewInstance(instanceIp)
}