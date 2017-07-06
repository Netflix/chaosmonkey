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

// Package term contains the logic for terminating instances
package term

import (
	"log"
	"math/rand"
	"time"

	"github.com/pkg/errors"

	"github.com/Netflix/chaosmonkey"
	"github.com/Netflix/chaosmonkey/deploy"
	"github.com/Netflix/chaosmonkey/deps"
	"github.com/Netflix/chaosmonkey/eligible"
	"github.com/Netflix/chaosmonkey/grp"
)

type leashedKiller struct {
}

func (l leashedKiller) Execute(trm chaosmonkey.Termination) error {
	log.Printf("leashed=true, not killing instance %s", trm.Instance.ID())
	return nil
}

// UnleashedInTestEnv is an error returned by Terminate if running unleashed in
// the test environment, which is not allowed
type UnleashedInTestEnv struct{}

func (err UnleashedInTestEnv) Error() string {
	return "not terminating: Chaos Monkey may not run unleashed in the test environment"
}

// Terminate executes the "terminate" command. This selects an instance
// based on the app, account, region, stack, cluster passed
//
// region, stack, and cluster may be blank
func Terminate(d deps.Deps, app string, account string, region string, stack string, cluster string) error {
	enabled, err := d.MonkeyCfg.Enabled()
	if err != nil {
		return errors.Wrap(err, "not terminating: could not determine if monkey is enabled")
	}

	if !enabled {
		log.Println("not terminating: enabled=false")
		return nil
	}

	problem, err := d.Ou.Outage()

	// If the check for ongoing outage fails, we err on the safe side nd don't terminate an instance
	if err != nil {
		return errors.Wrapf(err, "not terminating: problem checking if there is an outage")
	}

	if problem {
		log.Println("not terminating: outage in progress")
		return nil
	}

	accountEnabled, err := d.MonkeyCfg.AccountEnabled(account)

	if err != nil {
		return errors.Wrap(err, "not terminating: could not determine if account is enabled")
	}

	if !accountEnabled {
		log.Printf("Not terminating: account=%s is not enabled in Chaos Monkey", account)
		return nil
	}

	// create an instance group from the command-line parameters
	group := grp.New(app, account, region, stack, cluster)

	// do the actual termination
	return doTerminate(d, group)

}

// doTerminate does the actual termination
func doTerminate(d deps.Deps, group grp.InstanceGroup) error {
	leashed, err := d.MonkeyCfg.Leashed()

	if err != nil {
		return errors.Wrap(err, "not terminating: could not determine leashed status")
	}

	/*
		Do not allow running unleashed in the test environment.

		The prod deployment of chaos monkey is responsible for killing instances
		across environments, including test. We want to ensure that Chaos Monkey
		running in test cannot do harm.
	*/
	if d.Env.InTest() && !leashed {
		return UnleashedInTestEnv{}
	}

	var killer chaosmonkey.Terminator

	if leashed {
		killer = leashedKiller{}
	} else {
		killer = d.T
	}

	// get Chaos Monkey config info for this app
	appName := group.App()
	appCfg, err := d.ConfGetter.Get(appName)

	if err != nil {
		return errors.Wrapf(err, "not terminating: Could not retrieve config for app=%s", appName)
	}

	if !appCfg.Enabled {
		log.Printf("not terminating: enabled=false for app=%s", appName)
		return nil
	}

	if appCfg.Whitelist != nil {
		log.Printf("not terminating: app=%s has a whitelist which is no longer supported", appName)
		return nil
	}

	instance, ok := PickRandomInstance(group, *appCfg, d.Dep)
	if !ok {
		log.Printf("No eligible instances in group, nothing to terminate: %+v", group)
		return nil
	}

	log.Printf("Picked: %s", instance)

	loc, err := d.MonkeyCfg.Location()
	if err != nil {
		return errors.Wrap(err, "not terminating: could not retrieve location")
	}

	trm := chaosmonkey.Termination{Instance: instance, Time: d.Cl.Now(), Leashed: leashed}

	//
	// Check that we don't violate min time between terminations
	//
	err = d.Checker.Check(trm, *appCfg, d.MonkeyCfg.EndHour(), loc)
	if err != nil {
		return errors.Wrap(err, "not terminating: check for min time between terminations failed")
	}

	//
	// Record the termination with configured trackers
	//
	for _, tracker := range d.Trackers {
		err = tracker.Track(trm)
		if err != nil {
			return errors.Wrap(err, "not terminating: recording termination event failed")
		}
	}

	//
	// Actual instance termination happens here
	//
	err = killer.Execute(trm)
	if err != nil {
		return errors.Wrap(err, "termination failed")
	}

	return nil
}

// PickRandomInstance randomly selects an eligible instance from a group
func PickRandomInstance(group grp.InstanceGroup, cfg chaosmonkey.AppConfig, dep deploy.Deployment) (chaosmonkey.Instance, bool) {
	instances, err := eligible.Instances(group, cfg.Exceptions, dep)
	if err != nil {
		log.Printf("WARNING: eligible.Instances failed for %s: %v", group, err)
		return nil, false
	}
	if len(instances) == 0 {
		return nil, false
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	index := r.Intn(len(instances))
	return instances[index], true
}
