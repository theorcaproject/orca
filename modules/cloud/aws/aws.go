package aws

type AwsProvider struct {

}
var done = false

func (aws AwsProvider) NewInstance() string{
	if done {
		return "172.16.147.189"
	} else {
		done = true
		return "172.16.147.190"
	}
}