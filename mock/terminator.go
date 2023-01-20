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

package mock

import "github.com/Netflix/chaosmonkey/v2"

// Terminator implements term.terminator
type Terminator struct {
	Instance chaosmonkey.Instance
	Ncalls   int
	Error    error
}

// Execute pretends to terminate an instance
func (t *Terminator) Execute(trm chaosmonkey.Termination) error {
	// Records the most recent killed instance for assertion checking
	t.Instance = trm.Instance

	// Records how many times it's been invoked
	t.Ncalls++

	return t.Error
}
