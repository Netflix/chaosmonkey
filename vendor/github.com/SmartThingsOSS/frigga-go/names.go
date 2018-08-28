package frigga

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	nameChars       = "a-zA-Z0-9\\._\\^"
	nameHyphenChars = "\\-a-zA-Z0-9\\._\\^"

	// Deviation from Frigga: SmartThings allows for more than 3 characters in the
	// push format, as well as underscore values! This does not affect the test
	// results copied from Frigga.
	pushFormat          = "v([0-9_]+)"
	labeledVarSeparator = "0"
	labeledVariable     = "[a-zA-Z][" + labeledVarSeparator + "][a-zA-Z0-9]+"

	countriesKey    = "c"
	devPhaseKey     = "d"
	hardwareKey     = "h"
	partnersKey     = "p"
	revisionKey     = "r"
	usedByKey       = "u"
	redBlackSwapKey = "w"
	zoneKey         = "z"

	pushPattern        = "^([" + nameHyphenChars + "]*)-(" + pushFormat + ")$"
	labeledVarsPattern = "^([" + nameHyphenChars + "]*?)((-" + labeledVariable + ")*)$"
	namePattern        = "^([" + nameChars + "]+)(?:-([" + nameChars + "]*))?(?:-([" + nameHyphenChars + "]*?))?$"
)

var matchers = newRegexpMatchers()

type regexpMatchers struct {
	push        *regexp.Regexp
	labeledVars *regexp.Regexp
	name        *regexp.Regexp
}

func newRegexpMatchers() *regexpMatchers {
	return &regexpMatchers{
		push:        regexp.MustCompile(pushPattern),
		labeledVars: regexp.MustCompile(labeledVarsPattern),
		name:        regexp.MustCompile(namePattern),
	}
}

// Names is a deconstruction of a Frigga-specced name.
type Names struct {
	Group        string
	Cluster      string
	App          string
	Stack        string
	Detail       string
	Push         string
	Sequence     string
	Countries    string
	DevPhase     string
	Hardware     string
	Partners     string
	Revision     string
	UsedBy       string
	RedBlackSwap string
	Zone         string
}

// SequenceInt converts a string Sequence value into an int. This is a slight
// departure from the standard Frigga spec in that it doesn't enforce using a
// numerical sequence number.
func (n *Names) SequenceInt() int {
	s, err := strconv.Atoi(n.Sequence)
	if err != nil {
		fmt.Printf("Coudld not convert sequence '%s' to int", n.Sequence)
		return 0
	}
	return s
}

func (n *Names) String() string {
	return "Names [" + strings.Join([]string{
		"group=" + n.Group,
		"cluster=" + n.Cluster,
		"app=" + n.App,
		"stack" + n.Stack,
		"detail=" + n.Detail,
		"push=" + n.Push,
		"sequence=" + n.Sequence,
		"countries=" + n.Countries,
		"devPhase=" + n.DevPhase,
		"hardware=" + n.Hardware,
		"partners=" + n.Partners,
		"revision=" + n.Revision,
		"usedBy=" + n.UsedBy,
		"redBlackSwap=" + n.RedBlackSwap,
		"zone=" + n.Zone,
	}, ", ") + "]"
}

// Parse converts a potentially Frigga string into a Names struct.
func Parse(name string) (*Names, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("Name '%s' is empty", name)
	}

	hasPush := matchers.push.MatchString(name)

	var theCluster string
	var push string
	var sequence string
	if hasPush {
		m := matchers.push.FindStringSubmatch(name)
		theCluster = m[1]
		push = m[2]
		sequence = fmt.Sprintf("%03s", m[3])
	} else {
		theCluster = name
	}

	labeledAndUnlabeledMatches := matchers.labeledVars.MatchString(theCluster)
	if !labeledAndUnlabeledMatches {
		return nil, fmt.Errorf("Name '%s' does not have any labeled or unlabeled matches", name)
	}

	vars := matchers.labeledVars.FindStringSubmatch(theCluster)
	unlabeledVars := vars[1]
	labeledVars := vars[2]

	nameMatches := matchers.name.FindStringSubmatch(unlabeledVars)
	app := nameMatches[1]
	stack := nameMatches[2]
	detail := nameMatches[3]

	names := &Names{
		Group:        name,
		Cluster:      theCluster,
		App:          app,
		Stack:        stack,
		Detail:       detail,
		Push:         push,
		Sequence:     sequence,
		Countries:    extractLabeledVariable(labeledVars, countriesKey),
		DevPhase:     extractLabeledVariable(labeledVars, devPhaseKey),
		Hardware:     extractLabeledVariable(labeledVars, hardwareKey),
		Partners:     extractLabeledVariable(labeledVars, partnersKey),
		Revision:     extractLabeledVariable(labeledVars, revisionKey),
		UsedBy:       extractLabeledVariable(labeledVars, usedByKey),
		RedBlackSwap: extractLabeledVariable(labeledVars, redBlackSwapKey),
		Zone:         extractLabeledVariable(labeledVars, zoneKey),
	}

	return names, nil
}

func extractLabeledVariable(labeledVariables string, labelKey string) string {
	if labeledVariables != "" {
		labelMatcher := regexp.MustCompile(".*?-" + labelKey + labeledVarSeparator + "([" + nameChars + "]*).*?$")
		if labelMatcher.MatchString(labeledVariables) {
			m := labelMatcher.FindStringSubmatch(labeledVariables)
			return m[1]
		}
	}
	return ""
}
