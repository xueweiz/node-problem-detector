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

package problemdaemon

import (
	"k8s.io/node-problem-detector/pkg/types"
)

var (
	handlers = make(map[types.ProblemDaemonType]types.ProblemDaemonHandler)
)

// Register registers a problem daemon factory method, which will be used to create the problem daemon.
func Register(problemDaemonType types.ProblemDaemonType, handler types.ProblemDaemonHandler) {
	handlers[problemDaemonType] = handler
}

// GetProblemDaemonNames retrieves the names of all available problem daemon types.
func GetProblemDaemonNames() []types.ProblemDaemonType {
	problemDaemons := []types.ProblemDaemonType{}
	for problemDaemonType := range handlers {
		problemDaemons = append(problemDaemons, problemDaemonType)
	}
	return problemDaemons
}

// GetProblemDaemonHandler retrieves the ProblemDaemonHandler for a specific type of problem daemon.
func GetProblemDaemonHandler(problemDaemonType types.ProblemDaemonType) types.ProblemDaemonHandler {
	return handlers[problemDaemonType]
}

// NewProblemDaemons creates all problem daemons based on the configurations provided.
func NewProblemDaemons(monitorConfigPaths types.ProblemDaemonConfigMap) []types.Monitor {
	problemDaemons := []types.Monitor{}
	for problemDaemonType, configs := range monitorConfigPaths {
		for _, config := range *configs {
			problemDaemon := handlers[problemDaemonType].Factory(config)
			if problemDaemon != nil {
				problemDaemons = append(problemDaemons, problemDaemon)
			}
		}
	}
	return problemDaemons
}
