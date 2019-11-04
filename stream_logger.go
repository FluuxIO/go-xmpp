package xmpp

import (
	"io"
)

// Mediated Read / Write on socket
// Used if logFile from Config is not nil
type streamLogger struct {
	socket  io.ReadWriter // Actual connection
	logFile io.Writer
}

func newStreamLogger(conn io.ReadWriter, logFile io.Writer) io.ReadWriter {
	if logFile == nil {
		return conn
	} else {
		return &streamLogger{conn, logFile}
	}
}

func (sl *streamLogger) Read(p []byte) (n int, err error) {
	n, err = sl.socket.Read(p)
	if n > 0 {
		sl.logFile.Write([]byte("RECV:\n")) // Prefix
		if n, err := sl.logFile.Write(p[:n]); err != nil {
			return n, err
		}
		sl.logFile.Write([]byte("\n\n")) // Separator
	}
	return
}

func (sl *streamLogger) Write(p []byte) (n int, err error) {
	sl.logFile.Write([]byte("SEND:\n")) // Prefix
	for _, w := range []io.Writer{sl.socket, sl.logFile} {
		n, err = w.Write(p)
		if err != nil {
			return
		}
		if n != len(p) {
			err = io.ErrShortWrite
			return
		}
	}
	sl.logFile.Write([]byte("\n\n")) // Separator
	return len(p), nil
}

/*
TODO: Make RECV, SEND prefixes +
*/
