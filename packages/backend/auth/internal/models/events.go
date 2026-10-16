package models

const (
	// EventTypeUserCreated is published by the Auth Service when a user
	// registers. The Account Service consumes it to provision the account.
	EventTypeUserCreated = "auth.user.created"

	EventVersionV1 = 1
)

// UserCreatedEvent announces a newly registered user. Each service defines its
// own copy of the shared event contract; the Auth Service produces it and the
// Account Service consumes it.
type UserCreatedEvent struct {
	EventID      string `json:"eventId"`
	EventType    string `json:"eventType"` // auth.user.created
	EventVersion int    `json:"eventVersion"`
	UserID       string `json:"userId"`
	AccountID    string `json:"accountId"`
	OccurredAt   string `json:"occurredAt"`
}
