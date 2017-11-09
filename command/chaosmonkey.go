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
	"log"
	"math"
	"os"
	"runtime/debug"
	"strings"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/Netflix/chaosmonkey"
	"github.com/Netflix/chaosmonkey/clock"
	"github.com/Netflix/chaosmonkey/config"
	"github.com/Netflix/chaosmonkey/config/param"
	"github.com/Netflix/chaosmonkey/deploy"
	"github.com/Netflix/chaosmonkey/deps"
	"github.com/Netflix/chaosmonkey/mysql"
	"github.com/Netflix/chaosmonkey/schedstore"
	"github.com/Netflix/chaosmonkey/schedule"
	"github.com/Netflix/chaosmonkey/spinnaker"
)

// Version is the version number
const Version = "2.0.2"

func printVersion() {
	fmt.Printf("%s\n", Version)
}

var (
	// configPaths is where Chaos Monkey will look for a chaosmonkey.toml
	// configuration file
	configPaths = [...]string{".", "/apps/chaosmonkey", "/etc", "/etc/chaosmonkey"}
)

// Usage prints usage
func Usage() {
	usage := `
Chaos Monkey

Usage:
	chaosmonkey <command> ...

command: migrate | schedule | terminate | fetch-schedule | outage | config  | email | eligible | intest

Install
-------
Installs chaosmonkey with all the setup required, e.g setting up the cron, appling database migration etc.

migrate
-------
Applies database migration to the database defined in the configuration file.

schedule [--max-apps=<N>] [--apps=foo,bar,baz] [--no-record-schedule]
--------------------------------------------------------------------
Generates a schedule of terminations for the day and installs the
terminations as local cron jobs that call "chaosmonkey terminate ..."

--apps=foo,bar,baz     Optionally specify an explicit list of apps to schedule.
                       This is primarily used for debugging.

--max-apps=<N>         Optionally specify the maximum number of apps that Chaos Monkey
					   will schedule. This is primarily used for debugging.

--no-record-schedule   Do not record the schedule with the database.
                       This is primarily used for debugging.


terminate <app> <account> [--region=<region>] [--stack=<stack>] [--cluster=<cluster>] [--leashed]
-----------------------------------------------------------------------------------------------------------------
Terminates an instance from a given app and account.

Optionally specify a region, stack, cluster.

The --leashed flag forces chaosmonkey to run in leashed mode. When leashed,
Chaos Monkey will check if an instance should be terminated, but will not
actually terminate it.

fetch-schedule
--------------
Queries the database to see if there is an existing schedule of
terminations for today. If so, downloads the schedule and sets up cron jobs to
implement the schedule.

outage
------
Output "true" if there is an ongoing outage, otherwise "false". Used for debugging.


config [<app>]
------------
Query Spinnaker for the config for a specific app and dump it to
standard out. This is only used for debugging.

If no app is specified, dump the Monkey-level configuration options to standard out.

Examples:

	chaosmonkey config chaosguineapig

	chaosmonkey config

eligible <app> <account> [--region=<region>] [--stack=<stack>] [--cluster=<cluster>]
-------------------------------------------------------------------------------------

Dump a list of instance-ids that are eligible for termination for a given app, account,
and optionally region, stack, and cluster.

intest
------

Outputs "true" on standard out if running within a test environment, otherwise outputs "false"


account <name>
--------------

Look up an cloud account ID by name.

Example:

	chaosmonkey account test


provider <name>
---------------

Look up the cloud provider by account name.

Example:

	chaosmonkey provider test


clusters <app> <account>
------------------------

List the clusters for a given app and account

Example:

	chaosmonkey clusters chaosguineapig test


regions <cluster> <account>
---------------------------

List the regions for a given cluster and account

Example:

	chaosmonkey regions chaosguineapig test
`
	fmt.Printf(usage)
}

func init() {
	// Prepend the pid to log statements
	log.SetPrefix(fmt.Sprintf("[%5d] ", os.Getpid()))
}

// Execute is the main entry point for the chaosmonkey cli.
func Execute() {
	regionPtr := flag.String("region", "", "region of termination group")
	stackPtr := flag.String("stack", "", "stack of termination group")
	clusterPtr := flag.String("cluster", "", "cluster of termination group")
	appsPtr := flag.String("apps", "", "comma-separated list of apps to schedule for termination")
	noRecordSchedulePtr := flag.Bool("no-record-schedule", false, "do not record schedule")
	versionPtr := flag.BoolP("version", "v", false, "show version")
	flag.Usage = Usage

	// These flags, if specified, override config values
	maxAppsFlag := "max-apps"
	leashedFlag := "leashed"
	flag.Int(maxAppsFlag, math.MaxInt32, "max number of apps to examine for termination")
	flag.Bool(leashedFlag, false, "force leashed mode")

	flag.Parse()
	if len(flag.Args()) == 0 {
		if *versionPtr {
			printVersion()
			os.Exit(0)
		}

		flag.Usage()
		os.Exit(1)
	}

	cmd := flag.Arg(0)

	cfg, err := getConfig()

	if err != nil {
		log.Fatalf("FATAL: failed to load config: %v", err)
	}

	// Associate config values with flags
	err = cfg.BindPFlag(param.MaxApps, flag.Lookup(maxAppsFlag))
	if err != nil {
		log.Fatalf("FATAL: failed to bind flag: --%s: %v", maxAppsFlag, err)
	}
	err = cfg.BindPFlag(param.Leashed, flag.Lookup(leashedFlag))
	if err != nil {
		log.Fatalf("FATAL: failed to bind flag: --%s: %v", leashedFlag, err)
	}

	spin, err := spinnaker.NewFromConfig(cfg)

	if err != nil {
		log.Fatalf("FATAL: spinnaker.New failed: %+v", err)
	}

	outage, err := deps.GetOutage(cfg)
	if err != nil {
		log.Fatalf("FATAL: deps.GetOutage fail: %+v", err)
	}

	sql, err := mysql.NewFromConfig(cfg)
	if err != nil {
		log.Fatalf("FATAL: could not initialize mysql connection: %+v", err)
	}

	cons, err := deps.GetConstrainer(cfg)
	if err != nil {
		log.Fatalf("FATAL: deps.GetConstrainer failed: %+v", err)
	}

	// Ensure mysql object gets closed
	defer func() {
		_ = sql.Close()
	}()

	switch cmd {
	case "install":
		executable := ChaosmonkeyExecutable{}
		Install(cfg, executable, sql)
	case "migrate":
		Migrate(sql)
	case "schedule":
		log.Println("chaosmonkey schedule starting")
		defer log.Println("chaosmonkey schedule done")

		var apps []string
		if *appsPtr != "" {
			// User explicitly specified list of apps on the command line
			apps = strings.Split(*appsPtr, ",")
		} else {
			// User did not explicitly specify list of apps, get 'em all
			var err error
			apps, err = spin.AppNames()
			if err != nil {
				log.Fatalf("FATAL: could not retrieve list of app names: %v", err)
			}
		}

		var schedStore schedstore.SchedStore

		schedStore = sql
		if *noRecordSchedulePtr {
			schedStore = nullSchedStore{}
		}

		Schedule(spin, schedStore, cfg, spin, cons, apps)
	case "fetch-schedule":
		FetchSchedule(sql, cfg)
	case "terminate":
		if len(flag.Args()) != 3 {
			flag.Usage()
			os.Exit(1)
		}
		app := flag.Arg(1)
		account := flag.Arg(2)
		trackers, err := deps.GetTrackers(cfg)
		if err != nil {
			log.Fatalf("FATAL: could not create trackers: %+v", err)
		}

		errCounter, err := deps.GetErrorCounter(cfg)
		if err != nil {
			log.Fatalf("FATAL: could not create error counter: %+v", err)
		}

		env, err := deps.GetEnv(cfg)
		if err != nil {
			log.Fatalf("FATAL: could not determine environment: %+v", err)
		}

		defer logOnPanic(errCounter) // Handler in case of panic
		deps := deps.Deps{
			MonkeyCfg:  cfg,
			Checker:    sql,
			ConfGetter: spin,
			Cl:         clock.New(),
			Dep:        spin,
			T:          spin,
			Trackers:   trackers,
			Ou:         outage,
			ErrCounter: errCounter,
			Env:        env,
		}
		Terminate(deps, app, account, *regionPtr, *stackPtr, *clusterPtr)
	case "outage":
		Outage(outage)
	case "config":
		if len(flag.Args()) != 2 {
			DumpMonkeyConfig(cfg)
			return
		}
		app := flag.Arg(1)
		DumpConfig(spin, app)
	case "eligible":
		if len(flag.Args()) != 3 {
			flag.Usage()
			os.Exit(1)
		}
		app := flag.Arg(1)
		account := flag.Arg(2)
		Eligible(spin, spin, app, account, *regionPtr, *stackPtr, *clusterPtr)
	case "intest":
		env, err := deps.GetEnv(cfg)
		if err != nil {
			log.Fatalf("FATAL: could not determine environment: %+v", err)
		}
		fmt.Println(env.InTest())
	case "account":
		if len(flag.Args()) != 2 {
			flag.Usage()
			os.Exit(1)
		}

		account := flag.Arg(1)
		id, err := spin.AccountID(account)
		if err != nil {
			fmt.Printf("ERROR: Could not retrieve id for account: %s. Reason: %v\n", account, err)
			return
		}
		fmt.Println(id)
	case "provider":
		if len(flag.Args()) != 2 {
			flag.Usage()
			os.Exit(1)
		}
		account := flag.Arg(1)
		provider, err := spin.CloudProvider(account)
		if err != nil {
			fmt.Printf("ERROR: Could not retrieve provider for account: %s. Reason: %v\n", account, err)
			return
		}
		fmt.Println(provider)
	case "clusters":
		if len(flag.Args()) != 3 {
			flag.Usage()
			os.Exit(1)
		}

		app := flag.Arg(1)
		account := flag.Arg(2)
		clusters, err := spin.GetClusterNames(app, deploy.AccountName(account))
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(1)
		}
		for _, cluster := range clusters {
			fmt.Println(cluster)
		}

	case "regions":
		if len(flag.Args()) != 3 {
			flag.Usage()
			os.Exit(1)
		}

		cluster := flag.Arg(1)
		account := flag.Arg(2)

		DumpRegions(cluster, account, spin)

	default:
		flag.Usage()
		os.Exit(1)
	}
}

func init() {
	// All logs to stdout
	log.SetOutput(os.Stdout)
}

// logOnPanic increments an error metric and logs if a panic happens
func logOnPanic(errCounter chaosmonkey.ErrorCounter) {
	if e := recover(); e != nil {
		log.Printf("FATAL: panic: %s: %s", e, debug.Stack())
		err := errCounter.Increment()
		if err != nil {
			log.Printf("failed to increment error counter: %s", err)
		}
	}
}

// return configuration info
func getConfig() (*config.Monkey, error) {
	cfg, err := config.Load(configPaths[:])
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// nullSchedStore is a no-op implementation of api.SchedStore
type nullSchedStore struct{}

// Retrieve implements api.SchedStore.Retrieve
func (n nullSchedStore) Retrieve(date time.Time) (*schedule.Schedule, error) {
	return nil, fmt.Errorf("nullSchedStore does not support Retrieve function")
}

// Publish implements api.SchedStore.Publish
func (n nullSchedStore) Publish(date time.Time, sched *schedule.Schedule) error {
	return nil
}
