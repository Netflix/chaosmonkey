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

	"github.com/Netflix/chaosmonkey/v2/config"
)

// DumpMonkeyConfig dumps the monkey-level config parameters to stdout
func DumpMonkeyConfig(cfg *config.Monkey) {
	var enabled, leashed, sched bool
	var accounts []string
	var err error

	if enabled, err = cfg.Enabled(); err != nil {
		fmt.Printf("ERROR getting enabled: %v", err)
	} else {
		fmt.Printf("enabled: %t\n", enabled)
	}

	if leashed, err = cfg.Leashed(); err != nil {
		fmt.Printf("ERROR getting leashed: %v", err)
	} else {
		fmt.Printf("leashed: %t\n", leashed)
	}

	if sched, err = cfg.ScheduleEnabled(); err != nil {
		fmt.Printf("ERROR getting schedule enabled: %v", err)
	} else {
		fmt.Printf("schedule enabled: %t\n", sched)
	}

	if accounts, err = cfg.Accounts(); err != nil {
		fmt.Printf("ERROR getting accounts: %v\n", err)
	} else {
		fmt.Printf("accounts: %v\n", accounts)
	}

	fmt.Printf("start hour: %d\n", cfg.StartHour())
	fmt.Printf("end hour: %d\n", cfg.EndHour())
	loc, _ := cfg.Location()
	fmt.Printf("location: %s\n", loc)
	fmt.Printf("cron path: %s\n", cfg.CronPath())
	fmt.Printf("term path: %s\n", cfg.TermPath())
	fmt.Printf("term account: %s\n", cfg.TermAccount())
	fmt.Printf("max apps: %d\n", cfg.MaxApps())
}
