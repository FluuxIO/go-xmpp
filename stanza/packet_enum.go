package stanza

type StanzaType string

// RFC 6120: part of A.5 Client Namespace and A.6 Server Namespace
const (
	IQTypeError  StanzaType = "error"
	IQTypeGet    StanzaType = "get"
	IQTypeResult StanzaType = "result"
	IQTypeSet    StanzaType = "set"

	MessageTypeChat      StanzaType = "chat"
	MessageTypeError     StanzaType = "error"
	MessageTypeGroupchat StanzaType = "groupchat"
	MessageTypeHeadline  StanzaType = "headline"
	MessageTypeNormal    StanzaType = "normal" // Default

	PresenceTypeError        StanzaType = "error"
	PresenceTypeProbe        StanzaType = "probe"
	PresenceTypeSubscribe    StanzaType = "subscribe"
	PresenceTypeSubscribed   StanzaType = "subscribed"
	PresenceTypeUnavailable  StanzaType = "unavailable"
	PresenceTypeUnsubscribe  StanzaType = "unsubscribe"
	PresenceTypeUnsubscribed StanzaType = "unsubscribed"
)
