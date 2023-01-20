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

package cal_test

import (
	"testing"
	"time"

	"github.com/Netflix/chaosmonkey/v2/cal"
)

var weekdayTests = []struct {
	date string
	pred bool
}{
	{"Mon Dec 14 11:33:58 PST 2015", true},
	{"Tue Dec 15 11:33:58 PST 2015", true},
	{"Wed Dec 16 11:33:58 PST 2015", true},
	{"Thu Dec 17 11:33:58 PST 2015", true},
	{"Fri Dec 18 11:33:58 PST 2015", true},
	{"Sat Dec 19 11:33:58 PST 2015", false},
	{"Sun Dec 20 11:33:58 PST 2015", false},
}

func TestIsWorkday(t *testing.T) {
	for _, tt := range weekdayTests {
		if got, want := cal.IsWorkday(parse(tt.date)), tt.pred; got != want {
			t.Fatalf("isWeekday(\"%s\")=%t, want %t", tt.date, got, want)
		}
	}
}

// parse returns a time formatted as the standard output of "date", e.g.:
// Thu Dec 17 15:18:30 PST 2015
func parse(s string) time.Time {
	t, err := time.Parse("Mon Jan  2 15:04:05 PST 2006", s)
	if err != nil {
		panic(err)
	}
	return t
}
