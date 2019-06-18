package xmpp

import (
	"fmt"

	"golang.org/x/xerrors"
)

type ConnError struct {
	frame xerrors.Frame
	err   error
	// Permanent will be true if error is not recoverable
	Permanent bool
}

func NewConnError(err error, permanent bool) ConnError {
	return ConnError{err: err, frame: xerrors.Caller(1), Permanent: permanent}
}

func (e ConnError) Format(s fmt.State, verb rune) {
	xerrors.FormatError(e, s, verb)
}

func (e ConnError) FormatError(p xerrors.Printer) error {
	e.frame.Format(p)
	return e.err
}

func (e ConnError) Error() string {
	return fmt.Sprint(e)
}

func (e ConnError) Unwrap() error { return e.err }
