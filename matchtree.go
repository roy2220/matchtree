package matchtree

import (
	"encoding/json"
	"fmt"
	"iter"
	"math"
	"slices"
)

// MatchTree is a generic tree structure for efficient pattern matching.
// It allows defining rules with various pattern types and searching for matching values based on keys.
type MatchTree[T any] struct {
	types  []MatchType
	values []T
	root   matchNode
}

// MatchType defines the type of data a pattern or key represents.
type MatchType int

const (
	// MatchNone is a placeholder, typically used for leaf nodes.
	MatchNone = MatchType(iota)
	// MatchString represents a string type.
	MatchString
	// MatchInteger represents an integer type.
	MatchInteger
	// MatchIntegerInterval represents an integer interval type.
	MatchIntegerInterval
	// MatchNumberInterval represents a floating-point number interval type.
	MatchNumberInterval
	// NumberOfMatchTypes indicates the total number of defined match types.
	NumberOfMatchTypes = int(iota)
)

var matchType2String = [NumberOfMatchTypes]string{
	MatchNone:            "NONE",
	MatchString:          "STRING",
	MatchInteger:         "INTEGER",
	MatchIntegerInterval: "INTEGER_INTERVAL",
	MatchNumberInterval:  "NUMBER_INTERVAL",
}

// String returns the string representation of a MatchType.
func (t MatchType) String() string {
	i := int(t)
	if i >= 0 && i < NumberOfMatchTypes {
		return matchType2String[t]
	}
	return fmt.Sprintf("UNKNOWN(%d)", i)
}

// ParseMatchType parses a string into a MatchType.
func ParseMatchType(s string) (MatchType, error) {
	for i, ss := range matchType2String {
		if ss == s {
			return MatchType(i), nil
		}
	}
	return 0, fmt.Errorf("unknown match type %q", s)
}

// MarshalJSON marshals the MatchType to its string representation.
func (t MatchType) MarshalJSON() ([]byte, error) { return json.Marshal(t.String()) }

// UnmarshalJSON unmarshals a JSON string into a MatchType.
func (t *MatchType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	var err error
	*t, err = ParseMatchType(s)
	return err
}

// NewMatchTree creates a new MatchTree with the specified sequence of MatchTypes.
// The order of types matters and defines the structure of the tree.
func NewMatchTree[T any](types []MatchType) *MatchTree[T] {
	for _, type1 := range types {
		switch type1 {
		case MatchString, MatchInteger, MatchIntegerInterval, MatchNumberInterval:
		default:
			panic(fmt.Sprintf("unknown match type: %v", type1))
		}
	}
	return &MatchTree[T]{
		types: types,
	}
}

// MatchRule represents a single rule to be added to the MatchTree.
// It consists of a sequence of patterns, a value to associate, and a priority.
type MatchRule[T any] struct {
	Patterns []MatchPattern `json:"patterns"`
	Value    T              `json:"value"`
	Priority int            `json:"priority"`
}

// MatchPattern defines a single pattern within a MatchRule.
// It can be an 'any' pattern, an 'inverse' pattern, or a specific value/interval pattern.
type MatchPattern struct {
	Type MatchType `json:"type"`

	// IsAny indicates if this pattern matches any value for its type.
	IsAny bool `json:"is_any"`

	// IsInverse indicates if this pattern matches any value NOT in its specified list/intervals.
	IsInverse bool `json:"is_inverse"`

	// Strings for MatchString type.
	Strings []string `json:"strings"`

	// Integers for MatchInteger type.
	Integers []int64 `json:"integers"`

	// IntegerIntervals for MatchIntegerInterval type.
	IntegerIntervals []IntegerInterval `json:"integer_intervals"`

	// NumberIntervals for MatchNumberInterval type.
	NumberIntervals []NumberInterval `json:"number_intervals"`

	// internal fields for pattern walking
	currentString          string
	currentInteger         int64
	currentIntegerInterval IntegerInterval
	currentNumberInterval  NumberInterval
}

// IntegerInterval represents a closed, open, or half-open interval for integers.
type IntegerInterval struct {
	Min           *int64 `json:"min"`
	MinIsExcluded bool   `json:"min_is_excluded"`
	Max           *int64 `json:"max"`
	MaxIsExcluded bool   `json:"max_is_excluded"`
}

// Int64Ptr is a helper function to create a pointer to an int64 value.
func Int64Ptr(x int64) *int64 { return &x }

// Equals checks if two IntegerIntervals are equal.
func (i IntegerInterval) Equals(other IntegerInterval) bool {
	if !((i.Min == nil) == (other.Min == nil) &&
		(i.Max == nil) == (other.Max == nil)) {
		return false
	}

	if i.Min != nil {
		if *i.Min != *other.Min {
			return false
		}
		if i.MinIsExcluded != other.MinIsExcluded {
			return false
		}
	}

	if i.Max != nil {
		if *i.Max != *other.Max {
			return false
		}
		if i.MaxIsExcluded != other.MaxIsExcluded {
			return false
		}
	}

	return true
}

// Contains checks if the given integer `x` falls within the interval.
func (i IntegerInterval) Contains(x int64) bool {
	if i.Min != nil {
		y := *i.Min
		if i.MinIsExcluded {
			if x <= y {
				return false
			}
		} else {
			if x < y {
				return false
			}
		}
	}
	if i.Max != nil {
		y := *i.Max
		if i.MaxIsExcluded {
			if x >= y {
				return false
			}
		} else {
			if x > y {
				return false
			}
		}
	}
	return true
}

// NumberInterval represents a closed, open, or half-open interval for floating-point numbers.
type NumberInterval struct {
	Min           *float64 `json:"min"`
	MinIsExcluded bool     `json:"min_is_excluded"`
	Max           *float64 `json:"max"`
	MaxIsExcluded bool     `json:"max_is_excluded"`
}

// Float64Ptr is a helper function to create a pointer to a float64 value.
func Float64Ptr(x float64) *float64 { return &x }

const epsilon = 1e-10

// Equals checks if two NumberIntervals are equal, considering floating-point precision.
func (i NumberInterval) Equals(other NumberInterval) bool {
	if !((i.Min == nil) == (other.Min == nil) &&
		(i.Max == nil) == (other.Max == nil)) {
		return false
	}

	if i.Min != nil {
		if math.Abs(*i.Min-*other.Min) >= epsilon {
			return false
		}
		if i.MinIsExcluded != other.MinIsExcluded {
			return false
		}
	}

	if i.Max != nil {
		if math.Abs(*i.Max-*other.Max) >= epsilon {
			return false
		}
		if i.MaxIsExcluded != other.MaxIsExcluded {
			return false
		}
	}

	return true
}

// Contains checks if the given floating-point number `x` falls within the interval,
// considering floating-point precision.
func (i NumberInterval) Contains(x float64) bool {
	if i.Min != nil {
		y := *i.Min
		if i.MinIsExcluded {
			if x <= y+epsilon {
				return false
			}
		} else {
			if x < y-epsilon {
				return false
			}
		}
	}
	if i.Max != nil {
		y := *i.Max
		if i.MaxIsExcluded {
			if x >= y-epsilon {
				return false
			}
		} else {
			if x > y+epsilon {
				return false
			}
		}
	}
	return true
}

// AddRule adds a new MatchRule to the MatchTree.
// It returns an error if the rule's patterns do not match the tree's defined types.
func (t *MatchTree[T]) AddRule(rule MatchRule[T]) error {
	if len(rule.Patterns) != len(t.types) {
		return fmt.Errorf("unexpected number of match patterns; expected=%v actual=%v", len(t.types), len(rule.Patterns))
	}
	for i, pattern := range rule.Patterns {
		if pattern.Type != t.types[i] {
			return fmt.Errorf("unexpected match type; expected=%v actual=%v", t.types[i], pattern.Type)
		}
	}

	valueIndex := len(t.values)
	t.values = append(t.values, rule.Value)

	patterns := slices.Clone(rule.Patterns)
	for i := range patterns {
		pattern := &patterns[i]
		pattern.Strings = cloneStrings(pattern.Strings)
		pattern.Integers = cloneIntegers(pattern.Integers)
		pattern.IntegerIntervals = cloneIntegerIntervals(pattern.IntegerIntervals)
		pattern.NumberIntervals = cloneNumberIntervals(pattern.NumberIntervals)

	}

	var walkPatterns func(int)
	walkPatterns = func(i int) {
		if i == len(patterns) {
			t.doAddRule(patterns, valueIndex, rule.Priority)
			return
		}

		pattern := &patterns[i]
		if pattern.IsAny {
			walkPatterns(i + 1)
			return
		}
		if pattern.IsInverse {
			walkPatterns(i + 1)
			return
		}

		switch pattern.Type {
		case MatchString:
			for _, v := range pattern.Strings {
				pattern.currentString = v
				walkPatterns(i + 1)
			}
		case MatchInteger:
			for _, v := range pattern.Integers {
				pattern.currentInteger = v
				walkPatterns(i + 1)
			}
		case MatchIntegerInterval:
			for _, v := range pattern.IntegerIntervals {
				pattern.currentIntegerInterval = v
				walkPatterns(i + 1)
			}
		case MatchNumberInterval:
			for _, v := range pattern.NumberIntervals {
				pattern.currentNumberInterval = v
				walkPatterns(i + 1)
			}
		default:
			panic("unreachable")
		}
	}
	walkPatterns(0)
	return nil
}

func cloneStrings(s []string) []string {
	clone := make([]string, 0, len(s))
	for _, v := range s {
		if slices.Contains(clone, v) {
			continue
		}
		clone = append(clone, v)
	}
	return clone
}

func cloneIntegers(s []int64) []int64 {
	clone := make([]int64, 0, len(s))
	for _, v := range s {
		if slices.Contains(clone, v) {
			continue
		}
		clone = append(clone, v)
	}
	return clone
}

func cloneIntegerIntervals(s []IntegerInterval) []IntegerInterval {
	clone := make([]IntegerInterval, 0, len(s))
	for _, v := range s {
		if slices.ContainsFunc(clone, v.Equals) {
			continue
		}
		clone = append(clone, v)
	}
	return clone
}

func cloneNumberIntervals(s []NumberInterval) []NumberInterval {
	clone := make([]NumberInterval, 0, len(s))
	for _, v := range s {
		if slices.ContainsFunc(clone, v.Equals) {
			continue
		}
		clone = append(clone, v)
	}
	return clone
}

func (t *MatchTree[T]) doAddRule(patterns []MatchPattern, valueIndex int, priority int) {
	getOrInsertNode := func(newNodeType MatchType) matchNode {
		node := t.root
		if node == nil {
			node = newMatchNode(newNodeType)
			t.root = node
		}
		return node
	}

	for i := range patterns {
		// non-leaf
		pattern := &patterns[i]
		node := getOrInsertNode(pattern.Type)

		getOrInsertNode = func(
			lastNode matchNode,
			lastPattern *MatchPattern,
		) func(MatchType) matchNode {
			return func(newNodeType MatchType) matchNode {
				return lastNode.GetOrInsertChild(lastPattern, newNodeType)
			}
		}(node, pattern)
	}

	// leaf
	node := getOrInsertNode(MatchNone)
	node.AddResult(matchResult{
		ValueIndex: valueIndex,
		Priority:   priority,
	})
}

// MatchKey represents a single key to search within the MatchTree.
// It specifies the type and the value for that key.
type MatchKey struct {
	Type MatchType `json:"type"`

	// String for MatchString type.
	String string `json:"string"`

	// Integer for MatchInteger, MatchIntegerInterval types.
	Integer int64 `json:"integer"`

	// Number for MatchNumberInterval type.
	Number float64 `json:"number"`
}

// Search traverses the MatchTree with the given keys and returns a slice of matching values.
// The returned values are sorted by priority (descending) and then by their insertion order.
// It returns an error if the keys do not match the tree's defined types.
func (t *MatchTree[T]) Search(keys []MatchKey) ([]T, error) {
	if len(keys) != len(t.types) {
		return nil, fmt.Errorf("unexpected number of match keys; expected=%v actual=%v", len(t.types), len(keys))
	}
	for i, key := range keys {
		if key.Type != t.types[i] {
			return nil, fmt.Errorf("unexpected match type; expected=%v actual=%v", t.types[i], key.Type)
		}
	}

	var nodes []matchNode
	if t.root != nil {
		nodes = []matchNode{t.root}
	}
	var nextNodes []matchNode
	for _, key := range keys {
		for _, node := range nodes {
			// non-leaf
			nextNodes = slices.AppendSeq(nextNodes, node.FindChildren(key))
		}
		nodes, nextNodes = nextNodes, nodes[:0]
	}
	if len(nodes) == 0 {
		return nil, nil
	}

	return t.extractValues(nodes), nil
}

func (t *MatchTree[T]) extractValues(nodes []matchNode) []T {
	n := 0
	for _, node := range nodes {
		n += len(node.GetResults())
	}
	if n == 1 {
		return []T{t.values[nodes[0].GetResults()[0].ValueIndex]}
	}

	results := make([]matchResult, 0, n)
	for _, node := range nodes {
		results = append(results, node.GetResults()...)
	}
	slices.SortFunc(results, func(x, y matchResult) int {
		delta := y.Priority - x.Priority
		if delta == 0 {
			delta = x.ValueIndex - y.ValueIndex
		}
		return delta
	})
	lastValueIndex := -1
	n = 0
	for _, result := range results {
		if result.ValueIndex == lastValueIndex {
			continue
		}
		results[n] = result
		n++
		lastValueIndex = result.ValueIndex
	}
	results = results[:n]

	values := make([]T, n)
	for i, result := range results {
		values[i] = t.values[result.ValueIndex]
	}
	return values
}

// matchNode is an interface that defines the behavior of nodes within the MatchTree.
type matchNode interface {
	// GetOrInsertChild retrieves an existing child node or inserts a new one based on the pattern and newChildType.
	GetOrInsertChild(pattern *MatchPattern, newChildType MatchType) matchNode
	// FindChildren finds child nodes that match the given key.
	FindChildren(key MatchKey) iter.Seq[matchNode]

	// AddResult adds a match result to a leaf node.
	AddResult(result matchResult)
	// GetResults returns the match results associated with a leaf node.
	GetResults() []matchResult
}

// matchResult stores the index of the matched value and its priority.
type matchResult struct {
	ValueIndex int
	Priority   int
}

var matchNodeFactories = [NumberOfMatchTypes]func() matchNode{
	MatchNone:            func() matchNode { return new(matchNodeOfNone) },
	MatchString:          func() matchNode { return new(matchNodeOfString) },
	MatchInteger:         func() matchNode { return new(matchNodeOfInteger) },
	MatchIntegerInterval: func() matchNode { return new(matchNodeOfIntegerInterval) },
	MatchNumberInterval:  func() matchNode { return new(matchNodeOfNumberInterval) },
}

func newMatchNode(type1 MatchType) matchNode { return matchNodeFactories[type1]() }

// ----- dummy match node -----

type dummyMatchNode struct{}

var _ matchNode = (*dummyMatchNode)(nil)

func (n dummyMatchNode) GetOrInsertChild(pattern *MatchPattern, newChildType MatchType) matchNode {
	panic("unreachable")
}
func (n dummyMatchNode) FindChildren(key MatchKey) iter.Seq[matchNode] { panic("unreachable") }
func (n dummyMatchNode) AddResult(result matchResult)                  { panic("unreachable") }
func (n dummyMatchNode) GetResults() []matchResult                     { panic("unreachable") }

// ----- match node of none -----

type matchNodeOfNone struct {
	dummyMatchNode

	results []matchResult
}

var _ matchNode = (*matchNodeOfNone)(nil)

func (n *matchNodeOfNone) AddResult(result matchResult) {
	n.results = append(n.results, result)
}
func (n *matchNodeOfNone) GetResults() []matchResult { return n.results }

// ----- match node of string -----

type matchNodeOfString struct {
	dummyMatchNode

	children            map[string]matchNode
	inverseChildren     []matchNodeWithRefCount
	inverseChildIndexes map[string][]int
	anyChild            matchNode
}

var _ matchNode = (*matchNodeOfString)(nil)

type stringAndMatchNode struct {
	String    string
	MatchNode matchNode
}

func (n *matchNodeOfString) GetOrInsertChild(pattern *MatchPattern, newChildType MatchType) matchNode {
	if pattern.IsAny {
		child := n.anyChild
		if child == nil {
			child = newMatchNode(newChildType)
			n.anyChild = child
		}
		return child
	}

	if pattern.IsInverse {
		refCounts := make([]int, len(n.inverseChildren))
		for _, v := range pattern.Strings {
			for _, childIndex := range n.inverseChildIndexes[v] {
				refCounts[childIndex]++
			}
		}
		maxRefCount := len(pattern.Strings)
		for childIndex, refCount := range refCounts {
			if refCount == maxRefCount && n.inverseChildren[childIndex].MaxRefCount == maxRefCount {
				return n.inverseChildren[childIndex].MatchNode
			}
		}
		newChild := newMatchNode(newChildType)
		newChildIndex := len(n.inverseChildren)
		n.inverseChildren = append(n.inverseChildren, matchNodeWithRefCount{
			MatchNode:   newChild,
			MaxRefCount: maxRefCount,
		})
		inverseChildIndexes := n.inverseChildIndexes
		if inverseChildIndexes == nil {
			inverseChildIndexes = make(map[string][]int, maxRefCount)
			n.inverseChildIndexes = inverseChildIndexes
		}
		for _, v := range pattern.Strings {
			inverseChildIndexes[v] = append(inverseChildIndexes[v], newChildIndex)
		}
		return newChild
	}

	children := n.children
	if children == nil {
		children = make(map[string]matchNode, 1)
		n.children = children
	}
	child, ok := children[pattern.currentString]
	if !ok {
		child = newMatchNode(newChildType)
		children[pattern.currentString] = child
	}
	return child
}

func (n *matchNodeOfString) FindChildren(key MatchKey) iter.Seq[matchNode] {
	return func(yield func(matchNode) bool) {
		if child, ok := n.children[key.String]; ok {
			if !yield(child) {
				return
			}
		}

		if len(n.inverseChildren) >= 1 {
			refCounts := make([]int, len(n.inverseChildren))
			for _, childIndex := range n.inverseChildIndexes[key.String] {
				refCounts[childIndex]++
			}
			for childIndex, refCount := range refCounts {
				if refCount >= 1 {
					continue
				}
				if !yield(n.inverseChildren[childIndex].MatchNode) {
					return
				}
			}
		}

		if child := n.anyChild; child != nil {
			if !yield(child) {
				return
			}
		}
	}
}

// ----- match node of integer -----

type matchNodeOfInteger struct {
	dummyMatchNode

	children            map[int64]matchNode
	inverseChildren     []matchNodeWithRefCount
	inverseChildIndexes map[int64][]int
	anyChild            matchNode
}

var _ matchNode = (*matchNodeOfInteger)(nil)

type integerAndMatchNode struct {
	Integer   int64
	MatchNode matchNode
}

func (n *matchNodeOfInteger) GetOrInsertChild(pattern *MatchPattern, newChildType MatchType) matchNode {
	if pattern.IsAny {
		child := n.anyChild
		if child == nil {
			child = newMatchNode(newChildType)
			n.anyChild = child
		}
		return child
	}

	if pattern.IsInverse {
		refCounts := make([]int, len(n.inverseChildren))
		for _, v := range pattern.Integers {
			for _, childIndex := range n.inverseChildIndexes[v] {
				refCounts[childIndex]++
			}
		}
		maxRefCount := len(pattern.Integers)
		for childIndex, refCount := range refCounts {
			if refCount == maxRefCount && n.inverseChildren[childIndex].MaxRefCount == maxRefCount {
				return n.inverseChildren[childIndex].MatchNode
			}
		}
		newChild := newMatchNode(newChildType)
		newChildIndex := len(n.inverseChildren)
		n.inverseChildren = append(n.inverseChildren, matchNodeWithRefCount{
			MatchNode:   newChild,
			MaxRefCount: maxRefCount,
		})
		inverseChildIndexes := n.inverseChildIndexes
		if inverseChildIndexes == nil {
			inverseChildIndexes = make(map[int64][]int, maxRefCount)
			n.inverseChildIndexes = inverseChildIndexes
		}
		for _, v := range pattern.Integers {
			inverseChildIndexes[v] = append(inverseChildIndexes[v], newChildIndex)
		}
		return newChild
	}

	children := n.children
	if children == nil {
		children = make(map[int64]matchNode, 1)
		n.children = children
	}
	child, ok := children[pattern.currentInteger]
	if !ok {
		child = newMatchNode(newChildType)
		children[pattern.currentInteger] = child
	}
	return child
}

func (n *matchNodeOfInteger) FindChildren(key MatchKey) iter.Seq[matchNode] {
	return func(yield func(matchNode) bool) {
		if child, ok := n.children[key.Integer]; ok {
			if !yield(child) {
				return
			}
		}

		if len(n.inverseChildren) >= 1 {
			refCounts := make([]int, len(n.inverseChildren))
			for _, childIndex := range n.inverseChildIndexes[key.Integer] {
				refCounts[childIndex]++
			}
			for childIndex, refCount := range refCounts {
				if refCount >= 1 {
					continue
				}
				if !yield(n.inverseChildren[childIndex].MatchNode) {
					return
				}
			}
		}

		if child := n.anyChild; child != nil {
			if !yield(child) {
				return
			}
		}
	}
}

// ----- match node of integer interval -----

type matchNodeOfIntegerInterval struct {
	dummyMatchNode

	children            []integerIntervalAndMatchNode
	inverseChildren     []matchNodeWithRefCount
	inverseChildIndexes []integerIntervalAndMatchNodeIndexes
	anyChild            matchNode
}

var _ matchNode = (*matchNodeOfIntegerInterval)(nil)

type integerIntervalAndMatchNode struct {
	IntegerInterval IntegerInterval
	MatchNode       matchNode
}

type integerIntervalAndMatchNodeIndexes struct {
	IntegerInterval  IntegerInterval
	MatchNodeIndexes []int
}

func (n *matchNodeOfIntegerInterval) GetOrInsertChild(pattern *MatchPattern, newChildType MatchType) matchNode {
	if pattern.IsAny {
		child := n.anyChild
		if child == nil {
			child = newMatchNode(newChildType)
			n.anyChild = child
		}
		return child
	}

	if pattern.IsInverse {
		refCounts := make([]int, len(n.inverseChildren))
		for _, v := range pattern.IntegerIntervals {
			i := slices.IndexFunc(n.inverseChildIndexes, func(x integerIntervalAndMatchNodeIndexes) bool {
				return x.IntegerInterval.Equals(v)
			})
			if i < 0 {
				continue
			}
			for _, childIndex := range n.inverseChildIndexes[i].MatchNodeIndexes {
				refCounts[childIndex]++
			}
		}
		maxRefCount := len(pattern.IntegerIntervals)
		for childIndex, refCount := range refCounts {
			if refCount == maxRefCount && n.inverseChildren[childIndex].MaxRefCount == maxRefCount {
				return n.inverseChildren[childIndex].MatchNode
			}
		}
		newChild := newMatchNode(newChildType)
		newChildIndex := len(n.inverseChildren)
		n.inverseChildren = append(n.inverseChildren, matchNodeWithRefCount{
			MatchNode:   newChild,
			MaxRefCount: maxRefCount,
		})
		for _, v := range pattern.IntegerIntervals {
			i := slices.IndexFunc(n.inverseChildIndexes, func(x integerIntervalAndMatchNodeIndexes) bool {
				return x.IntegerInterval.Equals(v)
			})
			if i < 0 {
				n.inverseChildIndexes = append(n.inverseChildIndexes, integerIntervalAndMatchNodeIndexes{
					IntegerInterval:  v,
					MatchNodeIndexes: []int{newChildIndex},
				})
				continue
			}
			n.inverseChildIndexes[i].MatchNodeIndexes = append(n.inverseChildIndexes[i].MatchNodeIndexes, newChildIndex)
		}
		return newChild
	}

	if childIndex := slices.IndexFunc(n.children, func(x integerIntervalAndMatchNode) bool {
		return x.IntegerInterval.Equals(pattern.currentIntegerInterval)
	}); childIndex >= 0 {
		return n.children[childIndex].MatchNode
	}
	newChild := newMatchNode(newChildType)
	n.children = append(n.children, integerIntervalAndMatchNode{
		IntegerInterval: pattern.currentIntegerInterval,
		MatchNode:       newChild,
	})
	return newChild
}

func (n *matchNodeOfIntegerInterval) FindChildren(key MatchKey) iter.Seq[matchNode] {
	return func(yield func(matchNode) bool) {
		for i := range n.children {
			if n.children[i].IntegerInterval.Contains(key.Integer) {
				if !yield(n.children[i].MatchNode) {
					return
				}
			}
		}

		if len(n.inverseChildren) >= 1 {
			refCounts := make([]int, len(n.inverseChildren))
			for _, v := range n.inverseChildIndexes {
				if !v.IntegerInterval.Contains(key.Integer) {
					continue
				}
				for _, childIndex := range v.MatchNodeIndexes {
					refCounts[childIndex]++
				}
			}
			for childIndex, refCount := range refCounts {
				if refCount >= 1 {
					continue
				}
				if !yield(n.inverseChildren[childIndex].MatchNode) {
					return
				}
			}
		}

		if child := n.anyChild; child != nil {
			if !yield(child) {
				return
			}
		}
	}
}

// ----- match node of number interval -----

type matchNodeOfNumberInterval struct {
	dummyMatchNode

	children            []numberIntervalAndMatchNode
	inverseChildren     []matchNodeWithRefCount
	inverseChildIndexes []numberIntervalAndMatchNodeIndexes
	anyChild            matchNode
}

var _ matchNode = (*matchNodeOfNumberInterval)(nil)

type numberIntervalAndMatchNode struct {
	NumberInterval NumberInterval
	MatchNode      matchNode
}

type numberIntervalAndMatchNodeIndexes struct {
	NumberInterval   NumberInterval
	MatchNodeIndexes []int
}

func (n *matchNodeOfNumberInterval) GetOrInsertChild(pattern *MatchPattern, newChildType MatchType) matchNode {
	if pattern.IsAny {
		child := n.anyChild
		if child == nil {
			child = newMatchNode(newChildType)
			n.anyChild = child
		}
		return child
	}

	if pattern.IsInverse {
		refCounts := make([]int, len(n.inverseChildren))
		for _, v := range pattern.NumberIntervals {
			i := slices.IndexFunc(n.inverseChildIndexes, func(x numberIntervalAndMatchNodeIndexes) bool {
				return x.NumberInterval.Equals(v)
			})
			if i < 0 {
				continue
			}
			for _, childIndex := range n.inverseChildIndexes[i].MatchNodeIndexes {
				refCounts[childIndex]++
			}
		}
		maxRefCount := len(pattern.NumberIntervals)
		for childIndex, refCount := range refCounts {
			if refCount == maxRefCount && n.inverseChildren[childIndex].MaxRefCount == maxRefCount {
				return n.inverseChildren[childIndex].MatchNode
			}
		}
		newChild := newMatchNode(newChildType)
		newChildIndex := len(n.inverseChildren)
		n.inverseChildren = append(n.inverseChildren, matchNodeWithRefCount{
			MatchNode:   newChild,
			MaxRefCount: maxRefCount,
		})
		for _, v := range pattern.NumberIntervals {
			i := slices.IndexFunc(n.inverseChildIndexes, func(x numberIntervalAndMatchNodeIndexes) bool {
				return x.NumberInterval.Equals(v)
			})
			if i < 0 {
				n.inverseChildIndexes = append(n.inverseChildIndexes, numberIntervalAndMatchNodeIndexes{
					NumberInterval:   v,
					MatchNodeIndexes: []int{newChildIndex},
				})
				continue
			}
			n.inverseChildIndexes[i].MatchNodeIndexes = append(n.inverseChildIndexes[i].MatchNodeIndexes, newChildIndex)
		}
		return newChild
	}

	if childIndex := slices.IndexFunc(n.children, func(x numberIntervalAndMatchNode) bool {
		return x.NumberInterval.Equals(pattern.currentNumberInterval)
	}); childIndex >= 0 {
		return n.children[childIndex].MatchNode
	}
	newChild := newMatchNode(newChildType)
	n.children = append(n.children, numberIntervalAndMatchNode{
		NumberInterval: pattern.currentNumberInterval,
		MatchNode:      newChild,
	})
	return newChild
}

func (n *matchNodeOfNumberInterval) FindChildren(key MatchKey) iter.Seq[matchNode] {
	return func(yield func(matchNode) bool) {
		for i := range n.children {
			if n.children[i].NumberInterval.Contains(key.Number) {
				if !yield(n.children[i].MatchNode) {
					return
				}
			}
		}

		if len(n.inverseChildren) >= 1 {
			refCounts := make([]int, len(n.inverseChildren))
			for _, v := range n.inverseChildIndexes {
				if !v.NumberInterval.Contains(key.Number) {
					continue
				}
				for _, childIndex := range v.MatchNodeIndexes {
					refCounts[childIndex]++
				}
			}
			for childIndex, refCount := range refCounts {
				if refCount >= 1 {
					continue
				}
				if !yield(n.inverseChildren[childIndex].MatchNode) {
					return
				}
			}
		}

		if child := n.anyChild; child != nil {
			if !yield(child) {
				return
			}
		}
	}
}

// ----- match node common -----

type matchNodeWithRefCount struct {
	MatchNode   matchNode
	MaxRefCount int
}
