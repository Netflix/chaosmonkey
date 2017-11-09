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

package constrainer

import (
	"github.com/Netflix/chaosmonkey/config"
	"github.com/Netflix/chaosmonkey/deps"
	"github.com/Netflix/chaosmonkey/schedule"
)

type NullConstrainer struct{}

func init() {
	deps.GetConstrainer = getNullConstrainer
}

// Filter implements schedule.Constrainer.Filter
// This is a no-op implementation
func (n NullConstrainer) Filter(s schedule.Schedule) schedule.Schedule {
	return s
}

func getNullConstrainer(cfg *config.Monkey) (schedule.Constrainer, error) {
	return NullConstrainer{}, nil
}
