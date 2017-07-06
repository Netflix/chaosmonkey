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

package spinnaker

import "fmt"

// appsUrl returns the Spinnaker endpoint for retrieving all applications
func (s Spinnaker) appsURL() string {
	return s.endpoint + "/applications"
}

// appUrl returns the Spinnaker endpoint for retrieving one application
func (s Spinnaker) appURL(appName string) string {
	return s.endpoint + "/applications/" + appName
}

// clustersUrl returns the Spinnaker endpoint for retrieving applications
func (s Spinnaker) clustersURL(appName string) string {
	return fmt.Sprintf("%s/applications/%s/clusters", s.endpoint, appName)
}

// clusterUrl returns the Spinnaker endpoint for retrieving info about a cluster
func (s Spinnaker) clusterURL(appName string, account string, clusterName string) string {
	return fmt.Sprintf("%s/applications/%s/clusters/%s/%s", s.endpoint, appName, account, clusterName)
}

// serverGroupsUrl returns the Spinnaker endpoint for retrieving server groups
func (s Spinnaker) serverGroupsURL(appName, account, clusterName string) string {
	return fmt.Sprintf("%s/applications/%s/clusters/%s/%s/serverGroups", s.endpoint, appName, account, clusterName)
}

// accountURL returns the Spinnaker endpoint for retrieving account info
func (s Spinnaker) accountURL(account string) string {
	return fmt.Sprintf("%s/credentials/%s", s.endpoint, account)
}

// instanceURL returns the spinnaker URL for an instance
func (s Spinnaker) instanceURL(account string, region string, id string) string {
	return fmt.Sprintf("%s/instances/%s/%s/%s", s.endpoint, account, region, id)
}

// activeASGURL returns the spinnaker URL for getting the active asg in a cluster
func (s Spinnaker) activeASGURL(appName, account, clusterName, cloudProvider, region string) string {
	return fmt.Sprintf("%s/applications/%s/clusters/%s/%s/%s/%s/serverGroups/target/CURRENT?onlyEnabled=true",
		s.endpoint, appName, account, clusterName, cloudProvider, region)
}
