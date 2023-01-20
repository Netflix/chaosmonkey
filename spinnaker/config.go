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

package spinnaker

import (
	"io/ioutil"
	"net/http"

	"github.com/Netflix/chaosmonkey/v2"

	"github.com/pkg/errors"
)

// Get implements chaosmonkey.Getter.Get
func (s Spinnaker) Get(app string) (c *chaosmonkey.AppConfig, err error) {
	// avoid expanding the response to avoid unneeded load
	url := s.appURL(app) + "?expand=false"
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, errors.Wrapf(err, "http get failed at %s", url)
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = errors.Wrapf(err, "body close failed at %s", url)
		}
	}()

	// should return a 200
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected response code (%d) from %s", resp.StatusCode, url)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "body read failed at %s", url)
	}

	return fromJSON(body)
}
