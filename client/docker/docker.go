package docker

import (
	Logger "gatoor/orca/rewriteTrainer/log"
	"gatoor/orca/client/types"
	"gatoor/orca/base"
	DockerClient "github.com/fsouza/go-dockerclient"
	"time"
	"bytes"
	"fmt"
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
	var buf bytes.Buffer
	imageOpt := DockerClient.PullImageOptions{
		Repository: appConf.DockerConfig.Repository,
		Tag: appConf.DockerConfig.Tag,
		OutputStream: &buf,
	}
	err := dockerCli.PullImage(imageOpt, DockerClient.AuthConfiguration{})
	if err != nil {
		DockerLogger.Errorf("Install of app %s:%s failed: %s", appConf.Name, appConf.Version, err)
		return false
	}
	DockerLogger.Errorf("Install of app %s:%s successful", appConf.Name, appConf.Version)
	return true
}


func (c *Client) RunApp(appId base.AppId, appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	DockerLogger.Infof("Running docker app %s - %s:%s", appId, appConf.Name, appConf.Version)

	config := DockerClient.Config{AttachStdout: true, AttachStdin: true, Image: fmt.Sprintf("%s:%s", appConf.Name, appConf.Version)}
	opts := DockerClient.CreateContainerOptions{Name: string(appId), Config: &config}
	container, containerErr := dockerCli.CreateContainer(opts)
	if containerErr != nil {
		DockerLogger.Errorf("Running docker app %s - %s:%s failed: %s", appId, appConf.Name, appConf.Version, containerErr)
		return false
	}

	err := dockerCli.StartContainer(container.ID, &DockerClient.HostConfig{})

	if err != nil {
		DockerLogger.Warnf("Running docker app %s - %s:%s failed: %s", appId, appConf.Name, appConf.Version)
		return false
	}
	DockerLogger.Infof("Running docker app %s - %s:%s successful", appId, appConf.Name, appConf.Version)
	return true
}


func (c *Client) QueryApp(appId base.AppId, appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	DockerLogger.Infof("Query docker app %s - %s:%s", appId, appConf.Name, appConf.Version)
	resp, err := dockerCli.InspectContainer(string(appId))
	if err != nil {
		DockerLogger.Infof("Query docker app %s - %s:%s failed: %s", appId, appConf.Name, appConf.Version)
		return false
	}
	DockerLogger.Infof("Query docker app %s - %s:%s successful %+v", appId, appConf.Name, appConf.Version, resp)
	return resp.State.Running
}

func (c *Client) StopApp(appId base.AppId, appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	DockerLogger.Infof("Stopping docker app %s - %s:%s", appId, appConf.Name, appConf.Version)
	err := dockerCli.StopContainer(fmt.Sprintf("%s", appId), 0)
	fail := false
	if err != nil {
		DockerLogger.Infof("Stopping docker app %s - %s:%s failed: %s", appId, appConf.Name, appConf.Version, err)
		fail = true
	}
	opts := DockerClient.RemoveContainerOptions{ID: string(appId)}
	err = dockerCli.RemoveContainer(opts)
	if err != nil {
		DockerLogger.Infof("Stopping docker app %s - %s:%s failed: %s", appId, appConf.Name, appConf.Version, err)
		fail = true
	}
	if fail {
		return false
	}
	DockerLogger.Infof("Stopping docker app %s - %s:%s successful", appId, appConf.Name, appConf.Version)
	return true
}

func (c *Client) DeleteApp(appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration) bool {
	DockerLogger.Infof("Deleting docker app %s:%s", appConf.Name, appConf.Version)
	err := dockerCli.RemoveImage(fmt.Sprintf("%s:%s", appConf.Name, appConf.Version))
	if err != nil {
		DockerLogger.Infof("Deleting docker app %s:%s failed: %s", appConf.Name, appConf.Version, err)
		return false
	}
	DockerLogger.Infof("Deleting docker app %s:%s successful", appConf.Name, appConf.Version)
	return true
}

func (c *Client) AppMetrics(appId base.AppId, appConf base.AppConfiguration, appsState *types.AppsState, conf *types.Configuration, metrics *types.AppsMetricsById) bool {
	DockerLogger.Infof("Getting AppMetrics for app %s %s:%s", appId, appConf.Name, appConf.Version)
	errC := make(chan error, 1)
	statsC := make(chan *DockerClient.Stats)
	done := make(chan bool)

	go func() {
		errC <- dockerCli.Stats(DockerClient.StatsOptions{ID: string(appId), Stats: statsC, Stream: true, Done: done})
		close(errC)
	}()
	var resultStats []*DockerClient.Stats
	count := 0
	for {
		count++
		DockerLogger.Info("ITER")
		stats, ok := <-statsC
		if !ok || count > 2 {
			close(done)
			break
		}
		resultStats = append(resultStats, stats)
	}
	err := <-errC
	if err != nil && len(resultStats) != 2 {
		DockerLogger.Infof("Getting AppMetrics for app %s %s:%s failed: %s. Only %d results", appId, appConf.Name, appConf.Version, err, len(resultStats))
		return false
	}
	DockerLogger.Infof("%+v", resultStats)
	appMet := parseDockerStats(resultStats[0], resultStats[1])
	//TODO test performance of app and add this to appMet.ResponsePerformance
	metrics.Add(appId, time.Now().UTC().Format(time.RFC3339Nano), appMet)
	DockerLogger.Infof("Getting AppMetrics for app %s %s:%s successful", appId, appConf.Name, appConf.Version)
	return true
}

func parseDockerStats(stat0 *DockerClient.Stats, stat1 *DockerClient.Stats) base.AppStats {
	var (
		cpuPercent = uint64(0)
		cpuDelta = float64(stat1.CPUStats.CPUUsage.TotalUsage) - float64(stat0.CPUStats.CPUUsage.TotalUsage)
		systemDelta = float64(stat1.CPUStats.SystemCPUUsage) - float64(stat0.CPUStats.SystemCPUUsage)
	)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = uint64((cpuDelta / systemDelta) * float64(len(stat1.CPUStats.CPUUsage.PercpuUsage)) * 100.0)
	}

	appStats := base.AppStats{}
	appStats.CpuUsage = base.Usage(cpuPercent)

	averageMem := (stat1.MemoryStats.Usage + stat0.MemoryStats.Usage) / 2
	appStats.MemoryUsage = base.Usage(averageMem)
	averageNet := (stat1.Network.RxBytes + stat0.Network.RxBytes) / 2
	if (stat1.Network.TxBytes + stat0.Network.TxBytes) / 2  > averageNet {
		averageNet = (stat1.Network.TxBytes + stat0.Network.TxBytes) / 2
	}
	appStats.NetworkUsage = base.Usage(averageNet)

	return appStats
}
