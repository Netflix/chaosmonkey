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

//go:build docker
// +build docker

// The tests in this package use docker to test against a mysql:5.6 database
// By default, the tests are off unless you pass the "-tags docker" flag
// when running the test.

package mysql_test

import (
	"testing"
	"time"

	c "github.com/Netflix/chaosmonkey/v2"
	"github.com/Netflix/chaosmonkey/v2/mock"
	"github.com/Netflix/chaosmonkey/v2/mysql"
)

var endHour = 15 // 3PM

// testSetup returns some values useful for test setup
func testSetup(t *testing.T) (ins c.Instance, loc *time.Location, appCfg c.AppConfig) {

	ins = mock.Instance{
		App:        "myapp",
		Account:    "prod",
		Stack:      "mystack",
		Cluster:    "mycluster",
		Region:     "us-east-1",
		ASG:        "myapp-mystack-mycluster-V123",
		InstanceID: "i-a96a0166",
	}

	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatalf(err.Error())
	}

	appCfg = c.AppConfig{
		Enabled:                        true,
		RegionsAreIndependent:          true,
		MeanTimeBetweenKillsInWorkDays: 5,
		MinTimeBetweenKillsInWorkDays:  1,
		Grouping:                       c.Cluster,
		Exceptions:                     nil,
	}

	return

}

// TestCheckPermitted verifies check succeeds when no previous terminations in database
func TestCheckPermitted(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	m, err := mysql.New("localhost", port, "root", password, "chaosmonkey")
	if err != nil {
		t.Fatal(err)
	}

	ins, loc, appCfg := testSetup(t)

	trm := c.Termination{Instance: ins, Time: time.Now(), Leashed: false}

	err = m.Check(trm, appCfg, endHour, loc)

	if err != nil {
		t.Fatal(err)
	}
}

// TestCheckPermitted verifies check fails if commit is too recent
func TestCheckForbidden(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	m, err := mysql.New("localhost", port, "root", password, "chaosmonkey")
	if err != nil {
		t.Fatal(err)
	}

	ins, loc, appCfg := testSetup(t)

	trm := c.Termination{Instance: ins, Time: time.Now(), Leashed: false}

	// First check should succeed
	err = m.Check(trm, appCfg, endHour, loc)

	if err != nil {
		t.Fatal(err)
	}

	// Second check should fail
	err = m.Check(trm, appCfg, endHour, loc)
	if err == nil {
		t.Fatal("Check() succeeded when it should have failed")
	}

	if _, ok := err.(c.ErrViolatesMinTime); !ok {
		t.Fatalf("Expected Err.ViolatesMinTime, got %v", err)
	}
}

// When we are going to commit an unleashed termination, we only care
// about unleashed previous terminations
func TestCheckLeashed(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	m, err := mysql.New("localhost", port, "root", password, "chaosmonkey")
	if err != nil {
		t.Fatal(err)
	}

	ins, loc, appCfg := testSetup(t)

	trm := c.Termination{Instance: ins, Time: time.Now(), Leashed: true}

	// First check should succeed
	err = m.Check(trm, appCfg, endHour, loc)

	if err != nil {
		t.Fatal(err)
	}

	trm = c.Termination{Instance: ins, Time: time.Now(), Leashed: false}

	// Second check should fail
	err = m.Check(trm, appCfg, endHour, loc)

	if err != nil {
		t.Fatalf("Should have allowed an unleashed termination after leashed: %v", err)
	}
}

// Check that only termination is permitted on concurrent attempts
func TestConcurrentChecks(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	m, err := mysql.New("localhost", port, "root", password, "chaosmonkey")
	if err != nil {
		t.Fatal(err)
	}

	ins, loc, appCfg := testSetup(t)

	trm := c.Termination{Instance: ins, Time: time.Now()}

	// Try to check twice. At least one should return an error
	ch := make(chan error, 2)

	go func() {
		// We use the "MySQL.CheckWithDelay" method which adds a delay between reading
		// from the database and writing to it, to increase the likelihood that
		// the two requests overlap
		ch <- m.CheckWithDelay(trm, appCfg, endHour, loc, 1*time.Second)
	}()

	go func() {
		ch <- m.Check(trm, appCfg, endHour, loc)
	}()

	var success int
	var txDeadlock int
	var violatesMinTime int
	for i := 0; i < 2; i++ {
		err := <-ch
		switch {
		case err == nil:
			success++
		case mysql.TxDeadlock(err):
			txDeadlock++
		case mysql.ViolatesMinTime(err):
			violatesMinTime++
		default:
			t.Fatalf("Unexpected error: %+v", err)
		}
	}

	if got, want := success, 1; got != want {
		t.Errorf("got %d succeses, want: %d", got, want)
	}
}

func TestCombinations(t *testing.T) {

	// Reference instance
	ins := mock.Instance{
		App:        "myapp",
		Account:    "prod",
		Stack:      "mystack",
		Cluster:    "mycluster",
		Region:     "us-east-1",
		ASG:        "myapp-mystack-mycluster-V123",
		InstanceID: "i-a96a0166",
	}

	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatalf(err.Error())
	}

	tests := []struct {
		desc    string
		grp     c.Group
		reg     bool // regions are independent
		ins     c.Instance
		allowed bool // true if we can kill this instance after previous
	}{
		{"same cluster, should fail", c.Cluster, true, mock.Instance{App: "myapp", Account: "prod", Stack: "mystack", Cluster: "mycluster", Region: "us-east-1", ASG: "myapp-mystack-mycluster-V123"}, false},

		{"different cluster, should succeed", c.Cluster, true, mock.Instance{App: "myapp", Account: "prod", Stack: "mystack", Cluster: "othercluster", Region: "us-east-1", ASG: "myapp-mystack-mycluster-V123"}, true},

		{"same stack should fail", c.Stack, true, mock.Instance{App: "myapp", Account: "prod", Stack: "mystack", Cluster: "othercluster", Region: "us-east-1", ASG: "myapp-mystack-mycluster-V123"}, false},

		{"different stack, should succeed", c.Stack, true, mock.Instance{App: "myapp", Account: "prod", Stack: "otherstack", Cluster: "othercluster", Region: "us-east-1", ASG: "myapp-otherstack-mycluster-V123"}, true},

		{"same app, should fail", c.App, true, mock.Instance{App: "myapp", Account: "prod", Stack: "mystack", Cluster: "othercluster", Region: "us-east-1", ASG: "myapp-mystack-mycluster-V123"}, false},

		{"different region, should succeed", c.Cluster, true, mock.Instance{App: "myapp", Account: "prod", Stack: "mystack", Cluster: "mycluster", Region: "us-west-2", ASG: "myapp-mystack-mycluster-V123"}, true},

		{"different region where regions are not independent, should fail", c.Cluster, false, mock.Instance{App: "myapp", Account: "prod", Stack: "mystack", Cluster: "mycluster", Region: "us-west-2", ASG: "myapp-mystack-mycluster-V123"}, false},
	}

	for _, tt := range tests {

		err := initDB()
		if err != nil {
			t.Fatal(err)
		}

		m, err := mysql.New("localhost", port, "root", password, "chaosmonkey")
		if err != nil {
			t.Fatal(err)
		}
		cfg := c.AppConfig{
			Enabled:                        true,
			RegionsAreIndependent:          tt.reg,
			MeanTimeBetweenKillsInWorkDays: 1,
			MinTimeBetweenKillsInWorkDays:  1,
			Grouping:                       tt.grp,
		}

		err = m.Check(c.Termination{Instance: ins, Time: time.Now()}, cfg, endHour, loc)

		if err != nil {
			t.Fatal(err)
		}

		term := c.Termination{Instance: tt.ins, Time: time.Now()}

		err = m.Check(term, cfg, endHour, loc)
		if tt.allowed && err != nil {
			t.Errorf("%s: got m.Check(%#v, %#v) = %+v, expected nil", tt.desc, term, cfg, err)
		}

		if !tt.allowed && err == nil {
			t.Errorf("%s: get m.Check(%#v, %#v) = nil, expected error", tt.desc, term, cfg)
		}

	}
}

func TestCheckMinTimeEnforced(t *testing.T) {

	cfg := c.AppConfig{
		Enabled:                        true,
		RegionsAreIndependent:          true,
		MeanTimeBetweenKillsInWorkDays: 5,
		MinTimeBetweenKillsInWorkDays:  2,
		Grouping:                       c.Cluster,
	}

	// The current kill time
	now := "Thu Dec 17 11:35:00 2015 -0800"

	// Since MinTimeBetweenKillsInWorkDays is 1 here, then the most recent
	// kill permitted is the day before at endHour
	endHour := 15

	// Tue Dec 15 15:00:00 2015 -0800

	// Any kills later than that time will not be permitted
	// Boundary value testing!

	// this is a magic date used by go for parsing strings
	refDate := "Mon Jan  2 15:04:05 2006 -0700"
	tnow, err := time.Parse(refDate, now)
	if err != nil {
		t.Fatal(err)
	}

	ins := mock.Instance{
		App:        "myapp",
		Account:    "prod",
		Stack:      "mystack",
		Cluster:    "mycluster",
		Region:     "us-east-1",
		ASG:        "myapp-mystack-mycluster-V123",
		InstanceID: "i-a96a0166",
	}

	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		last    string
		allowed bool
	}{
		{"Tue Dec 15 15:01:00 2015 -0800", false},
		{"Tue Dec 15 14:59:59 2015 -0800", true},
	}

	for _, tt := range tests {

		err := initDB()
		if err != nil {
			t.Fatal(err)
		}

		m, err := mysql.New("localhost", port, "root", password, "chaosmonkey")
		if err != nil {
			t.Fatal(err)
		}

		//
		// Write the initial termination
		//

		last, err := time.Parse("Mon Jan  2 15:04:05 2006 -0700", tt.last)
		if err != nil {
			t.Fatal(err)
		}
		err = m.Check(c.Termination{Instance: ins, Time: last}, cfg, endHour, loc)
		if err != nil {
			t.Fatalf("Failed to write the initial termination, should always succeed: %v", err)
		}

		//
		// Write today's termination
		//

		err = m.Check(c.Termination{Instance: ins, Time: tnow}, cfg, endHour, loc)

		switch err.(type) {
		case nil:
			if !tt.allowed {
				t.Fatalf("%s termination should have been forbidden, was allowed", tt.last)
			}
		case c.ErrViolatesMinTime:
			if tt.allowed {
				t.Errorf("%s termination should have been allowed, got: %v", tt.last, err)
			}
		default:
			t.Errorf("%s termination returned unexpected err: %v", tt.last, err)
		}
	}
}
