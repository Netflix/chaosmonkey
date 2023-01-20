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
	"github.com/Netflix/chaosmonkey/v2/config/param"
	"github.com/Netflix/chaosmonkey/v2/mock"
	"github.com/pkg/errors"
	"io/ioutil"
	"testing"
)

func assertHasSameContent(fileName string, expectedContent string) error {

	cronContent, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	actualContent := string(cronContent)
	if actualContent != expectedContent {
		return errors.Errorf("\nFile : %s\nExpected:\n%s\nActual:\n%s", fileName, expectedContent, actualContent)
	}
	return nil
}

func initInstallationConfig(script string, cron string, log string, term string) (*config.Monkey, error) {
	defaultConfig := config.Defaults()
	defaultConfig.Set(param.SchedulePath, script)
	defaultConfig.Set(param.ScheduleCronPath, cron)
	defaultConfig.Set(param.LogPath, log)
	defaultConfig.Set(param.StartHour, 9)
	defaultConfig.Set(param.TermAccount, "root")
	defaultConfig.Set(param.TermPath, term)
	return defaultConfig, nil
}

func TestInstallationWithDefaultCron(t *testing.T) {
	scriptPath := "/tmp/chaosmonkey-schedule.sh"
	termPath := "/tmp/chaosmonkey-terminate.sh"
	cronPath := "/tmp/chaosmonkey-schedule"
	execPath := "/tmp/chaosmonkey"
	logPath := "/var/log"

	defaultConfig, err := initInstallationConfig(scriptPath, cronPath, logPath, termPath)
	if err != nil {
		t.Error(err.Error())
		return
	}

	executable := mock.Executable{Path: execPath}
	InstallCron(defaultConfig, executable)

	expectedCron := fmt.Sprintf("0 7 * * 1-5 root %s\n", scriptPath)
	err = assertHasSameContent(cronPath, expectedCron)
	if err != nil {
		t.Error(err.Error())
		return
	}

	expectedScript := fmt.Sprintf(`#!/bin/bash
%s %s "$@" >> %s/chaosmonkey-%s.log 2>&1
`, execPath, "schedule", logPath, "schedule")
	err = assertHasSameContent(scriptPath, expectedScript)
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestInstallationWithUserDefinedCron(t *testing.T) {
	scriptPath := "/tmp/chaosmonkey-schedule.sh"
	termPath := "/tmp/chaosmonkey-terminate.sh"
	cronPath := "/tmp/chaosmonkey-schedule"
	execPath := "/tmp/chaosmonkey"
	logPath := "/var/log"
	userDefinedCron := "0 15 * * 1-5"

	defaultConfig, err := initInstallationConfig(scriptPath, cronPath, logPath, termPath)
	defaultConfig.Set(param.CronExpression, userDefinedCron)
	if err != nil {
		t.Error(err.Error())
		return
	}

	executable := mock.Executable{Path: execPath}
	InstallCron(defaultConfig, executable)

	expectedCron := fmt.Sprintf("%s root %s\n", userDefinedCron, scriptPath)
	err = assertHasSameContent(cronPath, expectedCron)
	if err != nil {
		t.Error(err.Error())
		return
	}

	expectedScript := fmt.Sprintf(`#!/bin/bash
%s %s "$@" >> %s/chaosmonkey-%s.log 2>&1
`, execPath, "schedule", logPath, "schedule")
	err = assertHasSameContent(scriptPath, expectedScript)
	if err != nil {
		t.Error(err.Error())
		return
	}
}
