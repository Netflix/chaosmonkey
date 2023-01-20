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

package mysql

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"

	"github.com/Netflix/chaosmonkey/v2"
	"github.com/Netflix/chaosmonkey/v2/cal"
	"github.com/Netflix/chaosmonkey/v2/config"
	"github.com/Netflix/chaosmonkey/v2/config/param"
	"github.com/Netflix/chaosmonkey/v2/deps"
	"github.com/Netflix/chaosmonkey/v2/grp"
	"github.com/Netflix/chaosmonkey/v2/migration"
	"github.com/Netflix/chaosmonkey/v2/schedstore"
	"github.com/Netflix/chaosmonkey/v2/schedule"
	"github.com/rubenv/sql-migrate"
	"log"
)

// MySQL represents a MySQL-backed store for schedules and terminations
type MySQL struct {
	db *sql.DB
}

// TxDeadlock returns true if the error is because of a transaction deadlock
func TxDeadlock(err error) bool {
	switch err := errors.Cause(err).(type) {
	case *mysql.MySQLError:
		// ER_LOCK_DEADLOCK
		// See: https://dev.mysql.com/doc/refman/5.6/en/error-messages-server.html
		return err.Number == 1213
	default:
		return false
	}
}

// ViolatesMinTime returns true if the error violates min time between
// terminations
func ViolatesMinTime(err error) bool {
	_, ok := errors.Cause(err).(chaosmonkey.ErrViolatesMinTime)
	return ok
}

// NewFromConfig creates a new MySQL taking config parameters from cfg
func NewFromConfig(cfg *config.Monkey) (MySQL, error) {

	if cfg.DatabaseHost() == "" {
		return MySQL{}, errors.Errorf("%s not specified", param.DatabaseHost)
	}

	encryptedPassword := cfg.DatabaseEncryptedPassword()

	decryptor, err := deps.GetDecryptor(cfg)
	if err != nil {
		return MySQL{}, err
	}

	password, err := decryptor.Decrypt(encryptedPassword)
	if err != nil {
		return MySQL{}, err
	}

	return New(cfg.DatabaseHost(), cfg.DatabasePort(), cfg.DatabaseUser(), password, cfg.DatabaseName())
}

// New creates a new MySQL
func New(host string, port int, user string, password string, dbname string) (MySQL, error) {
	db, err := sql.Open("mysql", dsn(host, port, user, password, dbname))
	if err != nil {
		return MySQL{}, errors.Wrap(err, "sql.Open failed")
	}

	return MySQL{db}, nil
}

// Close closes the underlying sql.DB
func (m MySQL) Close() error {
	return m.db.Close()
}

// utcDate takes a time.Time in a local time zone and returns a time.Time
// that has the same year/month/day as date, but is in UTC, at 12 PM
// We use this to work with MySQL DATE entries without having to worry about
// MySQL changing the value due to time conversion
func utcDate(date time.Time) time.Time {
	year, month, day := date.Date()
	return time.Date(year, month, day, 12, 0, 0, 0, time.UTC)
}

// Retrieve  retrieves the schedule for the given date
func (m MySQL) Retrieve(date time.Time) (sched *schedule.Schedule, err error) {
	rows, err := m.db.Query("SELECT time, app, account, region, stack, cluster FROM schedules WHERE date = DATE(?)", utcDate(date))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve schedule for %s", date)
	}

	sched = schedule.New()

	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = errors.Wrap(cerr, "rows.Close() failed")
		}
	}()

	for rows.Next() {
		var tm time.Time
		var app, account, region, stack, cluster string

		err = rows.Scan(&tm, &app, &account, &region, &stack, &cluster)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		sched.Add(tm, grp.New(app, account, region, stack, cluster))
	}

	err = rows.Err()
	if err != nil {
		return nil, errors.Wrap(err, "rows.Err() errored")
	}

	return sched, nil

}

// Publish publishes the schedule for the given date
func (m MySQL) Publish(date time.Time, sched *schedule.Schedule) error {
	return m.PublishWithDelay(date, sched, 0)
}

// PublishWithDelay publishes the schedule with a delay between checking the schedule
// exists and writing it. The delay is used only for testing race conditions
func (m MySQL) PublishWithDelay(date time.Time, sched *schedule.Schedule, delay time.Duration) (err error) {
	// First, we check to see if there is a schedule present
	tx, err := m.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	// We must either commit or rollback at the end
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		case schedstore.ErrAlreadyExists:
			// We want to return ErrAlreadyExists even if the transaction commit
			// fails
			_ = tx.Commit()
		default:
			_ = tx.Rollback()
		}
	}()

	exists, err := schedExists(tx, date)
	if err != nil {
		return err
	}

	if exists {
		return schedstore.ErrAlreadyExists
	}

	if delay > 0 {
		time.Sleep(delay)
	}
	query := "INSERT INTO schedules (date, time, app, account, region, stack, cluster) VALUES (?, ?, ?, ?, ?, ?, ?)"
	stmt, err := tx.Prepare(query)
	if err != nil {
		return errors.Wrapf(err, "failed to prepare sql statement: %s", query)
	}

	for _, entry := range sched.Entries() {
		var app, account, region, stack, cluster string
		app = entry.Group.App()
		account = entry.Group.Account()
		if val, ok := entry.Group.Region(); ok {
			region = val
		}
		if val, ok := entry.Group.Stack(); ok {
			stack = val
		}
		if val, ok := entry.Group.Cluster(); ok {
			cluster = val
		}

		_, err = stmt.Exec(utcDate(date), entry.Time.In(time.UTC), app, account, region, stack, cluster)
		if err != nil {
			return errors.Wrapf(err, "failed to execute prepared query")
		}
	}

	return nil
}

// schedExists returns true if a schedule has previously been
// published for this date
func schedExists(tx *sql.Tx, date time.Time) (result bool, err error) {
	rows, err := tx.Query("SELECT COUNT(*) FROM schedules WHERE date = DATE(?)", date)
	if err != nil {
		return false, errors.Wrapf(err, "failed to check if schedule exists for %s", date)
	}

	var count int

	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = errors.Wrap(err, "rows.Close() failed")
		}
	}()

	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return false, errors.Wrap(err, "failed to scan row")
		}
	}

	return (count > 0), nil

}

// dsn returns a MySQL TCP connection string (data source name)
// See: https://github.com/go-sql-driver/mysql#dsn-data-source-name
func dsn(host string, port int, user string, password string, dbname string) string {
	params := map[string]string{
		"tx_isolation": "SERIALIZABLE", // we need serializable transactions for atomic test & set behavior
		"parseTime":    "true",         // enable us to use sql.Rows.Scan to read time.Time objects from queries
		"loc":          "UTC",          // Scan'd time.Times should be treated as being in UTC time zone
		"time_zone":    "UTC",          // MySQL should interpret DATETIME values as being in UTC
	}

	var ss []string

	for k, v := range params {
		ss = append(ss, fmt.Sprintf("%s=%s", k, v))
	}

	query := strings.Join(ss, "&")

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", user, password, host, port, dbname, query)
}

// Check checks if a termination is permitted and, if so, records the
// termination time on the server
func (m MySQL) Check(term chaosmonkey.Termination, appCfg chaosmonkey.AppConfig, endHour int, loc *time.Location) error {
	return m.CheckWithDelay(term, appCfg, endHour, loc, 0)
}

// CheckWithDelay is the same as Check, but adds a delay between reading and
// writing to the database (used for testing only)
func (m MySQL) CheckWithDelay(term chaosmonkey.Termination, appCfg chaosmonkey.AppConfig, endHour int, loc *time.Location, delay time.Duration) error {
	tx, err := m.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			_ = tx.Rollback()
		}
	}()

	err = respectsMinTimeBetweenKills(tx, term.Time, term, appCfg, endHour, loc)
	if err != nil {
		return err
	}

	if delay > 0 {
		time.Sleep(delay)
	}

	err = recordTermination(tx, term, loc)
	return err

}

// respectsMinTimeBetweenKills checks if this termination will respect or
// violate the min time between kills value. If this termination is too close
// to the most recent one, this will return an error.
// If this termination would violate the min time, returns an ErrViolatesMinTime
func respectsMinTimeBetweenKills(tx *sql.Tx, now time.Time, term chaosmonkey.Termination, appCfg chaosmonkey.AppConfig, endHour int, loc *time.Location) (err error) {
	app := term.Instance.AppName()
	account := term.Instance.AccountName()
	threshold, err := noKillsSince(appCfg.MinTimeBetweenKillsInWorkDays, now, endHour, loc)
	if err != nil {
		return err
	}
	query := "SELECT instance_id, killed_at FROM terminations WHERE app = ? AND account = ? AND killed_at >= ?"

	var rows *sql.Rows

	args := []interface{}{app, account, threshold.In(time.UTC)}

	switch appCfg.Grouping {
	case chaosmonkey.App:
		// nothing to do
	case chaosmonkey.Stack:
		query += " AND stack = ?"
		args = append(args, term.Instance.StackName())
	case chaosmonkey.Cluster:
		query += " AND cluster = ?"
		args = append(args, term.Instance.ClusterName())
	default:
		return errors.Errorf("unknown group: %v", appCfg.Grouping)
	}

	if appCfg.RegionsAreIndependent {
		query += " AND region = ?"
		args = append(args, term.Instance.RegionName())
	}

	// For unleashed (real) terminations, we only care about previous
	// terminations that were also unleashed. That's because a previous
	// leashed termination wasn't a real one, so that wouldn't violate
	// the min time between terminations
	if !term.Leashed {
		query += " AND leashed = FALSE"
	}

	// We need at most one entry
	query += " LIMIT 1"

	rows, err = tx.Query(query, args...)

	if err != nil {
		return err
	}

	defer func() {
		cerr := rows.Close()
		if err == nil && cerr != nil {
			err = cerr
		}
	}()

	if rows.Next() {
		var instanceID string
		var killedAt time.Time
		err = rows.Scan(&instanceID, &killedAt)
		return chaosmonkey.ErrViolatesMinTime{InstanceID: instanceID, KilledAt: killedAt, Loc: loc}
	}

	return nil
}

// noKillsSince computes the date of the most recent kill
// that conforms to the min time between kills specified
// by days
//
// Note that the calculation is min time in work days, so it does not count weekends.
//
// chrono is an interface for returning the current time
// endHour is the hour of the end of a workday in 24-hour time. For example, if
// workday ends at 5PM, this would be 17
// loc is the location that corresponds to endHour, e.g. America/Los_Angeles for PST
//
// # The returned time will be in UTC
//
// If days=1, then we allow
// kills each day, so the most recent kill will be at the
// end of the previous workday. For example:
//
//	days: 1
//	endHour: 17 (i.e. work day ends at 5PM local time)
//	loc:  America/Los_Angeles (PST)
//	chrono.Now(): Wed, Dec. 16, 2015 2:30 PM PST
//	Output: Tue, Dec. 15, 2015 5:00 PM PST
//
// If days=0, returns the current date, with
// the time set to endHour. For example:
//
//	days: 0
//	endHour: 17 (i.e. work day ends at 5PM local time)
//	loc:  America/Los_Angeles (PST)
//	chrono.Now(): Wed, Dec. 16, 2015 2:30 PM PST
//	Output: Wed, Dec. 16, 2015 5:00 PM PST
//
// noKillsSince returns the a datetime that is the last allowed time that a kill
// is permitted to have happened.
func noKillsSince(days int, now time.Time, endHour int, loc *time.Location) (time.Time, error) {
	if days < 0 {
		return time.Time{}, errors.Errorf("noKillsSince passed illegal input: days=%d", days)
	}

	oneDay := time.Hour * 24

	// Tail-recursive helper function reads clearer than writing a
	// traditional loop
	//
	// It expects a time localized to the zone associated with endHour because
	// workday and year-month-day values depend on the local timezone
	var helper func(N int, tInLoc time.Time) time.Time

	helper = func(N int, tInLoc time.Time) time.Time {
		switch {
		case !cal.IsWorkday(tInLoc):
			return helper(N, tInLoc.Add(-oneDay))
		case N == 0:
			return time.Date(tInLoc.Year(), tInLoc.Month(), tInLoc.Day(), endHour, 0, 0, 0, loc).UTC()
		default:
			return helper(N-1, tInLoc.Add(-oneDay))
		}
	}

	return helper(days, now.In(loc)), nil
}

func recordTermination(tx *sql.Tx, term chaosmonkey.Termination, loc *time.Location) (err error) {

	i := term.Instance

	_, err = tx.Exec("INSERT INTO terminations (app, account, stack, cluster, region, asg, instance_id, killed_at, leashed) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		i.AppName(), i.AccountName(), i.StackName(), i.ClusterName(), i.RegionName(), i.ASGName(), i.ID(), term.Time.In(time.UTC), term.Leashed)

	return err
}

var migrationSource = &migrate.AssetMigrationSource{
	Asset:    migration.Asset,
	AssetDir: migration.AssetDir,
	Dir:      "migration/mysql",
}

var databaseDialect = "mysql"

// Migrate upgrades a database to the latest database schema version.
func Migrate(mysqlDb MySQL) error {
	migrationCount, err := migrate.Exec(mysqlDb.db, databaseDialect, migrationSource, migrate.Up)
	if err != nil {
		return errors.Wrap(err, "database migration failed")
	}
	log.Println("Successfully applied database migrations. Number of migrations applied: ", migrationCount)

	return nil
}
