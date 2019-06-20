package xmpp

// PresenceType is a Enum of presence attribute type
type PresenceType string

// RFC 6120: part of A.5 Client Namespace and A.6 Server Namespace
const (
	PresenceTypeError        PresenceType = "error"
	PresenceTypeProbe        PresenceType = "probe"
	PresenceTypeSubscribe    PresenceType = "subscribe"
	PresenceTypeSubscribed   PresenceType = "subscribed"
	PresenceTypeUnavailable  PresenceType = "unavailable"
	PresenceTypeUnsubscribe  PresenceType = "unsubscribe"
	PresenceTypeUnsubscribed PresenceType = "unsubscribed"
)

// PresenceShow is a Enum of presence element show
type PresenceShow string

// RFC 6120: part of A.5 Client Namespace and A.6 Server Namespace
const (
	PresenceShowAway PresenceShow = "away"
	PresenceShowChat PresenceShow = "chat"
	PresenceShowDND  PresenceShow = "dnd"
	PresenceShowXA   PresenceShow = "xa"
)
