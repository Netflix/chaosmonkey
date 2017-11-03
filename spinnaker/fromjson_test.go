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

package spinnaker

import (
	"testing"

	"github.com/Netflix/chaosmonkey"
)

func TestFromJSON(t *testing.T) {
	input := `
	{
		  "name": "abc",
		  "attributes": {
			  "chaosMonkey": {
				  "enabled": true,
				  "meanTimeBetweenKillsInWorkDays": 5,
				  "minTimeBetweenKillsInWorkDays": 1,
				  "grouping": "cluster",
				  "regionsAreIndependent": true,
				  "exceptions" : [
				  {
					  "account": "test",
					  "stack": "*",
					  "detail": "*",
					  "region": "*"
				  },
				  {
					  "account": "prod",
					  "stack": "*",
					  "detail": "*",
					  "region": "eu-west-1"
				  }
				  ]
			  }
		  }
	  }
  `
	actual, err := fromJSON([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if !actual.Enabled {
		t.Error("Expected enabled to be true")
	}

	if actual.MeanTimeBetweenKillsInWorkDays != 5 {
		t.Errorf("Expected mean time: 5. acutal mean time: %d", actual.MeanTimeBetweenKillsInWorkDays)
	}

	if !actual.RegionsAreIndependent {
		t.Error("Expected regions to be independent")
	}

	if actual.Grouping != chaosmonkey.Cluster {
		t.Errorf("Expected grouping to be Cluster, was %s", actual.Grouping)
	}

	expectedEx := []chaosmonkey.Exception{
		{Account: "test", Stack: "*", Detail: "*", Region: "*"},
		{Account: "prod", Stack: "*", Detail: "*", Region: "eu-west-1"},
	}

	actualEx := actual.Exceptions

	if len(actualEx) != len(expectedEx) {
		t.Fatalf("Expected number of exceptions: %d. Actual number of exceptions: %d", len(expectedEx), len(actualEx))
	}

	if actual.Whitelist != nil {
		t.Fatalf("Expected whitelist to be nil when not specified, was: %v", actual.Whitelist)
	}

	for i := range expectedEx {
		var expected, actual string
		expected = expectedEx[i].Account
		actual = actualEx[i].Account
		if expected != actual {
			t.Errorf("i: %d. Expected account: %s. Actual account: %s", i, expected, actual)
		}

		expected = expectedEx[i].Stack
		actual = actualEx[i].Stack
		if expected != actual {
			t.Errorf("i: %d. Expected stack: %s. Actual stack: %s", i, expected, actual)
		}

		expected = expectedEx[i].Detail
		actual = actualEx[i].Detail
		if expected != actual {
			t.Errorf("i: %d. Expected detail: %s. Actual detail: %s", i, expected, actual)
		}

		expected = expectedEx[i].Region
		actual = actualEx[i].Region
		if expected != actual {
			t.Errorf("i: %d. Expected region: %s. Actual region: %s", i, expected, actual)
		}
	}
}

func TestFromJSONDisabled(t *testing.T) {
	input := `
	{
		"name": "abc",
		"attributes": {
			"chaosMonkey": {
				"enabled": false
			}
		}
	}
	`

	actual, err := fromJSON([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if actual.Enabled {
		t.Error("Expected enabled to be false")
	}
}

func TestBadJSON(t *testing.T) {
	tests := []string{
		`{}`,
		`{"name": "abc"}`,
		`{"name": "abc", "attributes": {}}`,
		`{"name": "abc", "attributes": {"chaosMonkey": {}}}`,
		`{"name": "abc", "attributes": {"chaosMonkey": {}}}`,
		`{"name": "abc", "attributes": {"chaosMonkey": {"enabled": true}}}`, // if enabled, need valid grouping, mean, and min time.
		`{"name": "abc", "attributes": {"chaosMonkey": {"enabled": true, "grouping": app}}}`,
		`{"name": "abc", "attributes": {"chaosMonkey": {"enabled": true, "grouping": app, "meanTimeBetweenKillsInWorkDays": 1}}}`,
		`{"name": "abc", "attributes": {"chaosMonkey": {"enabled": true, "grouping": app, "minTimeBetweenKillsInWorkDays": 1}}}`,
		// mean time must be > 0
		`{"name": "abc", "attributes": {"chaosMonkey": {"enabled": true, "grouping": "app", "meanTimeBetweenKillsInWorkDays": 0, "minTimeBetweenKillsInWorkDays": 1}}}`,

		// exceptions must have a region field
		`
		{"name": "abc",
		 "attributes": {
			"chaosMonkey": {
				"enabled": true, "grouping": "app", "meanTimeBetweenKillsInWorkDays": 1, "minTimeBetweenKillsInWorkDays": 1,
				"exceptions": [{"account": "prod"}]
	    }}}`,

		// exceptions must have an account field
		`
		{"name": "abc",
		 "attributes": {
			"chaosMonkey": {
				"enabled": true, "grouping": "app", "meanTimeBetweenKillsInWorkDays": 1, "minTimeBetweenKillsInWorkDays": 1,
				"exceptions": [{"region": "*"}]
	    }}}`,
	}

	for _, input := range tests {
		_, err := fromJSON([]byte(input))
		if err == nil {
			t.Fatalf("Expected an error given missing config: %s", input)
		}
	}
}

func TestFromJSONEmptyWhitelist(t *testing.T) {
	input := `
	  {
		  "name": "abc",
		  "attributes": {
			  "chaosMonkey": {
				  "enabled": true,
				  "meanTimeBetweenKillsInWorkDays": 5,
				  "minTimeBetweenKillsInWorkDays": 1,
				  "grouping": "cluster",
				  "regionsAreIndependent": true,
				  "whitelist": [],
				  "exceptions" : [
				  {
					  "account": "test",
					  "stack": "*",
					  "detail": "*",
					  "region": "*"
				  },
				  {
					  "account": "prod",
					  "stack": "*",
					  "detail": "*",
					  "region": "eu-west-1"
				  }
				  ]
			  }
		  }
	  }
  `
	actual, err := fromJSON([]byte(input))
	if err != nil {
		t.Fatal(err)
	}

	if actual.Whitelist == nil {
		t.Fatal("Whitelist is not present")
	}

	wl := *actual.Whitelist
	if len(wl) != 0 {
		t.Errorf("Expected whitelist to be empty, was: %v", wl)
	}
}

func TestFromJSONPopulatedWhitelist(t *testing.T) {
	input := `
	  {
		  "name": "abc",
		  "attributes": {
			  "chaosMonkey": {
				  "enabled": true,
				  "meanTimeBetweenKillsInWorkDays": 5,
				  "minTimeBetweenKillsInWorkDays": 1,
				  "grouping": "cluster",
				  "regionsAreIndependent": true,
				  "exceptions": [],
				  "whitelist" : [
				  {
					  "account": "test",
					  "stack": "*",
					  "detail": "*",
					  "region": "*"
				  },
				  {
					  "account": "prod",
					  "stack": "*",
					  "detail": "*",
					  "region": "eu-west-1"
				  }
				  ]
			  }
		  }
	  }
  `
	actual, err := fromJSON([]byte(input))
	if err != nil {
		t.Fatal(err)
	}

	if actual.Whitelist == nil {
		t.Fatal("Whitelist is not present")
	}

	actualWl := *actual.Whitelist

	expectedWl := []chaosmonkey.Exception{
		{Account: "test", Stack: "*", Detail: "*", Region: "*"},
		{Account: "prod", Stack: "*", Detail: "*", Region: "eu-west-1"},
	}

	if len(actualWl) != len(expectedWl) {
		t.Fatalf("Expected whitelist size: %d. Actual whitelist size: %d", len(expectedWl), len(actualWl))
	}

	for i := range expectedWl {
		var expected, actual string
		expected = expectedWl[i].Account
		actual = actualWl[i].Account
		if expected != actual {
			t.Errorf("i: %d. Expected account: %s. Actual account: %s", i, expected, actual)
		}

		expected = expectedWl[i].Stack
		actual = actualWl[i].Stack
		if expected != actual {
			t.Errorf("i: %d. Expected stack: %s. Actual stack: %s", i, expected, actual)
		}

		expected = expectedWl[i].Detail
		actual = actualWl[i].Detail
		if expected != actual {
			t.Errorf("i: %d. Expected detail: %s. Actual detail: %s", i, expected, actual)
		}

		expected = expectedWl[i].Region
		actual = actualWl[i].Region
		if expected != actual {
			t.Errorf("i: %d. Expected region: %s. Actual region: %s", i, expected, actual)
		}
	}
}
