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
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/pkg/errors"

	"github.com/Netflix/chaosmonkey"
)

const terminateType string = "terminateInstances"

type (
	// killPayload is the POST request body for Spinnaker instance terminations
	killPayload struct {
		Application string  `json:"application"`
		Description string  `json:"description"`
		Job         []kpJob `json:"job"`
	}

	// kpJob is the "job" of killPayload
	kpJob struct {
		User            string   `json:"user"`
		Type            string   `json:"type"`
		Credentials     string   `json:"credentials"`
		Region          string   `json:"region"`
		ServerGroupName string   `json:"serverGroupName"`
		InstanceIDs     []string `json:"instanceIds"`
		CloudProvider   string   `json:"cloudProvider"`
	}

	// fakeTerminator implements term.Terminator, but it just logs the http requests rather than actually
	// making them
	fakeTerminator struct{}
)

// NewFakeTerm returns a fake Terminator that prints out what API calls it would make against Spinnaker
func NewFakeTerm() chaosmonkey.Terminator {
	return fakeTerminator{}
}

// tasksURL returns the Spinnaker tasks URL associated with an app
func (s Spinnaker) tasksURL(appName string) string {
	return s.appURL(appName) + "/tasks"
}

// Kill implements term.Terminator.Kill
func (t fakeTerminator) Execute(trm chaosmonkey.Termination) error {
	return nil
}

// Execute implements term.Terminator.Execute
func (s Spinnaker) Execute(trm chaosmonkey.Termination) (err error) {
	ins := trm.Instance
	url := s.tasksURL(ins.AppName())

	otherID, err := s.OtherID(ins)
	if err != nil {
		return errors.Wrap(err, "retrieve other id failed")
	}

	payload := killJSONPayload(ins, otherID, s.user)
	resp, err := s.client.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("POST to %s failed, (body '%s')", url, string(payload)))
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = errors.Wrap(cerr, fmt.Sprintf("failed to close response body of %s", url))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Unexpected response: %d", resp.StatusCode)
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "failed to read response body")
		}
		return fmt.Errorf("unexpected response code: %d, body: %s", resp.StatusCode, string(contents))
	}

	return nil
}

// killJsonPayload generates the JSON request body for terminating an instance
// otherID is an optional second instance ID, as some backends may have a second
// identifer.
func killJSONPayload(ins chaosmonkey.Instance, otherID string, spinnakerUser string) []byte {
	var desc string
	if otherID != "" {
		desc = fmt.Sprintf("Chaos Monkey terminate instance: %s %s (%s, %s, %s)", ins.ID(), otherID, ins.AccountName(), ins.RegionName(), ins.ASGName())
	} else {
		desc = fmt.Sprintf("Chaos Monkey terminate instance: %s (%s, %s, %s)", ins.ID(), ins.AccountName(), ins.RegionName(), ins.ASGName())
	}

	p := killPayload{
		Application: ins.AppName(),
		Description: desc,
		Job: []kpJob{
			{
				User:            spinnakerUser,
				Type:            terminateType,
				Credentials:     ins.AccountName(),
				Region:          ins.RegionName(),
				ServerGroupName: ins.ASGName(),
				InstanceIDs:     []string{ins.ID()},
				CloudProvider:   ins.CloudProvider(),
			},
		},
	}

	result, err := json.Marshal(p)
	if err != nil {
		log.Fatalf("chronos.jsonPayload could not marshal data into json: %v", err)
	}

	return result
}

// OtherID returns the alternate instance id of an instance, if it exists
// If there is no alternate instance id, it returns an empty string
// This is used by Titus, where we also report the uuid
func (s Spinnaker) OtherID(ins chaosmonkey.Instance) (otherID string, err error) {
	url := s.instanceURL(ins.AccountName(), ins.RegionName(), ins.ID())
	resp, err := s.client.Get(url)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("get failed on %s", url))
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = errors.Wrap(cerr, fmt.Sprintf("failed to close response body from %s", url))
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("body read failed at %s", url))
	}

	// Example of response body:
	/*
		{
			...
			"health": [
				{
					"type": "Titus",
					"healthClass": "platform",
					"state": "Up"
				},
				{
					"instanceId": "55fe33ab-5b66-450a-85f7-f3129806b87f",
					"titusTaskId": "Titus-123456-worker-0-0",
					...
				}
			],
		}
	*/

	var fields struct {
		Health []map[string]interface{} `json:"health"`
		Error  string                   `json:"error"`
	}

	err = json.Unmarshal(body, &fields)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("json unmarshal failed, body: %s", body))
	}

	if resp.StatusCode != http.StatusOK {
		if fields.Error == "" {
			return "", fmt.Errorf("unexpected status code: %d. body: %s", resp.StatusCode, body)
		}

		return "", fmt.Errorf("unexpected status code: %d. error: %s", resp.StatusCode, fields.Error)
	}

	// In some cases, an instance may be missing health information.
	// We just return a blank otherID in that case
	if len(fields.Health) < 2 {
		return "", nil
	}

	otherID, ok := fields.Health[1]["instanceId"].(string)
	if !ok {
		return "", nil
	}

	// If the instance id is the same, there is no alternate
	if ins.ID() == otherID {
		return "", nil
	}

	return otherID, nil
}
