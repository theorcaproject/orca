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
	"context"
	"fmt"
	"github.com/google/uuid"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	"log"
	"orca/trainer/model"
	"strings"
	"time"
)

type GcpCloudEngine struct {
	ProjectId       string
	CredentialsFile string
	Zone            string

	PemFile   string
	PublicKey string
	User      string
	ImageUrl  string
}

func (engine *GcpCloudEngine) Init(projectId string, zone string, credentialsFile string, user string, publicKey string,
	imageUrl string, pemFile string) {

	engine.ProjectId = projectId
	engine.Zone = zone
	engine.CredentialsFile = credentialsFile
	engine.User = user
	engine.PublicKey = publicKey
	engine.ImageUrl = imageUrl
	engine.PemFile= pemFile

	}

func (a *GcpCloudEngine) GetComputeClient() *compute.Service {
	ctx := context.Background()
	service, err := compute.NewService(ctx, option.WithCredentialsFile(a.CredentialsFile))

	if err != nil {
		fmt.Println("Unable to create Compute service: %v", err)
		return nil
	}

	return service
}

func (a *GcpCloudEngine) GetIp(hostId string) string {
	service := a.GetComputeClient()
	inst, _ := service.Instances.Get(a.ProjectId, a.Zone, hostId).Do()

	// Find a nat ip, if the box is inside an internal VPC then its not going to have one.
	internalIp := inst.NetworkInterfaces[0].NetworkIP
	if len(inst.NetworkInterfaces[0].AccessConfigs) > 0 {
		internalIp = inst.NetworkInterfaces[0].AccessConfigs[0].NatIP
	}

	return internalIp
}

func (a *GcpCloudEngine) GetSubnet(hostId string) string {
	service := a.GetComputeClient()
	inst, _ := service.Instances.Get(a.ProjectId, a.Zone, hostId).Do()
	return inst.NetworkInterfaces[0].Subnetwork
}

func (a *GcpCloudEngine) GetHostInfo(hostId HostId) (string, string, []model.SecurityGroup, bool, string, string) {
	ip := a.GetIp(string(hostId))
	subnetId := a.GetSubnet(string(hostId))
	instanceType := a.GetInstanceType(hostId)

	/* GCP does not have the notion of security groups like AWS. So we return an empty sg array */
	return ip, subnetId, []model.SecurityGroup{}, false, "", string(instanceType)
}

func (a *GcpCloudEngine) waitOnInstanceReady(hostId HostId) bool {
	fmt.Println("Waiting on instance to be ready");
	service := a.GetComputeClient()
	for true {
		inst, _ := service.Instances.Get(a.ProjectId, a.Zone, string(hostId)).Do()
		fmt.Println("Status is: %s", inst.Status);
		if inst.Status == "RUNNING" {
			break
		}

		time.Sleep(2 * time.Millisecond)
	}

	return true
}

func (engine *GcpCloudEngine) GetInstanceType(hostId HostId) InstanceType {
	service := engine.GetComputeClient()
	inst, err := service.Instances.Get(engine.ProjectId, engine.Zone, string(hostId)).Do()
	if err != nil {
		return ""
	}

	return InstanceType(inst.MachineType)
}

func (engine *GcpCloudEngine) SpawnInstanceSync(change *model.ChangeServer) *model.Host {
	service := engine.GetComputeClient()
	prefix := "https://www.googleapis.com/compute/v1/projects/" + engine.ProjectId
	instanceNameUuid, _ := uuid.NewUUID()
	s := strings.Split(instanceNameUuid.String(), "-")
	instanceName := s[0]
	instanceName = "orca-" + instanceName
	googlifiedPublicKey := engine.GetUsername() + ":" + engine.GetPublicKey()

	// Show the current images that are available.
	instance := &compute.Instance{
		Name:        instanceName,
		Description: "orca instance",
		MachineType: prefix + "/zones/" + engine.Zone + "/machineTypes/n1-standard-1",
		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				Type:       "PERSISTENT",
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskName:    instanceName,
					SourceImage: engine.ImageUrl,
				},
			},
		},
		NetworkInterfaces: []*compute.NetworkInterface{
			{
				AccessConfigs: []*compute.AccessConfig{
					{
						Type: "ONE_TO_ONE_NAT",
						Name: "External NAT",
					},
				},
				Network: prefix + "/global/networks/default",
			},
		},
		ServiceAccounts: []*compute.ServiceAccount{
			{
				Email: "default",
				Scopes: []string{
					compute.DevstorageFullControlScope,
					compute.ComputeScope,
				},
			},
		},
		Metadata: &compute.Metadata{
			Items: []*compute.MetadataItems{
				{
					Key:   "sshKeys",
					Value: &googlifiedPublicKey,
				},
			},
		},
	}

	op, err := service.Instances.Insert(engine.ProjectId, engine.Zone, instance).Do()
	log.Printf("Got compute.Operation, err: %#v, %v", op, err)

	inst, err := service.Instances.Get(engine.ProjectId, engine.Zone, instanceName).Do()
	log.Printf("Got compute.Instance, err: %#v, %v", inst, err)

	host := &model.Host{
		Id:             instanceName,
		SecurityGroups: change.SecurityGroups,
		Network:        change.Network,
	}

	hostId := HostId(instanceName)
	if !engine.waitOnInstanceReady(hostId) {
		return &model.Host{}
	}

	return host;
}

func (engine *GcpCloudEngine) TerminateInstance(hostId HostId) bool {
	service := engine.GetComputeClient()
	_, err := service.Instances.Delete(engine.ProjectId, engine.Zone, string(hostId)).Do()
	if err != nil {
		log.Printf("Terminate instance had an issue, %s", err)
		return false
	}

	return true
}

func (aws *GcpCloudEngine) GetPem() string {
	return aws.PemFile
}

func (aws *GcpCloudEngine) GetUsername() string {
	return aws.User
}

func (aws *GcpCloudEngine) GetPublicKey() string {
	return aws.PublicKey
}

func (engine *GcpCloudEngine) SpawnSpotInstanceSync(change *model.ChangeServer) *model.Host {
	/* At this moment, do not support spot instances as spots. Just spawn normals */
	return engine.SpawnInstanceSync(change)
}

func (engine *GcpCloudEngine) GetSpotInstanceHostId(spotId string) (HostId, error) {
	return "", nil
}

func (engine *GcpCloudEngine) SanityCheckHosts(hosts map[string]*model.Host) {
	for _, host := range hosts {
		engine.doSanityCheck(host)
	}
}

func (engine *GcpCloudEngine) doSanityCheck(host *model.Host) {
}

func (engine *GcpCloudEngine) WasSpotInstanceTerminatedDueToPrice(spotRequestId string) (bool, string) {
	return false, ""
}

func (engine *GcpCloudEngine) GetTag(tagKey string, newHostId string) string {
	service := engine.GetComputeClient()
	inst, _ := service.Instances.Get(engine.ProjectId, engine.Zone, newHostId).Do()

	value, ok := inst.Labels[tagKey]
	if ok {
		return value
	}

	return ""
}

func (engine *GcpCloudEngine) SetTag(newHostId string, tagKey string, tagValue string) {
	service := engine.GetComputeClient()
	inst, _ := service.Instances.Get(engine.ProjectId, engine.Zone, newHostId).Do()
	existingLabels := inst.Labels
	if existingLabels == nil {
		existingLabels = make(map[string]string)
	}

	existingLabels[tagKey] = tagValue
	labels := compute.InstancesSetLabelsRequest{
		Labels:           existingLabels,
		LabelFingerprint: inst.LabelFingerprint,
	}
	_, err := service.Instances.SetLabels(engine.ProjectId, engine.Zone, newHostId, &labels).Do()
	if err != nil {
		log.Printf("Set Labels failed, %s", err)
	}
}

func (engine *GcpCloudEngine) AddNameTag(newHostId string, appName string) {
	currentTag := engine.GetTag("Name", newHostId)
	splices := strings.Split(currentTag, "_")
	splices = append(splices, appName)

	engine.SetTag(newHostId, "Name", strings.Join(splices, "_"))
}

func (engine *GcpCloudEngine) RemoveNameTag(newHostId string, appName string) {
	newTags := make([]string, 0)
	currentTag := engine.GetTag("Name", newHostId)
	splices := strings.Split(currentTag, "_")

	for _, tag := range splices {
		if tag != appName {
			newTags = append(newTags, tag)
		}
	}

	engine.SetTag(newHostId, "Name", strings.Join(newTags, "_"))
}

func (engine *GcpCloudEngine) BackupConfiguration(configuration string) bool {
	return true
}

func (engine *GcpCloudEngine) CreateDataQueue(name string, rogueName string) {
}

func (engine *GcpCloudEngine) MonitorDataQueue(name string) int {
	return 0
}

func (engine *GcpCloudEngine) RegisterWithLb(hostId string, lbId string) {
	return
}

func (engine *GcpCloudEngine) DeRegisterWithLb(hostId string, lbId string) {
	return
}
