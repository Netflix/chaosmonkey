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

package errorcounter

import (
	"github.com/Netflix/chaosmonkey/v2"
	"github.com/Netflix/chaosmonkey/v2/config"
	"github.com/Netflix/chaosmonkey/v2/deps"
	"github.com/pkg/errors"
)

// Netflix uses Atlas for tracking error events.
// In the open-source build, we currently only support a null (no-op) error
// counter

type nullErrorCounter struct{}

func (n nullErrorCounter) Increment() error {
	return nil
}

func init() {
	deps.GetErrorCounter = getNullErrorCounter
}

func getNullErrorCounter(cfg *config.Monkey) (chaosmonkey.ErrorCounter, error) {
	kind := cfg.ErrorCounter()
	if kind != "" {
		return nil, errors.Errorf("unsupported error counter: %s", kind)
	}

	return nullErrorCounter{}, nil
}
