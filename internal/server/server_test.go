package server

import (
	"bytes"
	"errors"
	"io"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hisashiishihara/gopher-ng/internal/protocol"
)

var errHandlerFailure = errors.New("storage unavailable")

func TestServeConn(t *testing.T) {
	tests := []struct {
		name         string
		request      string
		handler      Handler
		wantResponse string
		wantCalls    int32
		wantError    error
	}{
		{
			name:    "success",
			request: "/pet/123\r\n",
			handler: func(selector string) ([]protocol.Record, error) {
				if selector != "/pet/123" {
					t.Errorf("handler selector = %q, want /pet/123", selector)
				}
				return []protocol.Record{{Type: protocol.TypeFact, Fields: []string{"pet:name", "Moko"}}}, nil
			},
			wantResponse: "FACT\tpet:name\tMoko\r\n.\r\n",
			wantCalls:    1,
		},
		{
			name:         "invalid selector",
			request:      "pet/123\r\n",
			handler:      func(string) ([]protocol.Record, error) { return nil, nil },
			wantResponse: "ERROR\tBAD_SELECTOR\r\n.\r\n",
			wantError:    protocol.ErrInvalidSelector,
		},
		{
			name:    "handler failure",
			request: "/pet/123\r\n",
			handler: func(string) ([]protocol.Record, error) {
				return nil, errHandlerFailure
			},
			wantResponse: "ERROR\tTEMPORARY_FAILURE\r\n.\r\n",
			wantCalls:    1,
			wantError:    errHandlerFailure,
		},
		{
			name:    "handler ERROR record",
			request: "/missing\r\n",
			handler: func(string) ([]protocol.Record, error) {
				return []protocol.Record{{Type: protocol.TypeError, Fields: []string{"NOT_FOUND"}}}, nil
			},
			wantResponse: "ERROR\tNOT_FOUND\r\n.\r\n",
			wantCalls:    1,
		},
		{
			name:    "FACT opaque",
			request: "/pet/123\r\n",
			handler: func(string) ([]protocol.Record, error) {
				return []protocol.Record{{Type: protocol.TypeFact, Fields: []string{"pet:medical-record", "https://vet.example/records/123"}}}, nil
			},
			wantResponse: "FACT\tpet:medical-record\thttps://vet.example/records/123\r\n.\r\n",
			wantCalls:    1,
		},
		{
			name:    "LINK validated",
			request: "/pet/123\r\n",
			handler: func(string) ([]protocol.Record, error) {
				return []protocol.Record{{Type: protocol.TypeLink, Fields: []string{"pet:veterinarian", "http://vet.example:7070/pet/123"}}}, nil
			},
			wantCalls: 1,
			wantError: protocol.ErrInvalidRecord,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			serverConn, clientConn := net.Pipe()
			defer clientConn.Close()

			var calls int32
			handler := func(selector string) ([]protocol.Record, error) {
				atomic.AddInt32(&calls, 1)
				return test.handler(selector)
			}
			errCh := make(chan error, 1)
			go func() { errCh <- ServeConn(serverConn, handler) }()

			if _, err := io.WriteString(clientConn, test.request); err != nil {
				t.Fatalf("write request: %v", err)
			}
			response, err := io.ReadAll(clientConn)
			if err != nil {
				t.Fatalf("read response: %v", err)
			}
			if got := string(response); got != test.wantResponse {
				t.Fatalf("response = %q, want %q", got, test.wantResponse)
			}
			if got := atomic.LoadInt32(&calls); got != test.wantCalls {
				t.Fatalf("handler calls = %d, want %d", got, test.wantCalls)
			}
			err = <-errCh
			if test.wantError == nil && err != nil {
				t.Fatalf("ServeConn() error = %v", err)
			}
			if test.wantError != nil && !errors.Is(err, test.wantError) {
				t.Fatalf("ServeConn() error = %v, want %v", err, test.wantError)
			}
		})
	}
}

func TestServeConnProcessesOneSelector(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	var calls int32
	errCh := make(chan error, 1)
	go func() {
		errCh <- ServeConn(serverConn, func(string) ([]protocol.Record, error) {
			atomic.AddInt32(&calls, 1)
			return nil, nil
		})
	}()

	if _, err := io.WriteString(clientConn, "/one\r\n/two\r\n"); err != nil {
		t.Fatalf("write requests: %v", err)
	}
	response, err := io.ReadAll(clientConn)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if got, want := string(response), ".\r\n"; got != want {
		t.Fatalf("response = %q, want %q", got, want)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("handler calls = %d, want 1", got)
	}
	if err := <-errCh; err != nil {
		t.Fatalf("ServeConn() error = %v", err)
	}
}

func TestServeConnDoesNotRespondToReadError(t *testing.T) {
	readErr := errors.New("connection reset")
	conn := &readErrorConn{readErr: readErr}
	called := false

	err := ServeConn(conn, func(string) ([]protocol.Record, error) {
		called = true
		return nil, nil
	})
	if !errors.Is(err, readErr) {
		t.Fatalf("ServeConn() error = %v, want read error", err)
	}
	if called {
		t.Fatal("handler was invoked after a read error")
	}
	if got := conn.writes.String(); got != "" {
		t.Fatalf("response = %q, want no response", got)
	}
	if !conn.closed {
		t.Fatal("connection was not closed")
	}
}

type readErrorConn struct {
	readErr error
	writes  bytes.Buffer
	closed  bool
}

func (c *readErrorConn) Read([]byte) (int, error)         { return 0, c.readErr }
func (c *readErrorConn) Write(p []byte) (int, error)      { return c.writes.Write(p) }
func (c *readErrorConn) Close() error                     { c.closed = true; return nil }
func (c *readErrorConn) LocalAddr() net.Addr              { return testAddr("local") }
func (c *readErrorConn) RemoteAddr() net.Addr             { return testAddr("remote") }
func (c *readErrorConn) SetDeadline(time.Time) error      { return nil }
func (c *readErrorConn) SetReadDeadline(time.Time) error  { return nil }
func (c *readErrorConn) SetWriteDeadline(time.Time) error { return nil }

type testAddr string

func (a testAddr) Network() string { return "test" }
func (a testAddr) String() string  { return string(a) }
