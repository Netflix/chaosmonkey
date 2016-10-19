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
	"strings"

	"github.com/Netflix/chaosmonkey"
	"github.com/Netflix/chaosmonkey/deploy"
	"github.com/Netflix/chaosmonkey/grp"
)

// EligibleInstances returns a list of instances that belong to group that are eligible for termination
// It does not include any instances that match the list of exceptions
func EligibleInstances(group grp.InstanceGroup, cfg chaosmonkey.AppConfig, app *deploy.App) []*deploy.Instance {
	if !cfg.Enabled {
		return nil
	}

	/*
		Pipeline: emit -> filterGroup -> filterWhitelist -> filterException -> filterCanaries -> toInstances

		emit: generates all of the asgs for this app
		filterGroup: filters out asgs that don't match the group
		filterWhitelist: filters out asgs that don't match whitelist (deprecated, will be removed in the future)
		filterExceptions: filters out asgs based on exception list
		filterCanaries: filters out asgs that are part of canary deploys to avoid interfering with canary analysis
		toInstances: converts ASGs to instances
	*/

	egchan := make(chan *deploy.ASG)      // (emit -> filterGroup) channel
	gwchan := make(chan *deploy.ASG)      // (filterGroup -> filterWhitelist) channel
	gechan := make(chan *deploy.ASG)      // (filterGroup -> filterException) channel
	ecchan := make(chan *deploy.ASG)      // (filterExceptions -> filterCanaries) channel
	cichan := make(chan *deploy.ASG)      // (filterCanaries -> toInstance) channel
	tichan := make(chan *deploy.Instance) // (toInstance -> <result> ) channel

	go emit(app, egchan)
	go filterGroup(egchan, gwchan, group)
	go filterWhitelist(gwchan, gechan, cfg.Whitelist)
	go filterExceptions(gechan, ecchan, cfg.Exceptions)
	go filterCanaries(ecchan, cichan)
	go toInstances(cichan, tichan)

	result := []*deploy.Instance{}
	for instance := range tichan {
		result = append(result, instance)
	}

	return result

}

// emit takes all of the ASGs associated with an App and
// pushes them one at a time over a channel
func emit(app *deploy.App, dst chan<- *deploy.ASG) {
	defer close(dst)

	for _, account := range app.Accounts() {
		for _, cluster := range account.Clusters() {
			for _, asg := range cluster.ASGs() {
				dst <- asg
			}
		}
	}
}

// filterWhitelist receives ASGs from src and pushes them through dst if they
// match at least one element in the whitelist. If there's no whitelist,
// they all get through
func filterWhitelist(src <-chan *deploy.ASG, dst chan<- *deploy.ASG, pwl *[]chaosmonkey.Exception) {
	defer close(dst)

	for asg := range src {
		if isWhitelisted(pwl, asg) {
			dst <- asg
		}
	}
}

// isWhitelisted returns true if instances from the ASG
// match any of the elements of the whitelist
//
// If the whitelist pointer is null, it always returns true
func isWhitelisted(pwl *[]chaosmonkey.Exception, asg *deploy.ASG) bool {
	return (pwl == nil) || isException(*pwl, asg)
}

// filterExceptions receives ASGs from src and pushes them through dst, unless
// there's an exception that matches, in which case it does not push that ASG
// through
func filterExceptions(src <-chan *deploy.ASG, dst chan<- *deploy.ASG, exs []chaosmonkey.Exception) {
	defer close(dst)

	for asg := range src {
		if isException(exs, asg) {
			continue
		}

		dst <- asg
	}
}

// filterCanaries receives ASGs from src and pushes them through dst,
// unless the ASG is involved in canarying (name ends with -baseline or -canary)
func filterCanaries(src <-chan *deploy.ASG, dst chan<- *deploy.ASG) {
	defer close(dst)

	for asg := range src {
		if isCanary(asg) {
			continue
		}

		dst <- asg
	}

}

// Returns true if asg is part of a canary deployment
func isCanary(asg *deploy.ASG) bool {
	cluster := asg.ClusterName()

	// If cluster name ends with an element of the blacklist, it's a canary
	//
	// TODO: Specify these in a configuration file instead of hard-coding
	blacklist := []string{"-canary", "-baseline", "-citrus", "-citrusproxy"}

	for _, suffix := range blacklist {
		if strings.HasSuffix(cluster, suffix) {
			return true
		}
	}

	return false
}

// filterGroup receives ASGs from src, and sends ASGs
// to dst if the IntsanceGroup contains the ASG
func filterGroup(src <-chan *deploy.ASG, dst chan<- *deploy.ASG, group grp.InstanceGroup) {
	defer close(dst)

	for asg := range src {
		if contains(group, asg) {
			dst <- asg
		}
	}
}

// matches return true if the group contains the asg
func contains(group grp.InstanceGroup, asg *deploy.ASG) bool {
	return grp.Contains(group, asg.AppName(), asg.AccountName(), asg.RegionName(), asg.StackName(), asg.ClusterName())
}

// toInstances reads ASGs from src, extracts the Instances and writes them to dst
func toInstances(src <-chan *deploy.ASG, dst chan<- *deploy.Instance) {
	defer close(dst)

	for asg := range src {
		for _, instance := range asg.Instances() {
			dst <- instance
		}
	}
}

// isException returns true if instances from the ASG match
// any of the exceptions
func isException(exs []chaosmonkey.Exception, asg *deploy.ASG) bool {
	for _, ex := range exs {
		account := asg.AccountName()
		stack := asg.StackName()
		detail := asg.DetailName()
		region := asg.RegionName()
		if ex.Matches(account, stack, detail, region) {
			return true
		}
	}
	return false
}
