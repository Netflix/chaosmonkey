// Copyright 2017 Netflix, Inc.
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

package command

import (
	"fmt"
	"github.com/Netflix/chaosmonkey/deploy"
	"github.com/Netflix/chaosmonkey/spinnaker"
	"github.com/SmartThingsOSS/frigga-go"
	"os"
)

// DumpRegions lists the regions that a cluster is in
func DumpRegions(cluster, account string, spin spinnaker.Spinnaker) {

	names, err := frigga.Parse(cluster)
	if err != nil {
		fmt.Printf("ERROR: %s", err)
		os.Exit(1)
	}

	regions, err := spin.GetRegionNames(names.App, deploy.AccountName(account), deploy.ClusterName(cluster))
	if err != nil {
		fmt.Printf("ERROR: %v", err)
		os.Exit(1)
	}

	for _, region := range regions {
		fmt.Println(region)
	}

}
