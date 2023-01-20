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
	"encoding/json"
	"testing"

	"github.com/Netflix/chaosmonkey/v2/mock"
)

func TestKillJSONPayload(t *testing.T) {
	ins := mock.Instance{
		App:        "foo",
		Account:    "test",
		Stack:      "beta",
		Cluster:    "foo-beta",
		Region:     "us-west-2",
		ASG:        "foo-beta-v052",
		InstanceID: "i-703a0439",
	}

	otherID := "" // some backends may have a second instance identifier

	payload := killJSONPayload(ins, otherID, "user@example.com")

	/*
		{
		  "application": "foo",
		  "description": "Chaos Monkey terminate instance: i-703a0439 (test, us-west-2, foo-beta-v052)",
		  "job": [
			{
			  "user": "user@example.com"
			  "type": "terminateInstances",
			  "credentials": "test",
			  "region": "us-west-2",
			  "serverGroupName": "foo-beta-v052",
			  "instanceIds": [
				"i-703a0439"
			  ],
			}
		  ]
		}
	*/

	var f interface{}
	err := json.Unmarshal(payload, &f)
	if err != nil {
		t.Log(string(payload))
		t.Fatal(err)
	}

	m := f.(map[string]interface{})
	if m == nil {
		t.Fatalf("payload is not a JSON object: %s", payload)
	}

	tests := []struct {
		name, value string
	}{
		{"application", "foo"},
		{"description", "Chaos Monkey terminate instance: i-703a0439 (test, us-west-2, foo-beta-v052)"},
	}

	for _, tt := range tests {
		if _, ok := m[tt.name].(string); !ok {
			t.Fatalf("Missing field: %s", tt.name)
		}

		if got, want := m[tt.name].(string), tt.value; got != want {
			t.Errorf("got ['%s']=%s, want %s", tt.name, got, want)
		}
	}

	var jobs []interface{}
	var ok bool

	if jobs, ok = m["job"].([]interface{}); !ok {
		t.Fatalf("jobs is not an array: %s", payload)
	}

	if got, want := len(jobs), 1; got != want {
		t.Fatalf("got len(jobs)=%d, want: %d", got, want)
	}

	var job map[string]interface{}

	if job, ok = jobs[0].(map[string]interface{}); !ok {
		t.Fatalf("job[0] is not a json object: %s", payload)
	}

	tests = []struct {
		name, value string
	}{
		{"type", "terminateInstances"},
		{"serverGroupName", "foo-beta-v052"},
		{"region", "us-west-2"},
		{"credentials", "test"},
		{"user", "user@example.com"},
	}

	for _, tt := range tests {
		if got, want := job[tt.name].(string), tt.value; got != want {
			t.Errorf("got obj['%s']=%s, want %s", tt.name, got, want)
		}
	}

	ids, ok := job["instanceIds"].([]interface{})
	if !ok {
		t.Fatalf("No job.instanceIds field: %s", payload)
	}

	if len(ids) != 1 {
		t.Fatalf("job.instanceIds field is not 1: %v", payload)
	}

	id, ok := ids[0].(string)
	if !ok {
		t.Fatalf("job.InstanceIds[0] is not a string: %v", payload)
	}

	if got, want := id, "i-703a0439"; got != want {
		t.Fatalf("Wrong instance id. got: %s, want: %s", got, want)
	}
}

func TestKillJSONPayloadWithOtherID(t *testing.T) {
	ins := mock.Instance{
		App:        "foo",
		Account:    "other",
		Stack:      "beta",
		Cluster:    "foo-beta",
		Region:     "us-west-2",
		ASG:        "foo-beta-v052",
		InstanceID: "custom-id-123",
	}

	otherID := "39033754-c0ac-423d-aab7-2736548acf65"
	payload := killJSONPayload(ins, otherID, "user@example.com")

	var f interface{}
	err := json.Unmarshal(payload, &f)
	if err != nil {
		t.Log(string(payload))
		t.Fatal(err)
	}

	m := f.(map[string]interface{})
	if m == nil {
		t.Fatalf("payload is not a JSON object: %s", payload)
	}

	want := "Chaos Monkey terminate instance: custom-id-123 39033754-c0ac-423d-aab7-2736548acf65 (other, us-west-2, foo-beta-v052)"

	if got := m["description"]; got != want {
		t.Errorf("got: %s, want: %s", got, want)
	}
}
