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

// +build docker

// The tests in this package use docker to test against a mysql:5.6 database
// By default, the tests are off unless you pass the "-tags docker" flag
// when running the test.

package mysql_test

import (
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/Netflix/chaosmonkey/grp"
	"github.com/Netflix/chaosmonkey/mysql"
	"github.com/Netflix/chaosmonkey/schedstore"
	"github.com/Netflix/chaosmonkey/schedule"
)

// Test we can publish and then retrieve a schedule
func TestPublishRetrieve(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	m, err := mysql.New("localhost", port, "root", password, "chaosmonkey")
	if err != nil {
		t.Fatal(err)
	}

	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}

	sched := schedule.New()

	t1 := time.Date(2016, time.June, 20, 11, 40, 0, 0, loc)
	sched.Add(t1, grp.New("chaosguineapig", "test", "us-east-1", "", "chaosguineapig-test"))

	date := time.Date(2016, time.June, 20, 0, 0, 0, 0, loc)

	// Code under test:
	err = m.Publish(date, sched)
	if err != nil {
		t.Fatal(err)
	}
	sched, err = m.Retrieve(date)
	if err != nil {
		t.Fatal(err)
	}

	entries := sched.Entries()
	if got, want := len(entries), 1; got != want {
		t.Fatalf("got len(entries)=%d, want %d", got, want)
	}

	entry := entries[0]

	if !t1.Equal(entry.Time) {
		t.Errorf("%s != %s", t1, entry.Time)
	}
}

func NewMySQL() (mysql.MySQL, error) {
	return mysql.New("localhost", port, "root", password, "chaosmonkey")
}

func TestPublishRetrieveMultipleEntries(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	m, err := NewMySQL()
	if err != nil {
		t.Fatal(err)
	}

	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}

	psched := schedule.New()

	pEntries := []schedule.Entry{
		{Time: time.Date(2016, time.June, 20, 11, 40, 0, 0, loc), Group: grp.New("doesnotexist", "test", "us-east-1", "", "doesnotexist-foo-bar")},
		{Time: time.Date(2016, time.June, 20, 12, 35, 0, 0, loc), Group: grp.New("foobar", "other", "us-west-2", "", "foobar-baz-quux")},
		{Time: time.Date(2016, time.June, 20, 9, 7, 0, 0, loc), Group: grp.New("chaosguineapig", "prod", "us-east-1", "", "chaosguineapig-prod")},
	}

	for _, v := range pEntries {
		psched.Add(v.Time, v.Group)
	}

	date := time.Date(2016, time.June, 20, 0, 0, 0, 0, loc)

	// Code under test:
	err = m.Publish(date, psched)
	if err != nil {
		t.Fatal(err)
	}
	rsched, err := m.Retrieve(date)
	if err != nil {
		t.Fatal(err)
	}

	rEntries := rsched.Entries()
	if got, want := len(rEntries), len(pEntries); got != want {
		t.Fatalf("got len(entries)=%d, want %d", got, want)
	}

	for i := range pEntries {
		if got, want := rEntries[i], pEntries[i]; !got.Equal(&want) {
			t.Errorf("got entry[%d]=%v, want %v", i, got, want)
		}
	}
}

func TestScheduleAlreadyExists(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	m, err := NewMySQL()
	if err != nil {
		t.Fatal(err)
	}

	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}

	psched1 := schedule.New()
	psched1.Add(time.Date(2016, time.June, 20, 11, 40, 0, 0, loc), grp.New("imaginaryproject", "test", "us-west-2", "", ""))

	date := time.Date(2016, time.June, 20, 0, 0, 0, 0, loc)

	// Create an initial schedule with one entry
	err = m.Publish(date, psched1)
	if err != nil {
		t.Fatal(err)
	}

	// Try to publish a new schedule

	pEntries := []schedule.Entry{
		{Time: time.Date(2016, time.June, 20, 11, 40, 0, 0, loc), Group: grp.New("doesnotexist", "test", "us-east-1", "", "doesnotexist-foo-bar")},
		{Time: time.Date(2016, time.June, 20, 12, 35, 0, 0, loc), Group: grp.New("foobar", "other", "us-west-2", "", "foobar-baz-quux")},
		{Time: time.Date(2016, time.June, 20, 9, 7, 0, 0, loc), Group: grp.New("chaosguineapig", "prod", "us-east-1", "", "chaosguineapig-prod")},
	}

	psched2 := schedule.New()
	for _, v := range pEntries {
		psched2.Add(v.Time, v.Group)
	}

	err = m.Publish(date, psched2)

	// This should return an error
	if got, want := err, schedstore.ErrAlreadyExists; got != want {
		t.Fatalf(`got m.Publish()="%v" want "%v"`, got, want)
	}
}

func TestScheduleAlreadyExistsConcurrency(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	m, err := NewMySQL()
	if err != nil {
		t.Fatal(err)
	}

	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}

	psched1 := schedule.New()
	psched1.Add(time.Date(2016, time.June, 20, 11, 40, 0, 0, loc), grp.New("imaginaryproject", "test", "us-west-2", "", ""))

	pEntries := []schedule.Entry{
		{Time: time.Date(2016, time.June, 20, 11, 40, 0, 0, loc), Group: grp.New("doesnotexist", "test", "us-east-1", "", "doesnotexist-foo-bar")},
		{Time: time.Date(2016, time.June, 20, 12, 35, 0, 0, loc), Group: grp.New("foobar", "other", "us-west-2", "", "foobar-baz-quux")},
		{Time: time.Date(2016, time.June, 20, 9, 7, 0, 0, loc), Group: grp.New("chaosguineapig", "prod", "us-east-1", "", "chaosguineapig-prod")},
	}

	psched2 := schedule.New()
	for _, v := range pEntries {
		psched2.Add(v.Time, v.Group)
	}

	// Try to publish the schedule twice. At least one schedule should return an
	// error
	ch := make(chan error, 2)

	date := time.Date(2016, time.June, 20, 0, 0, 0, 0, loc)

	go func() {
		ch <- m.PublishWithDelay(date, psched1, 3*time.Second)
	}()

	go func() {
		ch <- m.PublishWithDelay(date, psched2, 0)
	}()

	// Retrieve the two error values from the two calls

	var success int
	var txDeadlock int
	for i := 0; i < 2; i++ {
		err := <-ch
		switch {
		case err == nil:
			success++
		case mysql.TxDeadlock(err):
			txDeadlock++
		default:
			t.Fatalf("Unexpected error: %+v", err)
		}
	}

	if got, want := success, 1; got != want {
		t.Errorf("got %d succeses, want: %d", got, want)
	}

	// Should cause a deadlock
	if got, want := txDeadlock, 1; got != want {
		t.Errorf("got %d txDeadlock, want: %d", got, want)
	}
}

func TestOnlyReturnsFromDayRequested(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	m, err := NewMySQL()
	if err != nil {
		t.Fatal(err)
	}

	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}

	// Day 1: 6/20/2016: 1 entry
	psched1 := schedule.New()
	d1 := time.Date(2016, time.June, 20, 0, 0, 0, 0, loc)
	psched1.Add(time.Date(2016, time.June, 20, 11, 40, 0, 0, loc), grp.New("imaginaryproject", "test", "us-west-2", "", ""))

	err = m.Publish(d1, psched1)
	if err != nil {
		t.Fatal(err)
	}

	// Day 2: 6/21/2016: 3 entries
	psched2 := schedule.New()
	d2 := time.Date(2016, time.June, 21, 0, 0, 0, 0, loc)
	pEntries := []schedule.Entry{
		{Time: time.Date(2016, time.June, 21, 11, 40, 0, 0, loc), Group: grp.New("doesnotexist", "test", "us-east-1", "", "doesnotexist-foo-bar")},
		{Time: time.Date(2016, time.June, 21, 12, 35, 0, 0, loc), Group: grp.New("foobar", "other", "us-west-2", "", "foobar-baz-quux")},
		{Time: time.Date(2016, time.June, 21, 9, 7, 0, 0, loc), Group: grp.New("chaosguineapig", "prod", "us-east-1", "", "chaosguineapig-prod")},
	}

	for _, v := range pEntries {
		psched2.Add(v.Time, v.Group)
	}

	m.Publish(d2, psched2)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		date    time.Time
		entries int
	}{
		{d1, 1},
		{d2, 3},
	}

	for _, tt := range tests {
		sched, err := m.Retrieve(tt.date)
		if err != nil {
			t.Fatal(err)
		}

		if got, want := len(sched.Entries()), tt.entries; got != want {
			t.Fatalf("got len(entries)=%d, want %d", got, want)
		}
	}
}

func TestNoScheduleRetrievedOnWrongDay(t *testing.T) {
	err := initDB()
	if err != nil {
		t.Fatal(err)
	}

	m, err := NewMySQL()
	if err != nil {
		t.Fatal(err)
	}

	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}

	// Day 1: 6/20/2016: 1 entry
	psched := schedule.New()
	d := time.Date(2016, time.June, 20, 0, 0, 0, 0, loc)
	psched.Add(time.Date(2016, time.June, 20, 11, 40, 0, 0, loc), grp.New("imaginaryproject", "test", "us-west-2", "", ""))

	m.Publish(d, psched)

	tests := []struct {
		date    time.Time
		entries int
	}{
		{time.Date(2016, time.June, 19, 0, 0, 0, 0, loc), 0},
		{time.Date(2016, time.June, 20, 0, 0, 0, 0, loc), 1},
		{time.Date(2016, time.June, 21, 0, 0, 0, 0, loc), 0},
	}

	for _, tt := range tests {
		sched, err := m.Retrieve(tt.date)
		if err != nil {
			t.Fatal(err)
		}

		if got, want := len(sched.Entries()), tt.entries; got != want {
			t.Fatalf("got len(entries)=%d, want %d", got, want)
		}
	}
}

func TestPublishDateDifferentTimes(t *testing.T) {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		ptime time.Time
		rtime time.Time
	}{
		{time.Date(2016, time.June, 20, 0, 0, 0, 0, loc), time.Date(2016, time.June, 20, 0, 0, 0, 0, loc)},
		{time.Date(2016, time.June, 20, 12, 0, 0, 0, loc), time.Date(2016, time.June, 20, 12, 0, 0, 0, loc)},
		{time.Date(2016, time.June, 20, 0, 0, 0, 0, loc), time.Date(2016, time.June, 20, 16, 59, 59, 0, loc)},
		{time.Date(2016, time.June, 20, 0, 0, 0, 0, loc), time.Date(2016, time.June, 20, 17, 0, 0, 0, loc)}, // UTC boundary
		{time.Date(2016, time.June, 20, 0, 0, 0, 0, loc), time.Date(2016, time.June, 20, 17, 0, 1, 0, loc)},
		{time.Date(2016, time.June, 20, 0, 0, 0, 0, loc), time.Date(2016, time.June, 20, 23, 59, 59, 0, loc)},
		{time.Date(2016, time.June, 20, 23, 59, 59, 0, loc), time.Date(2016, time.June, 20, 23, 59, 59, 0, loc)},
		{time.Date(2016, time.June, 20, 0, 0, 0, 0, loc), time.Date(2016, time.June, 20, 23, 59, 59, 0, loc)},
		{time.Date(2016, time.June, 20, 23, 59, 59, 0, loc), time.Date(2016, time.June, 20, 0, 0, 0, 0, loc)},
	}

	for _, tt := range tests {
		err := initDB()
		if err != nil {
			t.Fatal(err)
		}

		m, err := NewMySQL()
		if err != nil {
			t.Fatal(err)
		}

		psched := schedule.New()
		psched.Add(time.Date(2016, time.June, 20, 11, 40, 0, 0, loc), grp.New("imaginaryproject", "test", "us-west-2", "", ""))

		m.Publish(tt.ptime, psched)

		sched, err := m.Retrieve(tt.rtime)
		if err != nil {
			t.Fatal(err)
		}

		if got, want := len(sched.Entries()), 1; got != want {
			t.Fatalf("publish date:%v, retrieve date:%v, got len(entries)=%d, want %d", tt.ptime, tt.rtime, got, want)
		}

	}

}
