# Python

## Self-Comparison Detection

Detects redundant comparisons where a variable is compared to itself.

### What it does

Identifies instances where a variable is being compared to itself, which is typically a logical error or code smell.

### Why it matters

Self-comparisons are usually bugs and can indicate copying errors or incomplete refactoring. They can lead to logic errors since they always evaluate to true.

### Examples

**Dangerous:**
```python
# Don't do this
x = 5
if x == x:
    print("This is redundant")

# Don't do this either
value = 42
result = value is value
```

**Safe:**
```python
# Do this instead
x = 5
if x == expected_value:
    print("This is meaningful")

# For object identity checks
if id(value) == id(other_value):
    print("Same object")
```

### Prevention
1. Always compare variables against other values
2. Use proper comparison operators
3. Review comparisons during code review

### Rule Configuration
```yml
language: py
name: bad_eq_eq
message: "Comparing '@a' to itself is useless"
pattern: |
  (
    (comparison_operator (identifier)@a (identifier) @b )
    (#eq? @a @b)
  )  @bad_eq_eq
filters:
  - pattern-not-inside: |
      (
        (function_definition name:(identifier) @fn_name)
        (#match? @fn_name "__(init|eq)__")
      )
  - pattern-not-inside: |
      ((call function:(_)@fn_name)
        (#any-of? @fn_name  "assertTrue" "assertFalse" "assert"))

exclude:
  - venv/*
  - build/*
  - bin/*
```

### Special Cases
The rule includes exceptions for:
1. Special Methods: `__init__` and `__eq__`
2. Test Assertions: `assertTrue`, `assertFalse`, and `assert`
3. Excluded Directories: `venv/`, `build/`, `bin/`

### Further Reading
- [Python Identity vs Equality](https://realpython.com/python-is-identity-vs-equality/)
