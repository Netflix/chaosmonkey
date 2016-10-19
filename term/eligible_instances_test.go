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

package term

import (
	"testing"

	"github.com/Netflix/chaosmonkey"
	D "github.com/Netflix/chaosmonkey/deploy"
	"github.com/Netflix/chaosmonkey/grp"
)

// mockApp returns a deploy.App object mock for testing
func mockApp() *D.App {
	usEast1 := D.RegionName("us-east-1")
	usWest2 := D.RegionName("us-west-2")
	return D.NewApp("mock", D.AppMap{
		D.AccountName("prod"): {
			CloudProvider: "aws",
			Clusters: D.ClusterMap{
				D.ClusterName("mock-prod-a"): {
					usEast1: {
						D.ASGName("mock-prod-a-v123"): []D.InstanceID{"i-4a003cd0"},
					},
					usWest2: {
						D.ASGName("mock-prod-a-v111"): []D.InstanceID{"i-efdc42dc"},
					},
				},
				D.ClusterName("mock-prod-b"): {
					usEast1: {
						D.ASGName("mock-prod-b-v002"): []D.InstanceID{"i-115ccc27"},
					},
					usWest2: {
						D.ASGName("mock-prod-b-v001"): []D.InstanceID{"i-7881287e"},
					},
				},
				D.ClusterName("mock-staging-a"): {
					usEast1: {
						D.ASGName("mock-staging-a-v123"): []D.InstanceID{"i-ff8e7e4b"},
					},
					usWest2: {
						D.ASGName("mock-staging-a-v111"): []D.InstanceID{"i-6eed18a4"},
					},
				},
				D.ClusterName("mock-staging-b"): {
					usEast1: {
						D.ASGName("mock-staging-b-v002"): []D.InstanceID{"i-13770e40"},
					},
					usWest2: {
						D.ASGName("mock-staging-b-v001"): []D.InstanceID{"i-afb7595e"},
					},
				},
			},
		},
		D.AccountName("test"): {
			CloudProvider: "aws",
			Clusters: D.ClusterMap{
				D.ClusterName("mock-test-a"): {
					usEast1: {
						D.ASGName("mock-test-a-v123"): []D.InstanceID{"i-23b61f89"},
					},
					usWest2: {
						D.ASGName("mock-test-a-v111"): []D.InstanceID{"i-fe7a0827"},
					},
				},
				D.ClusterName("mock-test-b"): {
					usEast1: {
						D.ASGName("mock-test-b-v002"): []D.InstanceID{"i-f581d5c3"},
					},
					usWest2: {
						D.ASGName("mock-test-b-v001"): []D.InstanceID{"i-986e988a"},
					},
				},
				D.ClusterName("mock-beta-a"): {
					usEast1: {
						D.ASGName("mock-beta-a-v123"): []D.InstanceID{"i-4b359d5d"},
					},
					usWest2: {
						D.ASGName("mock-beta-a-v111"): []D.InstanceID{"i-e751bdd2"},
					},
				},
				D.ClusterName("mock-beta-b"): {
					usEast1: {
						D.ASGName("mock-beta-b-v002"): []D.InstanceID{"i-e5eeba5e"},
					},
					usWest2: {
						D.ASGName("mock-beta-b-v001"): []D.InstanceID{"i-76013ffb"},
					},
				},
			},
		},
	})
}

func testConfig(grouping chaosmonkey.Group) chaosmonkey.AppConfig {
	return chaosmonkey.AppConfig{
		Enabled:                        true,
		RegionsAreIndependent:          true,
		MeanTimeBetweenKillsInWorkDays: 5,
		MinTimeBetweenKillsInWorkDays:  1,
		Grouping:                       grouping,
	}
}

func TestGetEligibleInstances(t *testing.T) {
	app := mockApp()
	group := grp.New("mock", "prod", "us-east-1", "", "mock-prod-a")
	cfg := testConfig(chaosmonkey.Cluster)

	instances := EligibleInstances(group, cfg, app)
	got, want := len(instances), 1
	if got != want {
		t.Fatalf("len(eligibleInstances(group, cfg, app))=%v, want %v", got, want)
	}

	if instances[0].ID() != "i-4a003cd0" {
		t.Fatal("Expected id i-4a003cd0, got", instances[0].ID())
	}
}

func TestDisabled(t *testing.T) {
	app := mockApp()
	group := grp.New("mock", "prod", "us-east-1", "", "mock-prod-a")
	cfg := testConfig(chaosmonkey.Cluster)
	cfg.Enabled = false

	instances := EligibleInstances(group, cfg, app)
	got, want := len(instances), 0
	if got != want {
		t.Fatalf("len(eligibleInstances(group, cfg, app))=%v, want %v", got, want)
	}
}

func TestSimpleException(t *testing.T) {
	app := mockApp()
	group := grp.New("mock", "prod", "us-east-1", "", "mock-prod-a")
	cfg := testConfig(chaosmonkey.Cluster)
	cfg.Exceptions = []chaosmonkey.Exception{{Account: "prod", Stack: "prod", Detail: "a", Region: "us-east-1"}}
	instances := EligibleInstances(group, cfg, app)
	got, want := len(instances), 0
	if got != want {
		t.Fatalf("len(eligibleInstances(group, cfg, app))=%v, want %v", got, want)
	}
}

func TestMultipleExceptions(t *testing.T) {
	app := abcloudMockApp()
	// Group across everything in prod
	group := grp.New("abcloud", "prod", "", "", "")
	cfg := testConfig(chaosmonkey.Cluster)
	cfg.Exceptions = []chaosmonkey.Exception{
		{Account: "prod", Stack: "batch", Detail: "", Region: "eu-west-1"},
		{Account: "prod", Stack: "ecom", Detail: "", Region: "us-west-2"},
		{Account: "prod", Stack: "", Detail: "", Region: "us-west-2"},
	}

	instances := EligibleInstances(group, cfg, app)
	got, want := len(instances), 6
	if got != want {
		t.Fatalf("len(eligibleInstances(group, cfg, app))=%v, want %v", got, want)
	}

	// Ensure none of the excepted instances are in the list
	for _, instance := range instances {
		if instance.ID() == "i-8a1bd7ac" || instance.ID() == "i-2910a0e4" || instance.ID() == "i-b28a69c8" {
			t.Errorf("excepted instance is present: %v", instance)
		}
	}
}

// mockApp based on actual structure of abcloud
func abcloudMockApp() *D.App {
	usEast1 := D.RegionName("us-east-1")
	usWest2 := D.RegionName("us-west-2")
	euWest1 := D.RegionName("eu-west-1")
	m := D.AppMap{
		D.AccountName("prod"): {
			CloudProvider: "aws",
			Clusters: D.ClusterMap{
				D.ClusterName("abcloud"): {
					usEast1: {
						D.ASGName("abcloud-v123"): []D.InstanceID{"i-7921a2f8"},
					},
					usWest2: {
						D.ASGName("abcloud-v123"): []D.InstanceID{"i-8a1bd7ac"},
					},
					euWest1: {
						D.ASGName("abcloud-v123"): []D.InstanceID{"i-87a90e92"},
					},
				},
				D.ClusterName("abcloud-batch"): {
					usEast1: {
						D.ASGName("abcloud-batch-v123"): []D.InstanceID{"i-2c25ab60"},
					},
					usWest2: {
						D.ASGName("abcloud-batch-v123"): []D.InstanceID{"i-3bc40bdb"},
					},
					euWest1: {
						D.ASGName("abcloud-batch-v123"): []D.InstanceID{"i-2910a0e4"},
					},
				},
				D.ClusterName("abcloud-ecom"): {
					usEast1: {
						D.ASGName("abcloud-ecom-v123"): []D.InstanceID{"i-ab9a4f10"},
					},
					usWest2: {
						D.ASGName("abcloud-ecom-v123"): []D.InstanceID{"i-b28a69c8"},
					},
					euWest1: {
						D.ASGName("abcloud-ecom-v123"): []D.InstanceID{"i-4fa09365"},
					},
				},
			},
		},
	}
	return D.NewApp("abcloud", m)
}
