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
	"io/ioutil"
	"log"
	"testing"

	"github.com/netflix/chaosmonkey"
	"github.com/netflix/chaosmonkey/grp"
)

// Test the whitelist logic
// uses the mockApp function defined in eligible_instances_test.go

func TestWhitelist(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	cfg := testConfig(chaosmonkey.Cluster)
	app := mockApp()
	group := grp.New("mock", "prod", "", "", "")

	var tests = []struct {
		whitelist *[]chaosmonkey.Exception
		count     int
	}{
		{nil, 8},                        // no whitelist
		{&[]chaosmonkey.Exception{}, 0}, // empty whitelist
		{&[]chaosmonkey.Exception{chaosmonkey.Exception{Account: "prod", Region: "us-east-1", Stack: "*", Detail: "*"}}, 4},
		{&[]chaosmonkey.Exception{chaosmonkey.Exception{Account: "prod", Region: "us-east-1", Stack: "prod", Detail: "*"}}, 2},
		{&[]chaosmonkey.Exception{chaosmonkey.Exception{Account: "prod", Region: "us-east-1", Stack: "prod", Detail: "b"}}, 1},

		{&[]chaosmonkey.Exception{
			chaosmonkey.Exception{Account: "prod", Region: "us-east-1", Stack: "prod", Detail: "b"},
			chaosmonkey.Exception{Account: "prod", Region: "us-west-2", Stack: "staging", Detail: "*"},
		}, 3},

		{&[]chaosmonkey.Exception{
			chaosmonkey.Exception{Account: "prod", Region: "us-east-1", Stack: "doesnotexist", Detail: "*"},
			chaosmonkey.Exception{Account: "prod", Region: "us-east-1", Stack: "prod", Detail: "b"},
		}, 1},
		{&[]chaosmonkey.Exception{chaosmonkey.Exception{Account: "*", Region: "*", Stack: "doesnotexist", Detail: "*"}}, 0},
	}

	for _, tt := range tests {
		cfg.Whitelist = tt.whitelist
		instances := EligibleInstances(group, cfg, app)
		got, want := len(instances), tt.count
		if got != want {
			t.Fatalf("len(eligibleInstances(group, cfg, app))=%v, want %v", got, want)
		}
	}
}
