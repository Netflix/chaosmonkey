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

// Package schedule implements a schedule of terminations
package schedule

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"time"

	"github.com/Netflix/chaosmonkey/v2"
	"github.com/Netflix/chaosmonkey/v2/config"
	"github.com/Netflix/chaosmonkey/v2/deploy"
	"github.com/Netflix/chaosmonkey/v2/grp"
)

// Populate populates the termination schedule with the random
// terminations for a list of apps. If the specified list of apps is empty,
// then it will
func (s *Schedule) Populate(d deploy.Deployment, getter chaosmonkey.AppConfigGetter, chaosConfig *config.Monkey, apps []string) error {
	c := make(chan *deploy.App)

	// If the caller explicitly a set of apps, use those
	// If they did not, do all apps
	if len(apps) == 0 {
		var err error
		apps, err = d.AppNames()
		if err != nil {
			return fmt.Errorf("could not retrieve list of apps: %v", err)
		}
	}

	go d.Apps(c, apps)
	i := 0 // number of apps already processed
	for app := range c {
		if i >= chaosConfig.MaxApps() {
			break
		}

		i++

		cfg, err := getter.Get(app.Name())

		if err != nil {
			log.Printf("WARNING: Could not retrieve config for app=%s. %s", app.Name(), err)
			continue
		}
		doScheduleApp(s, app, *cfg, chaosConfig)
	}

	return nil
}

// Add schedules a termination for group at time tm
func (s *Schedule) Add(tm time.Time, group grp.InstanceGroup) {
	s.entries = append(s.entries, Entry{Group: group, Time: tm})
}

// Entries returns the list of schedule entries
func (s *Schedule) Entries() []Entry {
	return s.entries
}

// doScheduleApp populates the termination schedule for one app
func doScheduleApp(schedule *Schedule, app *deploy.App, cfg chaosmonkey.AppConfig, chaosConfig *config.Monkey) {

	if !cfg.Enabled {
		log.Printf("app=%s disabled\n", app.Name())
		return
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	startHour := chaosConfig.StartHour()
	endHour := chaosConfig.EndHour()
	location, err := chaosConfig.Location()

	if err != nil {
		panic(fmt.Sprintf("Could not get Location for time zone calculation: %s", err.Error()))
	}

	groups := app.EligibleInstanceGroups(cfg)

	if len(groups) == 0 {
		log.Printf("app=%s no eligible instance groups", app.Name())
	}

	for _, group := range groups {
		kill := shouldKillInstance(cfg.MeanTimeBetweenKillsInWorkDays, r)
		log.Printf("%s mtbk=%d kill=%t\n", grp.String(group), cfg.MeanTimeBetweenKillsInWorkDays, kill)
		if kill {
			time := chooseTerminationTime(time.Now(), startHour, endHour, location)
			schedule.Add(time, group)
		}
	}
}

// chooseTerminationTime Randomly selects a time to terminate an instance
// on the same date as now, between startHour:00 and endHour:00 in the same
// timezone as location
// Panics if endHour <= startHour
//
// Note that there is no guarantee that the selected termination time will be in
// the future
//
// now is passed as an argument to simplify testing
func chooseTerminationTime(now time.Time, startHour int, endHour int, location *time.Location) time.Time {
	if endHour <= startHour {
		panic(fmt.Sprintf("ChooseTermination called with startHour <= endHour, startHour: %d. endHour: %d", startHour, endHour))
	}

	// Compute the number of minutes in the interval between start and end,
	// pick a random one in there, and then add it to the start time as an
	// offset
	minutesInTimeInterval := (endHour - startHour) * 60
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	sample := r.Intn(minutesInTimeInterval)

	// Convert the sample to duration in minutes
	offset := time.Duration(sample) * time.Minute

	year, month, day := now.Date()
	startTime := time.Date(year, month, day, startHour, 0, 0, 0, location)

	return startTime.Add(offset)
}

// float64Rand generates random floats on [0, 1)
type float64Rand interface {

	// Return a random float64 on [0, 1)
	Float64() float64
}

// ShouldKillInstance randomly determines whether an instance should
// be terminated today by flipping a biased coin.
//
// It uses the meanTimeBetweenKillsInWorkDays to determine the probability
// of a kill
func shouldKillInstance(meanTimeBetweenKillsInWorkDays int, r float64Rand) bool {

	if meanTimeBetweenKillsInWorkDays <= 0 {
		panic("meanTimeBetweenKillsInWorkDays is zero or negative")
	}

	var pkill = 1.0 / float64(meanTimeBetweenKillsInWorkDays)

	// Sample uniformly over [0,1)
	sample := r.Float64()

	return pkill >= sample

}

// Entry is an entry a termination schedule.
// It contains the instance group that the terminator will randomly select from
// as well as the time of termination.
type Entry struct {
	Group grp.InstanceGroup `json:"group"`
	Time  time.Time         `json:"time"`
}

// apiGroup represents group representation passed by the API
type apiGroup struct {
	App, Account, Region, Stack, Cluster string
}

// UnmarshalJSON implements Unmarshaler.UnmarshalJSON
func (e *Entry) UnmarshalJSON(b []byte) (err error) {

	var ce struct {
		Group apiGroup
		Time  time.Time
	}

	err = json.Unmarshal(b, &ce)
	if err != nil {
		return err
	}

	g := &ce.Group
	e.Group = grp.New(g.App, g.Account, g.Region, g.Stack, g.Cluster)
	e.Time = ce.Time
	return nil

}

// Equal checks that two entries are equal
func (e *Entry) Equal(o *Entry) bool {
	return grp.Equal(e.Group, o.Group) && e.Time.Equal(o.Time)
}

// Crontab returns a termination command for the Entry, in crontab format.
// It takes as arguments:
//   - the path to the termination executable
//   - the account that should execute the job
//
// The returned string is not terminated by a newline.
func (e *Entry) Crontab(termPath, account string) string {
	// From https://en.wikipedia.org/wiki/Cron
	// # * * * * *  account command to execute
	// # │ │ │ │ │
	// # │ │ │ │ │
	// # │ │ │ │ └───── day of week (0 - 6) (0 to 6 are Sunday to Saturday, or use names; 7 is Sunday, the same as 0)
	// # │ │ │ └────────── month (1 - 12)
	// # │ │ └─────────────── day of month (1 - 31)
	// # │ └──────────────────── hour (0 - 23)
	// # └───────────────────────── min (0 - 59)
	t := e.Time.UTC()
	return fmt.Sprintf("%d %d %d %d %d %s %s", t.Minute(), t.Hour(), t.Day(), t.Month(), t.Weekday(), account, terminateCommand(termPath, e.Group))
}

// terminateCommand returns the string for terminating an instance
// given the path to the chaosmonkey termination executable and an instance to terminate
func terminateCommand(termPath string, group grp.InstanceGroup) string {
	cmd := fmt.Sprintf("%s %s %s", termPath, group.App(), group.Account())
	if cluster, ok := group.Cluster(); ok {
		cmd = fmt.Sprintf("%s --cluster=%s", cmd, cluster)
	}

	if stack, ok := group.Stack(); ok {
		cmd = fmt.Sprintf("%s --stack=%s", cmd, stack)
	}

	if region, ok := group.Region(); ok {
		cmd = fmt.Sprintf("%s --region=%s", cmd, region)
	}

	return cmd
}

// logRedirect returns a string to append to a shell command so it redirects
// stdout and stderr to a logfile
// Example output: ">> /path/to/log 2>&1"
func logRedirect(logPath string) string {
	return fmt.Sprintf(">> %s 2>&1", logPath)
}

// Schedule is a collection of termination entries.
type Schedule struct {
	entries []Entry
}

// New returns a new Schedule
func New() *Schedule {
	return &Schedule{
		// We need a zero-element slice instead of a nil slice so that
		// it will JSON-marshall into '[ ]' instead of 'null'
		make([]Entry, 0),
	}
}

// ByTime implements sort.Interface for []Entry based on the time field
type ByTime []Entry

func (t ByTime) Len() int           { return len(t) }
func (t ByTime) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t ByTime) Less(i, j int) bool { return t[i].Time.Before(t[j].Time) }

// Crontab returns a schedule of termination commands in crontab format
// It takes as arguments:
//   - the path to the executable that terminates an instance
//   - the account that should execute the job
func (s Schedule) Crontab(exPath string, account string) []byte {
	var result bytes.Buffer

	// In-place sort the entries before generating the table
	sort.Sort(ByTime(s.entries))

	for _, entry := range s.entries {
		_, err := result.WriteString(entry.Crontab(exPath, account))
		if err != nil {
			panic(fmt.Sprintf("Could not generate string with crontab: %s", err.Error()))
		}
		_, err = result.WriteString("\n")
		if err != nil {
			panic(fmt.Sprintf("Could not generate string with crontab: %s", err.Error()))
		}

	}
	return result.Bytes()
}

// MarshalJSON implements Marshaler.MarshalJSON
func (s Schedule) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.entries)
}

// UnmarshalJSON implements Unmarshaler.UnmarshalJSON
func (s *Schedule) UnmarshalJSON(b []byte) (err error) {
	return json.Unmarshal(b, &s.entries)
}
