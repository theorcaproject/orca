package cloud


import (
	"testing"
	"gatoor/orca/base"
	"time"
	"gatoor/orca/rewriteTrainer/state/cloud"
)

//These Tests require an AWS Account. Configure it via ORCA_AWS_KEY and ORCA_AWS_SECRET environment variables


func before () AWSProvider {
	aws := AWSProvider{}
	CurrentProviderConfig = ProviderConfiguration{
		Type: PROVIDER_AWS, MinInstances: 1, MaxInstances: 3,
		AWSConfiguration: AWSConfiguration{
			Region: "us-west-2",
			AMI: "unknown",
			SecurityGroupId: "sg-cdf3cfb4",
			InstanceTypes: []InstanceType{},
			InstanceCost: make(map[InstanceType]Cost),
			InstanceResources: make(map[InstanceType]state_cloud.InstanceResources),
			InstanceSafety: make(map[InstanceType]SafeInstance),
			SuitableInstanceSafetyFactor: 2.0,
		},
	}
	aws.Init()
	return aws
}

func TestAWSProvider_CheckCredentials(t *testing.T) {
	aws := before()

	if !aws.CheckCredentials() {
		t.Error("Invalid Credentials")
	}
}


func TestAWSProvider_SpawnInstance_TerminateInstance(t *testing.T) {
        aws := before()

	if aws.SpawnInstance("xyz") != "" {
		t.Error()
	}

	if aws.SpawnInstance("t2.nano") != "" {
		t.Error()
	}

	CurrentProviderConfig.AWSConfiguration.AMI = "ami-3df75e5d"

	if aws.SpawnInstance("xyz") != "" {
		t.Error()
	}

	id := aws.SpawnInstance("t2.nano")
	if id == "" {
		t.Error()
	}
	log := aws.GetSpawnLog()
	for _, h := range log {
		if h != id {
			t.Error(log)
		}
	}
	res := aws.CheckInstance(id)
	if res != INSTANCE_STATUS_DEAD {
		t.Error(res)
	}

	aws.waitOnInstanceReady(id)

	res2 := aws.CheckInstance(id)
	if res2 != INSTANCE_STATUS_HEALTHY {
		t.Error(res2)
	}
	log2 := aws.GetSpawnLog()
	for _, h := range log2 {
		if h != id {
			t.Error(log2)
		}
	}

	if !aws.TerminateInstance(base.HostId(id)) {
		t.Error()
	}
	log3 := aws.GetSpawnLog()
	if len(log3) != 0 {
		t.Error(log3)
	}

	time.Sleep(time.Duration(5 * time.Second))
	res3 := aws.CheckInstance(id)
	if res3 != INSTANCE_STATUS_DEAD {
		t.Error(res3)
	}

	if aws.TerminateInstance("unkown") {
		t.Error()
	}

	id2 := aws.SpawnInstanceSync("t2.nano")

	if id2 == "" {
		t.Error()
	}
	log4 := aws.GetSpawnLog()
	for _, h := range log4 {
		if h != id2 {
			t.Error(log4)
		}
	}

	res4 := aws.CheckInstance(id2)
	if res4 != INSTANCE_STATUS_HEALTHY {
		t.Error(res4)
	}

	ip := aws.GetIp(id2)

	if ip == "" {
		t.Error()
	}

	ty := aws.GetInstanceType(id2)

	if ty != InstanceType("t2.nano") {
		t.Error(ty)
	}

	if !aws.TerminateInstance(id2) {
		t.Error()
	}

	log5 := aws.GetSpawnLog()
	if len(log5) != 0 {
		t.Error(log5)
	}
}

func TestAWSProvider_SuitableInstanceTypes(t *testing.T) {
	aws := before()

	CurrentProviderConfig.AWSConfiguration.InstanceTypes = []InstanceType{"i1", "i10", "i20", "i100", "i50"}
	CurrentProviderConfig.AWSConfiguration.SuitableInstanceSafetyFactor = 2.0
	CurrentProviderConfig.AWSConfiguration.InstanceResources["i1"] = state_cloud.InstanceResources{TotalCpuResource: 1, TotalMemoryResource: 1, TotalNetworkResource: 1}
	CurrentProviderConfig.AWSConfiguration.InstanceResources["i10"] = state_cloud.InstanceResources{TotalCpuResource: 10, TotalMemoryResource: 10, TotalNetworkResource: 10}
	CurrentProviderConfig.AWSConfiguration.InstanceResources["i20"] = state_cloud.InstanceResources{TotalCpuResource: 20, TotalMemoryResource: 20, TotalNetworkResource: 20}
	CurrentProviderConfig.AWSConfiguration.InstanceResources["i50"] = state_cloud.InstanceResources{TotalCpuResource: 50, TotalMemoryResource: 50, TotalNetworkResource: 50}
	CurrentProviderConfig.AWSConfiguration.InstanceResources["i100"] = state_cloud.InstanceResources{TotalCpuResource: 100, TotalMemoryResource: 100, TotalNetworkResource: 100}

	CurrentProviderConfig.AWSConfiguration.InstanceCost["i1"] = 1
	CurrentProviderConfig.AWSConfiguration.InstanceCost["i10"] = 10
	CurrentProviderConfig.AWSConfiguration.InstanceCost["i20"] = 20
	CurrentProviderConfig.AWSConfiguration.InstanceCost["i50"] = 5
	CurrentProviderConfig.AWSConfiguration.InstanceCost["i100"] = 100

	instances := aws.SuitableInstanceTypes(state_cloud.InstanceResources{TotalCpuResource: 10, TotalMemoryResource: 10, TotalNetworkResource: 10})

	if len(instances) != 3 {
		t.Error(instances)
	}
	if instances[0] != "i50" || instances[1] != "i20" || instances[2] != "i100" {
		t.Error(instances)
	}

	instances = aws.SuitableInstanceTypes(state_cloud.InstanceResources{TotalCpuResource: 60, TotalMemoryResource: 10, TotalNetworkResource: 10})
	if len(instances) != 0 {
		t.Error(instances)
	}

	instances = aws.SuitableInstanceTypes(state_cloud.InstanceResources{TotalCpuResource: 10, TotalMemoryResource: 60, TotalNetworkResource: 10})
	if len(instances) != 0 {
		t.Error(instances)
	}

	instances = aws.SuitableInstanceTypes(state_cloud.InstanceResources{TotalCpuResource: 10, TotalMemoryResource: 10, TotalNetworkResource: 60})
	if len(instances) != 0 {
		t.Error(instances)
	}

	instances = aws.SuitableInstanceTypes(state_cloud.InstanceResources{TotalCpuResource: 20, TotalMemoryResource: 10, TotalNetworkResource: 10})
	if len(instances) != 2 {
		t.Error(instances)
	}
	instances = aws.SuitableInstanceTypes(state_cloud.InstanceResources{TotalCpuResource: 0, TotalMemoryResource: 0, TotalNetworkResource: 0})
	if len(instances) != 5 {
		t.Error(instances)
	}
}


