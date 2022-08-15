package handler

//go:generate mockery --dir=../service --name=WordOfWisdom --case underscore

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/laonix/pow-word-of-wisdom/handler/mocks"
)

func TestWordOfWisdomHandler_ServeTCP_correct(t *testing.T) {
	log := setupLogMock(t)

	svc := mocks.NewWordOfWisdom(t)
	svc.On("Quote").Return("random quote", nil)

	handler := NewWordOfWisdomHandler(svc, log)

	conn := setupConnMock(t)
	conn.On("Write", []byte("random quote")).Return(len([]byte("random quote")), nil)

	cancellingCtx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(5*time.Millisecond, cancel)

	handler.ServeTCP(cancellingCtx, conn)

	log.AssertNumberOfCalls(t, "Info", 1)  // on write quote to conn, no errors
	log.AssertNumberOfCalls(t, "Debug", 1) // on closing conn, no errors
	log.AssertNumberOfCalls(t, "Warn", 0)  // ctx hasn't been cancelled
	log.AssertNumberOfCalls(t, "Error", 0) // no errors
}

func TestWordOfWisdomHandler_ServeTCP_internal_error(t *testing.T) {
	log := setupLogMock(t)

	svc := mocks.NewWordOfWisdom(t)
	svc.On("Quote").Return("", errors.New("get random quote id"))

	handler := NewWordOfWisdomHandler(svc, log)

	conn := setupConnMock(t)
	conn.On("Write", []byte("cannot get a quote")).Return(len([]byte("cannot get a quote")), nil)

	cancellingCtx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(5*time.Millisecond, cancel)

	handler.ServeTCP(cancellingCtx, conn)

	log.AssertNumberOfCalls(t, "Info", 1)  // on write internal error message to conn, no errors
	log.AssertNumberOfCalls(t, "Debug", 1) // on closing conn, no errors
	log.AssertNumberOfCalls(t, "Warn", 0)  // ctx hasn't been cancelled
	log.AssertNumberOfCalls(t, "Error", 1) // log internal error
}

func TestWordOfWisdomHandler_ServeTCP_context_cancelled(t *testing.T) {
	log := setupLogMock(t)

	svc := mocks.NewWordOfWisdom(t)
	svc.On("Quote").Maybe().Run(func(_ mock.Arguments) {
		time.Sleep(10 * time.Millisecond)
	}).Return("random quote", nil)

	handler := NewWordOfWisdomHandler(svc, log)

	conn := setupConnMock(t)
	conn.On("Write", []byte("random quote")).Maybe().Return(len([]byte("random quote")), nil)
	conn.On("Write", []byte("context done")).Maybe().Return(len([]byte("context done")), nil)

	cancellingCtx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(1*time.Millisecond, cancel)

	handler.ServeTCP(cancellingCtx, conn)

	log.AssertNumberOfCalls(t, "Info", 1)  // write to conn on context done, no errors
	log.AssertNumberOfCalls(t, "Debug", 1) // on closing conn, no errors
	log.AssertNumberOfCalls(t, "Warn", 1)  // on ctx done
	log.AssertNumberOfCalls(t, "Error", 0) // no errors
}

var skip = mock.Anything

func setupLogMock(t *testing.T) *mocks.Logger {
	skippedLogArgs := []interface{}{skip, skip, skip, skip, skip}

	log := mocks.NewLogger(t)
	log.On("Info", skippedLogArgs...).Maybe()
	log.On("Debug", skippedLogArgs...).Maybe()
	log.On("Warn", skippedLogArgs...).Maybe()
	log.On("Error", skippedLogArgs...).Maybe()

	return log
}

func setupConnMock(t *testing.T) *mocks.Conn {
	conn := mocks.NewConn(t)
	conn.On("Close").Return(nil)
	conn.On("RemoteAddr").Return(func() net.Addr { return &net.TCPAddr{Port: 80} })

	return conn
}
