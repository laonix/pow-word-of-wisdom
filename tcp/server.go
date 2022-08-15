package tcp

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/laonix/pow-word-of-wisdom/logger"
)

const NetworkTcp = "tcp"

// Handler is a contract to serve a TCP connection.
type Handler interface {
	ServeTCP(ctx context.Context, conn Conn)
}

// Server holds settings and handler to serve accepted TCP connections.
type Server struct {
	addr    string
	handler Handler
	log     logger.Logger
}

// NewServer returns a new instance of Server.
func NewServer(addr string, handler Handler, log logger.Logger) *Server {
	return &Server{
		addr:    addr,
		handler: handler,
		log:     log,
	}
}

// ListenAndServe listens for a new TCP connections on a declared port.
//
// Once the connection accepted control hands over to the underlying Handler.
// If the context is cancelled, TCP connections listener closes.
func (s *Server) ListenAndServe(ctx context.Context) error {
	addr, err := net.ResolveTCPAddr(NetworkTcp, s.addr)
	if err != nil {
		return fmt.Errorf("resolve TCP address: %w", err)
	}

	l, err := net.ListenTCP(NetworkTcp, addr)
	if err != nil {
		return fmt.Errorf("listen TCP: %w", err)
	}

	host, port, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return fmt.Errorf("get listened host and port: %w", err)
	}
	s.log.Info("listening for TCP connections", "host", host, "port", port)

	// while listening for accepting connections we might get context cancellation
	for {
		select {
		case <-ctx.Done(): // handle context cancellation
			{
				s.log.Debug("close TCP connections listener")
				if err := l.Close(); err != nil {
					s.log.Error(err, "action", "close TCP connections listener")
				}
				return nil
			}
		default: // waiting for connections to accept
			{
				// to loop over we set a short deadline to the listener
				if err := l.SetDeadline(time.Now().Add(time.Second)); err != nil {
					return fmt.Errorf("set TCP listener deadline: %w", err)
				}

				conn, err := l.Accept()
				if err != nil {
					if os.IsTimeout(err) { // if the error is received due to timeout keep looping
						continue
					}
					return fmt.Errorf("accept connection: %w", err)
				}

				wrapped := &ConnWrapper{conn: conn}

				go s.handler.ServeTCP(ctx, wrapped)
			}
		}
	}
}
