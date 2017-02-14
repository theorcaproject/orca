/*
Copyright Alex Mack (al9mack@gmail.com) and Michael Lawson (michael@sphinix.com)
This file is part of Orca.

Orca is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Orca is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with Orca.  If not, see <http://www.gnu.org/licenses/>.
*/

package cloud

import (
	"orca/trainer/model"
	orcaSSh "orca/util"
	"orca/trainer/state"
	"fmt"
)

type CloudProvider struct {
	Engine      CloudEngine
	Changes     []*model.ChangeServer

	apiEndpoint string
	loggingEndpoint string
	sshUser string
}

func (cloud* CloudProvider) Init(engine CloudEngine, sshUser string, apiEndpoint string, loggingEndpoint string) {
	cloud.Engine = engine
	cloud.apiEndpoint = apiEndpoint
	cloud.sshUser = sshUser
	cloud.loggingEndpoint= loggingEndpoint
}

func (cloud*CloudProvider) ActionChange(change *model.ChangeServer, stateStore *state.StateStore) {
	/* First push this change onto the change queue for the cloud provider */
	cloud.AddChange(change)

	go func() {
		/* Here we can spawn a new server */
		if change.Type == "new_server" {
			var newHost *model.Host
			if change.RequiresReliableInstance {
				newHost = cloud.Engine.SpawnInstanceSync("", change.Network, change.SecurityGroups)
			} else {
				newHost = cloud.Engine.SpawnSpotInstanceSync("", change.Network, change.SecurityGroups)
			}

			if newHost.Id != "" {
				state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
					Message: fmt.Sprintf("Beginning installation of orcahostd to server %s", newHost.Id),
					Details:map[string]string{
						"host": newHost.Id,
					}})

				stateStore.HostInit(newHost)
				/* If the change times out we need to nuke it */
				change.NewHostId = string(newHost.Id)
				change.InstanceLaunched = true

				/* A new server was created, wahoo */
				/* Next we should install some stuff to it */
				ipAddr := cloud.Engine.GetIp(newHost.Id)
				sshKeyPath := cloud.Engine.GetPem()
				if ipAddr == "" {
					state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__ERROR,
						Message: fmt.Sprintf("Missing IP address for host %s, cannot deploy package to instance", newHost.Id),
						Details:map[string]string{
						}})

					return
				}
				for {
					session, addr := orcaSSh.Connect(cloud.sshUser, string(ipAddr) + ":22", sshKeyPath)
					if session == nil {
						state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__ERROR,
							Message: fmt.Sprintf("Could not connect to host %s to deploy orcahostd. Giving up!", newHost.Id),
							Details:map[string]string{
							}})

						return
					}

					SUPERVISOR_CONFIG := "'[unix_http_server]\\nfile=/var/run/supervisor.sock\\nchmod=0770\\nchown=root:supervisor\\n[supervisord]\\nlogfile=/var/log/supervisor/supervisord.log\\npidfile=/var/run/supervisord.pid\\nchildlogdir=/var/log/supervisor\\n[rpcinterface:supervisor]\\nsupervisor.rpcinterface_factory = supervisor.rpcinterface:make_main_rpcinterface\\n[supervisorctl]\\nserverurl=unix:///var/run/supervisor.sock\\n[include]\\nfiles = /etc/supervisor/conf.d/*.conf' > /etc/supervisor/supervisord.conf"
					ORCA_SUPERVISOR_CONFIG := "'[program:orca_client]\\ncommand=/orca/bin/orcahostd --interval 30 --hostid " + string(newHost.Id) + " --traineruri " + cloud.apiEndpoint + "\\nautostart=true\\nautorestart=true\\nstartretries=2\\nuser=root\\nredirect_stderr=true\\nstdout_logfile=/orca/log/client.log\\nstdout_logfile_maxbytes=50MB\\n' > /etc/supervisor/conf.d/orca.conf"
					RSYSLOG_CONFIG := "'module(load=\"imfile\" PollingInterval=\"10\")\\ninput(type=\"imfile\"\\nFile=\"/var/log/syslog\")\\nmodule(load=\"imuxsock\")\\ntemplate(name=\"ForwardFormat\" type=\"list\") {nconstant(value=\"<\")\\nproperty(name=\"pri\")\\nconstant(value=\">\")\\nproperty(name=\"timestamp\" dateFormat=\"rfc3339\")\\nproperty(name=\"syslogtag\" position.from=\"1\" position.to=\"32\")\\nconstant(value=\" "+ newHost.Id +": \")\\nproperty(name=\"msg\" spifno1stsp=\"on\")\\nproperty(name=\"msg\")\\n}\\n$ModLoad imuxsock\\n*.* @@" + cloud.loggingEndpoint + ";ForwardFormat' > /etc/rsyslog.d/orca.conf"

					instance := []string{
						"echo orca | sudo -S addgroup --system supervisor",
						"echo orca | sudo -S apt-get update",
						"echo orca | sudo -S apt-get install -y git golang supervisor docker.io rsyslog",
						"echo orca | sudo -S sh -c \"echo " + SUPERVISOR_CONFIG + "\"",
						"echo orca | sudo -S sh -c \"echo " + ORCA_SUPERVISOR_CONFIG + "\"",
						"echo orca | sudo -S sh -c \"echo " + RSYSLOG_CONFIG + "\"",
						"echo orca | sudo -S rm -rf /orca",
						"echo orca | sudo -S mkdir -p /orca",
						"echo orca | sudo -S mkdir -p /orca/apps",
						"echo orca | sudo -S mkdir -p /orca/log",
						"echo orca | sudo -S mkdir -p /orca/client",
						"echo orca | sudo -S mkdir -p /orca/client/data",
						"echo orca | sudo -S mkdir -p /orca/client/config",
						"echo orca | sudo -S chmod -R 777 /orca",

						"rm -rf /orca/src && mkdir -p /orca/src && cd /orca/src && git clone https://github.com/theorcaproject/orcahostd.git",
						"GOPATH=/orca bash -c 'cd /orca/src/orcahostd && go get github.com/Sirupsen/logrus && go get golang.org/x/crypto/ssh && go get github.com/gorilla/mux'",
						"GOPATH=/orca bash -c 'cd /orca/src/orcahostd && go get orcahostd && go build && go install'",
						"echo orca | sudo -S service supervisor restart",
					}

					for _, cmd := range instance {
						res := orcaSSh.ExecuteSshCommand(session, addr, cmd)
						if !res {
							state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__ERROR,
								Message: fmt.Sprintf("Could not execute command '%s' on host '%s'. Giving up now!", cmd, newHost.Id),
								Details:map[string]string{
								}})
						}
					}

					change.InstalledPackages = true

					state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
						Message: fmt.Sprintf("Finished installation of orcahostd to server %s", newHost.Id),
						Details:map[string]string{
							"host": newHost.Id,
						}})
					break
				}
			}
		} else if change.Type == "remove" {
			hostToRemove, err := stateStore.GetConfiguration(change.NewHostId)
			if err == nil {
				hostToRemove.State = "terminating"
				cloud.Engine.TerminateInstance(HostId(change.NewHostId))
			}
			cloud.RemoveChange(change.Id, true)
			stateStore.RemoveHost(change.NewHostId)

		}else if change.Type == "loadbalancer_join" {
			cloud.Engine.RegisterWithLb(change.NewHostId, change.LoadBalancerName)
			cloud.RemoveChange(change.Id, true)

		}else if change.Type == "loadbalancer_leave" {
			cloud.Engine.DeRegisterWithLb(change.NewHostId, change.LoadBalancerName)
			cloud.RemoveChange(change.Id, true)
		}
	}()
}

func (cloud *CloudProvider) NotifyHostCheckIn(host *model.Host) {
	/* Search for changes related to this instance */
	for _, change := range cloud.Changes {
		if change.Type == "new_server" {
			if change.NewHostId == host.Id {
				host.SpotInstance = !change.RequiresReliableInstance
				cloud.RemoveChange(change.Id, true)
			}
		}
	}
}

func (cloud *CloudProvider) HasChanges() bool {
	return len(cloud.Changes) > 0;
}

func (cloud *CloudProvider) GetAllChanges() []*model.ChangeServer {
	return cloud.Changes
}

func (cloud*CloudProvider) RemoveChange(changeId string, success bool) {
	newChanges := make([]*model.ChangeServer, 0)
	for _, change := range cloud.Changes {
		if change.Id != changeId {
			newChanges = append(newChanges, change)
		}
	}
	cloud.Changes = newChanges
}

func (cloud*CloudProvider) AddChange(change *model.ChangeServer) {
	cloud.Changes = append(cloud.Changes, change)
}

func (cloud*CloudProvider) GetChange(changeId string) *model.ChangeServer {
	for _, change := range cloud.Changes {
		if change.Id == changeId {
			return change
		}
	}
	return nil
}