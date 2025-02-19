# JavaScript Unsafe DOM Manipulation Checker

Cross-Site Scripting (XSS) vulnerabilities often occur through unsafe DOM manipulation methods. This checker detects potentially dangerous DOM operations that could lead to XSS attacks.

## Step 1: Writing the test file

First, let's create a test file that covers various unsafe DOM manipulation patterns. Create `.globstar/js_unsafe_dom.test.js`:

```javascript
function test_dangerous_dom_operations() {
  const userInput = getUserInput();
  const element = document.getElementById('content');

  // These should be flagged

  // <expect-error>
  element.innerHTML = userInput;

  // <expect-error>
  element.innerHTML = "<div>" + userInput + "</div>";

  // <expect-error>
  element.insertAdjacentHTML('beforeend', `${userInput}`);

  // These are safe and should not be flagged

  // Safe - using textContent
  element.textContent = userInput;

  // Safe - using createElement
  const div = document.createElement('div');
  div.textContent = userInput;
  element.appendChild(div);

  // Safe - using static HTML
  element.innerHTML = '<div>Static content</div>';
}

function test_edge_cases() {
  const element = document.querySelector('.content');

  // Should not flag property access
  const currentHTML = element.innerHTML;

  // Should not flag non-HTML string concatenation
  const message = "Hello, " + username;

  // Should not flag commented code
  // element.innerHTML = userInput;
}

// Helper function to simulate user input
function getUserInput() {
  return "user provided content";
}
```

Our test file:
1. Includes common unsafe DOM manipulation patterns
2. Shows safe alternatives using proper DOM methods
3. Covers edge cases to avoid false positives
4. Uses `<expect-error>` to mark lines that should trigger the checker

## Step 2: Writing the checker

Now that we have our test file ready, let's create the checker in `.globstar/js_unsafe_dom.yml`:

```yaml
language: javascript
name: js_unsafe_dom
message: "Possible XSS vulnerability: Use textContent, createElement, or proper HTML sanitization instead of innerHTML."
category: security
severity: critical
pattern: >
  [
    (assignment_expression
      left: (member_expression
        property: (property_identifier) @prop)
      right: [
        (identifier)
        (binary_expression)
        (template_string)
      ]
      (#eq? @prop "innerHTML")) @js_unsafe_dom

    (call_expression
      function: (member_expression
        property: (property_identifier) @method)
      arguments: (arguments
        [
          (identifier)
          (template_string)
          (binary_expression)
        ])
      (#eq? @method "insertAdjacentHTML")) @js_unsafe_dom
  ]
filters:
  - pattern-inside: (function_declaration)
exclude:
  - "test/**"
  - "*_test.js"
  - "tests/**"
  - "__tests__/**"
description: |
  Cross-Site Scripting (XSS) vulnerabilities can occur when using unsafe DOM manipulation
  methods with unsanitized input. Methods like innerHTML and insertAdjacentHTML can
  execute arbitrary JavaScript if they contain malicious content.
```

Let's break down how this checker matches our test cases:

1. **First Pattern Block**
   ```
   (assignment_expression
     left: (member_expression
       property: (property_identifier) @prop)
     right: [...])
   ```
   This matches assignments to innerHTML, checking for various types of potentially unsafe values.

2. **Second Pattern Block**
   ```
   (call_expression
     function: (member_expression
       property: (property_identifier) @method)
     arguments: (arguments ...))
   ```
   This matches calls to insertAdjacentHTML with potentially unsafe arguments.

## Testing the checker

Run the checker against your test file:

```bash
globstar test
```

## Further Reading
- [OWASP XSS Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html)
- [CWE-79: Cross-site Scripting](https://cwe.mitre.org/data/definitions/79.html)
- [Content Security Policy (CSP)](https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP)
