/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package selfmonitor

import (
	"encoding/json"
	"io/ioutil"
	"time"
	"os"

	"github.com/golang/glog"
	"github.com/shirou/gopsutil/process"

	"k8s.io/node-problem-detector/pkg/problemdaemon"
	"k8s.io/node-problem-detector/pkg/selfmonitor/config"
	"k8s.io/node-problem-detector/pkg/types"
	"k8s.io/node-problem-detector/pkg/util/tomb"
	"k8s.io/node-problem-detector/pkg/util/metrics"
)

const SelfMonitorName = "self-monitor"

func init() {
	problemdaemon.Register(SelfMonitorName, types.ProblemDaemonHandler{
		CreateProblemDaemonOrDie: NewSelfMonitorOrDie,
		CmdOptionDescription:     "Set to config file paths."})
}

type selfMonitor struct {
	configPath    string
	config        config.SelfMonitorConfig
	mCPUTotal     *metrics.Float64Metric
	mMemory       *metrics.Int64Metric
	tomb          *tomb.Tomb
	pid           int32
}

// NewSelfMonitorOrDie creates a self monitor.
func NewSelfMonitorOrDie(configPath string) types.Monitor {
	mon := selfMonitor{
		configPath: configPath,
		tomb:       tomb.NewTomb(),
	}

	// Apply configurations.
	f, err := ioutil.ReadFile(configPath)
	if err != nil {
		glog.Fatalf("Failed to read configuration file %q: %v", configPath, err)
	}
	err = json.Unmarshal(f, &mon.config)
	if err != nil {
		glog.Fatalf("Failed to unmarshal configuration file %q: %v", configPath, err)
	}

	err = mon.config.ApplyConfiguration()
	if err != nil {
		glog.Fatalf("Failed to apply configuration for %q: %v", configPath, err)
	}

	err = mon.config.Validate()
	if err != nil {
		glog.Fatalf("Failed to validate %s configuration %+v: %v", mon.configPath, mon.config, err)
	}

	mon.mCPUTotal, err = metrics.NewFloat64Metric(
		mon.config.MetricsConfigs["npd/cpu_total"].DisplayName,
		"The CPU time used by NPD",
		"s",
		metrics.LastValue,
		[]string{})
	if err != nil {
		glog.Fatalf("Error initializing metric for npd/cpu_total: %v", err)
	}

	mon.mMemory, err = metrics.NewInt64Metric(
		mon.config.MetricsConfigs["npd/memory"].DisplayName,
		"The memory used by NPD",
		"Bytes",
		metrics.LastValue,
		[]string{})
	if err != nil {
		glog.Fatalf("Error initializing metric for npd/memory: %v", err)
	}

	mon.pid = int32(os.Getpid())

	return &mon
}

func (mon *selfMonitor) Start() (<-chan *types.Status, error) {
	glog.Infof("Start self monitor %s", mon.configPath)
	go mon.monitorLoop()
	return nil, nil
}

func (mon *selfMonitor) monitorLoop() {
	defer mon.tomb.Done()

	// runTicker := time.NewTicker(mon.config.InvokeInterval)
	runTicker := time.NewTicker(time.Second * 10)
	defer runTicker.Stop()

	select {
	case <-mon.tomb.Stopping():
		glog.Infof("Self stopped: %s", mon.configPath)
		return
	default:
		mon.collect()
	}

	for {
		select {
		case <-runTicker.C:
			mon.collect()
		case <-mon.tomb.Stopping():
			glog.Infof("Self monitor stopped: %s", mon.configPath)
			return
		}
	}
}

func (mon *selfMonitor) collect() {
	if mon == nil {
		return
	}

	tags := map[string]string{}

	proc, err := process.NewProcess(mon.pid)
	if err != nil {
		glog.Errorf("Failed to find NPD's self process %d: %v", mon.pid, err)
	}

	if mon.mCPUTotal != nil {
		cpuStat, err := proc.Times()
		if err != nil {
			glog.Errorf("Failed to retrieve NPD's CPU stat: %v", err)
		}
		mon.mCPUTotal.Record(tags, cpuStat.Total())
	}

	if mon.mMemory != nil {
		memInfo, err := proc.MemoryInfo()
		if err != nil {
			glog.Errorf("Failed to retrieve NPD's memory information: %v", err)
		}

		mon.mMemory.Record(tags, int64(memInfo.RSS))
	}

}

func (mon *selfMonitor) Stop() {
	glog.Infof("Stop self monitor %s", mon.configPath)
	mon.tomb.Stop()
}
