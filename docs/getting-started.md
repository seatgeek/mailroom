# Getting Started

Mailroom is a framework that simplifies the creation, routing, and delivery of user notifications based on events from external systems. This guide will help you get up and running quickly.

## Quick Start

Create a new Go project that uses Mailroom as a dependency:

```go
mkdir my-project
cd my-project
go mod init my-project
go get github.com/seatgeek/mailroom
```

Copy [the code example](../internal/example.go) into your project as `main.go`, modifying it to suit your needs, and then run it:

```bash
go run main.go
```

## Architecture Overview

Mailroom provides an HTTP server that accepts incoming webhooks from external systems. When an event is received, Mailroom generates notifications and sends them to users based on their preferences:

![Flow diagram](./flow.png)

See the [Core Concepts](./core-concepts.md) documentation for more details.
