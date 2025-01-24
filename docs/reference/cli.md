# Globstar CLI Reference

Refer to this document for information on how to use the Globstar CLI, including commands, flags, and configuration. You can also run `globstar --help` if you need a quick reference.

## Commands

### `check`

Run Globstar analysis in the current directory. By default, it runs all local (in your `.globstar` folder) and builtin checkers.

```bash
globstar check [flags]
```

#### Flags

- `--ignore, -i <pattern>`: Ignore file paths that match the specified pattern.
- `--checkers, -c <mode>`: Specify which checkers to run. Available modes:
  - `local`: Run only checkers from the `.globstar` directory
  - `builtin`: Run only built-in checkers
  - `all`: Run both local and built-in checkers (default)

### `test`

Test all checkers in the `.globstar` directory. This is useful for testing checker behaviour before running them on your codebase.

```bash
globstar test
```

### `help`

Display help information.

```bash
globstar help
```

## Configuration

Globstar looks for a `.globstar.yml` configuration file in the project root, which can be used to configure defaults. Read more in the [Configuration](#configuration) section.