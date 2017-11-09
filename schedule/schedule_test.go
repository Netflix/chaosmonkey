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

package schedule_test

import (
	"bytes"
	"testing"

	"github.com/Netflix/chaosmonkey"
	"github.com/Netflix/chaosmonkey/config"
	"github.com/Netflix/chaosmonkey/config/param"
	"github.com/Netflix/chaosmonkey/mock"
	"github.com/Netflix/chaosmonkey/schedule"
)

func TestPopulate(t *testing.T) {
	// Setup
	s := schedule.New()
	// mock deployment returns 4 single-cluster apps, 3 in prod and one in test
	d := mock.Dep()

	// mockConfigGetter configures each app for App-level grouping
	getter := new(mockConfigGetter)

	cfg := config.Defaults()
	cfg.Set(param.ScheduleEnabled, true)

	// Code under test
	err := s.Populate(d, getter, cfg, nil)

	if err != nil {
		t.Fatalf("%v", err)
	}

	// Assertions
	expectedCount := 4

	dontCare := "dontcare"
	actualCount := countEntries(s.Crontab(dontCare, dontCare))

	if actualCount != expectedCount {
		t.Errorf("\nExpected:\n%d\nActual:\n%d", expectedCount, actualCount)
	}

}

// mockConfigGetter implements chaosmonkey.Getter
// returns configs for apps
type mockConfigGetter struct {
}

// Get implements chaosmonkey.Getter.Get
// Configures each app for app-level grouping
// configures mean time between work days to 1, which ensures
// a kill on each day
func (g mockConfigGetter) Get(app string) (*chaosmonkey.AppConfig, error) {
	cfg := chaosmonkey.NewAppConfig(nil)
	cfg.Grouping = chaosmonkey.App
	cfg.MeanTimeBetweenKillsInWorkDays = 1
	return &cfg, nil
}

// countEntries counts the number of entries in a cron file's contents
func countEntries(buf []byte) int {
	return bytes.Count(buf, []byte("\n"))
}
