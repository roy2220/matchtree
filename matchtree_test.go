package matchtree_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/roy2220/matchtree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestSuite struct {
	Scenario   string                        `json:"scenario"`
	MatchTypes []matchtree.MatchType         `json:"match_types"`
	MatchRules []matchtree.MatchRule[string] `json:"match_rules"`
	Cases      []TestCase                    `json:"cases"`
}

type TestCase struct {
	MatchKeys []matchtree.MatchKey `json:"match_keys"`
	Values    []string             `json:"values"`
}

func TestMatchTree_Search(t *testing.T) {
	data, err := os.ReadFile("testsuites.json")
	require.NoError(t, err)

	var suites []TestSuite
	err = json.Unmarshal(data, &suites)
	require.NoError(t, err)

	for _, suite := range suites {
		matchTree := matchtree.NewMatchTree[string](suite.MatchTypes)
		for _, matchRule := range suite.MatchRules {
			err = matchTree.AddRule(matchRule)
			require.NoError(t, err)
		}
		for i, case1 := range suite.Cases {
			t.Run(fmt.Sprintf("%s#%d", suite.Scenario, i+1), func(t *testing.T) {
				values, err := matchTree.Search(case1.MatchKeys)
				require.NoError(t, err)
				assert.Equal(t, case1.Values, values)
			})
		}
	}
}
