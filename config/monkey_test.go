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

package config

import (
	"fmt"
	"github.com/Netflix/chaosmonkey/v2/config/param"
	"testing"
)

func TestDefaultCron(t *testing.T) {
	monkey := Defaults()
	monkey.Set(param.StartHour, 9)

	actual, err := monkey.CronExpression()
	if err != nil {
		t.Error(err.Error())
		return
	}

	expected := fmt.Sprintf("0 %d * * 1-5", 7)
	if actual != expected {
		t.Errorf("\nExpected:\n%s\nActual:\n%s", expected, actual)
	}
}

func TestDefaultCronForStartHourMidnight(t *testing.T) {
	monkey := Defaults()
	monkey.Set(param.StartHour, 0)

	actual, err := monkey.CronExpression()
	if err != nil {
		t.Error(err.Error())
		return
	}

	expected := fmt.Sprintf("0 %d * * 1-5", 22)
	if actual != expected {
		t.Errorf("\nExpected:\n%s\nActual:\n%s", expected, actual)
	}
}

func TestDefaultCronForStartHourOneAM(t *testing.T) {
	monkey := Defaults()
	monkey.Set(param.StartHour, 1)

	actual, err := monkey.CronExpression()
	if err != nil {
		t.Error(err.Error())
		return
	}

	expected := fmt.Sprintf("0 %d * * 1-5", 23)
	if actual != expected {
		t.Errorf("\nExpected:\n%s\nActual:\n%s", expected, actual)
	}
}

func TestDefaultCronForStartHourBeforeClockStart(t *testing.T) {
	monkey := Defaults()
	monkey.Set(param.StartHour, -1)

	_, err := monkey.CronExpression()
	if err == nil {
		t.Error("Expected InstalledCronExpression to return an error as start hour is before clock start hour")
		return
	}
}

func TestDefaultCronForStartHourAfterClockEnd(t *testing.T) {
	monkey := Defaults()
	monkey.Set(param.StartHour, 24)

	_, err := monkey.CronExpression()
	if err == nil {
		t.Error("Expected InstalledCronExpression to return an error as start hour is after clock end hour")
		return
	}
}
