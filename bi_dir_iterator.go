package xmpp

type BiDirIterator interface {
	// Next returns the next element of this iterator, if a response is available within t milliseconds
	Next(t int) (BiDirIteratorElt, error)
	// Previous returns the previous element of this iterator, if a response is available within t milliseconds
	Previous(t int) (BiDirIteratorElt, error)
}

type BiDirIteratorElt interface {
	NoOp()
}
