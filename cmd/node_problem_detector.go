/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"os"

	"github.com/golang/glog"
	"github.com/spf13/pflag"

	"k8s.io/node-problem-detector/cmd/options"
	"k8s.io/node-problem-detector/pkg/custompluginmonitor"
	"k8s.io/node-problem-detector/pkg/k8sexporter"
	"k8s.io/node-problem-detector/pkg/problemdaemon"
	"k8s.io/node-problem-detector/pkg/problemdetector"
	"k8s.io/node-problem-detector/pkg/systemlogmonitor"
	"k8s.io/node-problem-detector/pkg/types"
	"k8s.io/node-problem-detector/pkg/version"
)

func main() {
	npdo := options.NewNodeProblemDetectorOptions()
	npdo.AddFlags(pflag.CommandLine)

	pflag.Parse()

	if npdo.PrintVersion {
		version.PrintVersion()
		os.Exit(0)
	}

	npdo.SetNodeNameOrDie()

	npdo.ValidOrDie()

	// Initialize problem daemons.
	problemDaemons := problemdaemon.NewProblemDaemons(npdo.MonitorsConfigs)

	// Legacy per-monitor configuration.
	// This section is kept so that existing command line option still works on NPD.
	// TODO(xueweiz): when NPD's refactoring is finished, remove legacy configuration.
	for _, config := range npdo.SystemLogMonitorConfigPaths {
		problemDaemons = append(problemDaemons, systemlogmonitor.NewLogMonitorOrDie(config))
	}
	for _, config := range npdo.CustomPluginMonitorConfigPaths {
		problemDaemons = append(problemDaemons, custompluginmonitor.NewCustomPluginMonitorOrDie(config))
	}

	if len(problemDaemons) == 0 {
		glog.Fatalf("No problem daemon is successfully setup")
	}

	// Initialize exporters.
	exporters := []types.Exporter{}
	if ke := k8sexporter.NewExporterOrDie(npdo); ke != nil {
		exporters = append(exporters, ke)
	}

	if len(exporters) == 0 {
		glog.Fatalf("No exporter is successfully setup")
	}

	// Initialize NPD core.
	p := problemdetector.NewProblemDetector(problemDaemons, exporters)

	if err := p.Run(); err != nil {
		glog.Fatalf("Problem detector failed with error: %v", err)
	}
}
