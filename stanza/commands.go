package stanza

import "encoding/xml"

// Implements the XEP-0050 extension

const (
	CommandActionCancel   = "cancel"
	CommandActionComplete = "complete"
	CommandActionExecute  = "execute"
	CommandActionNext     = "next"
	CommandActionPrevious = "prev"

	CommandStatusCancelled = "canceled"
	CommandStatusCompleted = "completed"
	CommandStatusExecuting = "executing"

	CommandNoteTypeErr  = "error"
	CommandNoteTypeInfo = "info"
	CommandNoteTypeWarn = "warn"
)

type Command struct {
	XMLName xml.Name `xml:"http://jabber.org/protocol/commands command"`

	CommandElement CommandElement

	BadAction       *struct{} `xml:"bad-action,omitempty"`
	BadLocale       *struct{} `xml:"bad-locale,omitempty"`
	BadPayload      *struct{} `xml:"bad-payload,omitempty"`
	BadSessionId    *struct{} `xml:"bad-sessionid,omitempty"`
	MalformedAction *struct{} `xml:"malformed-action,omitempty"`
	SessionExpired  *struct{} `xml:"session-expired,omitempty"`

	// Attributes
	Action    string `xml:"action,attr,omitempty"`
	Node      string `xml:"node,attr"`
	SessionId string `xml:"sessionid,attr,omitempty"`
	Status    string `xml:"status,attr,omitempty"`
	Lang      string `xml:"lang,attr,omitempty"`

	// Result sets
	ResultSet *ResultSet `xml:"set,omitempty"`
}

func (c *Command) Namespace() string {
	return c.XMLName.Space
}

func (c *Command) GetSet() *ResultSet {
	return c.ResultSet
}

type CommandElement interface {
	Ref() string
}

type Actions struct {
	Prev     *struct{} `xml:"prev,omitempty"`
	Next     *struct{} `xml:"next,omitempty"`
	Complete *struct{} `xml:"complete,omitempty"`

	Execute string `xml:"execute,attr,omitempty"`
}

func (a *Actions) Ref() string {
	return "actions"
}

type Note struct {
	Text string `xml:",cdata"`
	Type string `xml:"type,attr,omitempty"`
}

func (n *Note) Ref() string {
	return "note"
}
func (f *Form) Ref() string { return "form" }

func (n *Node) Ref() string {
	return "node"
}

func (c *Command) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	c.XMLName = start.Name

	// Extract packet attributes
	for _, attr := range start.Attr {
		if attr.Name.Local == "action" {
			c.Action = attr.Value
		}
		if attr.Name.Local == "node" {
			c.Node = attr.Value
		}
		if attr.Name.Local == "sessionid" {
			c.SessionId = attr.Value
		}
		if attr.Name.Local == "status" {
			c.Status = attr.Value
		}
		if attr.Name.Local == "lang" {
			c.Lang = attr.Value
		}
	}

	// decode inner elements
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}

		switch tt := t.(type) {

		case xml.StartElement:
			// Decode sub-elements
			var err error
			switch tt.Name.Local {

			case "affiliations":
				a := Actions{}
				err = d.DecodeElement(&a, &tt)
				c.CommandElement = &a
			case "configure":
				nt := Note{}
				err = d.DecodeElement(&nt, &tt)
				c.CommandElement = &nt
			case "x":
				f := Form{}
				err = d.DecodeElement(&f, &tt)
				c.CommandElement = &f
			default:
				n := Node{}
				err = d.DecodeElement(&n, &tt)
				c.CommandElement = &n
				if err != nil {
					return err
				}
			}
			if err != nil {
				return err
			}
		case xml.EndElement:
			if tt == start.End() {
				return nil
			}
		}
	}
}

func init() {
	TypeRegistry.MapExtension(PKTIQ, xml.Name{Space: "http://jabber.org/protocol/commands", Local: "command"}, Command{})
}
