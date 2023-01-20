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

// Package outage provides a default no-op outage implementation
package outage

import (
	"github.com/Netflix/chaosmonkey/v2"
	"github.com/Netflix/chaosmonkey/v2/config"
	"github.com/Netflix/chaosmonkey/v2/deps"
	"github.com/pkg/errors"
)

// NullOutage is a no-op outage checker
type NullOutage struct{}

// Outage always returns false
func (n NullOutage) Outage() (bool, error) {
	return false, nil
}

func init() {
	deps.GetOutage = GetOutage
}

// GetOutage returns a do-nothing outage checker
func GetOutage(cfg *config.Monkey) (chaosmonkey.Outage, error) {
	checker := cfg.OutageChecker()
	if checker != "" {
		return nil, errors.Errorf("unknown outage provider: %s", checker)
	}
	return NullOutage{}, nil
}
