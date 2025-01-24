# Checker YAML Interface

All Yaml lints are present inside a `.globstar` repo in the project root.

## Schema

- `language`: Key specifying which language this lint is meant for. Check the [table](TODO) for possible languages.

- `name`: A unique ID by which this rule will be identified.
- `message`: The error message shown by globstar when this lint is raised.
    - Use the `@` token to use contents from a captured AST node in the error message.
- `pattern`: A tree-sitter S-Expression query
- `filters`: Removes AST nodes after they're matched, depending on their parent nodes.

### Filters

Currently, two kinds of filters are supported: `parent-inside` and `parent-not-inside`.
A filter like this:

```yml
pattern: "(integer)"
filters:
  - parent-inside: (function_definition)
  - parent-not-inside: (if_statement)
```

will only match an AST node that:
1. Is an `integer`
2. Is inside a function definition, and
3. Is not inside an `if` statement.

```py
def foo():
  1 # <- match

def bar():
  if baz:
      1 # NO match

if cond:
  1 # match
```

### Testing Lints

After a `lint.yml` file has been created, make a corresponding `lint.test.py` file:

```py
# .globstar/lint.test.py
def foo():
  # <expect-error>
  return 1

def foo2(x):
  if x:
    return 1 # OK

if x:
  print(1) # OK
```

Use Use an `<expect-error>` (Or `<expect-error>`) comment to assert that the lint gets raised
on a specific line. Note that the comment should be placed *above* the line you expect to match the lint.