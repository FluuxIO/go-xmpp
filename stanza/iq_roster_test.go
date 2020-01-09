package stanza

import (
	"encoding/xml"
	"reflect"
	"testing"
)

func TestRosterBuilder(t *testing.T) {
	iq := NewIQ(Attrs{Type: IQTypeResult, From: "romeo@montague.net/orchard"})
	var noGroup []string

	iq.RosterItems().AddItem("xl8ceawrfu8zdneomw1h6h28d@crypho.com",
		SubscriptionBoth,
		"",
		"xl8ceaw",
		[]string{"0flucpm8i2jtrjhxw01uf1nd2",
			"bm2bajg9ex4e1swiuju9i9nu5",
			"rvjpanomi4ejpx42fpmffoac0"}).
		AddItem("9aynsym60zbu78jbdvpho7s68@crypho.com",
			SubscriptionBoth,
			"",
			"9aynsym60",
			[]string{"mzaoy73i6ra5k502182zi1t97"}).
		AddItem("admin@crypho.com",
			SubscriptionBoth,
			"",
			"admin",
			noGroup)

	parsedIQ, err := checkMarshalling(t, iq)
	if err != nil {
		return
	}

	// Check result
	pp, ok := parsedIQ.Payload.(*RosterItems)
	if !ok {
		t.Errorf("Parsed stanza does not contain correct IQ payload")
	}

	// Check items
	items := []RosterItem{
		{
			XMLName:      xml.Name{},
			Name:         "xl8ceaw",
			Ask:          "",
			Jid:          "xl8ceawrfu8zdneomw1h6h28d@crypho.com",
			Subscription: SubscriptionBoth,
			Groups: []string{"0flucpm8i2jtrjhxw01uf1nd2",
				"bm2bajg9ex4e1swiuju9i9nu5",
				"rvjpanomi4ejpx42fpmffoac0"},
		},
		{
			XMLName:      xml.Name{},
			Name:         "9aynsym60",
			Ask:          "",
			Jid:          "9aynsym60zbu78jbdvpho7s68@crypho.com",
			Subscription: SubscriptionBoth,
			Groups:       []string{"mzaoy73i6ra5k502182zi1t97"},
		},
		{
			XMLName:      xml.Name{},
			Name:         "admin",
			Ask:          "",
			Jid:          "admin@crypho.com",
			Subscription: SubscriptionBoth,
			Groups:       noGroup,
		},
	}
	if len(pp.Items) != len(items) {
		t.Errorf("List length mismatch: %#v", pp.Items)
	} else {
		for i, item := range pp.Items {
			if item.Jid != items[i].Jid {
				t.Errorf("Jid Mismatch (expected: %s): %s", items[i].Jid, item.Jid)
			}
			if !reflect.DeepEqual(item.Groups, items[i].Groups) {
				t.Errorf("Node Mismatch (expected: %s): %s", items[i].Jid, item.Jid)
			}
			if item.Name != items[i].Name {
				t.Errorf("Name Mismatch (expected: %s): %s", items[i].Jid, item.Jid)
			}
			if item.Ask != items[i].Ask {
				t.Errorf("Name Mismatch (expected: %s): %s", items[i].Jid, item.Jid)
			}
			if item.Subscription != items[i].Subscription {
				t.Errorf("Name Mismatch (expected: %s): %s", items[i].Jid, item.Jid)
			}
		}
	}
}

func checkMarshalling(t *testing.T, iq IQ) (*IQ, error) {
	// Marshall
	data, err := xml.Marshal(iq)
	if err != nil {
		t.Errorf("cannot marshal iq: %s\n%#v", err, iq)
		return nil, err
	}

	// Unmarshall
	var parsedIQ IQ
	err = xml.Unmarshal(data, &parsedIQ)
	if err != nil {
		t.Errorf("Unmarshal returned error: %s\n%s", err, data)
	}
	return &parsedIQ, err
}
