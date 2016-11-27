package main
//
//import (
//	"gatoor/orca/base"
//	"github.com/docker/docker/client"
//	"github.com/docker/docker/api/types"
//	"context"
//	"fmt"
//)
//
//type DockerMetrics struct {}
//
//func getContainerId(appId base.AppId) {
//
//}
//
//var dockerClient *client.Client
//
//func init() {
//	cli, err := client.NewEnvClient()
//	if err != nil {
//		panic(err)
//	}
//	dockerClient = cli
//	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
//	if err != nil {
//		panic(err)
//	}
//
//	for _, container := range containers {
//		fmt.Printf("%s %s\n", container.ID[:10], container.Image)
//	}
//}
//
//func (d DockerMetrics) HostMetrics() base.HostStats {
//	metrics := base.HostStats{}
//	return metrics
//}
//
//func (d DockerMetrics) AppMetrics(app base.AppName) base.AppStats {
//	metrics := base.AppStats{}
//	return metrics
//}
