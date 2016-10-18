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

import frigga "github.com/SmartThingsOSS/frigga-go"

// ASG identifies an autoscaling group in the deployment
type ASG struct {
	name      string
	region    string
	instances []*Instance
	cluster   *Cluster
}

// NewASG creates a new ASG
func NewASG(name, region string, instanceIDs []string, cluster *Cluster) *ASG {
	result := ASG{
		name:      name,
		region:    region,
		instances: make([]*Instance, len(instanceIDs)),
		cluster:   cluster,
	}

	for i, id := range instanceIDs {
		result.instances[i] = &Instance{id, &result}
	}

	return &result
}

// Instances returns a slice of the instances associated with the ASG
func (a *ASG) Instances() []*Instance {
	return a.instances
}

// Empty returns true if the ASG does not contain any instances
func (a *ASG) Empty() bool {
	return len(a.instances) == 0
}

// AppName returns the name of the app associated with this ASG
func (a *ASG) AppName() string {
	return a.cluster.AppName()
}

// AccountName returns the name of the AWS account associated with the ASG
func (a *ASG) AccountName() string {
	return a.cluster.AccountName()
}

// ClusterName returns the name of the cluster associated with the ASG
func (a *ASG) ClusterName() string {
	return a.cluster.name
}

// DetailName returns the name of the detail field associated with the ASG
func (a *ASG) DetailName() string {

	asgName := a.Name()

	if a.missingPushNumber() {
		/*
			ASGs that were launched before Spinnaker existed may be missing the -vXXX
			push number at the end of the ASG. If this happens, we need to guard
			against the case where the detail field happens to match the push
			field syntax.

			In this case, we work around it by appending a phony push number before
			parsing with frigga.
		*/
		asgName += "-v000"
	}

	names, err := frigga.Parse(asgName)
	if err != nil {
		panic(err)
	}

	return names.Detail
}

// missingPushNumber returns true if the ASG does not have an associated push
// number
func (a *ASG) missingPushNumber() bool {
	return a.Name() == a.ClusterName()
}

// RegionName returns the name of the region associated with the ASG
func (a *ASG) RegionName() string {
	return a.region
}

// Name returns the name of the ASG
func (a *ASG) Name() string {
	return a.name
}

// StackName returns the name of the stack
func (a *ASG) StackName() string {
	return a.cluster.StackName()
}

// CloudProvider returns the cloud provider (e.g., "aws")
func (a *ASG) CloudProvider() string {
	return a.cluster.CloudProvider()
}
