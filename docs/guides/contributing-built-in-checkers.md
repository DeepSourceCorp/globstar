---
outline: 2
---

# Contributing built-in checkers

This guide explains how to contribute built-in checkers to Globstar. While the [Writing a checker in YAML](/guides/writing-yaml-checker) guide covers the basics of creating checkers, this document focuses on the specific requirements and best practices for contributing to Globstar's built-in checker collection.

## Repository Structure

Built-in checkers live in the `checkers` directory of the Globstar repository. The directory structure is organized by language:

```
checkers/
├── python/
│   ├── dangerous_eval.yml
│   ├── dangerous_eval.test.py
├── javascript/
│   ├── ...
└── ruby/
    ├── ...
```

Each language folder contains:
- Checker definition files (`.yml`)
- Corresponding test files (`.test.{extension}`)

## Contribution Guidelines

### 1. Checker naming and organization

- Each checker should have a unique combination of name and language. Please scan through existing checkers to ensure uniqueness.
- Use descriptive names that indicate the issue being detected
- Follow the naming convention: `{issue_type}_{specific_pattern}.yml`
  ```
  dangerous_eval.yml      ✓ Good
  eval_checker.yml        ✗ Too vague
  dangerous-eval.yml      ✗ Wrong format (use underscore)
  ```

### 2. One checker per pull request

- Submit each new checker as a separate pull request. This allows for focused review and discussion.
- Include both the checker file and its test file, and ensure that the test file covers all possible scenarios.
- Ensure `globstar test` passes before submitting.

### 3. Writing quality descriptions

Your checker should have clear, actionable messages and descriptions:

```yaml
# ✓ Good
message: "Dangerous use of eval() detected. Use ast.literal_eval() for parsing data structures."
description: |
  Using eval() with untrusted input can lead to remote code execution vulnerabilities.
  Attackers can inject malicious Python code that will be executed by eval().

# ✗ Too vague
message: "Don't use eval"
description: |
  eval is dangerous
```

Guidelines for messages:
- Be specific about what was found
- Suggest the correct alternative
- Keep it concise but informative

Guidelines for descriptions:
- Explain why the pattern is problematic
- Provide concrete examples if helpful
- Include any relevant security context
- Link to relevant documentation when applicable
- Markdown formatting is allowed

## Finding Patterns for Checkers

We focus primarily on security-related checkers across all [supported languages](/supported-languages). Here are valuable resources for finding patterns:

### Security Resources

- **OWASP**
  - [OWASP Top 10](https://owasp.org/www-project-top-ten/)
  - [OWASP Cheat Sheet Series](https://cheatsheetseries.owasp.org/)

- **CWE Database**
  - [CWE Top 25](https://cwe.mitre.org/top25/archive/2024/2024_cwe_top25.html)
  - [CWE/SANS Top 25](https://www.sans.org/top25-software-errors/)
  - Language-specific weaknesses in their database

- **Language Security Resources**
  - Python: [Bandit](https://bandit.readthedocs.io/en/latest/plugins/index.html) rules
  - JavaScript: [NodeJS Security Best Practices](https://nodejs.org/en/docs/guides/security/)
  - Ruby: [Ruby Security Guide](https://guides.rubyonrails.org/security.html)

### Common Security Patterns

1. **Input Validation**
   - Dangerous function calls (`eval`, `exec`, etc.)
   - Command injection vectors
   - Unsafe deserialization

2. **Authentication & Authorization**
   - Hardcoded credentials
   - Weak cryptographic practices
   - Insecure session management

3. **Data Protection**
   - Unencrypted sensitive data
   - Insecure random number generation
   - Weak hashing algorithms

4. **Code Injection**
   - SQL injection patterns
   - XSS vulnerabilities
   - Template injection

5. **System Security**
   - Unsafe file operations
   - Insecure permissions
   - Uncontrolled resource consumption

### Prioritizing Patterns

When choosing patterns to implement, consider:

1. **Impact**
   - How severe are the security implications?
   - How common is the vulnerability?
   - What's the potential for exploitation?

2. **False Positives**
   - Can you reliably detect the pattern?
   - Are there legitimate use cases that should be excluded?
   - How specific can you make the detection?

3. **Scope**
   - Is this relevant to multiple frameworks/libraries?
   - Does it apply to both small and large codebases?
   - Is it language-specific or a general pattern?

## Submission Checklist

Before submitting your pull request, ensure:

- [ ] Checker is in the correct language directory
- [ ] Filename matches the pattern naming convention
- [ ] Test file is included and covers edge cases
- [ ] `globstar test` passes locally
- [ ] Message and description are clear and actionable
- [ ] Pattern has been tested against real-world code
- [ ] No similar checker exists for the same pattern

## Getting Help
- Use [GitHub Discussions](https://github.com/DeepSourceCorp/globstar/discussions) for pattern ideas
- Tag your issues with `checker-proposal` when suggesting new patterns. While we recommend this, you can also directly submit a pull request with a proposed checker.

Our goal is to build a comprehensive collection of reliable, high-impact security checkers. Quality is more important than quantity, so take the time to make each checker as robust and useful as possible.
