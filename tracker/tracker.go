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

// Package tracker provides an entry point for instantiating Trackers
package tracker

import (
	"github.com/Netflix/chaosmonkey/v2"
	"github.com/Netflix/chaosmonkey/v2/config"
	"github.com/Netflix/chaosmonkey/v2/deps"
	"github.com/pkg/errors"
)

func init() {
	deps.GetTrackers = getTrackers
}

// getTrackers returns a list of trackers specified in the configuration
func getTrackers(cfg *config.Monkey) ([]chaosmonkey.Tracker, error) {
	var result []chaosmonkey.Tracker

	kinds, err := cfg.Trackers()
	if err != nil {
		return nil, err
	}

	for _, kind := range kinds {
		tr, err := getTracker(kind, cfg)
		if err != nil {
			return nil, err
		}
		result = append(result, tr)
	}
	return result, nil
}

// getTracker returns a tracker by name
// No trackers have been implemented yet
func getTracker(kind string, cfg *config.Monkey) (chaosmonkey.Tracker, error) {
	switch kind {
	// As trackers are contributed to the open source project, they should
	// be instantiated here
	default:
		return nil, errors.Errorf("unsupported tracker: %s", kind)
	}
}
