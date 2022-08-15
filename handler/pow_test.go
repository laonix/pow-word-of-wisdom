package handler

//go:generate mockery --dir=../tcp --name=Conn --case underscore
//go:generate mockery --dir=../logger --name=Logger --case underscore
//go:generate mockery --dir=../tcp --name=Handler --case underscore
//go:generate mockery --dir=../pow --name=ChallengeFunc --case underscore
//go:generate mockery --dir=../pow --name=VerifyFunc --case underscore

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/laonix/pow-word-of-wisdom/handler/mocks"
)

func TestProofOfWork_ServeTCP_correct(t *testing.T) {
	challengeStr := "1:12:2208082121:d778f1e9-d0a8-485e-ab51-053a12e9b397::cRvZdlXCCIrWoQ==:NDAwMjk4NDM4NTU1MTUyNDEzOA=="
	calculatedStr := "1:12:2208082121:d778f1e9-d0a8-485e-ab51-053a12e9b397::cRvZdlXCCIrWoQ==:NDAwMjk4NDM4NTU1MTUyODcyOA=="

	log := setupLogMock(t)

	challenge := mocks.NewChallengeFunc(t)
	challenge.On("Execute", mock.AnythingOfType("uint"), mock.AnythingOfType("string")).
		Return(challengeStr, nil)

	verify := mocks.NewVerifyFunc(t)
	verify.On("Execute", calculatedStr, challengeStr).Return(true, nil)

	settings := ProofOfWorkSettings{
		Challenge:  challenge.Execute,
		Verify:     verify.Execute,
		Complexity: 20,
		WaitPOW:    1 * time.Minute,
	}

	cancellingCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn := mocks.NewConn(t)
	conn.On("Close").Return(nil)
	conn.On("RemoteAddr").Return(func() net.Addr { return &net.TCPAddr{Port: 80} })
	conn.On("Read", mock.AnythingOfType("[]uint8")).Return([]byte("ping"), nil).Once()
	conn.On("Write", []byte(challengeStr)).Return(len([]byte(challengeStr)), nil).Once()
	conn.On("Read", mock.AnythingOfType("[]uint8")).Return([]byte(calculatedStr), nil).Once()

	mockHandler := mocks.NewHandler(t)
	mockHandler.On("ServeTCP", cancellingCtx, conn).Run(func(args mock.Arguments) {
		conn.Close()
	}).Once()

	handler := NewProofOfWork(mockHandler, settings, log)

	handler.ServeTCP(cancellingCtx, conn)

	log.AssertNumberOfCalls(t, "Info", 3)  // on read from and write to conn, no errors
	log.AssertNumberOfCalls(t, "Debug", 1) // on get header to verify, no errors
	log.AssertNumberOfCalls(t, "Warn", 0)  // ctx hasn't been cancelled
	log.AssertNumberOfCalls(t, "Error", 0) // no errors
}

func TestProofOfWork_ServeTCP_verification_timeout(t *testing.T) {
	challengeStr := "1:12:2208082121:d778f1e9-d0a8-485e-ab51-053a12e9b397::cRvZdlXCCIrWoQ==:NDAwMjk4NDM4NTU1MTUyNDEzOA=="
	calculatedStr := "1:12:2208082121:d778f1e9-d0a8-485e-ab51-053a12e9b397::cRvZdlXCCIrWoQ==:NDAwMjk4NDM4NTU1MTUyODcyOA=="

	log := setupLogMock(t)

	challenge := mocks.NewChallengeFunc(t)
	challenge.On("Execute", mock.AnythingOfType("uint"), mock.AnythingOfType("string")).
		Return(challengeStr, nil)

	verify := mocks.NewVerifyFunc(t)
	verify.On("Execute", mock.AnythingOfType("string"), challengeStr).Maybe().Return(true, nil)

	settings := ProofOfWorkSettings{
		Challenge:  challenge.Execute,
		Verify:     verify.Execute,
		Complexity: 20,
		WaitPOW:    1 * time.Nanosecond,
	}

	cancellingCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn := mocks.NewConn(t)
	conn.On("Close").Return(nil)
	conn.On("RemoteAddr").Return(func() net.Addr { return &net.TCPAddr{Port: 80} })
	conn.On("Read", mock.AnythingOfType("[]uint8")).Return([]byte("ping"), nil).Once()
	conn.On("Write", []byte(challengeStr)).Return(len([]byte(challengeStr)), nil).Once()
	conn.On("Read", mock.AnythingOfType("[]uint8")).Maybe().Return([]byte(calculatedStr), nil).Once()
	conn.On("Write", []byte("context done")).Return(len([]byte("context done")), nil).Once()

	mockHandler := mocks.NewHandler(t)

	handler := NewProofOfWork(mockHandler, settings, log)

	handler.ServeTCP(cancellingCtx, conn)

	log.AssertNumberOfCalls(t, "Info", 3)  // on read from and write to conn, no errors
	log.AssertNumberOfCalls(t, "Debug", 2) // on read calc result and close conn, no errors
	log.AssertNumberOfCalls(t, "Warn", 1)  // ctx has been cancelled
	log.AssertNumberOfCalls(t, "Error", 0) // no errors
}

func TestProofOfWork_ServeTCP_context_cancelled(t *testing.T) {
	challengeStr := "1:12:2208082121:d778f1e9-d0a8-485e-ab51-053a12e9b397::cRvZdlXCCIrWoQ==:NDAwMjk4NDM4NTU1MTUyNDEzOA=="
	calculatedStr := "1:12:2208082121:d778f1e9-d0a8-485e-ab51-053a12e9b397::cRvZdlXCCIrWoQ==:NDAwMjk4NDM4NTU1MTUyODcyOA=="

	log := setupLogMock(t)

	challenge := mocks.NewChallengeFunc(t)
	challenge.On("Execute", mock.AnythingOfType("uint"), mock.AnythingOfType("string")).
		Return(challengeStr, nil)

	verify := mocks.NewVerifyFunc(t)
	verify.On("Execute", mock.AnythingOfType("string"), challengeStr).Maybe().Return(true, nil)

	settings := ProofOfWorkSettings{
		Challenge:  challenge.Execute,
		Verify:     verify.Execute,
		Complexity: 20,
		WaitPOW:    1 * time.Minute,
	}

	cancellingCtx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(1*time.Nanosecond, cancel)

	conn := mocks.NewConn(t)
	conn.On("Close").Return(nil)
	conn.On("RemoteAddr").Return(func() net.Addr { return &net.TCPAddr{Port: 80} })
	conn.On("Read", mock.AnythingOfType("[]uint8")).Return([]byte("ping"), nil).Once()
	conn.On("Write", []byte(challengeStr)).Return(len([]byte(challengeStr)), nil).Once()
	conn.On("Read", mock.AnythingOfType("[]uint8")).Maybe().Return([]byte(calculatedStr), nil).Once()
	conn.On("Write", []byte("context done")).Return(len([]byte("context done")), nil).Once()

	mockHandler := mocks.NewHandler(t)

	handler := NewProofOfWork(mockHandler, settings, log)

	handler.ServeTCP(cancellingCtx, conn)

	log.AssertNumberOfCalls(t, "Info", 3)  // on read from and write to conn, no errors
	log.AssertNumberOfCalls(t, "Debug", 2) // on read calc result and close conn, no errors
	log.AssertNumberOfCalls(t, "Warn", 1)  // ctx has been cancelled
	log.AssertNumberOfCalls(t, "Error", 0) // no errors
}

func TestProofOfWork_ServeTCP_verification_failed(t *testing.T) {
	challengeStr := "1:12:2208082121:d778f1e9-d0a8-485e-ab51-053a12e9b397::cRvZdlXCCIrWoQ==:NDAwMjk4NDM4NTU1MTUyNDEzOA=="
	calculatedStr := "1:12:2208082121:d778f1e9-d0a8-485e-ab51-053a12e9b397::cRvZdlXCCIrWoQ==:NDAwMjk4NDM4NTU1MTUyODcyNw=="

	log := setupLogMock(t)

	challenge := mocks.NewChallengeFunc(t)
	challenge.On("Execute", mock.AnythingOfType("uint"), mock.AnythingOfType("string")).
		Return(challengeStr, nil)

	verify := mocks.NewVerifyFunc(t)
	verify.On("Execute", calculatedStr, challengeStr).Return(false, nil)

	settings := ProofOfWorkSettings{
		Challenge:  challenge.Execute,
		Verify:     verify.Execute,
		Complexity: 20,
		WaitPOW:    1 * time.Minute,
	}

	cancellingCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn := mocks.NewConn(t)
	conn.On("Close").Return(nil)
	conn.On("RemoteAddr").Return(func() net.Addr { return &net.TCPAddr{Port: 80} })
	conn.On("Read", mock.AnythingOfType("[]uint8")).Return([]byte("ping"), nil).Once()
	conn.On("Write", []byte(challengeStr)).Return(len([]byte(challengeStr)), nil).Once()
	conn.On("Read", mock.AnythingOfType("[]uint8")).Return([]byte(calculatedStr), nil).Once()
	conn.On("Write", []byte("PoW verification failed")).Return(len([]byte("PoW verification failed")), nil).Once()

	mockHandler := mocks.NewHandler(t)

	handler := NewProofOfWork(mockHandler, settings, log)

	handler.ServeTCP(cancellingCtx, conn)

	log.AssertNumberOfCalls(t, "Info", 3)  // on read from and write to conn, no errors
	log.AssertNumberOfCalls(t, "Debug", 2) // on read calc result and close conn, no errors
	log.AssertNumberOfCalls(t, "Warn", 1)  // PoW verification failed
	log.AssertNumberOfCalls(t, "Error", 0) // no errors
}
