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
	"github.com/Netflix/chaosmonkey/v2/mysql"
	"io/ioutil"
	"log"
	"os"
)

const (
	scheduleCommand  = "schedule"
	terminateCommand = "terminate"
	scriptContent    = `#!/bin/bash
%s %s "$@" >> %s/chaosmonkey-%s.log 2>&1
`
)

// CurrentExecutable provides an interface to extract information about the current executable
type CurrentExecutable interface {
	// ExecutablePath returns the path to current executable
	ExecutablePath() (string, error)
}

// Install installs chaosmonkey and runs database migration
func Install(cfg *config.Monkey, exec CurrentExecutable, db mysql.MySQL) {
	InstallCron(cfg, exec)
	Migrate(db)
	log.Println("installation done!")
}

// InstallCron installs chaosmonkey schedule generation cron
func InstallCron(cfg *config.Monkey, exec CurrentExecutable) {
	executablePath, err := exec.ExecutablePath()
	if err != nil {
		log.Fatalf("FATAL: %v", err)
	}
	err = setupTerminationScript(cfg, executablePath)
	if err != nil {
		log.Fatalf("FATAL: %v", err)
	}

	err = setupCron(cfg, executablePath)
	if err != nil {
		log.Fatalf("FATAL: %v", err)
	}

	log.Println("chaosmonkey cron is installed successfully")
}

func setupCron(cfg *config.Monkey, executablePath string) error {
	err := EnsureFileAbsent(cfg.SchedulePath())
	if err != nil {
		return err
	}

	err = EnsureFileAbsent(cfg.ScheduleCronPath())
	if err != nil {
		return err
	}

	var scriptPerms os.FileMode = 0755 // -rwx-rx--rx-- : scripts should be executable
	log.Printf("Creating %s\n", cfg.SchedulePath())

	content, err := generateScriptContent(scheduleCommand, cfg, executablePath)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(cfg.SchedulePath(), content, scriptPerms)
	if err != nil {
		return err
	}

	cronExpr, err := cfg.CronExpression()
	if err != nil {
		return err
	}

	crontab := fmt.Sprintf("%s %s %s\n", cronExpr, cfg.TermAccount(), cfg.SchedulePath())
	var cronPerms os.FileMode = 0644 // -rw-r--r-- : cron config file shouldn't have write perm
	log.Printf("Creating %s\n", cfg.ScheduleCronPath())
	err = ioutil.WriteFile(cfg.ScheduleCronPath(), []byte(crontab), cronPerms)
	return err
}

func setupTerminationScript(cfg *config.Monkey, executablePath string) error {
	err := EnsureFileAbsent(cfg.TermPath())
	if err != nil {
		return err
	}

	var perms os.FileMode = 0755 // -rwx-rx--rx-- : scripts should be executable
	log.Printf("Creating %s\n", cfg.TermPath())

	content, err := generateScriptContent(terminateCommand, cfg, executablePath)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(cfg.TermPath(), content, perms)
	return err
}

func generateScriptContent(cmdName string, cfg *config.Monkey, executablePath string) ([]byte, error) {
	content := fmt.Sprintf(scriptContent, executablePath, cmdName, cfg.LogPath(), cmdName)
	return []byte(content), nil
}
