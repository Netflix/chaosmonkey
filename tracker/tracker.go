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
	"github.com/netflix/chaosmonkey"
	"github.com/netflix/chaosmonkey/config"
	"github.com/netflix/chaosmonkey/deps"
	"github.com/pkg/errors"
)

func init() {
	deps.GetTrackers = noSupportedTrackers
}

// No trackers have been implemented yet

// noSupportedTrackers will return an error unless cfg.Trackers() is empty
// It is a placeholder function until the open-source version implements other
// trackers
func noSupportedTrackers(cfg *config.Monkey) ([]chaosmonkey.Tracker, error) {
	kinds, err := cfg.Trackers()
	if err != nil {
		return nil, err
	}

	for _, tracker := range kinds {
		return nil, errors.Errorf("unsupported tracker: %s", tracker)
	}

	return nil, nil
}
