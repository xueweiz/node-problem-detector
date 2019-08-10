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

package config

import (
	"fmt"
	"time"
)

var (
	defaultInvokeIntervalString = (60 * time.Second).String()
)

type MetricConfig struct {
	DisplayName string `json:"displayName"`
}

type SelfMonitorConfig struct {
	InvokeIntervalString string          `json:"invokeInterval"`
	InvokeInterval       time.Duration   `json:"-"`
	MetricsConfigs        map[string]MetricConfig `json:"metricsConfigs"`
}

// ApplyConfiguration applies default configurations.
func (smc *SelfMonitorConfig) ApplyConfiguration() error {
	if smc.InvokeIntervalString == "" {
		smc.InvokeIntervalString = defaultInvokeIntervalString
	}

	var err error
	smc.InvokeInterval, err = time.ParseDuration(smc.InvokeIntervalString)
	if err != nil {
		return fmt.Errorf("error in parsing InvokeIntervalString %q: %v", smc.InvokeIntervalString, err)
	}

	return nil
}

// Validate verifies whether the settings are valid.
func (smc *SelfMonitorConfig) Validate() error {
	if smc.InvokeInterval <= time.Duration(0) {
		return fmt.Errorf("InvokeInterval %v must be above 0s", smc.InvokeInterval)
	}
	// if smc.DiskConfig.LsblkTimeout <= time.Duration(0) {
	// 	return fmt.Errorf("LsblkTimeout %v must be above 0s", smc.DiskConfig.LsblkTimeout)
	// }

	return nil
}
