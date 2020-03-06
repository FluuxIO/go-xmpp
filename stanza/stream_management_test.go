package stanza_test

import (
	"gosrc.io/xmpp/stanza"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

// TODO : tests to add
// - Pop on nil or empty slice
// - PeekN (normal and too long)

func TestPushUnack(t *testing.T) {
	uaq := initUnAckQueue()
	toPush := stanza.UnAckedStz{
		Id: 3,
		Stz: `<iq type='submit'
    from='confucius@scholars.lit/home'
    to='registrar.scholars.lit'
    id='kj3b157n'
    xml:lang='en'>
  <query xmlns='jabber:iq:register'>
    <username>confucius</username>
    <first>Qui</first>
    <last>Kong</last>
  </query>
</iq>`,
	}

	err := uaq.Push(&toPush)
	if err != nil {
		t.Fatalf("could not push element to the queue : %v", err)
	}

	if len(uaq.Uslice) != 4 {
		t.Fatalf("push to the non-acked queue failed")
	}
	for i := 0; i < 4; i++ {
		if uaq.Uslice[i].Id != i+1 {
			t.Fatalf("indexes were not updated correctly. Expected %d got %d", i, uaq.Uslice[i].Id)
		}
	}

	// Check that the queue is a fifo : popped element should not be the one we just pushed.
	popped := uaq.Pop()
	poppedElt, ok := popped.(*stanza.UnAckedStz)
	if !ok {
		t.Fatalf("popped element is not a *stanza.UnAckedStz")
	}

	if reflect.DeepEqual(*poppedElt, toPush) {
		t.Fatalf("pushed element is at the top of the fifo queue when it should be at the bottom")
	}

}

func TestPeekUnack(t *testing.T) {
	uaq := initUnAckQueue()

	expectedPeek := stanza.UnAckedStz{
		Id: 1,
		Stz: `<iq type='set'
    from='romeo@montague.net/home'
    to='characters.shakespeare.lit'
    id='search2'
    xml:lang='en'>
  <query xmlns='jabber:iq:search'>
    <last>Capulet</last>
  </query>
</iq>`,
	}

	if !reflect.DeepEqual(expectedPeek, *uaq.Uslice[0]) {
		t.Fatalf("peek failed to return the correct stanza")
	}

}

func TestPopNUnack(t *testing.T) {
	uaq := initUnAckQueue()
	initLen := len(uaq.Uslice)
	randPop := rand.Int31n(int32(initLen))

	popped := uaq.PopN(int(randPop))

	if len(uaq.Uslice)+len(popped) != initLen {
		t.Fatalf("total length changed whith pop n operation : had %d found %d after pop", initLen, len(uaq.Uslice)+len(popped))
	}

	for _, elt := range popped {
		for _, oldElt := range uaq.Uslice {
			if reflect.DeepEqual(elt, oldElt) {
				t.Fatalf("pop n operation duplicated some elements")
			}
		}
	}
}

func TestPopNUnackTooLong(t *testing.T) {
	uaq := initUnAckQueue()
	initLen := len(uaq.Uslice)

	// Have a random number of elements to pop that's greater than the queue size
	randPop := rand.Int31n(int32(initLen)) + 1 + int32(initLen)

	popped := uaq.PopN(int(randPop))

	if len(uaq.Uslice)+len(popped) != initLen {
		t.Fatalf("total length changed whith pop n operation : had %d found %d after pop", initLen, len(uaq.Uslice)+len(popped))
	}

	for _, elt := range popped {
		for _, oldElt := range uaq.Uslice {
			if reflect.DeepEqual(elt, oldElt) {
				t.Fatalf("pop n operation duplicated some elements")
			}
		}
	}
}

func TestPopUnack(t *testing.T) {
	uaq := initUnAckQueue()
	initLen := len(uaq.Uslice)

	popped := uaq.Pop()

	if len(uaq.Uslice)+1 != initLen {
		t.Fatalf("total length changed whith pop operation : had %d found %d after pop", initLen, len(uaq.Uslice)+1)
	}
	for _, oldElt := range uaq.Uslice {
		if reflect.DeepEqual(popped, oldElt) {
			t.Fatalf("pop n operation duplicated some elements")
		}
	}

}

func initUnAckQueue() stanza.UnAckQueue {
	q := []*stanza.UnAckedStz{
		{
			Id: 1,
			Stz: `<iq type='set'
    from='romeo@montague.net/home'
    to='characters.shakespeare.lit'
    id='search2'
    xml:lang='en'>
  <query xmlns='jabber:iq:search'>
    <last>Capulet</last>
  </query>
</iq>`,
		},
		{Id: 2,
			Stz: `<iq type='get'
    from='juliet@capulet.com/balcony'
    to='characters.shakespeare.lit'
    id='search3'
    xml:lang='en'>
  <query xmlns='jabber:iq:search'/>
</iq>`},
		{Id: 3,
			Stz: `<iq type='set'
    from='juliet@capulet.com/balcony'
    to='characters.shakespeare.lit'
    id='search4'
    xml:lang='en'>
  <query xmlns='jabber:iq:search'>
    <x xmlns='jabber:x:data' type='submit'>
      <field type='hidden' var='FORM_TYPE'>
        <value>jabber:iq:search</value>
      </field>
      <field var='x-gender'>
        <value>male</value>
      </field>
    </x>
  </query>
</iq>`},
	}

	return stanza.UnAckQueue{Uslice: q}

}

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}
