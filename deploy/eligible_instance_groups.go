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
	"fmt"
	"log"

	"github.com/Netflix/chaosmonkey/v2"
	"github.com/Netflix/chaosmonkey/v2/grp"
)

// EligibleInstanceGroups returns a slice of InstanceGroups that represent
// groups of instances that are eligible for termination.
//
// Note that this code does not check for violations of minimum time between
// terminations. Chaos Monkey checks that precondition immediately before
// termination, not when considering groups of eligible instances.
//
// The way instances are divided into group will depend on
//   - the grouping configuration for the app (cluster, stack, app)
//   - whether regions are independent
//
// The returned InstanceGroups are guaranteed to contain at least one instance
// each
//
// Preconditions:
//   - app is enabled for Chaos Monkey
func (app *App) EligibleInstanceGroups(cfg chaosmonkey.AppConfig) []grp.InstanceGroup {
	if !cfg.Enabled {
		log.Fatalf("app %s unexpectedly disabled", app.Name())
	}

	grouping := cfg.Grouping
	indep := cfg.RegionsAreIndependent

	switch {
	case grouping == chaosmonkey.App && indep:
		return appIndep(app)
	case grouping == chaosmonkey.App && !indep:
		return appDep(app)
	case grouping == chaosmonkey.Stack && indep:
		return stackIndep(app)
	case grouping == chaosmonkey.Stack && !indep:
		return stackDep(app)
	case grouping == chaosmonkey.Cluster && indep:
		return clusterIndep(app)
	case grouping == chaosmonkey.Cluster && !indep:
		return clusterDep(app)
	default:
		panic(fmt.Sprintf("Unknown grouping: %d", grouping))
	}
}

// appindep returns a list of groups grouped by (app, account, region)
func appIndep(app *App) []grp.InstanceGroup {
	result := []grp.InstanceGroup{}
	for _, account := range app.accounts {
		for _, regionName := range account.RegionNames() {
			result = append(result, grp.New(app.Name(), account.Name(), regionName, "", ""))
		}
	}
	return result
}

// stackIndep returns a list of groups grouped by (app, account)
func appDep(app *App) []grp.InstanceGroup {
	result := []grp.InstanceGroup{}
	for _, account := range app.accounts {
		result = append(result, grp.New(app.Name(), account.Name(), "", "", ""))
	}
	return result
}

// stackIndep returns a list of groups grouped by (app, account, stack, region)
func stackIndep(app *App) []grp.InstanceGroup {

	type asr struct {
		account string
		stack   string
		region  string
	}

	set := make(map[asr]bool)

	for _, account := range app.Accounts() {
		for _, cluster := range account.Clusters() {
			stackName := cluster.StackName()
			for _, regionName := range cluster.RegionNames() {
				set[asr{account: account.Name(), stack: stackName, region: regionName}] = true
			}
		}
	}

	result := []grp.InstanceGroup{}
	for x := range set {
		result = append(result, grp.New(app.Name(), x.account, x.region, x.stack, ""))
	}

	return result
}

// stackDep returns a list of groups grouped by (app, account, stack)
func stackDep(app *App) []grp.InstanceGroup {
	result := []grp.InstanceGroup{}
	for _, account := range app.accounts {
		for _, stackName := range account.StackNames() {
			result = append(result, grp.New(app.Name(), account.Name(), "", stackName, ""))
		}
	}

	return result
}

// clusterDep returns a list of groups grouped by (app, account, cluster, region)
func clusterIndep(app *App) []grp.InstanceGroup {
	result := []grp.InstanceGroup{}
	for _, account := range app.accounts {
		for _, cluster := range account.Clusters() {
			for _, regionName := range cluster.RegionNames() {
				result = append(result, grp.New(app.Name(), account.Name(), regionName, "", cluster.Name()))
			}
		}
	}

	return result
}

// clusterDep returns a list of groups grouped by (app, account, cluster)
func clusterDep(app *App) []grp.InstanceGroup {
	result := []grp.InstanceGroup{}
	for _, account := range app.accounts {
		for _, cluster := range account.Clusters() {
			result = append(result, grp.New(app.Name(), account.Name(), "", "", cluster.Name()))
		}
	}

	return result
}
