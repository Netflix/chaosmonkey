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
	"bytes"
	"io/ioutil"
	"testing"
	"time"

	"github.com/Netflix/chaosmonkey"
	"github.com/Netflix/chaosmonkey/config"
	"github.com/Netflix/chaosmonkey/config/param"
	"github.com/Netflix/chaosmonkey/constrainer"
	"github.com/Netflix/chaosmonkey/mock"
	"github.com/Netflix/chaosmonkey/schedule"
)

// TestSchedule verifies the schedule command generates a cron file with
// the appropriate number of entries
func TestScheduleCommand(t *testing.T) {

	// Setup
	cronFile := "/tmp/chaoscron"
	err := EnsureFileAbsent(cronFile)
	if err != nil {
		t.Fatal(err)
	}

	d := mock.Dep() // mock that returns four apps
	a := new(mockAPI)
	cfg := config.Defaults()
	cfg.Set(param.Enabled, true)
	cfg.Set(param.CronPath, cronFile)
	cfg.Set(param.Accounts, []string{"prod", "test"})

	// Code under test
	appNames, err := d.AppNames()
	if err != nil {
		t.Fatalf("%v", err)
	}

	err = do(d, a, a, cfg, constrainer.NullConstrainer{}, appNames)

	if err != nil {
		t.Errorf("%v", err)
	}

	// Assertions
	expectedCount := 4

	cronFileContents, err := ioutil.ReadFile(cronFile)
	if err != nil {
		t.Fatal(err)
	}

	actualCount := countEntries(cronFileContents)

	if actualCount != expectedCount {
		t.Errorf("\nExpected:\n%d\nActual:\n%d", expectedCount, actualCount)
	}

}

// countEntries counts the number of entries in a cron file's contents
func countEntries(buf []byte) int {
	return bytes.Count(buf, []byte("\n"))
}

// mockAPI acts as a fake implementation of ChaosMonkeyAPI
type mockAPI struct {
}

// Publish implements ChaosMonkeyAPI.Publish
func (a mockAPI) Publish(date time.Time, sched *schedule.Schedule) error {
	return nil
}

func (a mockAPI) Retrieve(date time.Time) (*schedule.Schedule, error) {
	return nil, nil
}

// Get implements chaosmonkey.Getter.Get
func (a mockAPI) Get(name string) (*chaosmonkey.AppConfig, error) {
	cfg := chaosmonkey.NewAppConfig(nil)
	cfg.MeanTimeBetweenKillsInWorkDays = 1
	return &cfg, nil
}

// Check implements api.Checker.Check
func (a mockAPI) Check(term chaosmonkey.Termination, appCfg *chaosmonkey.AppConfig, endHour int, loc *time.Location) (bool, error) {
	return true, nil
}
