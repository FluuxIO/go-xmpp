package stanza

// PresenceShow is a Enum of presence element show
type PresenceShow string

// RFC 6120: part of A.5 Client Namespace and A.6 Server Namespace
const (
	PresenceShowAway PresenceShow = "away"
	PresenceShowChat PresenceShow = "chat"
	PresenceShowDND  PresenceShow = "dnd"
	PresenceShowXA   PresenceShow = "xa"
)
