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

package grp_test

import (
	"testing"

	"github.com/Netflix/chaosmonkey/grp"
)

func TestNewAppWithRegion(t *testing.T) {
	group := grp.New("myapp", "prod", "us-east-1", "", "")

	if group.App() != "myapp" {
		t.Error("Expected myapp, got", group.App())
	}

	if group.Account() != "prod" {
		t.Error("Expected prod, got", group.Account())
	}

	region, ok := group.Region()
	if !ok || region != "us-east-1" {
		t.Error("Expected us-east-1")
	}

	if _, ok := group.Stack(); ok {
		t.Error("Expected no stack")
	}

	if _, ok := group.Cluster(); ok {
		t.Error("Expected no cluster")
	}
}

func TestNewAppCrossRegion(t *testing.T) {
	group := grp.New("myapp", "prod", "", "", "")

	if group.App() != "myapp" {
		t.Error("Expected myapp, got", group.App())
	}

	if group.Account() != "prod" {
		t.Error("Expected prod, got", group.Account())
	}

	if _, ok := group.Region(); ok {
		t.Error("Expected no region")
	}

	if _, ok := group.Stack(); ok {
		t.Error("Expected no stack")
	}

	if _, ok := group.Cluster(); ok {
		t.Error("Expected no cluster")
	}
}

func TestNewStackWithRegion(t *testing.T) {
	group := grp.New("myapp", "prod", "us-east-1", "staging", "")

	if group.App() != "myapp" {
		t.Error("Expected myapp, got", group.App())
	}

	if group.Account() != "prod" {
		t.Error("Expected prod, got", group.Account())
	}

	region, ok := group.Region()
	if !ok || region != "us-east-1" {
		t.Error("Expected us-east-1")
	}

	stack, ok := group.Stack()
	if !ok || stack != "staging" {
		t.Error("Expected stack=staging, got stack=", stack)
	}

	if _, ok := group.Cluster(); ok {
		t.Error("Expected no cluster")
	}
}

func TestNewStackCrossRegion(t *testing.T) {
	group := grp.New("myapp", "prod", "", "staging", "")

	if group.App() != "myapp" {
		t.Error("Expected myapp, got", group.App())
	}

	if group.Account() != "prod" {
		t.Error("Expected prod, got", group.Account())
	}

	if _, ok := group.Region(); ok {
		t.Error("Expected no region")
	}

	stack, ok := group.Stack()
	if !ok || stack != "staging" {
		t.Error("Expected stack=staging, got stack=", stack)
	}

	if _, ok := group.Cluster(); ok {
		t.Error("Expected no cluster")
	}
}

func TestNewClusterWithRegion(t *testing.T) {
	group := grp.New("myapp", "prod", "us-east-1", "", "myapp-prod-foo")

	if group.App() != "myapp" {
		t.Error("Expected myapp, got", group.App())
	}

	if group.Account() != "prod" {
		t.Error("Expected prod, got", group.Account())
	}

	region, ok := group.Region()
	if !ok || region != "us-east-1" {
		t.Error("Expected us-east-1")
	}

	if _, ok := group.Stack(); ok {
		t.Error("Expected no stack")
	}

	cluster, ok := group.Cluster()
	if !ok || cluster != "myapp-prod-foo" {
		t.Error("Expected cluster myapp-prod-foo, got", cluster)
	}
}

func TestNewClusterCrossRegion(t *testing.T) {
	group := grp.New("myapp", "prod", "", "", "myapp-prod-foo")

	if group.App() != "myapp" {
		t.Error("Expected myapp, got", group.App())
	}

	if group.Account() != "prod" {
		t.Error("Expected prod, got", group.Account())
	}

	if _, ok := group.Region(); ok {
		t.Error("Expected no region")
	}

	if _, ok := group.Stack(); ok {
		t.Error("Expected no stack")
	}

	cluster, ok := group.Cluster()
	if !ok || cluster != "myapp-prod-foo" {
		t.Error("Expected cluster myapp-prod-foo, got", cluster)
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		group                    grp.InstanceGroup
		account, region, cluster string
		matches                  bool
	}{
		{grp.New("foo", "prod", "", "", ""), "prod", "us-east-1", "foo-staging-a", true},
		{grp.New("foo", "prod", "us-east-1", "", ""), "prod", "us-east-1", "foo-staging-a", true},
		{grp.New("foo", "prod", "us-east-1", "", ""), "prod", "us-east-1", "foo-staging-a", true},
		{grp.New("foo", "prod", "us-east-1", "", "foo-staging-a"), "prod", "us-east-1", "foo-staging-a", true},
		{grp.New("foo", "prod", "", "", ""), "prod", "us-east-1", "bar-staging-a", false},
		{grp.New("foo", "prod", "", "", ""), "test", "us-east-1", "foo-staging-a", false},
		{grp.New("foo", "prod", "us-east-1", "", "foo-staging-a"), "prod", "us-west-2", "foo-staging-a", false},
	}

	for _, tt := range tests {
		if grp.Contains(tt.group, tt.account, tt.region, tt.cluster) != tt.matches {
			t.Errorf("unexpected grp.Contains(account=%s, region=%s, cluster=%s). group=%+v. expected %t",
				tt.account, tt.region, tt.cluster, tt.group, tt.matches)
		}
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		g1   grp.InstanceGroup
		g2   grp.InstanceGroup
		want bool
	}{
		{grp.New("foo", "prod", "", "", ""), grp.New("foo", "prod", "", "", ""), true},
		{grp.New("foo", "prod", "", "", ""), grp.New("bar", "prod", "", "", ""), false},
		{grp.New("foo", "prod", "", "", ""), grp.New("foo", "test", "", "", ""), false},
		{grp.New("foo", "prod", "us-east-1", "", ""), grp.New("foo", "prod", "us-east-1", "", ""), true},
		{grp.New("foo", "prod", "us-east-1", "", ""), grp.New("foo", "prod", "us-west-2", "", ""), false},
		{grp.New("foo", "prod", "us-east-1", "", ""), grp.New("foo", "test", "us-east-1", "", ""), false},
		{grp.New("foo", "prod", "us-east-1", "", ""), grp.New("bar", "prod", "us-east-1", "", ""), false},
		{grp.New("foo", "prod", "us-east-1", "", ""), grp.New("foo", "prod", "", "", ""), false},
		{grp.New("foo", "prod", "", "", ""), grp.New("foo", "prod", "us-east-1", "", ""), false},
		{grp.New("foo", "prod", "us-east-1", "staging", ""), grp.New("foo", "prod", "us-east-1", "staging", ""), true},
		{grp.New("foo", "prod", "us-east-1", "staging", ""), grp.New("foo", "prod", "us-east-1", "canary", ""), false},
		{grp.New("foo", "prod", "us-east-1", "staging", ""), grp.New("foo", "prod", "us-west-2", "staging", ""), false},
		{grp.New("foo", "prod", "us-east-1", "staging", ""), grp.New("bar", "prod", "us-east-1", "staging", ""), false},
		{grp.New("foo", "prod", "us-east-1", "", "foo-staging-good"), grp.New("foo", "prod", "us-east-1", "", "foo-staging-good"), true},
		{grp.New("foo", "prod", "us-east-1", "", "foo-staging-good"), grp.New("foo", "prod", "us-east-1", "", "foo-staging-bad"), false},
		{grp.New("foo", "prod", "", "", "foo-staging-good"), grp.New("foo", "prod", "", "", "foo-staging-good"), true},
		{grp.New("foo", "prod", "", "", "foo-staging-good"), grp.New("foo", "prod", "us-east-1", "", "foo-staging-good"), false},
	}

	for _, tt := range tests {
		if got, want := grp.Equal(tt.g1, tt.g2), tt.want; got != want {
			t.Errorf("got Equal(%+v, %+v)=%t, want %t", tt.g1, tt.g2, got, want)
		}
	}
}
