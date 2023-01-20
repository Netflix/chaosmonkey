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

// Package env contains a no-op implementation of chaosmonkey.env
// where InTest() always returns false
package env

import (
	"github.com/Netflix/chaosmonkey/v2"
	"github.com/Netflix/chaosmonkey/v2/config"
	"github.com/Netflix/chaosmonkey/v2/deps"
)

// notTestEnv is an environment that does not report as a test env
type notTestEnv struct{}

// InTest implements chaosmonkey.Env.InTest
func (n notTestEnv) InTest() bool {
	return false
}

func init() {
	deps.GetEnv = getNotTestEnv
}

func getNotTestEnv(cfg *config.Monkey) (chaosmonkey.Env, error) {
	return notTestEnv{}, nil
}
