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

/*
Chaos Monkey randomly terminates instances.
*/
package main

import (
	"github.com/netflix/chaosmonkey/command"

	// These are anonymous imported so that the related Get* methods (e.g.,
	// GetDecryptor) are picked up.

	_ "github.com/netflix/chaosmonkey/decryptor"
	_ "github.com/netflix/chaosmonkey/env"
	_ "github.com/netflix/chaosmonkey/errorcounter"
	_ "github.com/netflix/chaosmonkey/outage"
	_ "github.com/netflix/chaosmonkey/tracker"
)

func main() {
	command.Execute()
}
