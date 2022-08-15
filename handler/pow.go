package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"

	"github.com/google/uuid"

	"github.com/laonix/pow-word-of-wisdom/logger"
	"github.com/laonix/pow-word-of-wisdom/pow"
	"github.com/laonix/pow-word-of-wisdom/tcp"
)

// ProofOfWork implements tcp.Handler
// to perform proof of work check before handing over control to the next tcp.Handler.
type ProofOfWork struct {
	challenge pow.ChallengeFunc
	verify    pow.VerifyFunc

	complexity int
	waitPOW    time.Duration

	handler tcp.Handler
	log     logger.Logger
}

// ProofOfWorkSettings holds ProofOfWork settings
// such as methods to create PoW challenge header and to verify client's calculation,
// challenge complexity, and PoW calculation result waiting time.
type ProofOfWorkSettings struct {
	Challenge pow.ChallengeFunc
	Verify    pow.VerifyFunc

	// Complexity is an upper limit for a randomly generated challenge header bits.
	//
	// Bits should vary in interval [10, Complexity).
	Complexity int
	WaitPOW    time.Duration
}

// NewProofOfWork returns a new instance of ProofOfWork.
func NewProofOfWork(handler tcp.Handler, settings ProofOfWorkSettings, log logger.Logger) *ProofOfWork {
	return &ProofOfWork{
		handler:    handler,
		challenge:  settings.Challenge,
		verify:     settings.Verify,
		complexity: settings.Complexity,
		waitPOW:    settings.WaitPOW,
		log:        log,
	}
}

// ServeTCP takes control over a newly accepted connection.
//
// It challenges a connected client with PoW header, waits for a calculation result and verifies it.
// If awaiting time exceeds a defined limit, this handler informs a client about operation context cancellation and
// closes the connection.
// If the received PoW calculation result cannot pass the verification,
// the handler informs the client about a verification failure and closes the connection.
func (h *ProofOfWork) ServeTCP(ctx context.Context, conn tcp.Conn) {
	// read initial message from connection
	// the message itself doesn't matter, it only flags about the intention to initiate the flow
	tmp := make([]byte, 1024)
	tmp, err := conn.Read(tmp)
	if err != nil && !errors.Is(err, io.EOF) {
		h.log.Error(err, "action", "read from connection")
		closeConn(conn, h.log)
		return
	}

	h.log.Info("got message", "message", string(tmp), "remote", conn.RemoteAddr().String())

	// send PoW challenge header to the client

	// bits should vary in interval [10, complexity)
	// it makes no sense to set bits less than 10 as PoW calculation appears too simple
	bits := rand.Intn(h.complexity-10) + 10
	// since we have no determined resource to access here (e.g. requested quotes should be randomly chosen)
	// let's set a resource as a random UUID string
	resource := uuid.NewString()

	challenge, err := h.challenge(uint(bits), resource)

	writeMessage(challenge, conn, h.log)

	// get PoW calculation result from the client
	verification := make(chan verificationResult)

	timeOut, cancel := context.WithTimeout(ctx, h.waitPOW) // set calculation result awaiting timeout
	defer cancel()

	go h.getVerificationResult(verification, challenge, conn)

	// while we wait for a calculation result we can either reach an awaiting timeout or get system interruption
loop:
	for {
		select {
		case <-timeOut.Done(): // handle system interruption or timeout
			{
				handleCtxDone(ctx, conn, h.log)
				cancel()
				return
			}
		case v := <-verification: // handle verification result
			{
				if v.err != nil {
					h.log.Error(err, "action", "verify PoW")
					writeMessage("internal error on verifying PoW", conn, h.log)
					closeConn(conn, h.log)
					return
				}
				if !v.ok {
					h.log.Warn("PoW verification failed", "header", v.header, "remote", conn.RemoteAddr().String())
					writeMessage("PoW verification failed", conn, h.log)
					closeConn(conn, h.log)
					return
				} else {
					h.log.Info("PoW verification passed", "header", v.header, "remote", conn.RemoteAddr().String())
					break loop
				}
			}
		}
	}

	// if PoW verification passed hand over control to the next handler
	h.handler.ServeTCP(ctx, conn)
}

type verificationResult struct {
	ok     bool
	header string
	err    error
}

func (h *ProofOfWork) getVerificationResult(v chan verificationResult, challenge string, conn tcp.Conn) {
	// read PoW calculation result from the client
	tmp := make([]byte, 1024)
	tmp, err := conn.Read(tmp)
	if err != nil {
		if errors.Is(err, io.EOF) {
			closeConn(conn, h.log)
			return
		}
		if err, ok := err.(net.Error); ok {
			v <- verificationResult{ok: false, header: "", err: fmt.Errorf("read from closed connection: %w", err)}
			return
		}

		h.log.Error(err, "action", "read from connection")
		return
	}

	header := string(tmp)
	h.log.Debug("header to verify", "header", header, "remote", conn.RemoteAddr())

	// verify a received calculation result
	ok, err := h.verify(header, challenge)

	// pass a verification result to the main handler flow
	v <- verificationResult{ok: ok, header: header, err: err}
}

func handleCtxDone(ctx context.Context, conn tcp.Conn, log logger.Logger) {
	log.Warn("context done", "err", ctx.Err())
	writeMessage("context done", conn, log)
	closeConn(conn, log)
}

func closeConn(conn tcp.Conn, log logger.Logger) {
	log.Debug("close TCP connection", "remote", conn.RemoteAddr())
	if err := conn.Close(); err != nil {
		log.Error(err, "action", "close TCP connection", "remote", conn.RemoteAddr())
	}
}

func writeMessage(message string, conn tcp.Conn, log logger.Logger) {
	log.Info("write message", "message", message, "remote", conn.RemoteAddr())
	if _, err := conn.Write([]byte(message)); err != nil {
		log.Error(err, "action", "write message", "message", message, "remote", conn.RemoteAddr())
	}
}
