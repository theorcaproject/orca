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
	"fmt"
	"orca/trainer/model"
	"orca/trainer/state"
	orcaSSh "orca/util"
	"time"
)

type CloudProvider struct {
	Engine  CloudEngine
	Changes []*model.ChangeServer

	apiEndpoint     string
	loggingEndpoint string
	sshUser         string

	lastSpotInstanceFailure time.Time
}

func (cloud *CloudProvider) Init(engine CloudEngine, sshUser string, apiEndpoint string, loggingEndpoint string) {
	cloud.Engine = engine
	cloud.apiEndpoint = apiEndpoint
	cloud.sshUser = sshUser
	cloud.loggingEndpoint = loggingEndpoint
}

func (cloud *CloudProvider) ActionChange(change *model.ChangeServer, stateStore *state.StateStore) {
	/* First push this change onto the change queue for the cloud provider */
	cloud.AddChange(change)

	go func() {
		/* Here we can spawn a new server */
		if change.Type == "new_server" {
			var newHost *model.Host
			if !change.RequiresReliableInstance && cloud.canLaunchSpotInstance() {
				change.SpotInstanceRequested = true
				newHost = cloud.Engine.SpawnSpotInstanceSync(change)
			} else {
				change.SpotInstanceRequested = false
				newHost = cloud.Engine.SpawnInstanceSync(change)
			}

			if newHost.Id != "" {
				state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
					Message: fmt.Sprintf("Beginning installation of orcahostd to server %s", newHost.Id),
					HostId:  newHost.Id,
				})

				newHost.GroupingTag = change.GroupingTag /* TODO Persist this guy as a tag*/

				stateStore.HostInit(newHost)

				/* If the change times out we need to nuke it */
				change.NewHostId = string(newHost.Id)
				change.InstanceLaunched = true
				cloud.Engine.SetTag(newHost.Id, "GroupingTag", newHost.GroupingTag)

				/* A new server was created, wahoo */
				/* Next we should install some stuff to it */
				ipAddr := cloud.Engine.GetIp(newHost.Id)
				sshKeyPath := cloud.Engine.GetPem()
				if ipAddr == "" {
					state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__ERROR,
						Message: fmt.Sprintf("Missing IP address for host %s, cannot deploy package to instance", newHost.Id),
					})

					return
				}
				for {
					session, addr := orcaSSh.Connect(cloud.sshUser, string(ipAddr)+":22", sshKeyPath)
					if session == nil {
						state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__ERROR,
							Message: fmt.Sprintf("Could not connect to host %s to deploy orcahostd. Giving up!", newHost.Id),
						})

						return
					}

					instance := []string{
						"sudo -S docker pull michaellawson/orcahostd:latest",
						"sudo -S docker run -e HOSTID='" + string(newHost.Id) + "' -e TRAINER_URL='"+cloud.apiEndpoint+"' -e DOCKER_SOCKET='/var/root/run/docker.sock' -v /var/run:/var/root/run/ --network='host' michaellawson/orcahostd",
					}

					for _, cmd := range instance {
						res := orcaSSh.ExecuteSshCommand(session, addr, cmd)
						if !res {
							state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__ERROR,
								Message: fmt.Sprintf("Could not execute command '%s' on host '%s'. Giving up now!", cmd, newHost.Id),
							})
							return
						}
					}

					change.InstalledPackages = true

					state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__INFO,
						Message: fmt.Sprintf("Finished installation of orcahostd to server %s", newHost.Id),
						HostId:  newHost.Id,
					})
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

		} else if change.Type == "loadbalancer_join" {
			cloud.Engine.RegisterWithLb(change.NewHostId, change.LoadBalancerName)
			cloud.RemoveChange(change.Id, true)

		} else if change.Type == "loadbalancer_leave" {
			cloud.Engine.DeRegisterWithLb(change.NewHostId, change.LoadBalancerName)
			cloud.RemoveChange(change.Id, true)

		} else if change.Type == "app_tag_add" {
			cloud.Engine.AddNameTag(change.NewHostId, change.LoadBalancerAppTarget)
			cloud.RemoveChange(change.Id, true)

		} else if change.Type == "app_tag_remove" {
			cloud.Engine.RemoveNameTag(change.NewHostId, change.LoadBalancerAppTarget)
			cloud.RemoveChange(change.Id, true)

		} else if change.Type == "retire_server" {
			hostToRemove, err := stateStore.GetConfiguration(change.NewHostId)
			if err == nil {
				hostToRemove.State = "terminating"
			}
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
	return len(cloud.Changes) > 0
}

func (cloud *CloudProvider) GetAllChanges() []*model.ChangeServer {
	return cloud.Changes
}

func (cloud *CloudProvider) RemoveChange(changeId string, success bool) {
	newChanges := make([]*model.ChangeServer, 0)
	for _, change := range cloud.Changes {
		if change.Id != changeId {
			newChanges = append(newChanges, change)
		}
	}
	cloud.Changes = newChanges
}

func (cloud *CloudProvider) AddChange(change *model.ChangeServer) {
	cloud.Changes = append(cloud.Changes, change)
}

func (cloud *CloudProvider) GetChange(changeId string) *model.ChangeServer {
	for _, change := range cloud.Changes {
		if change.Id == changeId {
			return change
		}
	}
	return nil
}

func (cloud *CloudProvider) NotifyHostTimedOut(host *model.Host) {
	if host.SpotInstance {
		terminate, reason := cloud.Engine.WasSpotInstanceTerminatedDueToPrice(host.SpotInstanceId)
		if terminate {
			state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__ERROR,
				Message: fmt.Sprintf("Spot instance was terminated by aws, reason provided was %s", reason),
			})
			cloud.lastSpotInstanceFailure = time.Now()
		}
	}
}

func (cloud *CloudProvider) NotifySpawnHostTimedOut(change *model.ChangeServer) {
	if change.SpotInstanceRequested {
		terminate, reason := cloud.Engine.WasSpotInstanceTerminatedDueToPrice(change.SpotInstanceId)

		if terminate {
			state.Audit.Insert__AuditEvent(state.AuditEvent{Severity: state.AUDIT__ERROR,
				Message: fmt.Sprintf("Could not launch new spot instance, aws kulled the request because %s", reason),
			})

			cloud.lastSpotInstanceFailure = time.Now()
		}
	}
}

func (cloud *CloudProvider) canLaunchSpotInstance() bool {
	return (time.Now().Unix() - cloud.lastSpotInstanceFailure.Unix()) > 60*60*2
}

func (cloud *CloudProvider) BackupConfiguration(configuration string) bool {
	return cloud.Engine.BackupConfiguration(configuration)
}

func (cloud *CloudProvider) MonitorQueue(queueName string) int {
	return cloud.Engine.MonitorDataQueue(queueName)
}

func (cloud *CloudProvider) CreateQueue(queueName string, rogueQueueName string) {
	cloud.Engine.CreateDataQueue(queueName, rogueQueueName)
}
