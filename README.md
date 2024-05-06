# ![mailroom](./mailroom.png)

A notification relay framework

## Usage

Coming soon!

See [`internal/example.go`](./internal/example.go) for an example of how to use mailroom.

## Terminology

- **Event**: An event is some action that occurs in an external system that we want to send a **notification** for.
- **Notification**: A notification is a string that should be sent to a **user** via some **transport**.
- **Source**: A source is any external system that sends webhook events to Mailroom.
- **Handler:** A handler is a function that processes an incoming HTTP request from a **source** and returns a list of **notifications** to send.
  - A **handler** can optionally be composed of separate **payload parser** and **notification generator** objects:
    - **Payload Parser**: The payload parser is responsible for:
      - Receiving a serialized **event** via an incoming HTTP request
      - Validating the request (e.g. verifying the signature or shared secret)
      - Parsing the raw request body into a useful **event** object if the event is of interest to us
    - **Notification Generator**: The notification generator takes the parsed **event** and generates one or more **notifications** from it
- **Transport**: A transport is a way to send a **notification** to a user. It could be email, Slack, Discord, or something else.
- **User**: A user is a person who wants to receive **notifications** from Mailroom and has **preferences** on how they'd like to receive them.
- **Identifier**: An identifier is a unique string that identifies an initiator or potential recipient (**user**) of some event. It could be an email address, a Slack user ID, or something else.
- **Preferences**: Preferences are how users specify which **notifications** they want, and which **transports** they prefer to receive them on.
- **User Store**: The user store is a database that stores user information, including their **identifiers** and **preferences**.

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for contribution guidelines.

Use `make` to run all linters and tests locally.
