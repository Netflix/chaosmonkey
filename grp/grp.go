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

// Package grp holds the InstanceGroup interface
package grp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/SmartThingsOSS/frigga-go"
	"log"
)

// New generates an InstanceGroup.
// region, stack, and cluster may be empty strings, in which case
// the group is cross-region, cross-stack, or cross-cluster
// Note that stack and cluster are mutually exclusive, can specify one
// but not both
func New(app, account, region, stack, cluster string) InstanceGroup {
	return group{
		app:     app,
		account: account,
		region:  region,
		stack:   stack,
		cluster: cluster,
	}
}

// InstanceGroup represents a group of instances
type InstanceGroup interface {
	// App returns the name of the app
	App() string

	// Account returns the name of the account
	Account() string

	// Region returns (region name, region present)
	// If the group is cross-region, the boolean will be false
	Region() (name string, ok bool)

	// Stack returns (region name, region present)
	// If the group is cross-stack, the boolean will be false
	Stack() (name string, ok bool)

	// Cluster returns (cluster name, cluster present)
	// If the group is cross-cluster, the boolean will be false
	Cluster() (name string, ok bool)

	// String outputs a stringified rep
	String() string
}

// Equal returns true if g1 and g2 represent the same group of instances
func Equal(g1, g2 InstanceGroup) bool {
	if g1 == g2 {
		return true
	}

	if g1.App() != g2.App() {
		return false
	}

	if g1.Account() != g2.Account() {
		return false
	}

	r1, ok1 := g1.Region()
	r2, ok2 := g2.Region()
	if ok1 != ok2 {
		return false
	}

	if ok1 && (r1 != r2) {
		return false
	}

	s1, ok1 := g1.Stack()
	s2, ok2 := g2.Stack()

	if ok1 != ok2 {
		return false
	}

	if ok1 && (s1 != s2) {
		return false
	}

	c1, ok1 := g1.Cluster()
	c2, ok2 := g2.Cluster()

	if ok1 != ok2 {
		return false
	}

	if ok1 && (c1 != c2) {
		return false
	}

	return true
}

// String outputs a string representation of InstanceGroup suitable for logging
func String(group InstanceGroup) string {
	var buffer bytes.Buffer
	writeString := func(s string) {
		_, _ = buffer.WriteString(s)
	}
	writeString("app=")
	writeString(group.App())
	writeString(" account=")
	writeString(group.Account())
	region, ok := group.Region()
	if ok {
		writeString(" region=")
		writeString(region)
	}
	stack, ok := group.Stack()
	if ok {
		writeString(" stack=")
		writeString(stack)
	}
	cluster, ok := group.Cluster()
	if ok {
		writeString(" cluster=")
		writeString(cluster)
	}

	return buffer.String()
}

type group struct {
	app, account, region, stack, cluster string
}

func (g group) String() string {
	return fmt.Sprintf("InstanceGroup{app=%s account=%s region=%s stack=%s cluster=%s}", g.app, g.account, g.region, g.stack, g.cluster)
}

func (g group) MarshalJSON() ([]byte, error) {
	var s = struct {
		App     string `json:"app"`
		Account string `json:"account"`
		Region  string `json:"region,omitempty"`
		Stack   string `json:"stack,omitempty"`
		Cluster string `json:"cluster,omitempty"`
	}{
		App:     g.app,
		Account: g.account,
		Region:  g.region,
		Stack:   g.stack,
		Cluster: g.cluster,
	}

	return json.Marshal(s)
}

// App implements InstanceGroup.App
func (g group) App() string {
	return g.app
}

// Account implements InstanceGroup.Account
func (g group) Account() string {
	return g.account
}

// Region implements InstanceGroup.Region
func (g group) Region() (string, bool) {
	if g.region == "" {
		return "", false
	}
	return g.region, true
}

// Stack implements InstanceGroup.Stack
func (g group) Stack() (string, bool) {
	if g.stack == "" {
		return "", false
	}
	return g.stack, true
}

// Cluster implements InstanceGroup.Cluster
func (g group) Cluster() (string, bool) {
	if g.cluster == "" {
		return "", false
	}
	return g.cluster, true
}

// AnyRegion is true if the group matches any region
func AnyRegion(g InstanceGroup) bool {
	_, specific := g.Region()
	return !specific
}

// AnyStack is true if the group matches any stack
func AnyStack(g InstanceGroup) bool {
	_, specific := g.Stack()
	return !specific
}

// AnyCluster is true if the group matches any cluster
func AnyCluster(g InstanceGroup) bool {
	_, specific := g.Cluster()
	return !specific
}

// Contains returns true if the (account, region, cluster) is within the instance group
func Contains(g InstanceGroup, account, region, cluster string) bool {
	names, err := frigga.Parse(cluster)
	if err != nil {
		log.Printf("WARNING: could not parse cluster name: %s", cluster)
		return false
	}

	return names.App == g.App() &&
		string(account) == g.Account() &&
		(AnyRegion(g) || string(region) == must(g.Region())) &&
		(AnyStack(g) || names.Stack == must(g.Stack())) &&
		(AnyCluster(g) || string(cluster) == must(g.Cluster()))
}

// must returns val if ok is true
// panics otherwise
func must(val string, specific bool) string {
	if !specific {
		panic("specific was unexpectedly false")
	}
	return val
}
