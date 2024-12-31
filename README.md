# ![mailroom](./docs/mailroom.png)

[![GitHub Release](https://img.shields.io/github/v/release/seatgeek/mailroom?style=flat-square)](https://github.com/seatgeek/mailroom/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/seatgeek/mailroom.svg?style=flat-square)](https://pkg.go.dev/github.com/seatgeek/mailroom)
[![go.mod](https://img.shields.io/github/go-mod/go-version/seatgeek/mailroom?style=flat-square)](go.mod)
[![LICENSE](https://img.shields.io/github/license/seatgeek/mailroom?style=flat-square)](LICENSE)
[![Build Status](https://img.shields.io/github/actions/workflow/status/seatgeek/mailroom/tests.yml?branch=main&style=flat-square)](https://github.com/seatgeek/mailroom/actions?query=workflow%3Atests+branch%3Amain)
[![Go Report Card](https://goreportcard.com/badge/github.com/seatgeek/mailroom?style=flat-square)](https://goreportcard.com/report/github.com/seatgeek/mailroom)
[![Codecov](https://img.shields.io/codecov/c/github/seatgeek/mailroom?style=flat-square)](https://codecov.io/gh/seatgeek/mailroom)

Mailroom is a framework that simplifies the creation, routing, and delivery of user notifications based on events from external systems.

![Flow diagram](./docs/flow.png)

Mailroom is designed to be flexible and extensible, allowing you to easily add new handlers and transports as your needs grow and evolve. Simply write a function to transform incoming events into notifications, and Mailroom will take care of the rest, including:

- Acting as the primary notification relay for incoming webhooks from external systems
- Sending notifications to the appropriate users based on their preferences (e.g. PR reviews go to email, but build failures go to Slack)
- Formatting notifications for different transports (e.g. email, Slack, etc.)
- Matching usernames, emails, IDs, etc. across different systems
- Logging, error handling, retries, and more

See [`internal/example.go`](./internal/example.go) for an example of how to use mailroom.

## Documentation

- [Getting Started](./docs/getting-started.md)
- [Core Concepts](./docs/core-concepts.md)
- [Integrations](./docs/integrations.md)

Also see the [GoDoc](https://pkg.go.dev/github.com/seatgeek/mailroom) for documentation.

## Stability

Mailroom is currently in alpha. It works well in production, but the API may change at any time. We will do our best to note breaking changes in the release notes.

## Contributing

See [CONTRIBUTING.md](./.github/CONTRIBUTING.md) for contribution guidelines.

Use `make` to run all linters and tests locally.
