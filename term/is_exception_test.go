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
	"testing"

	"github.com/netflix/chaosmonkey"
	D "github.com/netflix/chaosmonkey/deploy"
)

// mockASG creates a mock ASG
func mockASG() *D.ASG {
	app := D.NewApp("myapp", D.AppMap{
		D.AccountName("prod"): {
			CloudProvider: "aws",
			Clusters: D.ClusterMap{
				D.ClusterName("myapp-mystack-mydetail"): {
					D.RegionName("us-east-1"): {
						D.ASGName("myapp-mystack-mydetail-v123"): []D.InstanceID{"i-6643f925"},
					},
				},
			},
		},
	})

	return app.Accounts()[0].Clusters()[0].ASGs()[0]

}

func TestIsExceptionNoWildcards(t *testing.T) {
	exs := []chaosmonkey.Exception{
		chaosmonkey.Exception{Account: "prod", Stack: "mystack", Detail: "mydetail", Region: "us-east-1"},
	}

	asg := mockASG()

	if got, want := asg.DetailName(), "mydetail"; got != want {
		t.Fatalf("asg.DetailName()=%v, want %v", got, want)
	}

	if got, want := isException(exs, asg), true; got != want {
		t.Fatalf("isException(...)=%v, want %v", got, want)
	}
}
