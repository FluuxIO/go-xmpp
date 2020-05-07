package stanza

import (
	"errors"
	"strings"
	"time"
)

// Helper structures and functions to manage dates and timestamps as defined in
// XEP-0082: XMPP Date and Time Profiles (https://xmpp.org/extensions/xep-0082.html)

const dateLayoutXEP0082 = "2006-01-02"
const timeLayoutXEP0082 = "15:04:05+00:00"

var InvalidDateInput = errors.New("could not parse date. Input might not be in a supported format")
var InvalidDateOutput = errors.New("could not format date as desired")

type JabberDate struct {
	value time.Time
}

func (d JabberDate) DateToString() string {
	return d.value.Format(dateLayoutXEP0082)
}

func (d JabberDate) DateTimeToString(nanos bool) string {
	if nanos {
		return d.value.Format(time.RFC3339Nano)
	}
	return d.value.Format(time.RFC3339)
}

func (d JabberDate) TimeToString(nanos bool) (string, error) {
	if nanos {
		spl := strings.Split(d.value.Format(time.RFC3339Nano), "T")
		if len(spl) != 2 {
			return "", InvalidDateOutput
		}
		return spl[1], nil
	}
	spl := strings.Split(d.value.Format(time.RFC3339), "T")
	if len(spl) != 2 {
		return "", InvalidDateOutput
	}
	return spl[1], nil
}

func NewJabberDateFromString(strDate string) (JabberDate, error) {
	t, err := time.Parse(time.RFC3339, strDate)
	if err == nil {
		return JabberDate{value: t}, nil
	}

	t, err = time.Parse(time.RFC3339Nano, strDate)
	if err == nil {
		return JabberDate{value: t}, nil
	}

	t, err = time.Parse(dateLayoutXEP0082, strDate)
	if err == nil {
		return JabberDate{value: t}, nil
	}

	t, err = time.Parse(timeLayoutXEP0082, strDate)
	if err == nil {
		return JabberDate{value: t}, nil
	}

	return JabberDate{}, InvalidDateInput
}
