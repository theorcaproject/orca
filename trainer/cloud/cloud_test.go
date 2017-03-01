package cloud

import (
	"orca/trainer/configuration"
	"testing"
	"time"
)

func TestPlan_canLaunchSpots(t *testing.T) {
	config := configuration.ConfigurationStore{}
	config.Init("")

	cloud := CloudProvider{}
	if cloud.canLaunchSpotInstance() == false {
		t.Fail()
	}
}

func TestPlan_canLaunchSpots_Fails(t *testing.T) {
	config := configuration.ConfigurationStore{}
	config.Init("")

	cloud := CloudProvider{}
	if cloud.canLaunchSpotInstance() == false {
		t.Fail()
	}

	cloud.lastSpotInstanceFailure = time.Now()
	if cloud.canLaunchSpotInstance() == true {
		t.Fail()
	}
}

