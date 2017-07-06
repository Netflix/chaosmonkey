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

// App represents an application
type App struct {
	name     string
	accounts []*Account
}

// Name returns the name of an app
func (a App) Name() string {
	return a.name
}

// Accounts returns a slice of accounts
func (a App) Accounts() []*Account {
	return a.accounts
}

type (
	// AppName is the name of an app
	AppName string

	// AccountName is the name of a cloud account
	AccountName string

	// ClusterName is the app-stack-detail name of a cluster
	ClusterName string

	// StackName is the stack part of the cluster name
	StackName string

	// RegionName is the name of an AWS region
	RegionName string

	// ASGName is the app-stack-detail-sequence name of an ASG
	ASGName string

	// InstanceID is the i-xxxxxx name of an AWS instance or uuid of a container
	InstanceID string

	// CloudProvider is the name of the cloud backend (e.g., aws)
	CloudProvider string

	// ClusterMap maps cluster name to information about instances by region and
	// ASG
	ClusterMap map[ClusterName]map[RegionName]map[ASGName][]InstanceID

	// AccountInfo tracks the provider and the clusters
	AccountInfo struct {
		CloudProvider string
		Clusters      ClusterMap
	}

	// AppMap is a map that tracks info about an app
	AppMap map[AccountName]AccountInfo
)

// NewApp constructs a new App
func NewApp(name string, data AppMap) *App {
	app := App{name: name}
	for accountName, accountInfo := range data {
		account := Account{name: string(accountName), app: &app, cloudProvider: accountInfo.CloudProvider}
		app.accounts = append(app.accounts, &account)
		for clusterName, clusterValue := range accountInfo.Clusters {
			cluster := Cluster{name: string(clusterName), account: &account}
			account.clusters = append(account.clusters, &cluster)
			for regionName, regionValue := range clusterValue {
				for asgName, instanceIds := range regionValue {
					asg := ASG{
						name:    string(asgName),
						region:  string(regionName),
						cluster: &cluster,
					}
					cluster.asgs = append(cluster.asgs, &asg)
					for _, id := range instanceIds {
						instance := Instance{
							id:  string(id),
							asg: &asg,
						}
						asg.instances = append(asg.instances, &instance)
					}
				}
			}
		}
	}

	return &app
}
