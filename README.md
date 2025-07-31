# matchtree

A Go package for efficient pattern matching using a generic tree structure. It supports string, integer, and interval-based patterns with priority-based rule matching.

## Quick Start
1. **Create a MatchTree**:
   Define the sequence of `MatchType` (e.g., `MatchString`, `MatchInteger`, `MatchIntegerInterval`, `MatchNumberInterval`).
   ```go
   import "github.com/roy2220/matchtree"

   tree := matchtree.NewMatchTree[string]([]matchtree.MatchType{matchtree.MatchString, matchtree.MatchInteger})
   ```

2. **Add Rules**:
   Define rules with patterns, a value, and a priority.
   ```go
   rule := matchtree.MatchRule[string]{
       Patterns: []matchtree.MatchPattern{
           {Type: matchtree.MatchString, Strings: []string{"apple", "banana"}, IsAny: false},
           {Type: matchtree.MatchInteger, Integers: []int64{1, 2}, IsAny: false},
       },
       Value:    "fruit",
       Priority: 1,
   }
   err := tree.AddRule(rule)
   if err != nil {
       panic(err)
   }
   ```

3. **Search for Matches**:
   Provide keys to search and retrieve matching values, sorted by priority and insertion order.
   ```go
   keys := []matchtree.MatchKey{
       {Type: matchtree.MatchString, String: "apple"},
       {Type: matchtree.MatchInteger, Integer: 1},
   }
   results, err := tree.Search(keys)
   if err != nil {
       panic(err)
   }
   fmt.Println(results) // Output: [fruit]
   ```

## Features
- **Generic Type Support**: Use any type `T` for values.
- **Flexible Patterns**: Supports exact matches (`Strings`, `Integers`), intervals (`IntegerInterval`, `NumberInterval`), `IsAny`, and `IsInverse` patterns.
- **Priority-Based Matching**: Rules with higher priority are returned first.
- **JSON Serialization**: `MatchType`, `MatchPattern`, and intervals support JSON marshaling/unmarshaling.
- **Efficient Search**: Tree-based structure for fast pattern matching.

## Key Structures
- **MatchTree[T]**: The main tree structure holding patterns and values.
- **MatchType**: Defines pattern types (`MatchString`, `MatchInteger`, `MatchIntegerInterval`, `MatchNumberInterval`).
- **MatchRule[T]**: Combines patterns, a value, and a priority.
- **MatchPattern**: Specifies a pattern with type, values, and flags (`IsAny`, `IsInverse`).
- **MatchKey**: Defines a search key with type and value.
- **IntegerInterval/NumberInterval**: Supports closed, open, or half-open intervals for integers and floats.

## Example: Interval Matching
```go
tree := matchtree.NewMatchTree[string]([]matchtree.MatchType{matchtree.MatchNumberInterval})
rule := matchtree.MatchRule[string]{
    Patterns: []matchtree.MatchPattern{
        {
            Type: matchtree.MatchNumberInterval,
            NumberIntervals: []matchtree.NumberInterval{
                {Min: float64Ptr(0.0), Max: float64Ptr(10.0), MinIsExcluded: false, MaxIsExcluded: false},
            },
        },
    },
    Value:    "range 0-10",
    Priority: 1,
}
tree.AddRule(rule)
keys := []matchtree.MatchKey{{Type: matchtree.MatchNumberInterval, Number: 5.0}}
results, _ := tree.Search(keys)
fmt.Println(results) // Output: [range 0-10]
```

## Notes
- Ensure the number and types of patterns/keys match the tree's defined types to avoid errors.
- Use `IsAny` for wildcard patterns and `IsInverse` for excluding specific values.
- Floating-point intervals use an epsilon (`1e-10`) for precision.

## License
MIT
