# matchtree

[![Go Reference](https://pkg.go.dev/badge/github.com/roy2220/matchtree.svg)](https://pkg.go.dev/github.com/roy2220/matchtree)
[![Coverage](./.badges/coverage.svg)](#)

`matchtree` is a generic Go package that provides a **tree structure for efficient, multi-dimensional pattern matching**. It allows users to define rules based on a sequence of various data types (strings, integers, intervals, and regular expressions) and retrieve associated values using a sequence of keys.

---

## Features

* **Multi-Dimensional Matching:** Define matching rules across a sequence of different data types (dimensions).
* **Diverse Pattern Types:** Supports matching against:
    * String (exact match for `string`)
    * Integer (exact match for `int64`)
    * IntegerInterval (range match for `int64`)
    * NumberInterval (range match for `float64`)
    * Regexp (regular expression match for `string`)
* **Wildcard and Inverse Matching:** Supports **"match any"** and **"match none of these"** patterns.
* **Priority-Based Results:** Rules can be assigned a **priority**, and search results are sorted by priority (descending) and then insertion order.

---

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/roy2220/matchtree"
)

type Role struct {
    Name string
}

func main() {
    // Define tree structure: Gender -> Age -> Height
    tree := matchtree.NewMatchTree[Role]([]matchtree.MatchType{
        matchtree.MatchString,
        matchtree.MatchIntegerInterval,
        matchtree.MatchNumberInterval,
    })

    // Add rule #1:
    //   WHEN Gender=any Age=[18,30] Height=[165.0,185.0]
    //   THEN Role=Athlete
    tree.AddRule(matchtree.MatchRule[Role]{
        Patterns: []matchtree.MatchPattern{
            {Type: matchtree.MatchString, IsAny: true},
            {Type: matchtree.MatchIntegerInterval, IntegerIntervals: []matchtree.IntegerInterval{
                {Min: matchtree.Int64Ptr(18), Max: matchtree.Int64Ptr(30)},
            }},
            {Type: matchtree.MatchNumberInterval, NumberIntervals: []matchtree.NumberInterval{
                {Min: matchtree.Float64Ptr(165.0), Max: matchtree.Float64Ptr(185.0)},
            }},
        },
        Value:    Role{Name: "Athlete"},
        Priority: 10,
    })

    // Add rule #2:
    //   WHEN Gender=female Age=[25,40] Height=any
    //   THEN Role=Manager
    tree.AddRule(matchtree.MatchRule[Role]{
        Patterns: []matchtree.MatchPattern{
            {Type: matchtree.MatchString, Strings: []string{"female"}},
            {Type: matchtree.MatchIntegerInterval, IntegerIntervals: []matchtree.IntegerInterval{
                {Min: matchtree.Int64Ptr(25), Max: matchtree.Int64Ptr(40)},
            }},
            {Type: matchtree.MatchNumberInterval, IsAny: true},
        },
        Value:    Role{Name: "Manager"},
        Priority: 5,
    })

    // Search with keys: Gender=female Age=35 Height=170.0
    results, _ := tree.Search([]matchtree.MatchKey{
        {Type: matchtree.MatchString, String: "female"},
        {Type: matchtree.MatchIntegerInterval, Integer: 35},
        {Type: matchtree.MatchNumberInterval, Number: 170.0},
    })
    for _, role := range results {
        fmt.Println(role.Name) // Output: Manager
    }
}
````

-----

## Pattern Examples

### Exact Match

```go
// Match 'male' or 'female' strings
{Type: matchtree.MatchString, Strings: []string{"male", "female"}}

// Match integers 18, 21, or 25
{Type: matchtree.MatchInteger, Integers: []int64{18, 21, 25}}
```

### Interval Match

```go
// Closed integer interval [18, 65]
{Type: matchtree.MatchIntegerInterval, IntegerIntervals: []matchtree.IntegerInterval{
    {Min: matchtree.Int64Ptr(18), Max: matchtree.Int64Ptr(65)},
}}

// Open float interval (160.0, 180.0)
{Type: matchtree.MatchNumberInterval, NumberIntervals: []matchtree.NumberInterval{
    {Min: matchtree.Float64Ptr(160.0), MinIsExcluded: true,
     Max: matchtree.Float64Ptr(180.0), MaxIsExcluded: true},
}}

// Half-open interval [18, infinity)
{Type: matchtree.MatchIntegerInterval, IntegerIntervals: []matchtree.IntegerInterval{
    {Min: matchtree.Int64Ptr(18)},
}}
```

### Wildcard

```go
// Match any string
{Type: matchtree.MatchString, IsAny: true}
```

### Inverse Match

```go
// Match any string except "admin" and "root"
{Type: matchtree.MatchString, IsInverse: true, Strings: []string{"admin", "root"}}

// Match any integer except 0 and 100
{Type: matchtree.MatchInteger, IsInverse: true, Integers: []int64{0, 100}}

// Match any integer except in the intervals [0, 10] and [20, 30]
{Type: matchtree.MatchIntegerInterval, IsInverse: true, IntegerIntervals: []matchtree.IntegerInterval{
    {Min: matchtree.Int64Ptr(0), Max: matchtree.Int64Ptr(10)},
    {Min: matchtree.Int64Ptr(20), Max: matchtree.Int64Ptr(30)},
}}
```

### Regular Expression

```go
// Match strings starting with "user_" followed by one or more digits
{Type: matchtree.MatchRegexp, Regexp: "^user_[0-9]+$"}
```

-----

## Priority and Result Ordering

Results are sorted by:

1.  **Priority** (descending) - higher priority rules appear first.
2.  **Insertion order** (ascending) - earlier rules appear first when priorities are equal.

-----

## Options

### TreatEmptyPatternAsAny

```go
tree.AddRule(rule, matchtree.TreatEmptyPatternAsAny())
```

This option treats patterns that are `IsEmpty()` (i.e., `matchtree.MatchPattern{}`) as a **wildcard**. This allows for partial rule definitions where an omitted pattern means "match anything for this dimension."

-----

## License

MIT
