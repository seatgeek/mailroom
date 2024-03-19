// Copyright 2024 SeatGeek, Inc.
//
// Licensed under the terms of the Apache-2.0 license. See LICENSE file in project root for terms.

package common

// ServiceID is a type that identifies the source of an event; useful for routing certain payloads to different
// notifiers (for example, making ArgoCD notifications come from a different Slack bot than GitLab notifications)
type ServiceID string // eg. "gitlab"; "argocd"
// NotificationID is a type that identifies a specific notification type from some event within a service
type NotificationID string // eg. "assigned"; "deployed"

type EventID struct {
	ServiceID
	NotificationID
}

func (e EventID) String() string {
	return string(e.ServiceID) + "." + string(e.NotificationID)
}
