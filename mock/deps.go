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

import (
	"io/ioutil"
	"time"

	"github.com/Netflix/chaosmonkey"
	"github.com/Netflix/chaosmonkey/clock"
	"github.com/Netflix/chaosmonkey/config"
	"github.com/Netflix/chaosmonkey/config/param"
	"github.com/Netflix/chaosmonkey/deps"
)

type (
	// Checker implements deps.Checker
	Checker struct {
		Error error
	}

	// Tracker implements chaosmonkey.Tracker
	Tracker struct {
		Error error
	}

	// ErrorCounter implements chaosmonkey.Publisher
	ErrorCounter struct{}

	// Clock implements clock.Clock
	Clock struct {
		Time time.Time
	}

	// Env implements chaosmonkey.Env
	Env struct {
		IsInTest bool
	}
)

// Check implements deps.Checker.Check
func (c Checker) Check(term chaosmonkey.Termination, appCfg chaosmonkey.AppConfig, endHour int, loc *time.Location) error {
	return c.Error

}

// Track implements chaosmonkey.Tracker.Track
func (t Tracker) Track(trm chaosmonkey.Termination) error {
	return t.Error
}

// Increment implements chaosmonkey.ErrorCounter.Increment
func (e ErrorCounter) Increment() error {
	return nil
}

// Now implements clock.Clock.Now
func (c Clock) Now() time.Time {
	return c.Time
}

// InTest implements chaosmonkey.Env.InTest
func (e Env) InTest() bool {
	return e.IsInTest
}

// Deps returns a deps.Deps object that contains mocks.
// The mocks implement their interfaces by performing no-ops.
func Deps() deps.Deps {
	cfg := config.Defaults()
	cfg.Set(param.Enabled, true)
	cfg.Set(param.Leashed, false)
	cfg.Set(param.Accounts, []string{"prod", "test"})

	f, err := ioutil.TempFile("", "cm-test")
	if err != nil {
		panic(err)
	}

	// The ioutil.TempFile opens the file, but we
	// don't need it open, since we are just going
	// to pass the file name along via the CronPath
	// function, so we just close it
	err = f.Close()
	if err != nil {
		panic(err)
	}

	cfg.Set(param.CronPath, f.Name())

	return deps.Deps{
		MonkeyCfg:  cfg,
		Checker:    Checker{Error: nil},
		ConfGetter: DefaultConfigGetter(),
		Cl:         clock.New(),
		Dep:        Dep(),
		T:          new(Terminator),
		Ou:         Outage{},
		ErrCounter: ErrorCounter{},
		Env:        Env{false},
	}
}
