package main

import (
	"fmt"
	orcaSSh "gatoor/orca/util"
	log "gatoor/orca/base/log"
	"gatoor/orca/base"
)

func ubuntu1604(trainerIp base.IpAddr, hostId base.HostId) []string {
	return []string{
		"echo orca | sudo -S apt-get update",
		"echo orca | sudo -S apt-get install -y git golang supervisor",
		"echo orca | sudo -S mkdir -p /orca",
		"echo orca | sudo -S chmod -R 777 /orca",
		"echo orca | sudo -S mkdir -p /etc/orca",
		"echo orca | sudo -S sh -c \"echo '" + string(trainerIp) + " orcatrainer' >> /etc/hosts\"",
		"rm -rf /orca/src/gatoor && mkdir -p /orca/src/gatoor && cd /orca/src/gatoor && git clone https://github.com/gatoor/orca.git",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca && go get github.com/c9s/goprocinfo/linux && go get github.com/Sirupsen/logrus && go get golang.org/x/crypto/ssh'",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca/base && go build'",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca/base/log && go build'",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca/util && go build'",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca/host && go install'",
		"echo orca | sudo -S sh -c \"echo '{\\\"PollInterval\\\": 10, \\\"TrainerUrl\\\": \\\"http://orcatrainer:5000/stats\\\", \\\"HostId\\\":\\\"" + string(hostId) + "\\\"}' > /etc/orca/host.conf\"",
		"echo orca | sudo -S sh -c 'nohup /orca/bin/host >> /orca/log'",
	}
}

var InstallLogger = log.LoggerWithField(log.Logger, "Type", "install")

func installNewInstance(hostId base.HostId, ipAddr base.IpAddr) bool {
	HostLogger := log.LoggerWithField(InstallLogger, "HostId", string(hostId))
	HostLogger.Info(fmt.Sprintf("Starting Base Calf install at %s", ipAddr))
	orcaCloud.Desired.AddHost(hostId, ipAddr)
	userName := "orca"
	session, addr := orcaSSh.Connect(userName, string(ipAddr) + ":22")
	for _, cmd := range ubuntu1604("172.16.147.1", hostId) {
		res := orcaSSh.ExecuteSshCommand(session, addr, cmd)
		if !res {
			HostLogger.Info(fmt.Sprintf("Base Calf install failed for host %s", ipAddr))
			orcaCloud.Desired.DeleteHost(hostId)
			return false
		}
	}
	orcaCloud.Current.AddHost(hostId, ipAddr)
	HostLogger.Info(fmt.Sprintf("Base Calf install complete for host %s", hostId))
	return true
}

func createInstance() {
	InstallLogger.Info("Starting instance provisioning.")
	hostId, instanceIpAddr := cloudProvider.NewInstance()
	InstallLogger.Info(fmt.Sprintf("Instance provisioned as %s", hostId))
	installNewInstance(hostId, instanceIpAddr)
}