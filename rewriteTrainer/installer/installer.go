package installer

import (
	orcaSSh "gatoor/orca/util"
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/base"
	"gatoor/orca/client/types"
	"strconv"
)

var InstallerLogger = Logger.LoggerWithField(Logger.Logger, "module", "installer")

func TestClientConfig(id base.HostId) types.Configuration {
	return types.Configuration{
		HostId: id,
		Type: "test",
		AppStatusPollInterval: 1,
		MetricsPollInterval: 1,
		TrainerPollInterval: 5,
		TrainerUrl: "http://172.16.147.1:5000/push",
	}
}

func ubuntu1604(clientConfig types.Configuration) []string {
	const (
		SUPERVISOR_CONFIG = "'[unix_http_server]\\nfile=/var/run/supervisor.sock\\nchmod=0770\\nchown=root:supervisor\\n[supervisord]\\nlogfile=/var/log/supervisor/supervisord.log\\npidfile=/var/run/supervisord.pid\\nchildlogdir=/var/log/supervisor\\n[rpcinterface:supervisor]\\nsupervisor.rpcinterface_factory = supervisor.rpcinterface:make_main_rpcinterface\\n[supervisorctl]\\nserverurl=unix:///var/run/supervisor.sock\\n[include]\\nfiles = /etc/supervisor/conf.d/*.conf' > /etc/supervisor/supervisord.conf"
	        ORCA_SUPERVISOR_CONFIG = "'[program:orca_client]\\ncommand=/orca/bin/client\\nautostart=true\\nautorestart=true\\nstartretries=2\\nuser=root\\nredirect_stderr=true\\nstdout_logfile=/orca/log/client.log\\nstdout_logfile_maxbytes=50MB\\n' > /etc/supervisor/conf.d/orca.conf"
	)

	return []string{
		"echo orca | sudo -S addgroup --system supervisor",
		"echo orca | sudo -S apt-get update",
		"echo orca | sudo -S apt-get install -y git golang supervisor",
		"echo orca | sudo -S sh -c \"echo " + SUPERVISOR_CONFIG + "\"",
		"echo orca | sudo -S sh -c \"echo " + ORCA_SUPERVISOR_CONFIG + "\"",
		"echo orca | sudo -S rm -rf /orca",
		"echo orca | sudo -S mkdir -p /orca",
		"echo orca | sudo -S mkdir -p /orca/apps",
		"echo orca | sudo -S mkdir -p /orca/log",
		"echo orca | sudo -S mkdir -p /orca/client",
		"echo orca | sudo -S mkdir -p /orca/client/data",
		"echo orca | sudo -S mkdir -p /orca/client/config",
		"echo orca | sudo -S chmod -R 777 /orca",
		//"echo orca | sudo -S sh -c \"echo '" + string(trainerIp) + " orcatrainer' >> /etc/hosts\"",
		"rm -rf /orca/src/gatoor && mkdir -p /orca/src/gatoor && cd /orca/src/gatoor && git clone -b awstest https://github.com/gatoor/orca.git",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca && go get github.com/Sirupsen/logrus && go get golang.org/x/crypto/ssh && go get github.com/fsouza/go-dockerclient && go get github.com/gorilla/mux'",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca/base && go build'",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca/base/log && go build'",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca/util && go build'",
		"GOPATH=/orca bash -c 'cd /orca/src/gatoor/orca/client && go install'",
		"echo orca | sudo -S sh -c \"echo '{\\\"Type\\\": \\\"" + string(clientConfig.Type) + "\\\", \\\"TrainerPollInterval\\\": " + strconv.Itoa(clientConfig.TrainerPollInterval) + ", \\\"AppStatusPollInterval\\\": " + strconv.Itoa(clientConfig.AppStatusPollInterval) + ", \\\"MetricsPollInterval\\\": " + strconv.Itoa(clientConfig.MetricsPollInterval) + ", \\\"TrainerUrl\\\": \\\"" + clientConfig.TrainerUrl + "\\\", \\\"HostId\\\":\\\"" + string(clientConfig.HostId) + "\\\"}' > /orca/client/config/client.conf\"",
		"echo orca | sudo -S service supervisor restart",
		//"echo orca | sudo -S sh -c 'nohup /orca/bin/host >> /orca/log'",
	}
}

func InstallNewInstance(clientConfig types.Configuration, ipAddr base.IpAddr) bool {
	InstallerLogger.Infof("Starting install on host %s:%s", clientConfig.HostId, ipAddr)
	userName := "orca"
	session, addr := orcaSSh.Connect(userName, string(ipAddr) + ":22")
	if session == nil {
		InstallerLogger.Infof("Install on host %s:%s failed: No session", clientConfig.HostId, ipAddr)
		return false
	}
	instance := ubuntu1604(clientConfig)
	for _, cmd := range instance {
		res := orcaSSh.ExecuteSshCommand(session, addr, cmd)
		if !res {
			InstallerLogger.Infof("Install on host %s:%s failed", clientConfig.HostId, ipAddr)
			return false
		}
	}
	InstallerLogger.Infof("Install on host %s:%s success", clientConfig.HostId, ipAddr)
	return true
}
