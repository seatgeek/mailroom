# Core Concepts

This section covers the core concepts of the Mailroom framework. Familiarity with these concepts is essential for extending and customizing Mailroom to suit your needs.

![Flow diagram](./flow.png)

## Events and Event Parsing

An **Event** is some action that occurs in an external system that we want to send a **Notification** for. Events are the primary input to Mailroom and are used to generate notifications for users.

Typically, these will come in the form of a webhook sent by an external system. The raw HTTP request is first processed by an **Event Parser** to create an `event.Event` object.

### Event Parser

An **Event Parser** (implementing `event.Parser`) is responsible for the initial handling of an incoming HTTP request from an external system (a "source"). It validates the request (e.g., verifying a signature or shared secret), extracts relevant information, and converts it into a structured `event.Event` object.

### Event Object

The `event.Event` struct contains a `Context` (holding metadata like event type, source, and **Identifiers** of the initiator) and `Data` field. The `Data` field is of type `any` and holds the specific payload of the event (e.g., details of a GitHub pull request) which can be referenced by **Event Processors** to generate notifications.

## Event Processors

The **Event Processor** is a component that takes an `event.Event` and generates a list of **Notifications**. It can generate initial notifications, enrich existing ones, filter them, or perform other logic. Processors are designed to be chainable, allowing for a flexible and modular approach to notification handling.

Each `Processor` must implement the [`event.Processor` interface](https://pkg.go.dev/github.com/seatgeek/mailroom/pkg/processing#Processor):

```go
type Processor interface {
    Process(ctx context.Context, evt event.Event, notifications []event.Notification) ([]event.Notification, error)
}
```

The `Process` method takes the current `event.Event` and the list of `Notification`s accumulated so far. It returns an updated list of `Notification`s (or the same list if no changes were made) and an optional error.

## Notifications

A **Notification** (`event.Notification`) is an object representing a message that should be sent to a **User** via some **Transport**. It consists of metadata about the origins of the notification, the intended recipient (as an `identifier.Set`), and a method to render the message content for a specific transport.

Notifications are initially generated and then potentially modified or filtered by **Processors** in a processor chain. The final list of notifications from the chain is then passed to the **Notifier** for delivery.

Each instance of a notification object is targeted at a single user. If a message needs to be sent to multiple users, multiple `Notification` objects should be generated.

## Notifier

The **Notifier** is responsible for taking the generated **Notifications** and dispatching them to the appropriate **Transports** for delivery based on the **User**'s **Preferences**.

## Transports

A **Transport** is a way to send a **Notification** to a **User**. It could be email, Slack, Discord, or something else.

Each **Transport** must implement the `notifier.Transport` interface and can choose how to deliver the notification to the user.

## Users

A **User** is a person who wants to receive **Notifications** from Mailroom. They may have **Preferences** on how they'd like to receive them.

Each user has a set of **Identifiers** that uniquely identify them across different systems.

## Identifiers

An **Identifier** is a unique string that identifies an initiator or potential recipient (**User**) of some event. It could be an email address, a Slack user ID, or something else.

Each identifier is composed of three parts:

- **Namespace** (optional): The namespace of the identifier (e.g. `slack.com`, `github.com`)
- **Kind**: The kind of identifier (e.g. `email`, `username`, `id`)
- **Value**: The actual value of the identifier (e.g. `rufus@seatgeek.com`, `rufus`, `U123456`)

For example, the `slack.com/email:rufus@seatgeek.com` identifier means that Slack knows this user by the email address `rufus@seatgeek.com`.

Identifiers are an important concept in Mailroom - they allow us to match users across different systems and ensure that notifications are sent to the correct person.  This is because a user may be known by different identifiers in different systems; for example:

- Slack knows users by email address and Slack ID (e.g. `U123456`)
- GitHub knows users by username and email address
- etc.

Imagine we receive an event from GitHub that should notify a user in Slack. We may need to cross-reference their email addresses across both systems to determine their Slack ID given their GitHub username. Mailroom can automatically facilitate this matching process for you.

### Identifier Set

An **Identifier Set** is a collection of all known **Identifiers** that are associated with a single **User**.

The identifier set can only have one identifier of each namespace+kind.  This is why the optional namespace is included in the identifier - it allows us to differentiate between two identifiers of the same kind (like two different email addresses) that are associated with the same user.

## Preferences

Mailroom supports the ability for each **User** to specify which **Notifications** they want to receive, and which **Transports** they prefer to receive them on. This is done via **Preferences**.

These can be set via the Mailroom API and are stored in the **User Store**.

## User Store

The **User Store** is a database that stores user information, including their **Identifiers** and **Preferences**. It is used by Mailroom to look up user information when processing incoming events and generating notifications.
