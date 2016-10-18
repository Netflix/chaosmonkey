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

// Package clock provides the Clock interface for getting the current time
package clock

import "time"

// Clock provides an interface to the current time, useful for testing
type Clock interface {
	// Now returns the current time
	Now() time.Time
}

// New returns an implementation of Clock that uses the system time
func New() Clock {
	return SystemClock{}
}

// SystemClock uses the system clock to return the time
type SystemClock struct{}

// Now implements Clock.Now
func (cl SystemClock) Now() time.Time {
	return time.Now()
}
