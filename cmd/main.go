package main

import (
	"k8s.io/component-base/cli"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"
	"my-custom-scheduler/pkg/plugins/podHostNetwork"
	"os"
)

func main() {
	command := app.NewSchedulerCommand(
		app.WithPlugin(podHostNetwork.Name, podHostNetwork.New),
	)

	code := cli.Run(command)
	os.Exit(code)
}
