package handler

import (
	"context"

	"github.com/laonix/pow-word-of-wisdom/logger"
	"github.com/laonix/pow-word-of-wisdom/service"
	"github.com/laonix/pow-word-of-wisdom/tcp"
)

// WordOfWisdomHandler implements tcp.Handler
// to send a random word of wisdom quote to the client.
type WordOfWisdomHandler struct {
	srv service.WordOfWisdom
	log logger.Logger
}

// NewWordOfWisdomHandler returns a new instance of WordOfWisdomHandler.
func NewWordOfWisdomHandler(srv service.WordOfWisdom, log logger.Logger) *WordOfWisdomHandler {
	return &WordOfWisdomHandler{
		srv: srv,
		log: log,
	}
}

// ServeTCP writes a random word of wisdom quote to the client.
//
// If the server interrupts, it handles a correct connection closing (with client notification).
func (h *WordOfWisdomHandler) ServeTCP(ctx context.Context, conn tcp.Conn) {
	// get a random word of wisdom quote
	quote := make(chan quoteResult)

	go getQuoteResult(quote, h.srv)

	// while we're getting the quote we might receive a system interruption
	for {
		select {
		case <-ctx.Done(): // handle context cancellation
			{
				handleCtxDone(ctx, conn, h.log)
				return
			}
		case res := <-quote: // handle a retrieved quote
			{
				if res.err != nil {
					h.log.Error(res.err, "action", "get quote")
					writeMessage("cannot get a quote", conn, h.log)
					closeConn(conn, h.log)
					return
				}

				writeMessage(res.quote, conn, h.log)
				closeConn(conn, h.log)
				return
			}
		}
	}
}

type quoteResult struct {
	quote string
	err   error
}

func getQuoteResult(c chan quoteResult, srv service.WordOfWisdom) {
	quote, err := srv.Quote()

	c <- quoteResult{quote: quote, err: err}
}
