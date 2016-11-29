package docker

import (
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/client/types"
	"gatoor/orca/base"
	DockerClient "github.com/fsouza/go-dockerclient"
	//DockerTypes "github.com/docker/docker/api/types"
	//"context"
	//"fmt"
	"strings"
)

var DockerLogger = Logger.LoggerWithField(Logger.Logger, "module", "docker")
var dockerCli *DockerClient.Client
var dockerAuth DockerClient.AuthConfiguration

type Client struct {

}


func (c *Client) Init() {
	var err error
	//dockerAuth := DockerClient.AuthConfiguration{}
	dockerCli, err = DockerClient.NewClientFromEnv()
	if err != nil {
		DockerLogger.Fatalf("Docker client could not be instantiated: %v", err)
	}

}

func (c *Client) Type() types.ClientType {
	return types.DOCKER_CLIENT
}

func (c *Client) InstallApp(appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	DockerLogger.Infof("Installing docker app %s:%s", appConf.Name, appConf.Version)
	return false
	//imageOpt := DockerClient.PullImageOptions{
	//	Repository: appConf.DockerConfig.Repository,
	//	Registry: appConf.DockerConfig.Registry,
	//	RawJSONStream: true,
	//}
//	resp, err := dockerCli.PullImage(imageOpt, auth)
//	if err != nil {
//		DockerLogger.Errorf("Install of app %s:%s failed: %s", appConf.Name, appConf.Version)
//		return
//	}
//	fmt.Println("AA")
//	fmt.Println("AA")
//	resp.Close()
}


func (c *Client) RunApp(appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	return strings.Contains(string(appConf.Name), "fail")
}


func (c *Client) QueryApp(appId base.AppId, appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	return !strings.Contains(string(appConf.Version), "queryfail")
}

func (c *Client) StopApp(appId base.AppId, appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	return !strings.Contains(string(appConf.Version), "stopfail")
}

func (c *Client) DeleteApp(appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	return false
}

func (c *Client) AppMetrics(appId base.AppId, appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration, metrics *types.AppsMetricsById) bool {
	return false
}