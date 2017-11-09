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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/Netflix/chaosmonkey"
	"github.com/Netflix/chaosmonkey/config"
	"github.com/Netflix/chaosmonkey/deploy"
	"github.com/Netflix/chaosmonkey/schedstore"
	"github.com/Netflix/chaosmonkey/schedule"
)

// Schedule executes the "schedule" command. This defines the schedule
// of terminations for the day and records them as cron jobs
func Schedule(g chaosmonkey.AppConfigGetter, ss schedstore.SchedStore, cfg *config.Monkey, d deploy.Deployment, cons schedule.Constrainer, apps []string) {

	enabled, err := cfg.ScheduleEnabled()
	if err != nil {
		log.Fatalf("FATAL: cannot determine if schedule is enabled: %v", err)
	}
	if !enabled {
		log.Println("schedule disabled, not running")
		return
	}

	/*
	 Note: We don't check for the enable flag during scheduling, only
	 during terminations. That way, if chaos monkey is disabled during
	 scheduling time but later in the day becomes enabled, it still
	 functions correctly.
	*/
	err = do(d, g, ss, cfg, cons, apps)

	if err != nil {
		log.Fatalf("FATAL: %v", err)
	}

}

// do is the actual implementation for the Schedule function
func do(d deploy.Deployment, g chaosmonkey.AppConfigGetter, ss schedstore.SchedStore, cfg *config.Monkey, cons schedule.Constrainer, apps []string) error {

	s := schedule.New()
	err := s.Populate(d, g, cfg, apps)
	if err != nil {
		return fmt.Errorf("failed to populate schedule: %v", err)
	}

	// Filter out terminations that violate constrains
	sched := cons.Filter(*s)

	err = deploySchedule(&sched, ss, cfg)
	if err != nil {
		return fmt.Errorf("failed to deploy schedule: %v", err)
	}

	return nil
}

// deploySchedule publishes the schedule to chaosmonkey-api
// and registers the schedule with the local cron
func deploySchedule(s *schedule.Schedule, ss schedstore.SchedStore, cfg *config.Monkey) error {
	loc, err := cfg.Location()
	if err != nil {
		return fmt.Errorf("deploySchedule: could not retrieve local timezone: %v", err)
	}

	today := time.Now().In(loc)

	err = ss.Publish(today, s)

	if err != nil {
		return fmt.Errorf("deploySchedule: could not publish schedule: %v", err)
	}

	err = registerWithCron(s, cfg)
	return err
}

// registerWithCron registers the schedule of terminations with cron on the local machine
//
// Creates or overwrites the file specified by config.Chaos.CronPath()
func registerWithCron(s *schedule.Schedule, cfg *config.Monkey) error {
	crontab := s.Crontab(cfg.TermPath(), cfg.TermAccount())
	var perms os.FileMode = 0644 // -rw-r--r--
	log.Printf("Writing %s\n", cfg.CronPath())
	err := ioutil.WriteFile(cfg.CronPath(), crontab, perms)
	return err
}
