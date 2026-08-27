# Contributing to Coherence

Thank you for your interest in contributing! This document outlines how to get started and what to expect.

## Getting Started

1. **Fork** the repository on GitHub
2. **Clone** your fork locally: `git clone https://github.com/YOUR_USERNAME/coherence.git`
3. **Set up** the development environment: `make setup`
4. **Create a branch**: `git checkout -b feat/your-feature-name`

## Development Workflow

```bash
# Run everything locally
make start

# Run tests before committing
make test

# Format code
make fmt

# Lint
make lint
```

## Branch Naming

- `feat/` – new features
- `fix/` – bug fixes
- `docs/` – documentation changes
- `refactor/` – code refactoring
- `test/` – test additions

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(drift): add GCP support for Cloud SQL resources
fix(api): correct pagination offset in drift list endpoint
docs(readme): update quick start guide
```

## Pull Request Process

1. Ensure all tests pass: `make test`
2. Add tests for new functionality
3. Update documentation if needed
4. Fill in the PR template
5. Request review from a maintainer

## Code Style

- **Go**: `gofmt`, follow [Effective Go](https://go.dev/doc/effective_go)
- **TypeScript**: Prettier + ESLint config in repo
- **SQL**: UPPER CASE keywords, snake_case identifiers

## Reporting Bugs

Open a GitHub Issue with:
- Steps to reproduce
- Expected vs actual behaviour
- Environment details (OS, Go/Node version, cloud provider)

## Feature Requests

Open a GitHub Discussion or Issue with the `enhancement` label. Describe the use case and desired behaviour.

## License

By contributing, you agree your contributions will be licensed under the MIT License.
