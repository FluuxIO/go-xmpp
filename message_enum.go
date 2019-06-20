package xmpp

// MessageType is a Enum of message attribute type
type MessageType string

// RFC 6120: part of A.5 Client Namespace and A.6 Server Namespace
const (
	MessageTypeChat      MessageType = "chat"
	MessageTypeError     MessageType = "error"
	MessageTypeGroupchat MessageType = "groupchat"
	MessageTypeHeadline  MessageType = "headline"
	MessageTypeNormal    MessageType = "normal"
)
