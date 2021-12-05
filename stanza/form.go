package stanza

import "encoding/xml"

type FormType string

const (
	FormTypeCancel = "cancel"
	FormTypeForm   = "form"
	FormTypeResult = "result"
	FormTypeSubmit = "submit"
)

// See XEP-0004 and XEP-0068
// Pointer semantics
type Form struct {
	XMLName      xml.Name   `xml:"jabber:x:data x"`
	Instructions []string   `xml:"instructions"`
	Title        string     `xml:"title,omitempty"`
	Fields       []*Field   `xml:"field,omitempty"`
	Reported     *FormItem  `xml:"reported"`
	Items        []FormItem `xml:"item,omitempty"`
	Type         string     `xml:"type,attr"`
}

type FormItem struct {
	XMLName xml.Name
	Fields  []Field `xml:"field,omitempty"`
}

type Field struct {
	XMLName     xml.Name `xml:"field"`
	Description string   `xml:"desc,omitempty"`
	Required    *string  `xml:"required"`
	ValuesList  []string `xml:"value"`
	Options     []Option `xml:"option,omitempty"`
	Var         string   `xml:"var,attr,omitempty"`
	Type        string   `xml:"type,attr,omitempty"`
	Label       string   `xml:"label,attr,omitempty"`
}

func NewForm(fields []*Field, formType string) *Form {
	return &Form{
		Type:   formType,
		Fields: fields,
	}
}

type FieldType string

const (
	FieldTypeBool        = "boolean"
	FieldTypeFixed       = "fixed"
	FieldTypeHidden      = "hidden"
	FieldTypeJidMulti    = "jid-multi"
	FieldTypeJidSingle   = "jid-single"
	FieldTypeListMulti   = "list-multi"
	FieldTypeListSingle  = "list-single"
	FieldTypeTextMulti   = "text-multi"
	FieldTypeTextPrivate = "text-private"
	FieldTypeTextSingle  = "text-Single"
)

type Option struct {
	XMLName    xml.Name `xml:"option"`
	Label      string   `xml:"label,attr,omitempty"`
	ValuesList []string `xml:"value"`
}
