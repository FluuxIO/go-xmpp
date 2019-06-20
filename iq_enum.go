package xmpp

// IQType is a Enum of iq attribute type
type IQType string

// RFC 6120: part of A.5 Client Namespace and A.6 Server Namespace
const (
	IQTypeError  IQType = "error"
	IQTypeGet    IQType = "get"
	IQTypeResult IQType = "result"
	IQTypeSet    IQType = "set"
)
