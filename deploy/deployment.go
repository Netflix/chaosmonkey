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

// Package deploy contains information about all of the deployed instances, and how
// they are organized across accounts, apps, regions, clusters, and autoscaling
// groups.
package deploy

import (
	"fmt"

	"github.com/SmartThingsOSS/frigga-go"
)

// Deployment contains information about how apps are deployed
type Deployment interface {
	// Apps sends App objects over a channel
	Apps(c chan<- *App, appNames []string)

	// GetApp retrieves a single App
	GetApp(name string) (*App, error)

	// AppNames returns the names of all apps
	AppNames() ([]string, error)

	// GetInstanceIDs returns the ids for instances in a cluster
	GetInstanceIDs(app string, account AccountName, cloudProvider string, region RegionName, cluster ClusterName) (asgName ASGName, instances []InstanceID, err error)

	// GetClusterNames returns the list of cluster names
	GetClusterNames(app string, account AccountName) ([]ClusterName, error)

	// GetRegionNames returns the list of regions associated with a cluster
	GetRegionNames(app string, account AccountName, cluster ClusterName) ([]RegionName, error)

	// CloudProvider returns the provider associated with an account
	CloudProvider(account string) (provider string, err error)
}

// Account represents the set of clusters associated with an App that reside
// in one AWS account (e.g., "prod", "test").
type Account struct {
	name          string // e.g., "prod", "test"
	clusters      []*Cluster
	app           *App
	cloudProvider string // e.g., "aws"
}

// Name returns the name of the account associated with this account
func (a *Account) Name() string {
	return a.name
}

// Clusters returns a slice of clusters
func (a *Account) Clusters() []*Cluster {
	return a.clusters
}

// AppName returns the name of the app associated with this Account
func (a *Account) AppName() string {
	return a.app.name
}

// RegionNames returns the name of the regions that clusters in this account are
// running in
func (a *Account) RegionNames() []string {
	m := make(map[string]bool)

	// Get the region names of the clusters
	for _, cluster := range a.Clusters() {
		for _, name := range cluster.RegionNames() {
			m[name] = true
		}
	}

	result := make([]string, 0, len(m))
	for name := range m {
		result = append(result, name)
	}

	return result
}

// CloudProvider returns the cloud provider (e.g., "aws")
func (a *Account) CloudProvider() string {
	return a.cloudProvider
}

type stringSet map[string]bool

func (s *stringSet) add(val string) {
	(*s)[val] = true
}

// slice converts a stringSet to a string slice
func (s stringSet) slice() []string {
	result := []string{}
	for val := range s {
		result = append(result, val)
	}
	return result
}

// StackNames returns the names of the stacks associated with this account
func (a *Account) StackNames() []string {
	stacks := make(stringSet)

	for _, cluster := range a.Clusters() {
		stacks.add(cluster.StackName())
	}

	return stacks.slice()
}

// Cluster represents what Spinnaker refers to as a "cluster", which
// contains app-stack-detail.
// Every ASG is associated with exactly one cluster.
// Note that clusters can span regions
type Cluster struct {
	name    string
	asgs    []*ASG
	account *Account
}

// Name returns the name of the cluster, convention: app-stack-detail
func (c *Cluster) Name() string {
	return c.name
}

// AppName returns the name of the app associated with this cluster
func (c *Cluster) AppName() string {
	return c.account.AppName()
}

// StackName returns the name of the stack, following the app-stack-detail convention
func (c *Cluster) StackName() string {
	names, err := frigga.Parse(c.Name())
	if err != nil {
		panic(err)
	}
	return names.Stack
}

// AccountName returns the name of the account associated with this cluster
func (c *Cluster) AccountName() string {
	return c.account.Name()
}

// ASGs returns a slice of ASGs
func (c *Cluster) ASGs() []*ASG {
	return c.asgs
}

// RegionNames returns the name of the region that this cluster runs in
func (c *Cluster) RegionNames() []string {
	m := make(map[string]bool)
	for _, asg := range c.ASGs() {
		m[asg.RegionName()] = true
	}

	result := []string{}
	for name := range m {
		result = append(result, name)
	}

	return result
}

// CloudProvider returns the cloud provider (e.g., "aws")
func (c *Cluster) CloudProvider() string {
	return c.account.CloudProvider()
}

// Instance implements instance.Instance
type Instance struct {
	// instance id (e.g., "i-74e93ddb")
	id string

	// ASG that this instance is part of
	asg *ASG
}

func (i *Instance) String() string {
	return fmt.Sprintf("app=%s account=%s region=%s stack=%s cluster=%s asg=%s instance-id=%s",
		i.AppName(), i.AccountName(), i.RegionName(), i.StackName(), i.ClusterName(), i.ASGName(), i.ID())
}

// AppName returns the name of the app associated with this instance
func (i *Instance) AppName() string {
	return i.asg.AppName()
}

// AccountName returns the name of the AWS account associated with the instance
func (i *Instance) AccountName() string {
	return i.asg.AccountName()
}

// ClusterName returns the name of the cluster associated with the instance
func (i *Instance) ClusterName() string {
	return i.asg.ClusterName()
}

// RegionName returns the name of the region associated with the instance
func (i *Instance) RegionName() string {
	return i.asg.RegionName()
}

// ASGName returns the name of the ASG associated with the instance
func (i *Instance) ASGName() string {
	return i.asg.Name()
}

// StackName returns the name of the stack associated with the instance
func (i *Instance) StackName() string {
	return i.asg.StackName()
}

// CloudProvider returns the cloud provider (e.g., "aws")
func (i *Instance) CloudProvider() string {
	return i.asg.CloudProvider()
}

// ID returns the instance id
func (i *Instance) ID() string {
	return i.id
}
