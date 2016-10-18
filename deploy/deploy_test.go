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
	"runtime"
	"testing"
)

func TestASGAndClusters(t *testing.T) {
	nameOf := func(f interface{}) string {
		return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	}

	type tcase struct {
		appName     string
		accountName string
		regionName  string
		clusterName string
		asgName     string
		ids         []string
	}

	makeClusterASG := func(tc tcase) (*Cluster, *ASG) {
		var cluster Cluster
		var account Account
		var app App

		cloudProvider := "aws"

		asg := NewASG(tc.asgName, tc.regionName, tc.ids, &cluster)
		cluster = Cluster{tc.clusterName, []*ASG{asg}, &account}
		account = Account{tc.accountName, []*Cluster{&cluster}, &app, cloudProvider}
		app = App{tc.appName, []*Account{&account}}

		return &cluster, asg
	}

	type at struct {
		f    func(*ASG) string
		want string
	}

	type ct struct {
		f    func(*Cluster) string
		want string
	}

	var tests = []struct {
		scenario string
		t        tcase
		a        []at
		c        []ct
	}{
		{
			"stack and detail",
			tcase{"foo", "test", "us-east-1", "foo-staging-bar", "foo-staging-bar-v031", []string{"i-ff075688", "i-d9165a77"}},
			[]at{
				at{(*ASG).Name, "foo-staging-bar-v031"},
				at{(*ASG).AppName, "foo"},
				at{(*ASG).AccountName, "test"},
				at{(*ASG).RegionName, "us-east-1"},
				at{(*ASG).ClusterName, "foo-staging-bar"},
				at{(*ASG).StackName, "staging"},
				at{(*ASG).DetailName, "bar"},
			},
			[]ct{
				ct{(*Cluster).Name, "foo-staging-bar"},
				ct{(*Cluster).AppName, "foo"},
				ct{(*Cluster).AccountName, "test"},
				ct{(*Cluster).StackName, "staging"},
			},
		},
		{
			"no detail",
			tcase{"chaosguineapig", "prod", "eu-west-1", "chaosguineapig-staging", "chaosguineapig-staging-v000", []string{"i-7f40bbf5", "i-7a61d6f2"}},
			[]at{
				at{(*ASG).Name, "chaosguineapig-staging-v000"},
				at{(*ASG).AppName, "chaosguineapig"},
				at{(*ASG).AccountName, "prod"},
				at{(*ASG).RegionName, "eu-west-1"},
				at{(*ASG).ClusterName, "chaosguineapig-staging"},
				at{(*ASG).StackName, "staging"},
				at{(*ASG).DetailName, ""},
			},
			[]ct{
				ct{(*Cluster).Name, "chaosguineapig-staging"},
				ct{(*Cluster).AppName, "chaosguineapig"},
				ct{(*Cluster).AccountName, "prod"},
				ct{(*Cluster).StackName, "staging"},
			},
		},
		{
			"no stack",
			tcase{"chaosguineapig", "test", "eu-west-1", "chaosguineapig", "chaosguineapig-v030", []string{"i-7f40bbf5", "i-7a61d6f2"}},
			[]at{
				at{(*ASG).Name, "chaosguineapig-v030"},
				at{(*ASG).AppName, "chaosguineapig"},
				at{(*ASG).AccountName, "test"},
				at{(*ASG).RegionName, "eu-west-1"},
				at{(*ASG).ClusterName, "chaosguineapig"},
				at{(*ASG).StackName, ""},
				at{(*ASG).DetailName, ""},
			},
			[]ct{
				ct{(*Cluster).Name, "chaosguineapig"},
				ct{(*Cluster).AppName, "chaosguineapig"},
				ct{(*Cluster).AccountName, "test"},
				ct{(*Cluster).StackName, ""},
			},
		},
		{
			// We hit one case where there was a cluster with a name like foo-bar-v2, where the
			// asg had the same name: foo-bar-v2. The ASG had no push number, and the
			// detail looks like a push number.
			"detail looks like push number",
			tcase{"foo", "prod", "us-west-2", "foo-bar-v2", "foo-bar-v2", []string{"i-c7a513fc", "i-e06cfef1"}},
			[]at{
				at{(*ASG).Name, "foo-bar-v2"},
				at{(*ASG).AppName, "foo"},
				at{(*ASG).AccountName, "prod"},
				at{(*ASG).RegionName, "us-west-2"},
				at{(*ASG).ClusterName, "foo-bar-v2"},
				at{(*ASG).StackName, "bar"},
				at{(*ASG).DetailName, "v2"},
			},
			[]ct{
				ct{(*Cluster).Name, "foo-bar-v2"},
				ct{(*Cluster).AppName, "foo"},
				ct{(*Cluster).AccountName, "prod"},
				ct{(*Cluster).StackName, "bar"},
			},
		},
	}

	for _, tt := range tests {
		cluster, asg := makeClusterASG(tt.t)

		// ASG tests
		for _, att := range tt.a {
			if got, want := att.f(asg), att.want; got != want {
				t.Errorf("scenario %s: got %s()=%s, want: %s", tt.scenario, nameOf(att.f), got, want)
			}
		}

		// cluster tests
		for _, ctt := range tt.c {
			if got, want := ctt.f(cluster), ctt.want; got != want {
				t.Errorf("scenario %s: got %s()=%s, want: %s", tt.scenario, nameOf(ctt.f), got, want)
			}
		}
	}
}
