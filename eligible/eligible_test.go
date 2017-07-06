package eligible

import (
	"github.com/Netflix/chaosmonkey"
	D "github.com/Netflix/chaosmonkey/deploy"
	"github.com/Netflix/chaosmonkey/grp"
	"github.com/Netflix/chaosmonkey/mock"
	"sort"
	"testing"
)

func mockDeployment() D.Deployment {
	a := D.AccountName("prod")
	p := "aws"
	r1 := D.RegionName("us-east-1")
	r2 := D.RegionName("us-west-2")

	return &mock.Deployment{AppMap: map[string]D.AppMap{
		"foo": {a: D.AccountInfo{CloudProvider: p, Clusters: D.ClusterMap{
			"foo-crit": {
				r1: {"foo-crit-v001": []D.InstanceID{"i-11111111", "i-22222222"}},
				r2: {"foo-crit-v001": []D.InstanceID{"i-aaaaaaaa", "i-bbbbbbbb"}}},
			"foo-crit-lorin": {
				r1: {"foo-crit-lorin-v123": []D.InstanceID{"i-33333333", "i-44444444"}}},
			"foo-staging": {
				r1: {"foo-staging-v005": []D.InstanceID{"i-55555555", "i-66666666"}},
				r2: {"foo-staging-v005": []D.InstanceID{"i-cccccccc", "i-dddddddd"}},
			},
			"foo-staging-lorin": {r1: {"foo-crit-lorin-v117": []D.InstanceID{"i-77777777", "i-88888888"}}},
		}},
		}}}
}

// ids returns a sorted list of instance ids
func ids(instances []chaosmonkey.Instance) []string {
	result := make([]string, len(instances))
	for i, inst := range instances {
		result[i] = inst.ID()
	}

	sort.Strings(result)
	return result

}

func TestGroupings(t *testing.T) {
	tests := []struct {
		label string
		group grp.InstanceGroup
		wants []string
	}{
		{"cluster", grp.New("foo", "prod", "us-east-1", "", "foo-crit"), []string{"i-11111111", "i-22222222"}},
		{"stack", grp.New("foo", "prod", "us-east-1", "staging", ""), []string{"i-55555555", "i-66666666", "i-77777777", "i-88888888"}},
		{"app", grp.New("foo", "prod", "us-east-1", "", ""), []string{"i-11111111", "i-22222222", "i-33333333", "i-44444444", "i-55555555", "i-66666666", "i-77777777", "i-88888888"}},
		{"cluster, all regions", grp.New("foo", "prod", "", "", "foo-crit"), []string{"i-11111111", "i-22222222", "i-aaaaaaaa", "i-bbbbbbbb"}},
		{"stack, all regions", grp.New("foo", "prod", "", "staging", ""), []string{"i-55555555", "i-66666666", "i-77777777", "i-88888888", "i-cccccccc", "i-dddddddd"}},
		{"app, all regions", grp.New("foo", "prod", "", "", ""), []string{"i-11111111", "i-22222222", "i-33333333", "i-44444444", "i-55555555", "i-66666666", "i-77777777", "i-88888888", "i-aaaaaaaa", "i-bbbbbbbb", "i-cccccccc", "i-dddddddd"}},
	}

	// setup
	dep := mockDeployment()

	for _, tt := range tests {
		instances, err := Instances(tt.group, nil, dep)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		// assertions
		gots := ids(instances)

		if got, want := len(gots), len(tt.wants); got != want {
			t.Errorf("%s: len(eligible.Instances(group, cfg, app))=%v, want %v", tt.label, got, want)
			continue
		}

		for i, got := range gots {
			if want := tt.wants[i]; got != want {
				t.Errorf("%s: got=%v, want=%v", tt.label, got, want)
				break
			}
		}
	}
}

func TestExceptions(t *testing.T) {
	tests := []struct {
		label string
		exs   []chaosmonkey.Exception
		wants []string
	}{
		{"stack/detail/region", []chaosmonkey.Exception{{Account: "prod", Stack: "crit", Detail: "lorin", Region: "us-east-1"}}, []string{"i-11111111", "i-22222222", "i-55555555", "i-66666666", "i-77777777", "i-88888888"}},
		{"stack/detail", []chaosmonkey.Exception{{Account: "prod", Stack: "crit", Detail: "lorin", Region: "*"}}, []string{"i-11111111", "i-22222222", "i-55555555", "i-66666666", "i-77777777", "i-88888888"}},
		{"stack", []chaosmonkey.Exception{{Account: "prod", Stack: "crit", Detail: "*", Region: "*"}}, []string{"i-55555555", "i-66666666", "i-77777777", "i-88888888"}},
		{"detail", []chaosmonkey.Exception{{Account: "prod", Stack: "*", Detail: "lorin", Region: "*"}}, []string{"i-11111111", "i-22222222", "i-55555555", "i-66666666"}},
		{"all stacks", []chaosmonkey.Exception{{Account: "prod", Stack: "crit", Detail: "*", Region: "*"}, {Account: "prod", Stack: "staging", Detail: "*", Region: "*"}}, nil},
		{"blank stack", []chaosmonkey.Exception{{Account: "prod", Stack: "*", Detail: "", Region: "*"}}, []string{"i-33333333", "i-44444444", "i-77777777", "i-88888888"}},
		{"stack, detail", []chaosmonkey.Exception{{Account: "prod", Stack: "crit", Detail: "*", Region: "*"}, {Account: "prod", Stack: "*", Detail: "lorin", Region: "*"}}, []string{"i-55555555", "i-66666666"}},
	}

	// setup
	group := grp.New("foo", "prod", "us-east-1", "", "")
	dep := mockDeployment()

	for _, tt := range tests {
		instances, err := Instances(group, tt.exs, dep)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		// assertions
		gots := ids(instances)

		if got, want := len(gots), len(tt.wants); got != want {
			t.Errorf("%s: len(eligible.Instances(group, cfg, app))=%v, want %v", tt.label, got, want)
			continue
		}

		for i, got := range gots {
			if want := tt.wants[i]; got != want {
				t.Errorf("%s: got=%v, want=%v", tt.label, got, want)
				break
			}
		}
	}

}
