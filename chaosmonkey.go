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

// Package chaosmonkey contains our domain models
package chaosmonkey

import (
	"fmt"
	"time"
)

const (
	// App grouping: Chaos Monkey kills one instance per app per day
	App Group = iota
	// Stack grouping: Chaos Monkey kills one instance per stack per day
	Stack
	// Cluster grouping: Chaos Monkey kills one instance per cluster per day
	Cluster
)

type (

	// AppConfig contains app-specific configuration parameters for Chaos Monkey
	AppConfig struct {
		Enabled                        bool
		RegionsAreIndependent          bool
		MeanTimeBetweenKillsInWorkDays int
		MinTimeBetweenKillsInWorkDays  int
		Grouping                       Group
		Exceptions                     []Exception
		Whitelist                      *[]Exception
	}

	// Group describes what Chaos Monkey considers a group of instances
	// Chaos Monkey will randomly kill an instance from each group.
	// The group generally maps onto what the service owner considers
	// a "cluster", which is different from Spinnaker's notion of a cluster.
	Group int

	// Exception describes clusters that have been opted out of chaos monkey
	// If one of the members is a "*", it matches everything. That is the only
	// wildcard value
	// For example, this will opt-out all of the cluters in the test account:
	// Exception{ Account:"test", Stack:"*", Cluster:"*", Region: "*"}
	Exception struct {
		Account string
		Stack   string
		Detail  string
		Region  string
	}

	// Instance contains naming info about an instance
	Instance interface {
		// AppName is the name of the Netflix app
		AppName() string

		// AccountName is the name of the account the instance is running in (e.g., prod, test)
		AccountName() string

		// RegionName is the name of the AWS region (e.g., us-east-1
		RegionName() string

		// StackName returns the "stack" part of app-stack-detail in cluster names
		StackName() string

		// ClusterName is the full cluster name: app-stack-detail
		ClusterName() string

		// ASGName is the name of the ASG associated with the instance
		ASGName() string

		// ID is the instance ID, e.g. i-dbcba24c
		ID() string

		// CloudProvider returns the cloud provider (e.g., "aws")
		CloudProvider() string
	}

	// Termination contains information about an instance termination.
	Termination struct {
		Instance Instance  // The instance that will be terminated
		Time     time.Time // Termination time
		Leashed  bool      // If true, track the termination but do not execute it
	}

	// Tracker records termination events an a tracking system such as Chronos
	Tracker interface {
		// Track pushes a termination event to the tracking system
		Track(t Termination) error
	}

	// ErrorCounter counts when errors occur.
	ErrorCounter interface {
		Increment() error
	}

	// Decryptor decrypts encrypted text. It is used for decrypting
	// sensitive credentials that are stored encrypted
	Decryptor interface {
		Decrypt(ciphertext string) (string, error)
	}

	// Env provides information about the environment that Chaos Monkey has been
	// deployed to.
	Env interface {
		// InTest returns true if Chaos Monkey is running in a test environment
		InTest() bool
	}

	// AppConfigGetter retrieves App configuration info
	AppConfigGetter interface {
		// Get returns the App config info by app name
		Get(app string) (*AppConfig, error)
	}

	// Checker checks to see if a termination is permitted given min time between terminations
	//
	// if the termination is permitted, returns (true, nil)
	// otherwise, returns false with an error
	//
	// Returns ErrViolatesMinTime if violates min time between terminations
	//
	// Note that this call may change the state of the server: if the checker returns true, the termination will be recorded.
	Checker interface {
		// Check checks if a termination is permitted and, if so, records the
		// termination time on the server.
		// The endHour (hour time when Chaos Monkey stops killing) is in the
		// time zone specified by loc.
		Check(term Termination, appCfg AppConfig, endHour int, loc *time.Location) error
	}

	// Terminator provides an interface for killing instances
	Terminator interface {
		// Kill terminates a running instance
		Execute(trm Termination) error
	}

	// Outage provides an interface for checking if there is currently an outage
	// This provides a mechanism to check if there's an ongoing outage, since
	// Chaos Monkey doesn't run during outages
	Outage interface {
		// Outage returns true if there is an ongoing outage
		Outage() (bool, error)
	}

	// ErrViolatesMinTime represents an error when trying to record a termination
	// that violates the min time between terminations for that particular app
	ErrViolatesMinTime struct {
		InstanceID string         // the most recent terminated instance id
		KilledAt   time.Time      // the time that the most recent instance was terminated
		Loc        *time.Location // local time zone location
	}
)

// String returns a string representation for a Group
func (g Group) String() string {
	switch g {
	case App:
		return "app"
	case Stack:
		return "stack"
	case Cluster:
		return "cluster"
	}

	panic("Unknown Group value")
}

// NewAppConfig constructs a new app configuration with reasonable defaults
// with specified accounts enabled/disabled
func NewAppConfig(exceptions []Exception) AppConfig {
	result := AppConfig{
		Enabled:                        true,
		RegionsAreIndependent:          true,
		MeanTimeBetweenKillsInWorkDays: 5,
		Grouping:                       Cluster,
		Exceptions:                     exceptions,
	}

	return result
}

// Matches returns true if an exception matches an ASG
func (ex Exception) Matches(account, stack, detail, region string) bool {
	return exFieldMatches(ex.Account, account) &&
		exFieldMatches(ex.Stack, stack) &&
		exFieldMatches(ex.Detail, detail) &&
		exFieldMatches(ex.Region, region)
}

// exFieldMatches checks if an exception field matches a given value
// It's true if field is "*" or if the field is the same string as the value
func exFieldMatches(field, value string) bool {
	return field == "*" || field == value
}

func (e ErrViolatesMinTime) Error() string {
	s := fmt.Sprintf("Would violate min between kills: instance %s was killed at %s", e.InstanceID, e.KilledAt)

	// If we know the time zone, report that as well
	if e.Loc != nil {
		s += fmt.Sprintf(" (%s)", e.KilledAt.In(e.Loc))
	}

	return s
}
