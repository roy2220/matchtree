package matchtree

import (
	"fmt"
	"sort"
	"testing"
)

// Helper function to sort []any for consistent comparison in tests
func sortAnySlice(s []any) {
	sort.Slice(s, func(i, j int) bool {
		// Attempt to convert to string for comparison, as 'any' type cannot be directly compared
		// This is a simplification for test purposes, actual comparison might need more robust handling
		strI := fmt.Sprintf("%v", s[i])
		strJ := fmt.Sprintf("%v", s[j])
		return strI < strJ
	})
}

func TestNewMatchTree(t *testing.T) {
	tests := []struct {
		name         string
		types        []MatchType
		expectPanics bool
	}{
		{
			name:         "Valid types",
			types:        []MatchType{MatchString, MatchInteger, MatchNumberInterval},
			expectPanics: false,
		},
		{
			name:         "Invalid type",
			types:        []MatchType{MatchString, 99}, // 99 是一个未定义的 MatchType
			expectPanics: true,
		},
		{
			name:         "Empty types",
			types:        []MatchType{},
			expectPanics: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanics {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("NewMatchTree did not panic as expected for types: %v", tt.types)
					}
				}()
			}
			tree := NewMatchTree(tt.types)
			if !tt.expectPanics && tree == nil {
				t.Errorf("NewMatchTree returned nil for valid types: %v", tt.types)
			}
			if !tt.expectPanics && len(tree.types) != len(tt.types) {
				t.Errorf("NewMatchTree types mismatch, expected %v, got %v", tt.types, tree.types)
			}
		})
	}
}

func TestAddRuleAndSearch(t *testing.T) {
	// 定义一个多层级的 MatchTree
	tree := NewMatchTree([]MatchType{
		MatchString,          // level 0: user_group (string)
		MatchInteger,         // level 1: user_id (integer)
		MatchIntegerInterval, // level 2: age (integer interval)
		MatchNumberInterval,  // level 3: score (number interval)
	})

	// 测试用例结构体
	type testCase struct {
		name           string
		rulesToAdd     []MatchRule
		keysToSearch   []MatchKey
		expectedValues []any
		expectFound    bool
		addRulePanics  bool
		searchPanics   bool
	}

	tests := []testCase{
		// --- 基本精确匹配测试 ---
		{
			name: "Exact match for all levels",
			rulesToAdd: []MatchRule{
				{
					Patterns: []MatchPattern{
						{Type: MatchString, String: "admin"},
						{Type: MatchInteger, Integer: 101},
						{Type: MatchIntegerInterval, IntegerInterval: IntegerInterval{Min: 18, Max: 60}},
						{Type: MatchNumberInterval, NumberInterval: NumberInterval{Min: 70.0, Max: 100.0}},
					},
					Value: "Admin_Privileges",
				},
			},
			keysToSearch: []MatchKey{
				{Type: MatchString, String: "admin"},
				{Type: MatchInteger, Integer: 101},
				{Type: MatchIntegerInterval, Integer: 30}, // Within [18, 60]
				{Type: MatchNumberInterval, Number: 85.5}, // Within [70.0, 100.0]
			},
			expectedValues: []any{"Admin_Privileges"},
			expectFound:    true,
		},
		{
			name: "No match - string mismatch",
			keysToSearch: []MatchKey{
				{Type: MatchString, String: "user"}, // Mismatch
				{Type: MatchInteger, Integer: 101},
				{Type: MatchIntegerInterval, Integer: 30},
				{Type: MatchNumberInterval, Number: 85.5},
			},
			expectedValues: nil,
			expectFound:    false,
		},
		{
			name: "No match - integer mismatch",
			keysToSearch: []MatchKey{
				{Type: MatchString, String: "admin"},
				{Type: MatchInteger, Integer: 102}, // Mismatch
				{Type: MatchIntegerInterval, Integer: 30},
				{Type: MatchNumberInterval, Number: 85.5},
			},
			expectedValues: nil,
			expectFound:    false,
		},
		{
			name: "No match - age interval mismatch",
			keysToSearch: []MatchKey{
				{Type: MatchString, String: "admin"},
				{Type: MatchInteger, Integer: 101},
				{Type: MatchIntegerInterval, Integer: 10}, // Mismatch (out of 18-60)
				{Type: MatchNumberInterval, Number: 85.5},
			},
			expectedValues: nil,
			expectFound:    false,
		},
		{
			name: "No match - score interval mismatch",
			keysToSearch: []MatchKey{
				{Type: MatchString, String: "admin"},
				{Type: MatchInteger, Integer: 101},
				{Type: MatchIntegerInterval, Integer: 30},
				{Type: MatchNumberInterval, Number: 60.0}, // Mismatch (out of 70.0-100.0)
			},
			expectedValues: nil,
			expectFound:    false,
		},

		// --- 通配符 (IsAny) 测试 ---
		{
			name: "Match with string Any pattern",
			rulesToAdd: []MatchRule{
				{
					Patterns: []MatchPattern{
						{Type: MatchString, IsAny: true}, // Any string
						{Type: MatchInteger, Integer: 202},
						{Type: MatchIntegerInterval, IntegerInterval: IntegerInterval{Min: 10, Max: 20}},
						{Type: MatchNumberInterval, NumberInterval: NumberInterval{Min: 10.0, Max: 20.0}},
					},
					Value: "Any_String_Rule",
				},
			},
			keysToSearch: []MatchKey{
				{Type: MatchString, String: "guest"}, // Will match Any
				{Type: MatchInteger, Integer: 202},
				{Type: MatchIntegerInterval, Integer: 15},
				{Type: MatchNumberInterval, Number: 15.0},
			},
			expectedValues: []any{"Any_String_Rule"},
			expectFound:    true,
		},
		{
			name: "Match with integer Any pattern",
			rulesToAdd: []MatchRule{
				{
					Patterns: []MatchPattern{
						{Type: MatchString, String: "user"},
						{Type: MatchInteger, IsAny: true}, // Any integer
						{Type: MatchIntegerInterval, IntegerInterval: IntegerInterval{Min: 1, Max: 10}},
						{Type: MatchNumberInterval, NumberInterval: NumberInterval{Min: 1.0, Max: 10.0}},
					},
					Value: "Any_Integer_Rule",
				},
			},
			keysToSearch: []MatchKey{
				{Type: MatchString, String: "user"},
				{Type: MatchInteger, Integer: 999}, // Will match Any
				{Type: MatchIntegerInterval, Integer: 5},
				{Type: MatchNumberInterval, Number: 5.0},
			},
			expectedValues: []any{"Any_Integer_Rule"},
			expectFound:    true,
		},
		{
			name: "Match with interval Any pattern",
			rulesToAdd: []MatchRule{
				{
					Patterns: []MatchPattern{
						{Type: MatchString, String: "vip"},
						{Type: MatchInteger, Integer: 303},
						{Type: MatchIntegerInterval, IsAny: true}, // Any integer interval
						{Type: MatchNumberInterval, IsAny: true},  // Any number interval
					},
					Value: "Any_Interval_Rule",
				},
			},
			keysToSearch: []MatchKey{
				{Type: MatchString, String: "vip"},
				{Type: MatchInteger, Integer: 303},
				{Type: MatchIntegerInterval, Integer: 70},   // Will match Any
				{Type: MatchNumberInterval, Number: 123.45}, // Will match Any
			},
			expectedValues: []any{"Any_Interval_Rule"},
			expectFound:    true,
		},

		// --- 多个规则匹配测试 ---
		{
			name: "Multiple rules matching (exact and any)",
			rulesToAdd: []MatchRule{
				{
					Patterns: []MatchPattern{
						{Type: MatchString, String: "finance"},
						{Type: MatchInteger, Integer: 404},
						{Type: MatchIntegerInterval, IntegerInterval: IntegerInterval{Min: 25, Max: 50}},
						{Type: MatchNumberInterval, NumberInterval: NumberInterval{Min: 50.0, Max: 75.0}},
					},
					Value: "Finance_Specific_Rule",
				},
				{
					Patterns: []MatchPattern{
						{Type: MatchString, IsAny: true},
						{Type: MatchInteger, Integer: 404},
						{Type: MatchIntegerInterval, IntegerInterval: IntegerInterval{Min: 20, Max: 60}},
						{Type: MatchNumberInterval, NumberInterval: NumberInterval{Min: 40.0, Max: 80.0}},
					},
					Value: "Finance_Any_String_Rule",
				},
			},
			keysToSearch: []MatchKey{
				{Type: MatchString, String: "finance"},
				{Type: MatchInteger, Integer: 404},
				{Type: MatchIntegerInterval, Integer: 35}, // In both intervals
				{Type: MatchNumberInterval, Number: 60.0}, // In both intervals
			},
			expectedValues: []any{"Finance_Specific_Rule", "Finance_Any_String_Rule"}, // Order might vary, need sort
			expectFound:    true,
		},

		// --- 整数区间边界测试 ---
		{
			name: "Integer interval exact boundary (inclusive)",
			rulesToAdd: []MatchRule{
				{
					Patterns: []MatchPattern{
						{Type: MatchString, String: "range_test"},
						{Type: MatchInteger, Integer: 1},
						{Type: MatchIntegerInterval, IntegerInterval: IntegerInterval{Min: 10, Max: 20}},
						{Type: MatchNumberInterval, NumberInterval: NumberInterval{Min: 0.0, Max: 10.0}},
					},
					Value: "Inclusive_Rule",
				},
			},
			keysToSearch: []MatchKey{
				{Type: MatchString, String: "range_test"},
				{Type: MatchInteger, Integer: 1},
				{Type: MatchIntegerInterval, Integer: 10}, // Min boundary
				{Type: MatchNumberInterval, Number: 5.0},
			},
			expectedValues: []any{"Inclusive_Rule"},
			expectFound:    true,
		},
		{
			name: "Integer interval exact boundary (exclusive)",
			rulesToAdd: []MatchRule{
				{
					Patterns: []MatchPattern{
						{Type: MatchString, String: "range_test_exclusive"},
						{Type: MatchInteger, Integer: 2},
						{Type: MatchIntegerInterval, IntegerInterval: IntegerInterval{Min: 10, MinIsExcluded: true, Max: 20, MaxIsExcluded: true}}, // (10, 20)
						{Type: MatchNumberInterval, NumberInterval: NumberInterval{Min: 0.0, Max: 10.0}},
					},
					Value: "Exclusive_Rule",
				},
			},
			keysToSearch: []MatchKey{
				{Type: MatchString, String: "range_test_exclusive"},
				{Type: MatchInteger, Integer: 2},
				{Type: MatchIntegerInterval, Integer: 11}, // First valid integer
				{Type: MatchNumberInterval, Number: 5.0},
			},
			expectedValues: []any{"Exclusive_Rule"},
			expectFound:    true,
		},
		{
			name: "Integer interval outside exclusive min boundary",
			keysToSearch: []MatchKey{
				{Type: MatchString, String: "range_test_exclusive"},
				{Type: MatchInteger, Integer: 2},
				{Type: MatchIntegerInterval, Integer: 10}, // Should not match (10, 20)
				{Type: MatchNumberInterval, Number: 5.0},
			},
			expectedValues: nil,
			expectFound:    false,
		},

		// --- 浮点数区间边界测试 ---
		{
			name: "Number interval exact boundary (inclusive)",
			rulesToAdd: []MatchRule{
				{
					Patterns: []MatchPattern{
						{Type: MatchString, String: "float_test"},
						{Type: MatchInteger, Integer: 3},
						{Type: MatchIntegerInterval, IntegerInterval: IntegerInterval{Min: 0, Max: 100}},
						{Type: MatchNumberInterval, NumberInterval: NumberInterval{Min: 10.0, Max: 20.0}},
					},
					Value: "Float_Inclusive_Rule",
				},
			},
			keysToSearch: []MatchKey{
				{Type: MatchString, String: "float_test"},
				{Type: MatchInteger, Integer: 3},
				{Type: MatchIntegerInterval, Integer: 50},
				{Type: MatchNumberInterval, Number: 10.0}, // Min boundary
			},
			expectedValues: []any{"Float_Inclusive_Rule"},
			expectFound:    true,
		},
		{
			name: "Number interval exact boundary (exclusive)",
			rulesToAdd: []MatchRule{
				{
					Patterns: []MatchPattern{
						{Type: MatchString, String: "float_test_exclusive"},
						{Type: MatchInteger, Integer: 4},
						{Type: MatchIntegerInterval, IntegerInterval: IntegerInterval{Min: 0, Max: 100}},
						{Type: MatchNumberInterval, NumberInterval: NumberInterval{Min: 10.0, MinIsExcluded: true, Max: 20.0, MaxIsExcluded: true}}, // (10.0, 20.0)
					},
					Value: "Float_Exclusive_Rule",
				},
			},
			keysToSearch: []MatchKey{
				{Type: MatchString, String: "float_test_exclusive"},
				{Type: MatchInteger, Integer: 4},
				{Type: MatchIntegerInterval, Integer: 50},
				{Type: MatchNumberInterval, Number: 10.000001}, // Just inside exclusive min
			},
			expectedValues: []any{"Float_Exclusive_Rule"},
			expectFound:    true,
		},
		{
			name: "Number interval outside exclusive min boundary",
			keysToSearch: []MatchKey{
				{Type: MatchString, String: "float_test_exclusive"},
				{Type: MatchInteger, Integer: 4},
				{Type: MatchIntegerInterval, Integer: 50},
				{Type: MatchNumberInterval, Number: 10.0}, // Should not match (10.0, 20.0) due to epsilon
			},
			expectedValues: nil,
			expectFound:    false,
		},

		// --- AddRule 错误处理测试 ---
		{
			name: "AddRule - too few patterns (panic)",
			rulesToAdd: []MatchRule{
				{
					Patterns: []MatchPattern{ // Only 3 patterns, expected 4
						{Type: MatchString, String: "bad_rule"},
						{Type: MatchInteger, Integer: 1},
						{Type: MatchIntegerInterval, IntegerInterval: IntegerInterval{Min: 0, Max: 10}},
					},
					Value: "Bad_Rule_Value",
				},
			},
			addRulePanics: true,
		},
		{
			name: "AddRule - wrong pattern type (panic)",
			rulesToAdd: []MatchRule{
				{
					Patterns: []MatchPattern{
						{Type: MatchString, String: "bad_type_rule"},
						{Type: MatchIntegerInterval, IntegerInterval: IntegerInterval{Min: 0, Max: 10}}, // Wrong type here, expected MatchInteger
						{Type: MatchIntegerInterval, IntegerInterval: IntegerInterval{Min: 0, Max: 10}},
						{Type: MatchNumberInterval, NumberInterval: NumberInterval{Min: 0.0, Max: 10.0}},
					},
					Value: "Bad_Type_Value",
				},
			},
			addRulePanics: true,
		},

		// --- Search 错误处理测试 ---
		{
			name: "Search - too few keys (panic)",
			keysToSearch: []MatchKey{ // Only 3 keys, expected 4
				{Type: MatchString, String: "test"},
				{Type: MatchInteger, Integer: 1},
				{Type: MatchInteger, Integer: 5},
			},
			searchPanics: true,
		},
		{
			name: "Search - wrong key type (panic)",
			keysToSearch: []MatchKey{
				{Type: MatchString, String: "test"},
				{Type: MatchString, String: "wrong_key_type"}, // Expected MatchInteger
				{Type: MatchIntegerInterval, Integer: 5},
				{Type: MatchNumberInterval, Number: 5.0},
			},
			searchPanics: true,
		},
	}

	for _, tt := range tests {
		// Reset tree for each test that adds rules, to prevent rule pollution
		// For tests that only search, keep the tree state from previous rule additions
		if tt.rulesToAdd != nil && len(tt.rulesToAdd) > 0 {
			tree = NewMatchTree([]MatchType{
				MatchString,
				MatchInteger,
				MatchIntegerInterval,
				MatchNumberInterval,
			})
			for _, rule := range tt.rulesToAdd {
				t.Run(fmt.Sprintf("%s_AddRule", tt.name), func(t *testing.T) {
					if tt.addRulePanics {
						defer func() {
							if r := recover(); r == nil {
								t.Errorf("AddRule did not panic as expected for rule: %+v", rule)
							}
						}()
					}
					tree.AddRule(rule) // Pass tree as first argument
				})
			}
			if tt.addRulePanics {
				continue // If AddRule panicked, skip the search part for this test case
			}
		}

		t.Run(fmt.Sprintf("%s_Search", tt.name), func(t *testing.T) {
			if tt.searchPanics {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Search did not panic as expected for keys: %+v", tt.keysToSearch)
					}
				}()
			}
			values, found := tree.Search(tt.keysToSearch) // Pass tree as first argument

			if tt.searchPanics {
				return // If search is expected to panic, don't proceed with value checks
			}

			if found != tt.expectFound {
				t.Errorf("For keys %+v, expected found=%v, got %v", tt.keysToSearch, tt.expectFound, found)
				return
			}

			if tt.expectFound {
				sortAnySlice(values)
				sortAnySlice(tt.expectedValues)
				if !compareSlices(values, tt.expectedValues) {
					t.Errorf("For keys %+v, expected values %v, got %v", tt.keysToSearch, tt.expectedValues, values)
				}
			} else {
				if values != nil && len(values) > 0 {
					t.Errorf("For keys %+v, expected nil values, got %v", tt.keysToSearch, values)
				}
			}
		})
	}
}

// Utility to compare two slices of any
func compareSlices(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		// Use fmt.Sprintf to compare any values as strings for simplicity in tests
		// This might not be robust enough for all 'any' types, but good for primitives
		if fmt.Sprintf("%v", a[i]) != fmt.Sprintf("%v", b[i]) {
			return false
		}
	}
	return true
}

func TestIntegerIntervalContains(t *testing.T) {
	tests := []struct {
		name     string
		interval IntegerInterval
		value    int64
		expected bool
	}{
		{"Inclusive: inside", IntegerInterval{Min: 10, Max: 20}, 15, true},
		{"Inclusive: min boundary", IntegerInterval{Min: 10, Max: 20}, 10, true},
		{"Inclusive: max boundary", IntegerInterval{Min: 10, Max: 20}, 20, true},
		{"Inclusive: below min", IntegerInterval{Min: 10, Max: 20}, 9, false},
		{"Inclusive: above max", IntegerInterval{Min: 10, Max: 20}, 21, false},

		{"Exclusive min: inside", IntegerInterval{Min: 10, MinIsExcluded: true, Max: 20}, 15, true},
		{"Exclusive min: min boundary", IntegerInterval{Min: 10, MinIsExcluded: true, Max: 20}, 10, false},
		{"Exclusive min: just above min", IntegerInterval{Min: 10, MinIsExcluded: true, Max: 20}, 11, true},

		{"Exclusive max: inside", IntegerInterval{Min: 10, Max: 20, MaxIsExcluded: true}, 15, true},
		{"Exclusive max: max boundary", IntegerInterval{Min: 10, Max: 20, MaxIsExcluded: true}, 20, false},
		{"Exclusive max: just below max", IntegerInterval{Min: 10, Max: 20, MaxIsExcluded: true}, 19, true},

		{"Fully exclusive: inside", IntegerInterval{Min: 10, MinIsExcluded: true, Max: 20, MaxIsExcluded: true}, 15, true},
		{"Fully exclusive: min boundary", IntegerInterval{Min: 10, MinIsExcluded: true, Max: 20, MaxIsExcluded: true}, 10, false},
		{"Fully exclusive: max boundary", IntegerInterval{Min: 10, MinIsExcluded: true, Max: 20, MaxIsExcluded: true}, 20, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.interval.contains(tt.value)
			if actual != tt.expected {
				t.Errorf("IntegerInterval %+v contains %d: expected %v, got %v", tt.interval, tt.value, tt.expected, actual)
			}
		})
	}
}

func TestNumberIntervalContains(t *testing.T) {
	tests := []struct {
		name     string
		interval NumberInterval
		value    float64
		expected bool
	}{
		{"Inclusive: inside", NumberInterval{Min: 10.0, Max: 20.0}, 15.0, true},
		{"Inclusive: min boundary", NumberInterval{Min: 10.0, Max: 20.0}, 10.0, true},
		{"Inclusive: max boundary", NumberInterval{Min: 10.0, Max: 20.0}, 20.0, true},
		{"Inclusive: just below min", NumberInterval{Min: 10.0, Max: 20.0}, 9.999999, false},  // Outside due to epsilon
		{"Inclusive: just above max", NumberInterval{Min: 10.0, Max: 20.0}, 20.000001, false}, // Outside due to epsilon

		{"Exclusive min: inside", NumberInterval{Min: 10.0, MinIsExcluded: true, Max: 20.0}, 15.0, true},
		{"Exclusive min: min boundary", NumberInterval{Min: 10.0, MinIsExcluded: true, Max: 20.0}, 10.0, false},
		{"Exclusive min: just above min", NumberInterval{Min: 10.0, MinIsExcluded: true, Max: 20.0}, 10.000001, true}, // Just inside

		{"Exclusive max: inside", NumberInterval{Min: 10.0, Max: 20.0, MaxIsExcluded: true}, 15.0, true},
		{"Exclusive max: max boundary", NumberInterval{Min: 10.0, Max: 20.0, MaxIsExcluded: true}, 20.0, false},
		{"Exclusive max: just below max", NumberInterval{Min: 10.0, Max: 20.0, MaxIsExcluded: true}, 19.999999, true}, // Just inside

		{"Fully exclusive: inside", NumberInterval{Min: 10.0, MinIsExcluded: true, Max: 20.0, MaxIsExcluded: true}, 15.0, true},
		{"Fully exclusive: min boundary", NumberInterval{Min: 10.0, MinIsExcluded: true, Max: 20.0, MaxIsExcluded: true}, 10.0, false},
		{"Fully exclusive: max boundary", NumberInterval{Min: 10.0, MinIsExcluded: true, Max: 20.0, MaxIsExcluded: true}, 20.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.interval.contains(tt.value)
			if actual != tt.expected {
				t.Errorf("NumberInterval %+v contains %f: expected %v, got %v", tt.interval, tt.value, tt.expected, actual)
			}
		})
	}
}

// Ensure String and Integer intervals correctly handle equals
func TestIntegerIntervalEquals(t *testing.T) {
	i1 := IntegerInterval{Min: 1, Max: 10, MinIsExcluded: false, MaxIsExcluded: false}
	i2 := IntegerInterval{Min: 1, Max: 10, MinIsExcluded: false, MaxIsExcluded: false}
	i3 := IntegerInterval{Min: 1, Max: 10, MinIsExcluded: true, MaxIsExcluded: false}

	if !i1.equals(i2) { // Changed to Equals as per original code
		t.Errorf("Expected %v to equal %v", i1, i2)
	}
	if i1.equals(i3) { // Changed to Equals as per original code
		t.Errorf("Expected %v not to equal %v", i1, i3)
	}
}

func TestNumberIntervalEquals(t *testing.T) {
	n1 := NumberInterval{Min: 1.0, Max: 10.0, MinIsExcluded: false, MaxIsExcluded: false}
	n2 := NumberInterval{Min: 1.0, Max: 10.0, MinIsExcluded: false, MaxIsExcluded: false}
	n3 := NumberInterval{Min: 1.0, Max: 10.0, MinIsExcluded: true, MaxIsExcluded: false}
	n4 := NumberInterval{Min: 1.0000000000001, Max: 10.0000000000001, MinIsExcluded: false, MaxIsExcluded: false}

	if !n1.equals(n2) { // Changed to Equals as per original code
		t.Errorf("Expected %v to equal %v", n1, n2)
	}
	if n1.equals(n3) { // Changed to Equals as per original code
		t.Errorf("Expected %v not to equal %v", n1, n3)
	}
	if !n1.equals(n4) { // Changed to Equals as per original code
		t.Errorf("Expected %v to equal %v (due to epsilon)", n1, n4)
	}
}
