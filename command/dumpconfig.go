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
	"os"

	"github.com/Netflix/chaosmonkey/v2"
	"github.com/davecgh/go-spew/spew"
)

// DumpConfig dumps the config for an app to stdout
func DumpConfig(c chaosmonkey.AppConfigGetter, app string) {
	cfg, err := c.Get(app)
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}

	spew.Dump(cfg)
}
