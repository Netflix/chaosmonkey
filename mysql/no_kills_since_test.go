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

package mysql

// This file contains test for the config.NoKillsSince method

import (
	"testing"
	"time"
)

/*
Test scenarios to hit

- Zero days between kills
- one day
- N days
- Mid week
- Beginning of week
- Beginning of day
- End of day
- Daylight savings
- All day boundaries

*/

func TestZeroDaysBetweenKills(t *testing.T) {

	// Note: -0800 = PST
	//       -0700 = PDT
	tests := []struct {
		days  int
		now   string
		since string
	}{
		// 0 days means that kills are allowed on the same day
		{0, "Thu Dec 17 00:00:00 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},
		{0, "Thu Dec 17 00:00:01 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},
		{0, "Thu Dec 17 00:01:00 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},
		{0, "Thu Dec 17 08:59:00 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},
		{0, "Thu Dec 17 08:59:59 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},
		{0, "Thu Dec 17 09:00:00 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},
		{0, "Thu Dec 17 09:01:00 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},
		{0, "Thu Dec 17 14:18:30 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},

		// Test on the UTC boundary (midnight UTC = 4PM)
		{0, "Thu Dec 17 15:59:00 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},
		{0, "Thu Dec 17 15:59:59 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},
		{0, "Thu Dec 17 16:00:00 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},
		{0, "Thu Dec 17 16:00:01 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},
		{0, "Thu Dec 17 16:01:00 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},
		{0, "Thu Dec 17 16:59:59 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},
		{0, "Thu Dec 17 16:59:59 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},
		{0, "Thu Dec 17 15:00:00 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},
		{0, "Thu Dec 17 17:00:01 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},

		// Test on the midnight boundary
		{0, "Thu Dec 17 23:00:00 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},
		{0, "Thu Dec 17 23:59:00 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},
		{0, "Thu Dec 17 23:59:59 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},
		{0, "Fri Dec 18 00:00:00 2015 -0800", "Fri Dec 18 15:00:00 2015 -0800"},

		// Go back 1 day
		{1, "Thu Dec 17 00:00:00 2015 -0800", "Wed Dec 16 15:00:00 2015 -0800"},
		{1, "Thu Dec 17 00:00:01 2015 -0800", "Wed Dec 16 15:00:00 2015 -0800"},
		{1, "Thu Dec 17 00:01:00 2015 -0800", "Wed Dec 16 15:00:00 2015 -0800"},
		{1, "Thu Dec 17 08:59:00 2015 -0800", "Wed Dec 16 15:00:00 2015 -0800"},
		{1, "Thu Dec 17 08:59:59 2015 -0800", "Wed Dec 16 15:00:00 2015 -0800"},
		{1, "Thu Dec 17 09:00:00 2015 -0800", "Wed Dec 16 15:00:00 2015 -0800"},
		{1, "Thu Dec 17 09:01:00 2015 -0800", "Wed Dec 16 15:00:00 2015 -0800"},
		{1, "Thu Dec 17 15:00:00 2015 -0800", "Wed Dec 16 15:00:00 2015 -0800"},
		{1, "Thu Dec 17 15:00:01 2015 -0800", "Wed Dec 16 15:00:00 2015 -0800"},
		{1, "Thu Dec 17 15:18:30 2015 -0800", "Wed Dec 16 15:00:00 2015 -0800"},
		{1, "Thu Dec 17 15:59:00 2015 -0800", "Wed Dec 16 15:00:00 2015 -0800"},
		{1, "Thu Dec 17 15:59:59 2015 -0800", "Wed Dec 16 15:00:00 2015 -0800"},
		{1, "Thu Dec 17 16:00:00 2015 -0800", "Wed Dec 16 15:00:00 2015 -0800"},
		{1, "Thu Dec 17 16:00:01 2015 -0800", "Wed Dec 16 15:00:00 2015 -0800"},
		{1, "Thu Dec 17 16:01:00 2015 -0800", "Wed Dec 16 15:00:00 2015 -0800"},
		{1, "Thu Dec 17 16:59:59 2015 -0800", "Wed Dec 16 15:00:00 2015 -0800"},
		{1, "Thu Dec 17 16:59:59 2015 -0800", "Wed Dec 16 15:00:00 2015 -0800"},
		{1, "Thu Dec 17 23:00:00 2015 -0800", "Wed Dec 16 15:00:00 2015 -0800"},
		{1, "Thu Dec 17 23:59:00 2015 -0800", "Wed Dec 16 15:00:00 2015 -0800"},
		{1, "Thu Dec 17 23:59:59 2015 -0800", "Wed Dec 16 15:00:00 2015 -0800"},
		{1, "Fri Dec 18 00:00:00 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},
		{1, "Fri Dec 18 00:00:01 2015 -0800", "Thu Dec 17 15:00:00 2015 -0800"},

		// going back several days
		{1, "Thu Dec 17 15:18:30 2015 -0800", "Wed Dec 16 15:00:00 2015 -0800"},
		{2, "Thu Dec 17 15:18:30 2015 -0800", "Tue Dec 15 15:00:00 2015 -0800"},
		{3, "Thu Dec 17 15:18:30 2015 -0800", "Mon Dec 14 15:00:00 2015 -0800"},
		{4, "Thu Dec 17 15:18:30 2015 -0800", "Fri Dec 11 15:00:00 2015 -0800"},
		{5, "Thu Dec 17 15:18:30 2015 -0800", "Thu Dec 10 15:00:00 2015 -0800"},
		{6, "Thu Dec 17 15:18:30 2015 -0800", "Wed Dec  9 15:00:00 2015 -0800"},
		{7, "Thu Dec 17 15:18:30 2015 -0800", "Tue Dec  8 15:00:00 2015 -0800"},
		{8, "Thu Dec 17 15:18:30 2015 -0800", "Mon Dec  7 15:00:00 2015 -0800"},
		{9, "Thu Dec 17 15:18:30 2015 -0800", "Fri Dec  4 15:00:00 2015 -0800"},

		// beginning of week
		{2, "Mon Dec 14 00:00:00 2015 -0800", "Thu Dec 10 15:00:00 2015 -0800"},
		{2, "Mon Dec 14 08:59:59 2015 -0800", "Thu Dec 10 15:00:00 2015 -0800"},
		{2, "Mon Dec 14 09:00:00 2015 -0800", "Thu Dec 10 15:00:00 2015 -0800"},
		{2, "Mon Dec 14 09:01:00 2015 -0800", "Thu Dec 10 15:00:00 2015 -0800"},
		{2, "Mon Dec 14 15:00:00 2015 -0800", "Thu Dec 10 15:00:00 2015 -0800"},
		{2, "Mon Dec 14 15:01:00 2015 -0800", "Thu Dec 10 15:00:00 2015 -0800"},
		{2, "Mon Dec 14 23:59:59 2015 -0800", "Thu Dec 10 15:00:00 2015 -0800"},

		// daylight savings in 2016:
		// Sun Mar 13 02:00:00 2015 -0700
		// Sun Nov  6 02:00:00 2015 -0800

		// test inside of DST and on the boundaries
		{1, "Mon Mar 14 12:35:46 2016 -0700", "Fri Mar 11 15:00:00 2016 -0800"},
		{2, "Tue Apr 12 12:35:46 2016 -0700", "Fri Apr  8 15:00:00 2016 -0700"},
		{2, "Tue Nov  8 12:35:46 2016 -0800", "Fri Nov  4 15:00:00 2016 -0700"},

		// year boundary. Note: this'll break when we support holidays as
		// non-workdays
		{1, "Fri Jan  1 12:05:00 2016 -0800", "Thu Dec 31 15:00:00 2015 -0800"},

		// try a larger number
		{30, "Fri Dec 18 11:45:11 2015 -0800", "Fri Nov  6 15:00:00 2015 -0800"},
	}

	tz, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatal(err)
	}

	endHour := 15 // typically we run chaos monkey until 3PM
	for _, tt := range tests {
		got, err := noKillsSince(tt.days, parse(tt.now), endHour, tz)
		if err != nil {
			t.Fatal(err)
		}
		if want := parse(tt.since); got != want {
			t.Errorf("noKillsSince(%d, \"%s\")=\"%s\", want \"%s\"", tt.days, tt.now, format(got.In(tz)), format(want.In(tz)))
		}
	}
}

// parse returns a time formatted as the standard output of "date", e.g.:
// Thu Dec 17 15:18:30 PST 2015
func parse(s string) time.Time {
	t, err := time.Parse("Mon Jan  2 15:04:05 2006 -0700", s)
	if err != nil {
		panic(err)
	}
	return t.UTC()
}

// format returns a formatted string representing a time, to simplify debugging
func format(tm time.Time) string {
	return tm.Format("Mon Jan  2 15:04:05 2006 -0700")
}
