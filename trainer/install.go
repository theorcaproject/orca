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
		"echo orca | sudo -S sh -c \"echo '" + trainerIp + "' >> /etc/hosts\"",
	}
}

var InstallLogger = log.LoggerWithField(log.Logger, "Type", "install")

func installNewInstance(instanceIp string) bool {
	InstallLogger.Info(fmt.Sprintf("Starting Base Calf install at [%s]", instanceIp))
	addHostToLayout(cloudLayoutDesired, instanceIp)
	userName := "orca"
	session, addr := orcaSSh.Connect(userName, instanceIp + ":22")
	for _, cmd := range ubuntu1604("172.16.147.1") {
		res := orcaSSh.ExecuteSshCommand(session, addr, cmd)
		if !res {
			InstallLogger.Info(fmt.Sprintf("Base Calf install failed for host [%s].", instanceIp))
			deleteHostFromLayout(cloudLayoutDesired, instanceIp)
			return false
		}
	}
	addHostToLayout(cloudLayoutCurrent, instanceIp)
	return true
}

func createInstance() {
	InstallLogger.Info("Starting instance provisioning.")
	instanceIp := cloudProvider.NewInstance()
	InstallLogger.Info(fmt.Sprintf("Instance provisioned at %s", instanceIp))
	installNewInstance(instanceIp)
}