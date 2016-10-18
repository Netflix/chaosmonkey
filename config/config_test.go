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

import "testing"

func TestGetStringSlice(t *testing.T) {
	cfg := Defaults()
	cfg.Set("myparam1", `["foo", "bar"]`)
	cfg.Set("myparam2", []string{"foo", "bar"})
	cfg.Set("myparam3", []interface{}{interface{}("foo"), interface{}("bar")})

	for _, param := range []string{"myparam1", "myparam2", "myparam3"} {
		got, err := cfg.getStringSlice(param)
		if err != nil {
			t.Error(err)
		}

		if len(got) != 2 || got[0] != "foo" || got[1] != "bar" {
			t.Errorf(`param %s, got %+v want ["foo", "bar"]`, param, got)
		}
	}

}
