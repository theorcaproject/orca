package example

import (
	"gatoor/orca/rewriteTrainer/config"
	"gatoor/orca/base"
	"gatoor/orca/rewriteTrainer/needs"
)

func AwsJsonConfig() config.JsonConfiguration {
	conf := config.JsonConfiguration{}

	conf.Trainer.Port = 5000

	//conf.Habitats = []config.HabitatJsonConfiguration{
	//	{
	//		Name: "dockerHabitat",
	//		Version: "0.1",
	//		InstallCommands: []base.OsCommand{
	//			{base.EXEC_COMMAND, base.Command{"apt-get", "update"},},
	//			{base.EXEC_COMMAND, base.Command{"apt-get", "-y install apt-transport-https ca-certificates"},},
	//			{base.EXEC_COMMAND, base.Command{"apt-key", "adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys 58118E89F3A912897C070ADBF76221572C52609D"},},
	//			{base.FILE_COMMAND, base.Command{"/etc/apt/sources.list.d/docker.list", "deb https://apt.dockerproject.org/repo ubuntu-xenial main"},},
	//			{base.EXEC_COMMAND, base.Command{"apt-get", "update"},},
	//			{base.EXEC_COMMAND, base.Command{"apt-get", "-y install docker-engine"},},
	//			{base.EXEC_COMMAND, base.Command{"mkdir", "/orca/apps"},},
	//			{base.EXEC_COMMAND, base.Command{"chmod", "777 /orca/apps"},},
	//		},
	//	},
	//}

	conf.Apps = []config.AppJsonConfiguration{
		{
			Name: "nginx",
			Version: "0.1",
			Type: base.APP_HTTP,
			MinDeploymentCount: 2,
			MaxDeploymentCount: 10,
			InstallCommands: []base.OsCommand{
				{base.EXEC_COMMAND, base.Command{"mkdir", "/orca/apps/nginx"},},
				{base.FILE_COMMAND, base.Command{"/orca/apps/nginx/index.html", "HELLO ORCA!"},},
				{base.EXEC_COMMAND, base.Command{"docker", "run --name orca-nginx -p 80:80 -v /orca/apps/nginx:/usr/share/nginx/html nginx"},},
			},
			QueryStateCommand: base.OsCommand{
				Type: base.EXEC_COMMAND,
				Command: base.Command{"wget", "http://localhost"},
			},
			RemoveCommand: base.OsCommand{
				Type: base.EXEC_COMMAND,
				Command: base.Command{"docker", "stop orca-nginx"},
			},
			Needs: needs.AppNeeds{
				MemoryNeeds: needs.MemoryNeeds(5),
				CpuNeeds: needs.CpuNeeds(5),
				NetworkNeeds: needs.NetworkNeeds(5),
			},
		},
		{
			Name: "nginxworker",
			Version: "0.2",
			Type: base.APP_WORKER,
			MinDeploymentCount: 5,
			MaxDeploymentCount: 10,
			InstallCommands: []base.OsCommand{
				{base.EXEC_COMMAND, base.Command{"mkdir", "/orca/apps/nginxworker"},},
				{base.FILE_COMMAND, base.Command{"/orca/apps/nginxworker/index.html", "HELLO ORCA as worker!"},},
				{base.EXEC_COMMAND, base.Command{"docker", "run --name orca-nginxworker -v /orca/apps/nginxworker:/usr/share/nginx/html nginx"},},
			},
			QueryStateCommand: base.OsCommand{
				Type: base.EXEC_COMMAND,
				Command: base.Command{"wget", "http://localhost"},
			},
			RemoveCommand: base.OsCommand{
				Type: base.EXEC_COMMAND,
				Command: base.Command{"docker", "stop orca-nginxworker"},
			},
			Needs: needs.AppNeeds{
				MemoryNeeds: needs.MemoryNeeds(3),
				CpuNeeds: needs.CpuNeeds(3),
				NetworkNeeds: needs.NetworkNeeds(3),
			},
		},
	}
	return conf
}