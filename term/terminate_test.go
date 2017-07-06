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
	"errors"
	"testing"
	"time"

	"github.com/Netflix/chaosmonkey"
	"github.com/Netflix/chaosmonkey/clock"
	"github.com/Netflix/chaosmonkey/config"
	"github.com/Netflix/chaosmonkey/config/param"
	"github.com/Netflix/chaosmonkey/deps"
	"github.com/Netflix/chaosmonkey/mock"
)

func mockDeps() deps.Deps {
	monkeyCfg := config.Defaults()
	monkeyCfg.Set(param.Enabled, true)
	monkeyCfg.Set(param.Leashed, false)
	monkeyCfg.Set(param.Accounts, []string{"prod"})
	recorder := mock.Checker{Error: nil}
	confGetter := mock.DefaultConfigGetter()
	cl := clock.New()
	dep := mock.Dep()
	ttor := mock.Terminator{}
	ou := mock.Outage{}
	env := mock.Env{IsInTest: false}
	return deps.Deps{MonkeyCfg: monkeyCfg, Checker: recorder, ConfGetter: confGetter, Cl: cl, Dep: dep, T: &ttor, Ou: ou, Env: env}
}

// TestTerminateKills ensure the terminator actually gets invoked
func TestTerminateKills(t *testing.T) {

	deps := mockDeps()
	err := Terminate(deps, "foo", "prod", "us-east-1", "", "foo-prod")

	if err != nil {
		t.Fatal(err)
	}

	ttor := deps.T.(*mock.Terminator)
	ins := ttor.Instance

	if got, want := ttor.Ncalls, 1; got != want {
		t.Fatalf("Expected terminator to be called once, got ttor.Ncalls=%d", ttor.Ncalls)
	}

	if got, want := ins.AppName(), "foo"; got != want {
		t.Errorf("Expected ins.AppName()=%s. want %s", got, want)
	}

	if got, want := ins.AccountName(), "prod"; got != want {
		t.Errorf("Expected ins.AccountName()=%s. want %s", got, want)
	}

	if got, want := ins.RegionName(), "us-east-1"; got != want {
		t.Errorf("Expected ins.RegionName()=%s. want %s", got, want)
	}

	if got, want := ins.ClusterName(), "foo-prod"; got != want {
		t.Errorf("Expected ins.ClusterName()=%s. want %s", got, want)
	}
}

// TestTerminateOnlyKillsInProd ensures we don't kill in non-prod accounts
// This is temporary until we have full support for multiple accounts
func TestTerminateOnlyKillsInProd(t *testing.T) {
	deps := mockDeps()

	err := Terminate(deps, "quux", "test", "us-east-1", "", "quux-test")

	if err != nil {
		t.Fatal(err)
	}

	ttor := deps.T.(*mock.Terminator)
	if got, want := ttor.Ncalls, 0; got != want {
		t.Errorf("Expected terminator to not be called, got ttor.Ncalls=%d", ttor.Ncalls)
	}

}

func TestTerminateDoesntKillIfRecorderFails(t *testing.T) {
	deps := mockDeps()
	deps.Checker = mock.Checker{Error: chaosmonkey.ErrViolatesMinTime{InstanceID: "i-8703ada6", KilledAt: time.Now().Add(-1 * time.Hour)}}

	err := Terminate(deps, "foo", "prod", "us-east-1", "", "foo-prod")
	if err == nil {
		t.Fatal("Expected Terminate to fail, it succeeded")
	}

	ttor := deps.T.(*mock.Terminator)
	if got, want := ttor.Ncalls, 0; got != want {
		t.Errorf("Expected terminator to not be called, got ttor.Ncalls=%d", ttor.Ncalls)
	}
}

// TestTerminateDoesntKillInLeashedMode ensure terminator does not get invoked
// if leashed is enabled
func TestTerminateDoesntKillInLeashedMode(t *testing.T) {

	deps := mockDeps()
	cfg := config.Defaults()
	// Setting leashed explicitly for code clarity, default is leashed so
	// this isn't strictly neededj
	cfg.Set(param.Leashed, true)

	deps.MonkeyCfg = cfg

	err := Terminate(deps, "foo", "prod", "us-east-1", "", "foo-prod")

	if err != nil {
		t.Fatal(err)
	}

	ttor := deps.T.(*mock.Terminator)
	if got, want := ttor.Ncalls, 0; got != want {
		t.Errorf("Expected terminator to not be called, got ttor.Ncalls=%d", ttor.Ncalls)
	}

}

// TestNeverTerminateInTestEnv checks that unleasshed terms are not allowed in
// test
func TestNeverTerminateUnleashedInTestEnv(t *testing.T) {

	deps := mockDeps()
	deps.Env = mock.Env{IsInTest: true}

	err := Terminate(deps, "foo", "prod", "us-east-1", "", "foo-prod")

	if _, ok := err.(UnleashedInTestEnv); !ok {
		t.Fatalf("Expected Terminate to return an error when running unleashed in test mode")
	}

	ttor := deps.T.(*mock.Terminator)
	if got, want := ttor.Ncalls, 0; got != want {
		t.Errorf("Expected terminator to be called once, got ttor.Ncalls=%d", ttor.Ncalls)
	}

}

func TestDoesNotTerminateIfTrackerFails(t *testing.T) {
	deps := mockDeps()

	// We pass two trackers, the first one succeeds, the second returns an error
	deps.Trackers = []chaosmonkey.Tracker{
		mock.Tracker{},
		mock.Tracker{Error: errors.New("something went wrong")}}

	err := Terminate(deps, "foo", "prod", "us-east-1", "", "foo-prod")
	if err == nil {
		t.Fatal("Tracker failed but Terminate did not return an error")
	}

	ttor := deps.T.(*mock.Terminator)
	if got, want := ttor.Ncalls, 0; got != want {
		t.Errorf("Expected terminator to not be called, got ttor.Ncalls=%d", ttor.Ncalls)
	}

}

func TestDoesNotTerminateIfAppIsDisabled(t *testing.T) {
	deps := mockDeps()

	// Disable app
	deps.ConfGetter = mock.NewConfigGetter(chaosmonkey.AppConfig{
		Enabled:                        false,
		RegionsAreIndependent:          true,
		MeanTimeBetweenKillsInWorkDays: 5,
		MinTimeBetweenKillsInWorkDays:  1,
		Grouping:                       chaosmonkey.Cluster,
		Exceptions:                     nil,
	})

	err := Terminate(deps, "foo", "prod", "us-east-1", "", "foo-prod")
	if err != nil {
		t.Fatal(err)
	}

	ttor := deps.T.(*mock.Terminator)
	if got, want := ttor.Ncalls, 0; got != want {
		t.Errorf("Expected terminator to not be called, got ttor.Ncalls=%d", ttor.Ncalls)
	}
}
