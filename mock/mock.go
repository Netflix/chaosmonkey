// Copyright 2016 Netflix, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package mock contains helper functions for generating mock objects
// for testing
package mock

import D "github.com/Netflix/chaosmonkey/v2/deploy"

// AppFactory creates App objects used for testing
type AppFactory struct {
}

// App creates a mock App
func (factory AppFactory) App() *D.App {

	var m = D.AppMap{
		"prod": D.AccountInfo{
			CloudProvider: "aws",
			Clusters: D.ClusterMap{
				"abc-prod": {
					"us-east-1": {
						"abc-prod-v017": []D.InstanceID{"i-f60b22e8", "i-1b17963b", "i-7c0c8af4"},
					},
					"us-west-2": {
						"abc-prod-v017": []D.InstanceID{"i-8b42d04e", "i-52ead2f0", "i-b6261b80"},
					},
				},
			},
		},
		"test": D.AccountInfo{
			CloudProvider: "aws",
			Clusters: D.ClusterMap{
				"abc-beta": {
					"us-east-1": {
						"abc-beta-v031": []D.InstanceID{"i-c8a5458c", "i-61f55db3", "i-6a820363"},
						"abc-beta-v030": []D.InstanceID{"i-c41206b7", "i-c8a5458c", "i-6a820363"},
					},
				},
			},
		},
	}
	return D.NewApp("abc", m)
}
