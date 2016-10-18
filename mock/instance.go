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

package mock

// Instance implements instance.Instance
type Instance struct {
	App, Account, Stack, Cluster, Region, ASG, InstanceID string
}

// AppName implements instance.AppName
func (i Instance) AppName() string {
	return i.App
}

// AccountName implements instance.AccountName
func (i Instance) AccountName() string {
	return i.Account
}

// RegionName implements instance.RegionName
func (i Instance) RegionName() string {
	return i.Region
}

// StackName implements instance.StackName
func (i Instance) StackName() string {
	return i.Stack
}

// ClusterName implements instance.ClusterName
func (i Instance) ClusterName() string {
	return i.Cluster
}

// ASGName implements instance.ASGName
func (i Instance) ASGName() string {
	return i.ASG
}

// ID implements instance.ID
func (i Instance) ID() string {
	return i.InstanceID
}

// CloudProvider implements instance.IsContainer
func (i Instance) CloudProvider() string {
	return "aws"
}
