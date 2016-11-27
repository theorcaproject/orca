package installer

import (
	orcaSSh "gatoor/orca/util"
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/base"
	"gatoor/orca/rewriteTrainer/state/configuration"
)

var InstallerLogger = Logger.LoggerWithField(Logger.Logger, "module", "installer")

func ubuntu1604(trainerIp base.IpAddr, hostId base.HostId) []string {
	const (
		SUPERVISOR_CONFIG = "'[unix_http_server]\\nfile=/var/run/supervisor.sock\\nchmod=0770\\nchown=root:supervisor\\n[supervisord]\\nlogfile=/var/log/supervisor/supervisord.log\\npidfile=/var/run/supervisord.pid\\nchildlogdir=/var/log/supervisor\\n[rpcinterface:supervisor]\\nsupervisor.rpcinterface_factory = supervisor.rpcinterface:make_main_rpcinterface\\n[supervisorctl]\\nserverurl=unix:///var/run/supervisor.sock\\n[include]\\nfiles = /etc/supervisor/conf.d/*.conf' > /etc/supervisor/supervisord.conf"
	        ORCA_SUPERVISOR_CONFIG = "'[program:orca_client]\\ncommand=/orca/bin/host\\nautostart=true\\nautorestart=true\\nstartretries=2\\nuser=orca\\nredirect_stderr=true\\nstdout_logfile=/orca/log/host.log\\nstdout_logfile_maxbytes=50MB\\n' > /etc/supervisor/conf.d/orca.conf"
	)

	return []string{
		"echo orca | sudo -S addgroup --system supervisor",
		"echo orca | sudo -S adduser orca supervisor",
		"echo orca | sudo -S apt-get update",
		"echo orca | sudo -S apt-get install -y git golang supervisor",
		"echo orca | sudo -S sh -c \"echo " + SUPERVISOR_CONFIG + "\"",
		"echo orca | sudo -S sh -c \"echo " + ORCA_SUPERVISOR_CONFIG + "\"",
		"echo orca | sudo -S rm -rf /orca",
		"echo orca | sudo -S mkdir -p /orca",
		"echo orca | sudo -S mkdir -p /orca/apps",
		"echo orca | sudo -S mkdir -p /orca/log",
		"echo orca | sudo -S mkdir -p /orca/data",
		"echo orca | sudo -S mkdir -p /orca/data/host",
		"echo orca | sudo -S mkdir -p /orca/config",
		"echo orca | sudo -S mkdir -p /orca/config/host",
		"echo orca | sudo -S chmod -R 777 /orca",
		//"echo orca | sudo -S sh -c \"echo '" + string(trainerIp) + " orcatrainer' >> /etc/hosts\"",
		"rm -rf /orca/src/gatoor && mkdir -p /orca/src/gatoor && cd /orca/src/gatoor && git clone -b awstest https://github.com/gatoor/orca.git",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca && go get github.com/c9s/goprocinfo/linux && go get github.com/Sirupsen/logrus && go get golang.org/x/crypto/ssh'",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca/base && go build'",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca/base/log && go build'",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca/util && go build'",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca/host && go install'",
		"echo orca | sudo -S sh -c \"echo '{\\\"PollInterval\\\": 10, \\\"TrainerUrl\\\": \\\"http://" + string(trainerIp) + ":5000/push\\\", \\\"HostId\\\":\\\"" + string(hostId) + "\\\"}' > /orca/config/host/host.conf\"",
		"echo orca | sudo -S service supervisor restart",
		//"echo orca | sudo -S sh -c 'nohup /orca/bin/host >> /orca/log'",
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
	instance := ubuntu1604(state_configuration.GlobalConfigurationState.Trainer.Ip, hostId)
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
