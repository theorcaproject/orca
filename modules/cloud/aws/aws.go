package aws

import "gatoor/orca/base"

type AwsProvider struct {

}
var done = false

func (aws AwsProvider) NewInstance() (base.HostId, base.IpAddr) {
	if done {
		return base.HostId("HOST_1"), base.IpAddr("172.16.147.189")
	} else {
		done = true
		return base.HostId("HOST_2"), base.IpAddr("172.16.147.190")
	}
}