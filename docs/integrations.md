# Integrations

Although Mailroom is designed to be a framework where you implement your own handlers and transports, we have provided a few built-in integrations for common cases to help get you started.

## Transports

### Slack Transport

Use `slack.NewTransport()` to create a Mailroom transport that can send notifications via Slack.  It supports rich formatting (blocks, attachments, etc.).

## User Stores

### Postgres User Store

Use `postgres.NewUserStore()` to create a Mailroom user store that persists user information in a PostgreSQL database.

### In-Memory User Store

`user.NewInMemoryStore()` provides a simple yet complete in-memory user store implementation. It's especially useful for testing and development.

(It's stable enough to use for single-replica production deployments too, if needed - just realize that known identifiers and user preferences will be lost on restart).
