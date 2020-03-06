package stanza

import (
	"encoding/xml"
)

type StanzaErrorGroup interface {
	GroupErrorName() string
}

type BadFormat struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas bad-format"`
}

func (e *BadFormat) GroupErrorName() string { return "bad-format" }

type BadNamespacePrefix struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas bad-namespace-prefix"`
}

func (e *BadNamespacePrefix) GroupErrorName() string { return "bad-namespace-prefix" }

type Conflict struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas conflict"`
}

func (e *Conflict) GroupErrorName() string { return "conflict" }

type ConnectionTimeout struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas connection-timeout"`
}

func (e *ConnectionTimeout) GroupErrorName() string { return "connection-timeout" }

type HostGone struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas host-gone"`
}

func (e *HostGone) GroupErrorName() string { return "host-gone" }

type HostUnknown struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas host-unknown"`
}

func (e *HostUnknown) GroupErrorName() string { return "host-unknown" }

type ImproperAddressing struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas improper-addressing"`
}

func (e *ImproperAddressing) GroupErrorName() string { return "improper-addressing" }

type InternalServerError struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas internal-server-error"`
}

func (e *InternalServerError) GroupErrorName() string { return "internal-server-error" }

type InvalidForm struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas invalid-from"`
}

func (e *InvalidForm) GroupErrorName() string { return "invalid-from" }

type InvalidId struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas invalid-id"`
}

func (e *InvalidId) GroupErrorName() string { return "invalid-id" }

type InvalidNamespace struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas invalid-namespace"`
}

func (e *InvalidNamespace) GroupErrorName() string { return "invalid-namespace" }

type InvalidXML struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas invalid-xml"`
}

func (e *InvalidXML) GroupErrorName() string { return "invalid-xml" }

type NotAuthorized struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas not-authorized"`
}

func (e *NotAuthorized) GroupErrorName() string { return "not-authorized" }

type NotWellFormed struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas not-well-formed"`
}

func (e *NotWellFormed) GroupErrorName() string { return "not-well-formed" }

type PolicyViolation struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas policy-violation"`
}

func (e *PolicyViolation) GroupErrorName() string { return "policy-violation" }

type RemoteConnectionFailed struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas remote-connection-failed"`
}

func (e *RemoteConnectionFailed) GroupErrorName() string { return "remote-connection-failed" }

type Reset struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas reset"`
}

func (e *Reset) GroupErrorName() string { return "reset" }

type ResourceConstraint struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas resource-constraint"`
}

func (e *ResourceConstraint) GroupErrorName() string { return "resource-constraint" }

type RestrictedXML struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas restricted-xml"`
}

func (e *RestrictedXML) GroupErrorName() string { return "restricted-xml" }

type SeeOtherHost struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas see-other-host"`
}

func (e *SeeOtherHost) GroupErrorName() string { return "see-other-host" }

type SystemShutdown struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas system-shutdown"`
}

func (e *SystemShutdown) GroupErrorName() string { return "system-shutdown" }

type UndefinedCondition struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas undefined-condition"`
}

func (e *UndefinedCondition) GroupErrorName() string { return "undefined-condition" }

type UnsupportedEncoding struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas unsupported-encoding"`
}

type UnexpectedRequest struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas unexpected-request"`
}

func (e *UnexpectedRequest) GroupErrorName() string { return "unexpected-request" }

func (e *UnsupportedEncoding) GroupErrorName() string { return "unsupported-encoding" }

type UnsupportedStanzaType struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas unsupported-stanza-type"`
}

func (e *UnsupportedStanzaType) GroupErrorName() string { return "unsupported-stanza-type" }

type UnsupportedVersion struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas unsupported-version"`
}

func (e *UnsupportedVersion) GroupErrorName() string { return "unsupported-version" }

type XMLNotWellFormed struct {
	XMLName xml.Name `xml:"urn:ietf:params:xml:ns:xmpp-stanzas xml-not-well-formed"`
}

func (e *XMLNotWellFormed) GroupErrorName() string { return "xml-not-well-formed" }
