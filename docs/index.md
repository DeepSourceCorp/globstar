---
layout: home

hero:
  text: "The Open-Source Static Analysis Toolkit"
  tagline: Write SAST checkers with Globstar and run them in your CI with a single binary. It's fast, easy to write, and MIT-licensed.
  actions:
    - theme: brand
      text: What's Globstar?
      link: /introduction
    - theme: alt
      text: Quickstart
      link: /quickstart
    - theme: alt
      text: Writing a Checker
      link: /reference/checker-yaml
    - theme: alt
      text: Star on GitHub
      link: https://github.com/deepsourcecorp/globstar
features:
  - icon:
      src: /icon/fast.svg
      alt: Fast
    title: "Fast"
    details: "Globstar is designed to be fast. It's built with Go and uses native tree-sitter bindings for parsing, so you can run hundreds of checks in seconds."
  - icon:
      src: /icon/yml.svg
      alt: Easy to write
    title: "Easy to write"
    details: "Write checkers in a simple YAML file, with tree-sitter's S-expressions for matching patterns. No more custom DSL to learn."
  - icon:
      src: /icon/oss.svg
      alt: Open-source
    title: "MIT-licensed"
    details: "Use Globstar in your projects, commercial or otherwise. It's MIT-licensed, and we are committed to keeping it that way."
  - icon:
      src: /icon/blocks.svg
      alt: CI integration
    title: "Simple CI integration"
    details: "It's a single binary that you can run in your CI pipeline. No need to install dependencies or manage complex configurations."
  - icon:
      src: /icon/star.svg
      alt: Feature-rich
    title: "Feature-rich"
    details: "Use the Go API for advanced checkers, with full access to the tree-sitter AST, imports and scope resolution, cross-file analysis, and more."
  - icon:
      src: /icon/battery.svg
      alt: Batteries included
    title: "Batteries included"
    details: "Don't want to write your own checkers? Use the built-in checkers for common issues like security vulnerabilities, code quality, and more."
---
