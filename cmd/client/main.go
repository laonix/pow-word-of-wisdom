package main

import (
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/laonix/pow-word-of-wisdom/config"
	"github.com/laonix/pow-word-of-wisdom/logger"
	"github.com/laonix/pow-word-of-wisdom/pow"
)

func main() {
	// client setup
	cfg := initConfig()
	rand.Seed(time.Now().UnixNano())

	log := logger.NewZapLogger(logger.LevelOf(cfg.LoggingLevel))

	log.Info("client settings", "server", cfg.ServerAddr)

	// resolve server address
	tcpAddr, err := net.ResolveTCPAddr("tcp", cfg.ServerAddr)
	if err != nil {
		log.Error(err, "action", "resolve TCP address")
		os.Exit(1)
	}

	// get connection with server
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Error(err, "action", "dial TCP")
		os.Exit(1)
	}

	// send 'ping' message to server to initiate interaction
	log.Info("ping server", "server", conn.RemoteAddr())

	_, err = conn.Write([]byte("ping"))
	if err != nil {
		log.Error(err, "action", "ping server", "server", conn.RemoteAddr())
		closeConn(conn, log)
		os.Exit(1)
	}

	// receive PoW challenge header from server
	readBuffer := make([]byte, 1024)

	n, err := conn.Read(readBuffer)
	if err != nil {
		log.Error(err, "action", "read PoW challenge", "server", conn.RemoteAddr())
		closeConn(conn, log)
		os.Exit(1)
	}

	log.Info("got PoW challenge", "challenge", string(readBuffer[:n]), "server", conn.RemoteAddr())

	// start PoW result calculation
	powResChan := make(chan calcResult, 1)
	var powResult string

	go func() {
		res, err := pow.Calculate(string(readBuffer[:n]))
		powResChan <- calcResult{
			result: res,
			err:    err,
		}
	}()

	// while client calculates PoW result it might receive internal error
	// or context cancellation message (when calculation lasts longer than server waiting time) from server
calcLoop:
	for {
		select {
		case res := <-powResChan: // waiting for PoW calculation result
			{
				if res.err != nil {
					log.Error(err, "action", "calculate PoW result")
					closeConn(conn, log)
					os.Exit(1)
				}

				powResult = res.result

				// unset connection read deadline to proceed with the flow
				if err := conn.SetReadDeadline(time.Time{}); err != nil {
					log.Error(err, "action", "set connection read deadline")
				}

				break calcLoop
			}
		default: // waiting for messages from server during PoW calculation
			{
				// to loop over we set a short read deadline to connection
				if err := conn.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
					log.Error(err, "action", "set connection read deadline")
				}

				n, err = conn.Read(readBuffer)
				if err != nil {
					// if the error is connected with reaching a read deadline we loop over
					if os.IsTimeout(err) {
						continue
					}

					log.Error(err, "action", "read while calculating PoW result", "server", conn.RemoteAddr())
					closeConn(conn, log)
					os.Exit(1)
				}

				log.Info("got a message from server", "message", string(readBuffer[:n]))

				// a message from server received during PoW calculation flags us to wrap up the flow as we are done here
				closeConn(conn, log)
				return
			}
		}
	}

	// send PoW calculation result to server
	log.Info("PoW result calculated", "result", powResult)

	_, err = conn.Write([]byte(powResult))
	if err != nil {
		log.Error(err, "action", "send PoW result", "server", conn.RemoteAddr())
		closeConn(conn, log)
		os.Exit(1)
	}

	// read a word of wisdom from server
	n, err = conn.Read(readBuffer)
	if err != nil {
		log.Error(err, "action", "read quote", "server", conn.RemoteAddr())
		closeConn(conn, log)
		os.Exit(1)
	}

	log.Info("got a word of wisdom", "quote", string(readBuffer[:n]))

	closeConn(conn, log)
}

type calcResult struct {
	result string
	err    error
}

func initConfig() *config.ClientParameters {
	params := config.ClientParameters{}
	if err := env.Parse(&params); err != nil {
		panic(err)
	}

	return &params
}

func closeConn(conn net.Conn, log logger.Logger) {
	log.Debug("close TCP connection")
	if err := conn.Close(); err != nil {
		log.Error(err, "action", "close TCP connection")
	}
}
