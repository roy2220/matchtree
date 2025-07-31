package matchtree_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	. "github.com/roy2220/matchtree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestSuite struct {
	Scenario   string              `json:"scenario"`
	MatchTypes []MatchType         `json:"match_types"`
	MatchRules []MatchRule[string] `json:"match_rules"`
	Cases      []TestCase          `json:"cases"`
}

type TestCase struct {
	MatchKeys []MatchKey `json:"match_keys"`
	Values    []string   `json:"values"`
}

func TestMatchTree_Search(t *testing.T) {
	data, err := os.ReadFile("testsuites.json")
	require.NoError(t, err)

	var suites []TestSuite
	err = json.Unmarshal(data, &suites)
	require.NoError(t, err)

	for _, suite := range suites {
		matchTree := NewMatchTree[string](suite.MatchTypes)
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

func TestIntegerInterval_Equals(t *testing.T) {
	min1 := int64(1)
	max5 := int64(5)
	min10 := int64(10)

	tests := []struct {
		name string
		i1   IntegerInterval
		i2   IntegerInterval
		want bool
	}{
		{
			name: "equal open intervals",
			i1:   IntegerInterval{Min: &min1, MinIsExcluded: true, Max: &max5, MaxIsExcluded: true},
			i2:   IntegerInterval{Min: &min1, MinIsExcluded: true, Max: &max5, MaxIsExcluded: true},
			want: true,
		},
		{
			name: "equal closed intervals",
			i1:   IntegerInterval{Min: &min1, Max: &max5},
			i2:   IntegerInterval{Min: &min1, Max: &max5},
			want: true,
		},
		{
			name: "equal half-open intervals (left excluded)",
			i1:   IntegerInterval{Min: &min1, MinIsExcluded: true, Max: &max5},
			i2:   IntegerInterval{Min: &min1, MinIsExcluded: true, Max: &max5},
			want: true,
		},
		{
			name: "equal half-open intervals (right excluded)",
			i1:   IntegerInterval{Min: &min1, Max: &max5, MaxIsExcluded: true},
			i2:   IntegerInterval{Min: &min1, Max: &max5, MaxIsExcluded: true},
			want: true,
		},
		{
			name: "equal unbounded intervals",
			i1:   IntegerInterval{},
			i2:   IntegerInterval{},
			want: true,
		},
		{
			name: "equal lower bounded intervals",
			i1:   IntegerInterval{Min: &min1},
			i2:   IntegerInterval{Min: &min1},
			want: true,
		},
		{
			name: "equal upper bounded intervals",
			i1:   IntegerInterval{Max: &max5},
			i2:   IntegerInterval{Max: &max5},
			want: true,
		},
		{
			name: "different min values",
			i1:   IntegerInterval{Min: &min1, Max: &max5},
			i2:   IntegerInterval{Min: &min10, Max: &max5},
			want: false,
		},
		{
			name: "different max values",
			i1:   IntegerInterval{Min: &min1, Max: &max5},
			i2:   IntegerInterval{Min: &min1, Max: &min10},
			want: false,
		},
		{
			name: "different min exclusion",
			i1:   IntegerInterval{Min: &min1, MinIsExcluded: true, Max: &max5},
			i2:   IntegerInterval{Min: &min1, Max: &max5},
			want: false,
		},
		{
			name: "different max exclusion",
			i1:   IntegerInterval{Min: &min1, Max: &max5, MaxIsExcluded: true},
			i2:   IntegerInterval{Min: &min1, Max: &max5},
			want: false,
		},
		{
			name: "one min nil, other not",
			i1:   IntegerInterval{Max: &max5},
			i2:   IntegerInterval{Min: &min1, Max: &max5},
			want: false,
		},
		{
			name: "one max nil, other not",
			i1:   IntegerInterval{Min: &min1},
			i2:   IntegerInterval{Min: &min1, Max: &max5},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i1.Equals(tt.i2); got != tt.want {
				t.Errorf("IntegerInterval.Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntegerInterval_Contains(t *testing.T) {
	min1 := int64(1)
	max5 := int64(5)

	tests := []struct {
		name string
		i    IntegerInterval
		x    int64
		want bool
	}{
		{
			name: "closed interval, contains inside",
			i:    IntegerInterval{Min: &min1, Max: &max5},
			x:    3,
			want: true,
		},
		{
			name: "closed interval, contains min boundary",
			i:    IntegerInterval{Min: &min1, Max: &max5},
			x:    1,
			want: true,
		},
		{
			name: "closed interval, contains max boundary",
			i:    IntegerInterval{Min: &min1, Max: &max5},
			x:    5,
			want: true,
		},
		{
			name: "closed interval, does not contain below min",
			i:    IntegerInterval{Min: &min1, Max: &max5},
			x:    0,
			want: false,
		},
		{
			name: "closed interval, does not contain above max",
			i:    IntegerInterval{Min: &min1, Max: &max5},
			x:    6,
			want: false,
		},
		{
			name: "open interval, contains inside",
			i:    IntegerInterval{Min: &min1, MinIsExcluded: true, Max: &max5, MaxIsExcluded: true},
			x:    3,
			want: true,
		},
		{
			name: "open interval, does not contain min boundary",
			i:    IntegerInterval{Min: &min1, MinIsExcluded: true, Max: &max5, MaxIsExcluded: true},
			x:    1,
			want: false,
		},
		{
			name: "open interval, does not contain max boundary",
			i:    IntegerInterval{Min: &min1, MinIsExcluded: true, Max: &max5, MaxIsExcluded: true},
			x:    5,
			want: false,
		},
		{
			name: "half-open interval [min, max)",
			i:    IntegerInterval{Min: &min1, Max: &max5, MaxIsExcluded: true},
			x:    1,
			want: true,
		},
		{
			name: "half-open interval [min, max), does not contain max boundary",
			i:    IntegerInterval{Min: &min1, Max: &max5, MaxIsExcluded: true},
			x:    5,
			want: false,
		},
		{
			name: "half-open interval (min, max]",
			i:    IntegerInterval{Min: &min1, MinIsExcluded: true, Max: &max5},
			x:    5,
			want: true,
		},
		{
			name: "half-open interval (min, max], does not contain min boundary",
			i:    IntegerInterval{Min: &min1, MinIsExcluded: true, Max: &max5},
			x:    1,
			want: false,
		},
		{
			name: "unbounded interval, contains any number",
			i:    IntegerInterval{},
			x:    100,
			want: true,
		},
		{
			name: "lower bounded interval, contains above min",
			i:    IntegerInterval{Min: &min1},
			x:    10,
			want: true,
		},
		{
			name: "lower bounded interval, contains min",
			i:    IntegerInterval{Min: &min1},
			x:    1,
			want: true,
		},
		{
			name: "lower bounded interval, does not contain below min",
			i:    IntegerInterval{Min: &min1},
			x:    0,
			want: false,
		},
		{
			name: "upper bounded interval, contains below max",
			i:    IntegerInterval{Max: &max5},
			x:    0,
			want: true,
		},
		{
			name: "upper bounded interval, contains max",
			i:    IntegerInterval{Max: &max5},
			x:    5,
			want: true,
		},
		{
			name: "upper bounded interval, does not contain above max",
			i:    IntegerInterval{Max: &max5},
			x:    6,
			want: false,
		},
		{
			name: "lower bounded (excluded), contains above min",
			i:    IntegerInterval{Min: &min1, MinIsExcluded: true},
			x:    2,
			want: true,
		},
		{
			name: "lower bounded (excluded), does not contain min",
			i:    IntegerInterval{Min: &min1, MinIsExcluded: true},
			x:    1,
			want: false,
		},
		{
			name: "upper bounded (excluded), contains below max",
			i:    IntegerInterval{Max: &max5, MaxIsExcluded: true},
			x:    4,
			want: true,
		},
		{
			name: "upper bounded (excluded), does not contain max",
			i:    IntegerInterval{Max: &max5, MaxIsExcluded: true},
			x:    5,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.Contains(tt.x); got != tt.want {
				t.Errorf("IntegerInterval.Contains() for %v with x=%v = %v, want %v", tt.i, tt.x, got, tt.want)
			}
		})
	}
}

const epsilon = 1e-10

func TestNumberInterval_Equals(t *testing.T) {
	min1 := 1.0
	max5 := 5.0
	min10 := 10.0
	min1plusEpsilon := 1.0 + epsilon/2  // Slightly different but within epsilon
	min1minusEpsilon := 1.0 - epsilon/2 // Slightly different but within epsilon

	tests := []struct {
		name string
		i1   NumberInterval
		i2   NumberInterval
		want bool
	}{
		{
			name: "equal open intervals",
			i1:   NumberInterval{Min: &min1, MinIsExcluded: true, Max: &max5, MaxIsExcluded: true},
			i2:   NumberInterval{Min: &min1, MinIsExcluded: true, Max: &max5, MaxIsExcluded: true},
			want: true,
		},
		{
			name: "equal closed intervals",
			i1:   NumberInterval{Min: &min1, Max: &max5},
			i2:   NumberInterval{Min: &min1, Max: &max5},
			want: true,
		},
		{
			name: "equal intervals with min values within epsilon",
			i1:   NumberInterval{Min: &min1, Max: &max5},
			i2:   NumberInterval{Min: &min1plusEpsilon, Max: &max5},
			want: true,
		},
		{
			name: "equal intervals with min values within epsilon (other way)",
			i1:   NumberInterval{Min: &min1plusEpsilon, Max: &max5},
			i2:   NumberInterval{Min: &min1, Max: &max5},
			want: true,
		},
		{
			name: "equal intervals with max values within epsilon",
			i1:   NumberInterval{Min: &min1, Max: &max5},
			i2:   NumberInterval{Min: &min1, Max: func() *float64 { f := max5 + epsilon/2; return &f }()},
			want: true,
		},
		{
			name: "different min values (outside epsilon)",
			i1:   NumberInterval{Min: &min1, Max: &max5},
			i2:   NumberInterval{Min: &min10, Max: &max5},
			want: false,
		},
		{
			name: "different min values (just outside epsilon)",
			i1:   NumberInterval{Min: &min1, Max: &max5},
			i2:   NumberInterval{Min: &min1minusEpsilon, Max: &max5},
			want: true,
		},
		{
			name: "different min values (exactly epsilon)",
			i1:   NumberInterval{Min: &min1, Max: &max5},
			i2:   NumberInterval{Min: func() *float64 { f := min1 + 2*epsilon; return &f }(), Max: &max5},
			want: false,
		},
		{
			name: "different max values",
			i1:   NumberInterval{Min: &min1, Max: &max5},
			i2:   NumberInterval{Min: &min1, Max: &min10},
			want: false,
		},
		{
			name: "different min exclusion",
			i1:   NumberInterval{Min: &min1, MinIsExcluded: true, Max: &max5},
			i2:   NumberInterval{Min: &min1, Max: &max5},
			want: false,
		},
		{
			name: "different max exclusion",
			i1:   NumberInterval{Min: &min1, Max: &max5, MaxIsExcluded: true},
			i2:   NumberInterval{Min: &min1, Max: &max5},
			want: false,
		},
		{
			name: "one min nil, other not",
			i1:   NumberInterval{Max: &max5},
			i2:   NumberInterval{Min: &min1, Max: &max5},
			want: false,
		},
		{
			name: "one max nil, other not",
			i1:   NumberInterval{Min: &min1},
			i2:   NumberInterval{Min: &min1, Max: &max5},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i1.Equals(tt.i2); got != tt.want {
				t.Errorf("NumberInterval.Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNumberInterval_Contains(t *testing.T) {
	min1 := 1.0
	max5 := 5.0

	tests := []struct {
		name string
		i    NumberInterval
		x    float64
		want bool
	}{
		{
			name: "closed interval, contains inside",
			i:    NumberInterval{Min: &min1, Max: &max5},
			x:    3.0,
			want: true,
		},
		{
			name: "closed interval, contains min boundary",
			i:    NumberInterval{Min: &min1, Max: &max5},
			x:    1.0,
			want: true,
		},
		{
			name: "closed interval, contains max boundary",
			i:    NumberInterval{Min: &min1, Max: &max5},
			x:    5.0,
			want: true,
		},
		{
			name: "closed interval, contains min boundary slightly off",
			i:    NumberInterval{Min: &min1, Max: &max5},
			x:    1.0 + epsilon/2,
			want: true,
		},
		{
			name: "closed interval, contains max boundary slightly off",
			i:    NumberInterval{Min: &min1, Max: &max5},
			x:    5.0 - epsilon/2,
			want: true,
		},
		{
			name: "closed interval, does not contain below min (just outside)",
			i:    NumberInterval{Min: &min1, Max: &max5},
			x:    1.0 - 2*epsilon,
			want: false,
		},
		{
			name: "closed interval, does not contain above max (just outside)",
			i:    NumberInterval{Min: &min1, Max: &max5},
			x:    5.0 + 2*epsilon,
			want: false,
		},
		{
			name: "open interval, contains inside",
			i:    NumberInterval{Min: &min1, MinIsExcluded: true, Max: &max5, MaxIsExcluded: true},
			x:    3.0,
			want: true,
		},
		{
			name: "open interval, does not contain min boundary",
			i:    NumberInterval{Min: &min1, MinIsExcluded: true, Max: &max5, MaxIsExcluded: true},
			x:    1.0,
			want: false,
		},
		{
			name: "open interval, does not contain max boundary",
			i:    NumberInterval{Min: &min1, MinIsExcluded: true, Max: &max5, MaxIsExcluded: true},
			x:    5.0,
			want: false,
		},
		{
			name: "open interval, contains just above min",
			i:    NumberInterval{Min: &min1, MinIsExcluded: true, Max: &max5, MaxIsExcluded: true},
			x:    1.0 + 2*epsilon,
			want: true,
		},
		{
			name: "open interval, contains just below max",
			i:    NumberInterval{Min: &min1, MinIsExcluded: true, Max: &max5, MaxIsExcluded: true},
			x:    5.0 - 2*epsilon,
			want: true,
		},
		{
			name: "half-open interval [min, max)",
			i:    NumberInterval{Min: &min1, Max: &max5, MaxIsExcluded: true},
			x:    1.0,
			want: true,
		},
		{
			name: "half-open interval [min, max), does not contain max boundary",
			i:    NumberInterval{Min: &min1, Max: &max5, MaxIsExcluded: true},
			x:    5.0,
			want: false,
		},
		{
			name: "half-open interval [min, max), contains just below max",
			i:    NumberInterval{Min: &min1, Max: &max5, MaxIsExcluded: true},
			x:    5.0 - 2*epsilon,
			want: true,
		},
		{
			name: "half-open interval (min, max]",
			i:    NumberInterval{Min: &min1, MinIsExcluded: true, Max: &max5},
			x:    5.0,
			want: true,
		},
		{
			name: "half-open interval (min, max], does not contain min boundary",
			i:    NumberInterval{Min: &min1, MinIsExcluded: true, Max: &max5},
			x:    1.0,
			want: false,
		},
		{
			name: "half-open interval (min, max], contains just above min",
			i:    NumberInterval{Min: &min1, MinIsExcluded: true, Max: &max5},
			x:    1.0 + 2*epsilon,
			want: true,
		},
		{
			name: "unbounded interval, contains any number",
			i:    NumberInterval{},
			x:    100.0,
			want: true,
		},
		{
			name: "lower bounded interval, contains above min",
			i:    NumberInterval{Min: &min1},
			x:    10.0,
			want: true,
		},
		{
			name: "lower bounded interval, contains min",
			i:    NumberInterval{Min: &min1},
			x:    1.0,
			want: true,
		},
		{
			name: "lower bounded interval, does not contain below min",
			i:    NumberInterval{Min: &min1},
			x:    0.0,
			want: false,
		},
		{
			name: "upper bounded interval, contains below max",
			i:    NumberInterval{Max: &max5},
			x:    0.0,
			want: true,
		},
		{
			name: "upper bounded interval, contains max",
			i:    NumberInterval{Max: &max5},
			x:    5.0,
			want: true,
		},
		{
			name: "upper bounded interval, does not contain above max",
			i:    NumberInterval{Max: &max5},
			x:    6.0,
			want: false,
		},
		{
			name: "lower bounded (excluded), contains above min",
			i:    NumberInterval{Min: &min1, MinIsExcluded: true},
			x:    2.0,
			want: true,
		},
		{
			name: "lower bounded (excluded), does not contain min",
			i:    NumberInterval{Min: &min1, MinIsExcluded: true},
			x:    1.0,
			want: false,
		},
		{
			name: "lower bounded (excluded), does not contain min (just below)",
			i:    NumberInterval{Min: &min1, MinIsExcluded: true},
			x:    1.0 + epsilon/2, // within exclusion boundary, means not contained
			want: false,
		},
		{
			name: "upper bounded (excluded), contains below max",
			i:    NumberInterval{Max: &max5, MaxIsExcluded: true},
			x:    4.0,
			want: true,
		},
		{
			name: "upper bounded (excluded), does not contain max",
			i:    NumberInterval{Max: &max5, MaxIsExcluded: true},
			x:    5.0,
			want: false,
		},
		{
			name: "upper bounded (excluded), does not contain max (just above)",
			i:    NumberInterval{Max: &max5, MaxIsExcluded: true},
			x:    5.0 - epsilon/2, // within exclusion boundary, means not contained
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.i.Contains(tt.x); got != tt.want {
				t.Errorf("NumberInterval.Contains() for %v with x=%v = %v, want %v", tt.i, tt.x, got, tt.want)
			}
		})
	}
}
