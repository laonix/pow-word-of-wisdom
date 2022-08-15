package tcp

import "net"

// Conn is a contract to work with a generic stream-oriented network connection.
type Conn interface {
	Read(b []byte) (read []byte, err error)
	Write(b []byte) (n int, err error)
	Close() error
	RemoteAddr() net.Addr
}

// ConnWrapper is an implementation of Conn.
// It's a wrapper over net.Conn.
type ConnWrapper struct {
	conn net.Conn
}

// Read returns the result of reading from the connection.
//
// It returns read bytes slice instead of the number of read bytes.
func (w *ConnWrapper) Read(b []byte) (read []byte, err error) {
	n, err := w.conn.Read(b)

	return b[:n], err
}

// Write performs net.Conn#Write.
func (w *ConnWrapper) Write(b []byte) (n int, err error) {
	return w.conn.Write(b)
}

// Close performs net.Conn#Close.
func (w *ConnWrapper) Close() error {
	return w.conn.Close()
}

// RemoteAddr performs net.Conn#RemoteAddr.
func (w *ConnWrapper) RemoteAddr() net.Addr {
	return w.conn.RemoteAddr()
}
