package matchtree

import (
	"fmt"
	"iter"
	"math"
)

const epsilon = 1e-10

type MatchTree struct {
	types []MatchType
	root  matchNode
}

type MatchType int

const (
	MatchNone = MatchType(iota)
	MatchString
	MatchInteger
	MatchIntegerInterval
	MatchNumberInterval
)

func NewMatchTree(types []MatchType) *MatchTree {
	for _, type1 := range types {
		switch type1 {
		case MatchString, MatchInteger, MatchIntegerInterval, MatchNumberInterval:
		default:
			panic(fmt.Sprintf("unknown match type: %v", type1))
		}
	}
	return &MatchTree{
		types: types,
	}
}

type matchNode interface {
	// for non-leaf
	GetOrNewChild(pattern MatchPattern, newChild func() matchNode) matchNode
	FindChildren(key MatchKey) iter.Seq[matchNode]

	// for leaf
	AddValue(value any)
	GetValues() []any
}

type MatchRule struct {
	Patterns []MatchPattern
	Value    any
}

type MatchPattern struct {
	Type MatchType

	IsAny bool

	// MatchString
	String string

	// MatchInteger
	Integer int64

	// MatchIntegerInterval
	IntegerInterval IntegerInterval

	// MatchNumberInterval
	NumberInterval NumberInterval
}

type IntegerInterval struct {
	Min           int64
	MinIsExcluded bool
	Max           int64
	MaxIsExcluded bool
}

type NumberInterval struct {
	Min           float64
	MinIsExcluded bool
	Max           float64
	MaxIsExcluded bool
}

func (t *MatchTree) AddRule(rule MatchRule) {
	if len(rule.Patterns) != len(t.types) {
		panic(fmt.Sprintf("unexpected number of match patterns; expected=%v actual=%v", len(t.types), len(rule.Patterns)))
	}
	for i, pattern := range rule.Patterns {
		if pattern.Type != t.types[i] {
			panic(fmt.Sprintf("unexpected match type; expected=%v actual=%v", t.types[i], pattern.Type))
		}
	}

	getOrNewNode := func(newNode func() matchNode) matchNode {
		node := t.root
		if node == nil {
			node = newNode()
			t.root = node
		}
		return node
	}
	for _, pattern := range rule.Patterns {
		// non-leaf
		newNode := getMatchNodeFactory(pattern.Type)
		node := getOrNewNode(newNode)

		getOrNewNode = func(newNode func() matchNode) matchNode {
			return node.GetOrNewChild(pattern, newNode)
		}
	}
	{
		// leaf
		newNode := getMatchNodeFactory(MatchNone)
		node := getOrNewNode(newNode)
		node.AddValue(rule.Value)
	}
}

type MatchKey struct {
	Type MatchType

	// for MatchString
	String string

	// for MatchInteger, MatchIntegerInterval
	Integer int64

	// for MatchNumberInterval
	Number float64
}

func (t *MatchTree) Search(keys []MatchKey) ([]any, bool) {
	if len(keys) != len(t.types) {
		panic(fmt.Sprintf("unexpected number of match keys; expected=%v actual=%v", len(t.types), len(keys)))
	}
	for i, key := range keys {
		if key.Type != t.types[i] {
			panic(fmt.Sprintf("unexpected match type; expected=%v actual=%v", t.types[i], key.Type))
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
			for nextNode := range node.FindChildren(key) {
				nextNodes = append(nextNodes, nextNode)
			}
		}
		nodes, nextNodes = nextNodes, nodes[:0]
	}
	if len(nodes) == 0 {
		return nil, false
	}

	var values []any
	for _, node := range nodes {
		// leaf
		values = append(values, node.GetValues()...)
	}
	return values, true
}

// ---------- match node implementation ----------

func getMatchNodeFactory(type1 MatchType) func() matchNode {
	return [...]func() matchNode{
		MatchNone:            func() matchNode { return new(matchNodeOfNone) },
		MatchString:          func() matchNode { return new(matchNodeOfString) },
		MatchInteger:         func() matchNode { return new(matchNodeOfInteger) },
		MatchIntegerInterval: func() matchNode { return new(matchNodeOfIntegerInterval) },
		MatchNumberInterval:  func() matchNode { return new(matchNodeOfNumberInterval) },
	}[int(type1)]
}

type dummyMatchNode struct{}

var _ matchNode = (*dummyMatchNode)(nil)

func (n dummyMatchNode) GetOrNewChild(pattern MatchPattern, newChild func() matchNode) matchNode {
	return nil
}
func (n dummyMatchNode) FindChildren(key MatchKey) iter.Seq[matchNode] { return nil }
func (n dummyMatchNode) AddValue(value any)                            {}
func (n dummyMatchNode) GetValues() []any                              { return nil }

type matchNodeOfNone struct {
	dummyMatchNode

	values []any
}

var _ matchNode = (*matchNodeOfNone)(nil)

func (n *matchNodeOfNone) AddValue(value any) { n.values = append(n.values, value) }
func (n *matchNodeOfNone) GetValues() []any   { return n.values }

type matchNodeOfString struct {
	dummyMatchNode

	children map[string]matchNode
	anyChild matchNode
}

var _ matchNode = (*matchNodeOfString)(nil)

func (n *matchNodeOfString) GetOrNewChild(pattern MatchPattern, newChild func() matchNode) matchNode {
	if pattern.IsAny {
		child := n.anyChild
		if child == nil {
			child = newChild()
			n.anyChild = child
		}
		return child
	}
	children := n.children
	if children == nil {
		children = make(map[string]matchNode, 1)
		n.children = children
	}
	child, ok := children[pattern.String]
	if !ok {
		child = newChild()
		children[pattern.String] = child
	}
	return child
}

func (n *matchNodeOfString) FindChildren(key MatchKey) iter.Seq[matchNode] {
	return func(yield func(matchNode) bool) {
		if next, ok := n.children[key.String]; ok {
			if !yield(next) {
				return
			}
		}
		if next := n.anyChild; next != nil {
			if !yield(next) {
				return
			}
		}
	}
}

type matchNodeOfInteger struct {
	dummyMatchNode

	children map[int64]matchNode
	anyChild matchNode
}

var _ matchNode = (*matchNodeOfInteger)(nil)

func (n *matchNodeOfInteger) GetOrNewChild(pattern MatchPattern, newChild func() matchNode) matchNode {
	if pattern.IsAny {
		child := n.anyChild
		if child == nil {
			child = newChild()
			n.anyChild = child
		}
		return child
	}
	children := n.children
	if children == nil {
		children = make(map[int64]matchNode, 1)
		n.children = children
	}
	child, ok := children[pattern.Integer]
	if !ok {
		child = newChild()
		children[pattern.Integer] = child
	}
	return child
}

func (n *matchNodeOfInteger) FindChildren(key MatchKey) iter.Seq[matchNode] {
	return func(yield func(matchNode) bool) {
		if next, ok := n.children[key.Integer]; ok {
			if !yield(next) {
				return
			}
		}
		if next := n.anyChild; next != nil {
			if !yield(next) {
				return
			}
		}
	}
}

type matchNodeOfIntegerInterval struct {
	dummyMatchNode

	children []integerIntervalAndMatchNode
	anyChild matchNode
}

var _ matchNode = (*matchNodeOfIntegerInterval)(nil)

type integerIntervalAndMatchNode struct {
	IntegerInterval IntegerInterval
	MatchNode       matchNode
}

func (n *matchNodeOfIntegerInterval) GetOrNewChild(pattern MatchPattern, newChild func() matchNode) matchNode {
	if pattern.IsAny {
		child := n.anyChild
		if child == nil {
			child = newChild()
			n.anyChild = child
		}
		return child
	}
	i := 0
	j := len(n.children)
	for ; i < j; i++ {
		if pattern.IntegerInterval.equals(n.children[i].IntegerInterval) {
			break
		}
	}
	if i == j {
		n.children = append(n.children, integerIntervalAndMatchNode{
			IntegerInterval: pattern.IntegerInterval,
			MatchNode:       newChild(),
		})
	}
	return n.children[i].MatchNode
}

func (i IntegerInterval) equals(other IntegerInterval) bool {
	return i == other
}

func (n *matchNodeOfIntegerInterval) FindChildren(key MatchKey) iter.Seq[matchNode] {
	return func(yield func(matchNode) bool) {
		for i := range n.children {
			if n.children[i].IntegerInterval.contains(key.Integer) {
				if !yield(n.children[i].MatchNode) {
					return
				}
			}
		}
		if next := n.anyChild; next != nil {
			if !yield(next) {
				return
			}
		}
	}
}

func (i IntegerInterval) contains(x int64) bool {
	if i.MinIsExcluded {
		if x <= i.Min {
			return false
		}
	} else {
		if x < i.Min {
			return false
		}
	}
	if i.MaxIsExcluded {
		if x >= i.Max {
			return false
		}
	} else {
		if x > i.Max {
			return false
		}
	}
	return true
}

type matchNodeOfNumberInterval struct {
	dummyMatchNode

	children []numberIntervalAndMatchNode
	anyChild matchNode
}

var _ matchNode = (*matchNodeOfNumberInterval)(nil)

type numberIntervalAndMatchNode struct {
	NumberInterval NumberInterval
	MatchNode      matchNode
}

func (n *matchNodeOfNumberInterval) GetOrNewChild(pattern MatchPattern, newChild func() matchNode) matchNode {
	if pattern.IsAny {
		child := n.anyChild
		if child == nil {
			child = newChild()
			n.anyChild = child
		}
		return child
	}
	i := 0
	j := len(n.children)
	for ; i < j; i++ {
		if pattern.NumberInterval.equals(n.children[i].NumberInterval) {
			break
		}
	}
	if i == j {
		n.children = append(n.children, numberIntervalAndMatchNode{
			NumberInterval: pattern.NumberInterval,
			MatchNode:      newChild(),
		})
	}
	return n.children[i].MatchNode
}

func (i NumberInterval) equals(other NumberInterval) bool {
	return i.MinIsExcluded == other.MinIsExcluded &&
		i.MaxIsExcluded == other.MaxIsExcluded &&
		math.Abs(i.Min-other.Min) < epsilon &&
		math.Abs(i.Max-other.Max) < epsilon
}

func (n *matchNodeOfNumberInterval) FindChildren(key MatchKey) iter.Seq[matchNode] {
	return func(yield func(matchNode) bool) {
		for i := range n.children {
			if n.children[i].NumberInterval.contains(key.Number) {
				if !yield(n.children[i].MatchNode) {
					return
				}
			}
		}
		if next := n.anyChild; next != nil {
			if !yield(next) {
				return
			}
		}
	}
}

func (i NumberInterval) contains(x float64) bool {
	if i.MinIsExcluded {
		if x <= i.Min+epsilon {
			return false
		}
	} else {
		if x < i.Min-epsilon {
			return false
		}
	}
	if i.MaxIsExcluded {
		if x >= i.Max-epsilon {
			return false
		}
	} else {
		if x > i.Max+epsilon {
			return false
		}
	}
	return true
}
