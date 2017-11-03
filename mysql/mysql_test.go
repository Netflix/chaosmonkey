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

// +build docker

// The tests in this package use docker to test against a mysql:5.6 database
// By default, the tests are off unless you pass the "-tags docker" flag
// when running the test.
//
// By default, TestMain starts up a new mysql Docker container. However, if you
// already have a mysql docker container running, you can skip this by also
// passing the "dockerup" flag: -tags "docker dockerup"

package mysql_test

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/Netflix/chaosmonkey/mysql"
	"github.com/pkg/errors"
)

var (
	dbName   string = "chaosmonkey"
	password string = "password"
	port     int    = 3306
)

// inUse returns true if port accepts connections on localhsot
func inUse(port int) bool {
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}

	conn.Close()
	return true
}

func TestMain(m *testing.M) {

	//
	// Setup
	//

	var alwaysUp bool
	flag.BoolVar(&alwaysUp, "dockerup", false, "if true, won't start docker")
	flag.Parse()

	var cmd *exec.Cmd
	var err error

	if !alwaysUp {
		// Check to make sure the port isn't already in use
		if inUse(port) {
			panic(fmt.Sprintf("can't start mysql container: port %d currently in use", port))
		}
		cmd, err = startMySQLContainer()
		if err != nil {
			panic(err)
		}
	}

	//
	// Run tests
	//

	r := m.Run()

	//
	// Cleanup
	//

	if !alwaysUp {
		// Send a SIGTERM once we're done so mysql container shuts down
		cmd.Process.Signal(syscall.SIGTERM)

		// Wait for container to finish shutting down
		cmd.Wait()
	}

	os.Exit(r)
}

// startMySQLContainer starts a MySQL docker container
// Returns the Cmd object associated with the process
func startMySQLContainer() (*exec.Cmd, error) {
	cmd := exec.Command("docker", "run", "-e", "MYSQL_ROOT_PASSWORD="+password, fmt.Sprintf("-p3306:%d", port), "mysql:5.6")
	pipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	ch := make(chan int)

	readyString := "mysqld: ready for connections"

	go func() {
		reader := bufio.NewReader(pipe)

		// We loop until we see mysqld: ready for connections
		var s string

		for !strings.Contains(s, readyString) {
			s, err = reader.ReadString('\n')
			if err != nil {
				return nil, err
			}
			fmt.Print(s)
		}

		ch <- 0

	}()

	select {
	case <-ch:
		// noting to do
	case <-time.After(time.Second * 30):
		// timeout.
		return nil, errors.Errorf(`never saw "%s". (mysql container needs manual cleanup)`, readyString)
	}

	fmt.Println("Sleeping for 5 seconds")
	time.Sleep(5 * time.Second)

	return cmd, nil
}

// initDB initializes the "chaosmonkey" database with the chaosmonkey schemas
// It wipes out any existing database database with the same name
func initDB() error {
	db, err := sql.Open("mysql", fmt.Sprintf("root:%s@tcp(127.0.0.1:%d)/", password, port))
	if err != nil {
		return errors.Wrap(err, "sql.Open failed")
	}
	defer db.Close()

	_, err = db.Exec("DROP DATABASE IF EXISTS " + dbName)
	if err != nil {
		return errors.Wrap(err, "drop database failed")
	}

	_, err = db.Exec("CREATE DATABASE " + dbName)
	if err != nil {
		return errors.Wrap(err, "create database failed")
	}

	mysqlDb, dbErr := mysql.New("127.0.0.1", port, "root", password, dbName)
	if dbErr != nil {
		return errors.Wrap(err, "mysql.New failed")
	}
	defer mysqlDb.Close()

	// Get the "terminations" schema

	err = mysql.Migrate(mysqlDb)
	if err != nil {
		return errors.Wrap(err, "database migration failed")
	}

	return nil

}

func stopMySQLContainer(name string, t *testing.T) {

	// Dump the output just in case
	cmd := exec.Command("docker", "logs", name)
	data, _ := cmd.CombinedOutput()
	t.Log(string(data))

	cmd = exec.Command("docker", "kill", name)
	data, err := cmd.CombinedOutput()
	s := string(data)
	if err != nil {
		panic(fmt.Sprintf("docker kill errored (%v) with output: %s", err, s))
	}

	cmd = exec.Command("docker", "rm", name)
	data, err = cmd.CombinedOutput()
	s = string(data)
	if err != nil {
		panic(fmt.Sprintf("docker kill errored (%v) with output: %s", err, s))
	}
}
