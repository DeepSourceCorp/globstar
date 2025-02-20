# Contributing

Globstar is an MIT-licensed open-source project, and we welcome contributions from the community. Whether you're a developer, a security researcher, or a language enthusiast, there are many ways you can contribute to the project.

## Reporting issues

If you find a bug or have a feature request, please open an issue on [deepsourcecorp/globstar](https://github.com/DeepSourceCorp/globstar/issues) with as much detail as possible. If you're reporting a bug, please include the steps to reproduce it, and the version of Globstar you're using.

## What to contribute

Here are some ideas for contributions:

- **Writing checkers**: Globstar ships with a set of builtin checkers for common issues like security vulnerabilities, code quality, and more. We would like to expand this list with more checkers for different languages and frameworks. The builtin checkers are in the [.globstar](https://github.com/DeepSourceCorp/globstar/tree/master/checkers) directory of the upstream repository.

- **Improving documentation**: We're always looking to improve our documentation. If you find something that's unclear, or if you think we can add more information, please open a pull request.

- **Improving the CLI**: If you have ideas for improving the CLI, or if you find a bug, please open an issue or a pull request. Specifically, we're loking for ways to make the CLI more user-friendly, more performant, and add commonly requested features to make it usefult for running locally and in CI/CD pipelines.

- **Expanding the Go API**: The Go API is the most powerful way to write checkers for Globstar. We currently have advanced features like imports and scope resolution and cross-file analysis only for Python and JavaScript. We would like to expand this to more languages.

- **Adding examples**: Educating users on how to write checkers is important. If you've got a knack for writing tutorials, please consider adding examples to the documentation.
