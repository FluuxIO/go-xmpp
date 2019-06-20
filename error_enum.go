package xmpp

// ErrorType is a Enum of error attribute type
type ErrorType string

// RFC 6120: part of A.5 Client Namespace and A.6 Server Namespace
const (
	ErrorTypeAuth     ErrorType = "auth"
	ErrorTypeCancel   ErrorType = "cancel"
	ErrorTypeContinue ErrorType = "continue"
	ErrorTypeModify   ErrorType = "motify"
	ErrorTypeWait     ErrorType = "wait"
)
