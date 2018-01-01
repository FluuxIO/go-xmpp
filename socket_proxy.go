package xmpp // import "fluux.io/xmpp"

import (
	"io"
	"os"
)

// Mediated Read / Write on socket
// Used if logFile from Options is not nil
type socketProxy struct {
	socket  io.ReadWriter // Actual connection
	logFile *os.File
}

func newSocketProxy(conn io.ReadWriter, logFile *os.File) io.ReadWriter {
	if logFile == nil {
		return conn
	} else {
		return &socketProxy{conn, logFile}
	}
}

func (pl *socketProxy) Read(p []byte) (n int, err error) {
	n, err = pl.socket.Read(p)
	if n > 0 {
		pl.logFile.Write([]byte("RECV:\n")) // Prefix
		if n, err := pl.logFile.Write(p[:n]); err != nil {
			return n, err
		}
		pl.logFile.Write([]byte("\n\n")) // Separator
	}
	return
}

func (pl *socketProxy) Write(p []byte) (n int, err error) {
	pl.logFile.Write([]byte("SEND:\n")) // Prefix
	for _, w := range []io.Writer{pl.socket, pl.logFile} {
		n, err = w.Write(p)
		if err != nil {
			return
		}
		if n != len(p) {
			err = io.ErrShortWrite
			return
		}
	}
	pl.logFile.Write([]byte("\n\n")) // Separator
	return len(p), nil
}
