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

package spinnaker

import (
	"encoding/json"
	"fmt"

	"github.com/Netflix/chaosmonkey"

	"github.com/pkg/errors"
)

// FromJSON takes a Spinnaker JSON representation of an app
// and returns a Chaos Monkey config
// Example:
//   {
//       "name": "abc",
//       "attributes": {
//         "chaosMonkey": {
//         "enabled": true,
//           "meanTimeBetweenKillsInWorkDays": 5,
//           "minTimeBetweenKillsInWorkDays": 1,
//           "grouping": "cluster",
//           "regionsAreIndependent": false,
//         },
//         "exceptions" : [
//             {
//                 "account": "test",
//                 "stack": "*",
//                 "cluster": "*",
//                 "region": "*"
//             },
//             {
//                 "account": "prod",
//                 "stack": "*",
//                 "cluster": "*",
//                 "region": "eu-west-1"
//             },
//         ]
//       }
//   }
//
//
// Example of disabled app:
//   {
//       "name": "abc",
//       "attributes": {
//         "chaosMonkey": {
//         "enabled": false
//         }
//       }
//    }
//
//
// Example with whitelist
//
// 	  {
//  	  "enabled": true,
//  	  "grouping": "app",
//  	  "meanTimeBetweenKillsInWorkDays": 4,
//  	  "minTimeBetweenKillsInWorkDays": 1,
//  	  "regionsAreIndependent": true,
//  	  "exceptions": [
//  	  	{
//  	  	"account": "prod",
//  	  	"region": "us-west-2",
//  	  	"stack": "foo",
//  	  	"detail": "bar"
//  	  	}
//  	  ],
//  	  "whitelist": [
//  	  	{
//  	  	"account": "test",
//  	  	"stack": "*",
//  	  	"region": "*",
//  	  	"detail": "*"
//  	  	}
//  	  ]
// 	  }
//
func fromJSON(js []byte) (*chaosmonkey.AppConfig, error) {
	parsed := new(parsedJSON)
	err := json.Unmarshal(js, parsed)

	if err != nil {
		return nil, errors.Wrap(err, "json unmarshal failed")
	}

	if parsed.Attributes == nil {
		return nil, errors.New("'attributes' field missing")
	}

	if parsed.Attributes.ChaosMonkey == nil {
		return nil, errors.New("'attributes.chaosMonkey' field missing")
	}

	cm := parsed.Attributes.ChaosMonkey

	if cm.Enabled == nil {
		return nil, errors.New("'attributes.chaosMonkey.enabled' field missing")
	}

	// Check if mean time between kills is missing.
	// If not enabled, it's ok if it's missing
	if *cm.Enabled && cm.MeanTimeBetweenKillsInWorkDays == nil {
		return nil, errors.New("attributes.chaosMonkey.meanTimeBetweenKillsInWorkDays missing")
	}

	if *cm.Enabled && cm.MinTimeBetweenKillsInWorkDays == nil {
		return nil, errors.New("attributes.chaosMonkey.minTimeBetweenKillsInWorkDays missing")
	}

	if *cm.Enabled && (*cm.MeanTimeBetweenKillsInWorkDays <= 0) {
		return nil, fmt.Errorf("invalid attributes.chaosMonkey.meanTimeBetweenKillsInWorkDays: %d", cm.MeanTimeBetweenKillsInWorkDays)
	}

	grouping := chaosmonkey.Cluster

	switch cm.Grouping {
	case "app":
		grouping = chaosmonkey.App
	case "stack":
		grouping = chaosmonkey.Stack
	case "cluster":
		grouping = chaosmonkey.Cluster
	default:
		// If not enabled, the user may not have specified a grouping at all,
		// in which case we stick with the default
		if *cm.Enabled {
			return nil, errors.Errorf("Unknown grouping: %s", cm.Grouping)
		}
	}

	var meanTime int
	var minTime int

	if cm.MeanTimeBetweenKillsInWorkDays != nil {
		meanTime = *cm.MeanTimeBetweenKillsInWorkDays
	}

	if cm.MinTimeBetweenKillsInWorkDays != nil {
		minTime = *cm.MinTimeBetweenKillsInWorkDays
	}

	// Exceptions must have a non-blank region field
	for _, exception := range cm.Exceptions {
		if exception.Account == "" {
			return nil, errors.New("missing account field in exception")
		}

		if exception.Region == "" {
			return nil, errors.New("missing region field in exception")
		}
	}

	cfg := chaosmonkey.AppConfig{
		Enabled:                        *cm.Enabled,
		RegionsAreIndependent:          cm.RegionsAreIndependent,
		Grouping:                       grouping,
		MeanTimeBetweenKillsInWorkDays: meanTime,
		MinTimeBetweenKillsInWorkDays:  minTime,
		Exceptions:                     cm.Exceptions,
		Whitelist:                      cm.Whitelist,
	}

	return &cfg, nil
}

// parsedJson is the parsed JSON representation
type parsedJSON struct {
	Name       string      `json:"name"`
	Attributes *parsedAttr `json:"attributes"`
}

type parsedAttr struct {
	ChaosMonkey *parsedChaosMonkey `json:"chaosmonkey"`
}

type parsedChaosMonkey struct {
	Enabled                        *bool                    `json:"enabled"`
	Grouping                       string                   `json:"grouping"`
	MeanTimeBetweenKillsInWorkDays *int                     `json:"meanTimeBetweenKillsInWorkDays"`
	MinTimeBetweenKillsInWorkDays  *int                     `json:"minTimeBetweenKillsInWorkDays"`
	RegionsAreIndependent          bool                     `json:"regionsAreIndependent"`
	Exceptions                     []chaosmonkey.Exception  `json:"exceptions"`
	Whitelist                      *[]chaosmonkey.Exception `json:"whitelist"`
}
