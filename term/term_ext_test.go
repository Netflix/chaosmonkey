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

package term_test

import (
	"testing"

	"github.com/Netflix/chaosmonkey/config"
	"github.com/Netflix/chaosmonkey/config/param"
	D "github.com/Netflix/chaosmonkey/deploy"
	"github.com/Netflix/chaosmonkey/mock"
	"github.com/Netflix/chaosmonkey/term"
)

func TestEnabledAccounts(t *testing.T) {
	d := mock.Deps()
	d.Dep = mock.NewDeployment(
		map[string]D.AppMap{
			"foo": {
				D.AccountName("prod"): {CloudProvider: "aws", Clusters: D.ClusterMap{D.ClusterName("foo"): {D.RegionName("us-east-1"): {D.ASGName("foo-v001"): []D.InstanceID{"i-00000000"}}}}},
				D.AccountName("test"): {CloudProvider: "aws", Clusters: D.ClusterMap{D.ClusterName("foo"): {D.RegionName("us-east-1"): {D.ASGName("foo-v001"): []D.InstanceID{"i-00000001"}}}}},
				D.AccountName("mce"):  {CloudProvider: "aws", Clusters: D.ClusterMap{D.ClusterName("foo"): {D.RegionName("us-east-1"): {D.ASGName("foo-v001"): []D.InstanceID{"i-00000002"}}}}},
			},
		})

	app := "foo"
	region := "us-east-1"
	stack := ""
	cluster := ""

	tests := []struct {
		enabledAccounts []string
		killAccount     string
		want            bool
	}{
		{[]string{"prod"}, "prod", true},
		{[]string{"test"}, "test", true},
		{[]string{"mce"}, "mce", true},
		{[]string{"prod"}, "test", false},
		{[]string{"test"}, "prod", false},
		{[]string{"prod"}, "mce", false},
		{[]string{"prod", "test"}, "mce", false},
		{[]string{"mce", "prod", "test"}, "mce", true},
		{[]string{"prod", "mce", "test"}, "mce", true},
		{[]string{"prod", "test", "mce"}, "mce", true},
	}

	for _, test := range tests {
		account := test.killAccount

		// Set up the mock config that will use the list of accounts we pass it
		cfg := config.Defaults()
		cfg.Set(param.Enabled, true)
		cfg.Set(param.Leashed, false)
		cfg.Set(param.Accounts, test.enabledAccounts)

		d.MonkeyCfg = cfg

		// Set up the mock terminator that will track if a kill happened
		// create a new one each iteration so its state gets reset to zero
		mockT := new(mock.Terminator)
		d.T = mockT

		if err := term.Terminate(d, app, account, region, stack, cluster); err != nil {
			t.Fatal(err)
		}

		if got, want := mockT.Ncalls == 1, test.want; got != want {
			t.Errorf("kill? (account=%s, enabledAccounts=%v, got %t, want %t, mockT.Ncalls=%d", account, test.enabledAccounts, got, want, mockT.Ncalls)
		}

	}

}
