package main

import (
	"context"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/laonix/pow-word-of-wisdom/config"
	"github.com/laonix/pow-word-of-wisdom/handler"
	"github.com/laonix/pow-word-of-wisdom/logger"
	"github.com/laonix/pow-word-of-wisdom/pow"
	"github.com/laonix/pow-word-of-wisdom/service"
	"github.com/laonix/pow-word-of-wisdom/tcp"
)

func main() {
	// sever setup
	cfg := initConfig()
	rand.Seed(time.Now().UnixNano())

	log := logger.NewZapLogger(logger.LevelOf(cfg.LoggingLevel))

	// initiate a word of wisdom handler
	quoteGetter := service.NewFileGetter()
	wordOfWisdomSrv := service.NewWordOfWisdomService(quoteGetter)

	wordOfWisdomHandler := handler.NewWordOfWisdomHandler(wordOfWisdomSrv, log)

	// initiate a PoW handler
	settings := handler.ProofOfWorkSettings{
		Challenge:  pow.Challenge,
		Verify:     pow.Verify,
		Complexity: cfg.Complexity,
		WaitPOW:    cfg.WaitPOW,
	}
	powHandler := handler.NewProofOfWork(wordOfWisdomHandler, settings, log)

	// initiate TCP server
	tcpServer := tcp.NewServer(cfg.TCPAddr, powHandler, log)

	// create cancelling context to handle a graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// start TCP server
	go func() {
		if err := tcpServer.ListenAndServe(ctx); err != nil {
			log.Error(err, "action", "tcp listen and serve")
		}
	}()

	log.Info("server settings", "complexity", cfg.Complexity, "wait PoW duration", cfg.WaitPOW)

	// start listening for external signals to handle a server graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	log.Info("received system interruption", "signal", <-c)
	cancel()

	// wait for completion of graceful shutdown
	time.Sleep(3 * time.Second)
}

func initConfig() *config.ServerParameters {
	params := config.ServerParameters{}
	if err := env.Parse(&params); err != nil {
		panic(err)
	}

	return &params
}
