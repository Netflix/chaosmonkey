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

package command

import (
	"log"

	"github.com/Netflix/chaosmonkey/v2/deps"
	"github.com/Netflix/chaosmonkey/v2/term"
)

// Terminate executes the "terminate" command. This selects an instance
// based on the app, account, region, stack, cluster passed
//
// region, stack, and cluster may be blank
func Terminate(d deps.Deps, app string, account string, region string, stack string, cluster string) {
	err := term.Terminate(d, app, account, region, stack, cluster)
	if err != nil {
		cerr := d.ErrCounter.Increment()
		if cerr != nil {
			log.Printf("WARNING could not increment error counter: %v", cerr)
		}
		log.Fatalf("FATAL %v\n\nstack trace:\n%+v", err, err)
	}
}
