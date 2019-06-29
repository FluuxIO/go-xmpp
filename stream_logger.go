package xmpp

import (
	"io"
	"os"
)

// Mediated Read / Write on socket
// Used if logFile from Config is not nil
type streamLogger struct {
	socket  io.ReadWriter // Actual connection
	logFile *os.File
}

func newStreamLogger(conn io.ReadWriter, logFile *os.File) io.ReadWriter {
	if logFile == nil {
		return conn
	} else {
		return &streamLogger{conn, logFile}
	}
}

func (sp *streamLogger) Read(p []byte) (n int, err error) {
	n, err = sp.socket.Read(p)
	if n > 0 {
		sp.logFile.Write([]byte("RECV:\n")) // Prefix
		if n, err := sp.logFile.Write(p[:n]); err != nil {
			return n, err
		}
		sp.logFile.Write([]byte("\n\n")) // Separator
	}
	return
}

func (sp *streamLogger) Write(p []byte) (n int, err error) {
	sp.logFile.Write([]byte("SEND:\n")) // Prefix
	for _, w := range []io.Writer{sp.socket, sp.logFile} {
		n, err = w.Write(p)
		if err != nil {
			return
		}
		if n != len(p) {
			err = io.ErrShortWrite
			return
		}
	}
	sp.logFile.Write([]byte("\n\n")) // Separator
	return len(p), nil
}

/*
TODO: Make RECV, SEND prefixes +
*/
