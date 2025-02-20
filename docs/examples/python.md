# Python Command Injection Checker

Command injection vulnerabilities occur when shell commands are constructed using unvalidated user input. This checker detects potentially dangerous command execution patterns that could lead to command injection attacks.

## Step 1: Writing the test file

First, let's create a test file that covers various command injection patterns. Create `.globstar/command_injection.test.py`:

```python
import os
import subprocess
from pathlib import Path

def test_dangerous_os_commands():
    user_input = get_user_input()
    filename = "data.txt"

    # <expect-error>
    os.system("git clone " + user_input)

    # <expect-error>
    os.popen(f"cat {filename}")

    # <expect-error>
    subprocess.call("ping " + ip_addr, shell=True)

    # These are safe and should not be flagged

    # Safe - using lists for command arguments
    subprocess.run(["git", "clone", user_input])

    # Safe - using proper file operations
    Path(filename).read_text()

    # Safe - string literal without variables
    os.system("ls -l")

def test_edge_cases():
    # Should not flag non-command string concatenation
    greeting = "Hello, " + name

    # Should not flag multiline commands without variables
    subprocess.run("""
        ls -l
        pwd
    """)

    # Should not flag when shell=False
    subprocess.call("echo test", shell=False)

# Helper function to reduce noise in examples
def get_user_input():
    return input("Enter repository URL: ")
```

Our test file:
1. Includes clear examples of dangerous patterns to catch
2. Shows safe alternatives that shouldn't be flagged
3. Covers edge cases to avoid false positives
4. Uses `<expect-error>` to mark lines that should trigger the checker

## Step 2: Writing the checker

Now that we have our test file ready, let's create the checker in `.globstar/command_injection.yml`:

```yaml
language: python
name: command_injection
message: "Possible command injection vulnerability detected. Use subprocess.run() with a list of arguments instead."
category: security
severity: critical
pattern: >
  [
    (call
      function: (attribute
        object: (identifier) @obj
        attribute: (identifier) @func)
      arguments: (argument_list
        [(binary_operator
            left: (string)
            operator: "+")
          (string
            (interpolation))])
      (#eq? @obj "os")
      (#any-of? @func "system" "popen")) @command_injection

    (call
      function: (attribute
        object: (identifier) @sub
        attribute: (identifier) @call_func)
      arguments: (argument_list
        (binary_operator)
        .
        (keyword_argument
          name: (identifier) @shell_arg
          value: (true))
      (#eq? @sub "subprocess")
      (#eq? @call_func "call")
      (#eq? @shell_arg "shell"))) @command_injection
  ]
filters:
  - pattern-inside: (function_definition)
exclude:
  - "test/**"
  - "*_test.py"
  - "tests/**"
description: |
  Command injection vulnerabilities occur when shell commands are constructed using
  unvalidated user input. This can allow attackers to execute arbitrary commands
  on the system.

  Instead of using string concatenation or shell=True, use subprocess.run() with
  a list of arguments, which properly escapes command arguments.
```

Let's break down how this checker matches our test cases:

1. **First Pattern Block**
   ```
   (call
     function: (attribute
       object: (identifier) @obj
       attribute: (identifier) @func
     )
     arguments: (argument_list [...])
   ```
   This matches function calls like `os.system()` and `os.popen()`, capturing the object (`os`) and function name (`system`/`popen`).

2. **Second Pattern Block**
   ```
   (call
     function: (attribute
       object: (identifier) @sub
       attribute: (identifier) @call_func
     )
     arguments: (argument_list ... (keyword ...))
   ```
   This matches `subprocess.call()` with `shell=True`, checking for string concatenation or interpolation in the command argument.

3. **Filters**
   ```yaml
   filters:
     - pattern-inside: (function_definition)
   ```
   Ensures we only match patterns inside function definitions, reducing false positives.

## Testing the checker

Run the checker against your test file:

```bash
globstar test
```

## Further Reading
- [OWASP Command Injection](https://owasp.org/www-community/attacks/Command_Injection)
- [Python subprocess documentation](https://docs.python.org/3/library/subprocess.html)
- [CWE-78: OS Command Injection](https://cwe.mitre.org/data/definitions/78.html)
