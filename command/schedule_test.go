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

package command

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/Netflix/chaosmonkey/v2/config"
	"github.com/Netflix/chaosmonkey/v2/config/param"
	"github.com/Netflix/chaosmonkey/v2/grp"
	"github.com/Netflix/chaosmonkey/v2/schedule"
)

// addToSchedule schedules instanceId for termination at timeString
// where timeString is  formatted in RFC3339 format
func addToSchedule(t *testing.T, sched *schedule.Schedule, timeString string, group grp.InstanceGroup) {
	tm, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		t.Fatal("Could not parse time string:", tm, err.Error())
	}

	sched.Add(tm, group)
}

func newClusterGroup(app, account, cluster, region string) grp.InstanceGroup {
	return grp.New(app, account, region, "", cluster)
}

func TestRegisterWithCron(t *testing.T) {

	// setup

	// Ensure the file isn't there from a previous run
	fname := "/tmp/chaoscron"
	err := EnsureFileAbsent(fname)

	if err != nil {
		t.Error(err.Error())
		return
	}

	config := config.Defaults()
	config.Set(param.Enabled, true)
	config.Set(param.CronPath, fname)
	config.Set(param.Accounts, []string{"prod"})

	sched := schedule.New()

	// Thu Oct 1, 2015 10:15 AM PDT -> 17:15 UTC (7 hours)
	addToSchedule(t, sched, "2015-10-01T10:15:00-07:00", newClusterGroup("abc", "prod", "abc-prod", "us-east-1"))

	// Thu Oct 1, 2015 11:23 AM PDT -> 18:23 UTC (7 hours)
	addToSchedule(t, sched, "2015-10-01T11:23:00-07:00", newClusterGroup("abc", "prod", "abc-prod", "us-west-2"))

	// code under test
	err = registerWithCron(sched, config)

	if err != nil {
		t.Fatal(err.Error())
	}

	// assertions
	dat, err := ioutil.ReadFile(fname)
	if err != nil {
		t.Error(err.Error())
		return
	}

	actual := string(dat)
	expected := `15 17 1 10 4 root /apps/chaosmonkey/chaosmonkey-terminate.sh abc prod --cluster=abc-prod --region=us-east-1
23 18 1 10 4 root /apps/chaosmonkey/chaosmonkey-terminate.sh abc prod --cluster=abc-prod --region=us-west-2
`
	if actual != expected {
		t.Errorf("\nExpected:\n%s\nActual:\n%s", expected, actual)
	}
}

// same as TestRegisterWithCron, but reverses the order that
// things are added to schedule.
func TestCronOutputInSortedOrder(t *testing.T) {
	// setup

	// Ensure the file isn't there from a previous run
	fname := "/tmp/chaoscron"
	err := EnsureFileAbsent(fname)

	if err != nil {
		t.Fatal(err.Error())
	}

	config := config.Defaults()
	config.Set(param.Enabled, true)
	config.Set(param.CronPath, fname)
	config.Set(param.Accounts, []string{"prod"})

	schedule := schedule.New()

	// Thu Oct 1, 2015 11:23 AM PDT -> 18:23 UTC (7 hours)
	addToSchedule(t, schedule, "2015-10-01T11:23:00-07:00", newClusterGroup("abc", "prod", "abc-prod", "us-east-1"))

	// Thu Oct 1, 2015 10:15 AM PDT -> 17:15 UTC (7 hours)
	addToSchedule(t, schedule, "2015-10-01T10:15:00-07:00", newClusterGroup("abc", "prod", "abc-prod", "us-west-2"))

	// code under test
	err = registerWithCron(schedule, config)

	if err != nil {
		t.Fatal(err.Error())
	}

	// assertions
	dat, err := ioutil.ReadFile(fname)
	if err != nil {
		t.Error(err.Error())
		return
	}

	actual := string(dat)
	expected := `15 17 1 10 4 root /apps/chaosmonkey/chaosmonkey-terminate.sh abc prod --cluster=abc-prod --region=us-west-2
23 18 1 10 4 root /apps/chaosmonkey/chaosmonkey-terminate.sh abc prod --cluster=abc-prod --region=us-east-1
`
	if actual != expected {
		t.Errorf("\nExpected:\n%s\nActual:\n%s", expected, actual)
	}
}
