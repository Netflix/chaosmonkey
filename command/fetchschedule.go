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
	"time"

	"github.com/Netflix/chaosmonkey/v2/config"
	"github.com/Netflix/chaosmonkey/v2/schedstore"
)

// FetchSchedule executes the "fetch-schedule" command. This checks if there
// is an existing schedule for today that was previously registered
// in chaosmonkey-api. If so, it downloads the schedule from chaosmonkey-api
// and installs it locally.
func FetchSchedule(s schedstore.SchedStore, cfg *config.Monkey) {
	log.Println("chaosmonkey fetch-schedule starting")
	sched, err := s.Retrieve(today(cfg))
	if err != nil {
		log.Fatalf("FATAL: could not fetch schedule: %v", err)
	}

	if sched == nil {
		log.Println("no schedule to retrieve")
		return
	}

	err = registerWithCron(sched, cfg)
	if err != nil {
		log.Fatalf("FATAL: could not register with cron: %v", err)
	}

	defer log.Println("chaosmonkey fetch-schedule done")
}

// today returns a date in local time
func today(cfg *config.Monkey) time.Time {
	loc, err := cfg.Location()
	if err != nil {
		log.Fatalf("FATAL: Could not get local timezone: %v", err)
	}

	return time.Now().In(loc)
}
