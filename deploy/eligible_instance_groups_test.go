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

package deploy

import (
	"reflect"
	"testing"

	"github.com/Netflix/chaosmonkey/v2"
	"github.com/Netflix/chaosmonkey/v2/grp"
)

type groupList []grp.InstanceGroup

var grouptests = []struct {
	cfg    chaosmonkey.AppConfig
	groups []grp.InstanceGroup
}{
	{conf(chaosmonkey.App, false), groupList{
		grp.New("mock", "prod", "", "", ""),
		grp.New("mock", "test", "", "", ""),
	}},
	{conf(chaosmonkey.App, true), groupList{
		grp.New("mock", "prod", "us-east-1", "", ""),
		grp.New("mock", "prod", "us-west-2", "", ""),
		grp.New("mock", "test", "us-east-1", "", ""),
		grp.New("mock", "test", "us-west-2", "", ""),
	}},
	{conf(chaosmonkey.Stack, false), groupList{
		grp.New("mock", "prod", "", "prod", ""),
		grp.New("mock", "prod", "", "staging", ""),
		grp.New("mock", "test", "", "test", ""),
		grp.New("mock", "test", "", "beta", ""),
	}},
	{conf(chaosmonkey.Stack, true), groupList{
		grp.New("mock", "prod", "us-east-1", "prod", ""),
		grp.New("mock", "prod", "us-west-2", "prod", ""),
		grp.New("mock", "prod", "us-east-1", "staging", ""),
		grp.New("mock", "prod", "us-west-2", "staging", ""),
		grp.New("mock", "test", "us-east-1", "test", ""),
		grp.New("mock", "test", "us-west-2", "test", ""),
		grp.New("mock", "test", "us-east-1", "beta", ""),
		grp.New("mock", "test", "us-west-2", "beta", ""),
	}},
	{conf(chaosmonkey.Cluster, false), groupList{
		grp.New("mock", "prod", "", "", "mock-prod-a"),
		grp.New("mock", "prod", "", "", "mock-prod-b"),
		grp.New("mock", "prod", "", "", "mock-staging-a"),
		grp.New("mock", "prod", "", "", "mock-staging-b"),
		grp.New("mock", "test", "", "", "mock-test-a"),
		grp.New("mock", "test", "", "", "mock-test-b"),
		grp.New("mock", "test", "", "", "mock-beta-a"),
		grp.New("mock", "test", "", "", "mock-beta-b"),
	}},
	{conf(chaosmonkey.Cluster, true), groupList{
		grp.New("mock", "prod", "us-east-1", "", "mock-prod-a"),
		grp.New("mock", "prod", "us-west-2", "", "mock-prod-a"),
		grp.New("mock", "prod", "us-east-1", "", "mock-prod-b"),
		grp.New("mock", "prod", "us-west-2", "", "mock-prod-b"),
		grp.New("mock", "prod", "us-east-1", "", "mock-staging-a"),
		grp.New("mock", "prod", "us-west-2", "", "mock-staging-a"),
		grp.New("mock", "prod", "us-east-1", "", "mock-staging-b"),
		grp.New("mock", "prod", "us-west-2", "", "mock-staging-b"),
		grp.New("mock", "test", "us-east-1", "", "mock-test-a"),
		grp.New("mock", "test", "us-west-2", "", "mock-test-a"),
		grp.New("mock", "test", "us-east-1", "", "mock-test-b"),
		grp.New("mock", "test", "us-west-2", "", "mock-test-b"),
		grp.New("mock", "test", "us-east-1", "", "mock-beta-a"),
		grp.New("mock", "test", "us-west-2", "", "mock-beta-a"),
		grp.New("mock", "test", "us-east-1", "", "mock-beta-b"),
		grp.New("mock", "test", "us-west-2", "", "mock-beta-b"),
	}},
}

func TestEligibleInstanceGroups(t *testing.T) {
	for i, tt := range grouptests {
		groups := mockApp.EligibleInstanceGroups(tt.cfg)
		if len(tt.groups) != len(groups) {
			t.Errorf("test %d: incorrect number of groups. Expected: %d. Actual: %d", i, len(tt.groups), len(groups))
			continue
		}

		if !same(tt.groups, groups) {
			t.Errorf("test %d. Expected: %+v. Actual: %+v", i, tt.groups, groups)
		}
	}
}

//
// Test helper code
//

// conf creates a config file used for testing
func conf(grouping chaosmonkey.Group, regionsAreIndependent bool) chaosmonkey.AppConfig {
	return chaosmonkey.AppConfig{
		Enabled:                        true,
		RegionsAreIndependent:          regionsAreIndependent,
		MeanTimeBetweenKillsInWorkDays: 5,
		MinTimeBetweenKillsInWorkDays:  1,
		Grouping:                       grouping,
	}
}

type groupSet map[grp.InstanceGroup]bool

func (gs *groupSet) add(group grp.InstanceGroup) {
	(*gs)[group] = true
}

func (gl groupList) toSet() groupSet {
	result := make(groupSet)
	for _, group := range gl {
		result.add(group)
	}
	return result
}

// same return true if the two lists of groups contain the same elements,
// independent of order
func same(x, y groupList) bool {
	sx := x.toSet()
	sy := y.toSet()
	return reflect.DeepEqual(sx, sy)
}

var usEast1 = RegionName("us-east-1")
var usWest2 = RegionName("us-west-2")

var mockApp = NewApp("mock", AppMap{

	AccountName("prod"): {
		CloudProvider: "aws",
		Clusters: ClusterMap{
			ClusterName("mock-prod-a"): {
				usEast1: {
					ASGName("mock-prod-a-v123"): []InstanceID{"i-4a003cd0"},
				},
				usWest2: {
					ASGName("mock-prod-a-v111"): []InstanceID{"i-efdc42dc"},
				},
			},
			ClusterName("mock-prod-b"): {
				usEast1: {
					ASGName("mock-prod-b-v002"): []InstanceID{"i-115ccc27"},
				},
				usWest2: {
					ASGName("mock-prod-b-v001"): []InstanceID{"i-7881287e"},
				},
			},
			ClusterName("mock-staging-a"): {
				usEast1: {
					ASGName("mock-staging-a-v123"): []InstanceID{"i-ff8e7e4b"},
				},
				usWest2: {
					ASGName("mock-staging-a-v111"): []InstanceID{"i-6eed18a4"},
				},
			},
			ClusterName("mock-staging-b"): {
				usEast1: {
					ASGName("mock-staging-b-v002"): []InstanceID{"i-13770e40"},
				},
				usWest2: {
					ASGName("mock-staging-b-v001"): []InstanceID{"i-afb7595e"},
				},
			},
		},
	},
	AccountName("test"): {
		CloudProvider: "aws",
		Clusters: ClusterMap{
			ClusterName("mock-test-a"): {
				usEast1: {
					ASGName("mock-test-a-v123"): []InstanceID{"i-23b61f89"},
				},
				usWest2: {
					ASGName("mock-test-a-v111"): []InstanceID{"i-fe7a0827"},
				},
			},
			ClusterName("mock-test-b"): {
				usEast1: {
					ASGName("mock-test-b-v002"): []InstanceID{"i-f581d5c3"},
				},
				usWest2: {
					ASGName("mock-test-b-v001"): []InstanceID{"i-986e988a"},
				},
			},
			ClusterName("mock-beta-a"): {
				usEast1: {
					ASGName("mock-beta-a-v123"): []InstanceID{"i-4b359d5d"},
				},
				usWest2: {
					ASGName("mock-beta-a-v111"): []InstanceID{"i-e751bdd2"},
				},
			},
			ClusterName("mock-beta-b"): {
				usEast1: {
					ASGName("mock-beta-b-v002"): []InstanceID{"i-e5eeba5e"},
				},
				usWest2: {
					ASGName("mock-beta-b-v001"): []InstanceID{"i-76013ffb"},
				},
			},
		},
	},
})
