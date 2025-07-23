# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
Mailroom is a Go framework that simplifies the creation, routing, and delivery of user notifications based on events from external systems. It acts as a notification relay for incoming webhooks, handling user preferences, multiple transport formats (email, Slack, etc.), and cross-system user matching.

**Current Status:** Alpha (works in production but API may change)

## Key Technologies
- **Language:** Go 1.24
- **Main Dependencies:** Gorilla Mux, Slack SDK, GORM with PostgreSQL

## Architecture

### Core Components
1. **Server** (`server.go`) - Main application entry point that handles HTTP routing and coordination
2. **Event System** (`pkg/event/`) - Event parsing, processing, and notification generation
3. **Notifier System** (`pkg/notifier/`) - Transport abstraction and notification delivery
4. **User System** (`pkg/user/`) - User management, preferences, and identifier matching
5. **Transports** - Pluggable notification channels (Slack, console writer, etc.)

### Notification Pipeline
Webhooks → **Event Parsers** → **Event Processors** (chainable) → **Notifier** → **Transports** → Users

Key concepts:
- **Events** contain Context + Data from external systems
- **Processors** are chainable and can generate/modify/filter notifications
- **Identifiers** use namespace/kind/value format for cross-system user matching (e.g., `slack.com/email:user@company.com`)
- **Notifications** target single users (create multiple for multiple recipients)
- **User Store** manages identifiers and delivery preferences

## Development Workflow

### Build System
Uses Makefile with key targets:
- `make check` - Run all linters and tests (main target)
- `make lint-fix` - Auto-fix linting issues where possible
- `make generate` - Generate mocks and other code

### Testing
- Prefer table-based tests with testify assertions
- Use `t.Parallel()` for concurrent test execution
- Auto-generated Mockery mocks (via `make generate`)
- Integration tests using testcontainers (PostgreSQL)

## Directory Structure

(Keep this structure updated as the project evolves)

```
/
├── server.go                 # Main server implementation
├── internal/
│   └── example.go           # Usage example
├── pkg/                     # Main library packages
│   ├── event/              # Event system (parsers, processors, types)
│   ├── notifier/           # Notification delivery system
│   │   ├── slack/          # Slack transport implementation
│   │   └── preference/     # User preference management
│   ├── user/               # User management and storage
│   │   └── postgres/       # PostgreSQL user store
│   ├── identifier/         # Cross-system user identification
│   ├── notification/       # Notification building utilities
│   ├── server/            # HTTP server utilities
│   └── validation/        # Validation conventions
├── build/
│   ├── config/            # Configuration files for tools
│   └── reports/           # Generated test/coverage reports
├── docs/                  # Project documentation
└── test/                  # Test utilities and fixtures
```

## Key Patterns
- **Processor Chain** - Sequential processing where each processor receives accumulated notifications
- **Identifier Matching** - Cross-system user resolution using namespace/kind/value identifier sets
- **Options Pattern** - Server configuration via functional options

## Getting Started
Refer to the example in `internal/example.go` for a complete working implementation showing:
- Event parser implementation
- Notification generation
- User store setup
- Transport configuration
- Server startup

## Development Notes
- Use `make` for all build/test operations
- Generated files (mocks) are excluded from git but recreated during build
- Follow CloudEvents specification for event context structure
- Transport keys are used for user preference routing
- All major components implement validation interfaces where appropriate

### Core Interfaces
- `event.Parser` - HTTP request → Event conversion
- `event.Processor` - Event → Notification generation/modification (chainable)
- `notifier.Transport` - Notification delivery mechanism
- `user.Store` - User data persistence layer

### Important Implementation Details
- Each notification targets a single user (create multiple for multiple recipients)
- Processors receive accumulated notifications from previous processors in chain
- Identifier sets can only contain one identifier per namespace+kind combination
- User preferences determine which transports receive which notification types
