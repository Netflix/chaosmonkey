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

import D "github.com/Netflix/chaosmonkey/deploy"

const cloudProvider = "aws"

// Dep returns a mock implementation of deploy.Deployment
// Dep has 4 apps: foo, bar, baz, quux
// Each app runs in 1 account:
//    foo, bar, baz run in prod
//    quux runs in test
// Each app has one cluster: foo-prod, bar-prod, baz-prod
// Each cluster runs in one region: us-east-1
// Each cluster contains 1 AZ with two instances
func Dep() D.Deployment {
	prod := D.AccountName("prod")
	test := D.AccountName("test")
	usEast1 := D.RegionName("us-east-1")

	return &Deployment{map[string]D.AppMap{
		"foo":  {prod: D.AccountInfo{CloudProvider: cloudProvider, Clusters: D.ClusterMap{"foo-prod": {usEast1: {"foo-prod-v001": []D.InstanceID{"i-d3e3d611", "i-63f52e25"}}}}}},
		"bar":  {prod: D.AccountInfo{CloudProvider: cloudProvider, Clusters: D.ClusterMap{"bar-prod": {usEast1: {"bar-prod-v011": []D.InstanceID{"i-d7f06d45", "i-ce433cf1"}}}}}},
		"baz":  {prod: D.AccountInfo{CloudProvider: cloudProvider, Clusters: D.ClusterMap{"baz-prod": {usEast1: {"baz-prod-v004": []D.InstanceID{"i-25b86646", "i-573d46d5"}}}}}},
		"quux": {test: D.AccountInfo{CloudProvider: cloudProvider, Clusters: D.ClusterMap{"quux-test": {usEast1: {"quux-test-v004": []D.InstanceID{"i-25b866ab", "i-892d46d5"}}}}}},
	}}
}

// NewDeployment returns a mock implementation of deploy.Deployment
// Pass in a deploy.AppMap, for example:
//  map[string]deploy.AppMap{
// 		"foo":  deploy.AppMap{"prod": {"foo-prod": {"us-east-1": {"foo-prod-v001": []string{"i-d3e3d611", "i-63f52e25"}}}}},
// 		"bar":  deploy.AppMap{"prod": {"bar-prod": {"us-east-1": {"bar-prod-v011": []string{"i-d7f06d45", "i-ce433cf1"}}}}},
// 		"baz":  deploy.AppMap{"prod": {"baz-prod": {"us-east-1": {"baz-prod-v004": []string{"i-25b86646", "i-573d46d5"}}}}},
// 		"quux": deploy.AppMap{"test": {"quux-test": {"us-east-1": {"quux-test-v004": []string{"i-25b866ab", "i-892d46d5"}}}}},
// 	}
func NewDeployment(apps map[string]D.AppMap) D.Deployment {
	return &Deployment{apps}
}

// Deployment implements deploy.Deployment interface
type Deployment struct {
	AppMap map[string]D.AppMap
}

// Apps implements deploy.Deployment.Apps
func (d Deployment) Apps(c chan<- *D.App, apps []string) {
	defer close(c)

	for name, appmap := range d.AppMap {
		c <- D.NewApp(name, appmap)
	}
}

// GetClusterNames implements deploy.Deployment.GetClusterNames
func (d Deployment) GetClusterNames(app string, account D.AccountName) ([]D.ClusterName, error) {
	result := make([]D.ClusterName, 0)
	for cluster := range d.AppMap[app][account].Clusters {
		result = append(result, cluster)
	}

	return result, nil
}

// GetRegionNames implements deploy.Deployment.GetRegionNames
func (d Deployment) GetRegionNames(app string, account D.AccountName, cluster D.ClusterName) ([]D.RegionName, error) {
	result := make([]D.RegionName, 0)
	for region := range d.AppMap[app][account].Clusters[cluster] {
		result = append(result, region)
	}

	return result, nil
}

// AppNames implements deploy.Deployment.AppNames
func (d Deployment) AppNames() ([]string, error) {
	result := make([]string, len(d.AppMap), len(d.AppMap))
	i := 0
	for app := range d.AppMap {
		result[i] = app
		i++
	}

	return result, nil
}

// GetApp implements deploy.Deployment.GetApp
func (d Deployment) GetApp(name string) (*D.App, error) {
	return D.NewApp(name, d.AppMap[name]), nil
}

// CloudProvider implements deploy.Deployment.CloudProvider
func (d Deployment) CloudProvider(account string) (string, error) {
	return cloudProvider, nil
}

// GetInstanceIDs implements deploy.Deployment.GetInstanceIDs
func (d Deployment) GetInstanceIDs(app string, account D.AccountName, cloudProvider string, region D.RegionName, cluster D.ClusterName) (D.ASGName, []D.InstanceID, error) {
	// asgs associated with the cluster
	asgs := d.AppMap[app][account].Clusters[cluster][region]

	instances := make([]D.InstanceID, 0)

	// We assume there's only one asg, and retrieve the instances
	var asg D.ASGName
	var ids []D.InstanceID

	for asg, ids = range asgs {
		for _, id := range ids {
			instances = append(instances, id)
		}
	}

	return asg, instances, nil
}
