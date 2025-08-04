# Contributing to Zerops MCP Server

Thank you for your interest in contributing to the Zerops MCP Server! This document provides guidelines for contributing to the project.

## Code of Conduct

Please note that this project is released with a Contributor Code of Conduct. By participating in this project you agree to abide by its terms.

## How to Contribute

### Reporting Issues

- Use the GitHub issue tracker to report bugs
- Describe the issue clearly, including steps to reproduce
- Include system information (OS, Go version, etc.)
- Provide relevant logs or error messages

### Suggesting Features

- Open an issue describing the feature
- Explain the use case and why it would be valuable
- Be open to discussion and feedback

### Pull Requests

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`make test`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Development Setup

1. Ensure Go 1.21+ is installed
2. Clone the repository
3. Install dependencies: `go mod download`
4. Build: `make build`
5. Run tests: `make test`

### Code Style

- Follow standard Go conventions
- Use `gofmt` to format your code
- Keep functions focused and small
- Add comments for exported functions
- Write meaningful commit messages

### Testing

- Write tests for new features
- Ensure existing tests pass
- Aim for good test coverage
- Use table-driven tests where appropriate

### Adding New Tools

When adding a new tool:

1. Add the tool implementation in `internal/tools/`
2. Register it in `internal/tools/register.go`
3. Add integration tests
4. Update documentation in `docs/TOOL_REFERENCE.md`
5. Update the README.md tool count

### Documentation

- Keep documentation up to date
- Use clear, concise language
- Include examples where helpful
- Update TOOL_REFERENCE.md for new tools

## Questions?

Feel free to open an issue for any questions about contributing.