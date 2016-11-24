package installer

import (
	orcaSSh "gatoor/orca/util"
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/rewriteTrainer/base"
)

var InstallerLogger = Logger.LoggerWithField(Logger.Logger, "module", "installer")

func ubuntu1604(trainerIp base.IpAddr, hostId base.HostId) []string {
	return []string{
		"echo orca | sudo -S apt-get update",
		"echo orca | sudo -S apt-get install -y git golang supervisor",
		"echo orca | sudo -S mkdir -p /orca",
		"echo orca | sudo -S chmod -R 777 /orca",
		"echo orca | sudo -S mkdir -p /etc/orca",
		"echo orca | sudo -S sh -c \"echo '" + string(trainerIp) + " orcatrainer' >> /etc/hosts\"",
		"rm -rf /orca/src/gatoor && mkdir -p /orca/src/gatoor && cd /orca/src/gatoor && git clone -b awstest https://github.com/gatoor/orca.git",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca && go get github.com/c9s/goprocinfo/linux && go get github.com/Sirupsen/logrus && go get golang.org/x/crypto/ssh'",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca/base && go build'",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca/base/log && go build'",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca/util && go build'",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca/host && go install'",
		"echo orca | sudo -S sh -c \"echo '{\\\"PollInterval\\\": 10, \\\"TrainerUrl\\\": \\\"http://orcatrainer:5000/stats\\\", \\\"HostId\\\":\\\"" + string(hostId) + "\\\"}' > /etc/orca/host.conf\"",
		"echo orca | sudo -S sh -c 'nohup /orca/bin/host >> /orca/log'",
	}
}

func InstallNewInstance(hostId base.HostId, ipAddr base.IpAddr) bool {
	InstallerLogger.Infof("Starting install on host %s:%s", hostId, ipAddr)
	userName := "orca"
	session, addr := orcaSSh.Connect(userName, string(ipAddr) + ":22")
	if session == nil {
		InstallerLogger.Infof("Install on host %s:%s failed: No session", hostId, ipAddr)
		return false
	}
	instance := ubuntu1604(ipAddr, hostId)
	for _, cmd := range instance {
		res := orcaSSh.ExecuteSshCommand(session, addr, cmd)
		if !res {
			InstallerLogger.Infof("Install on host %s:%s failed", hostId, ipAddr)
			return false
		}
	}
	InstallerLogger.Infof("Install on host %s:%s success", hostId, ipAddr)
	return true
}
