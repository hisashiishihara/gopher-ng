package main

import (
	"bytes"
	"errors"
	"io"
	"net"
	"reflect"
	"strings"
	"testing"

	"github.com/hisashiishihara/gopher-ng/internal/protocol"
	"github.com/hisashiishihara/gopher-ng/internal/server"
)

func TestFetch(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		wantSelector string
		response     []protocol.Record
	}{
		{
			name:         "root",
			path:         "/",
			wantSelector: "/",
			response: []protocol.Record{{
				Type:   protocol.TypeEntity,
				Fields: []string{"gopher-ng:Server", "gopher-ng:root"},
			}},
		},
		{
			name:         "decoded path",
			path:         "/pet/Moko%20Chan",
			wantSelector: "/pet/Moko Chan",
			response:     []protocol.Record{{Type: protocol.TypeFact, Fields: []string{"pet:name", "Moko Chan"}}},
		},
		{
			name:         "NOT_FOUND response",
			path:         "/missing",
			wantSelector: "/missing",
			response:     []protocol.Record{{Type: protocol.TypeError, Fields: []string{"NOT_FOUND"}}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			address, wait := serveOnce(t, func(selector string) ([]protocol.Record, error) {
				if selector != test.wantSelector {
					t.Errorf("selector = %q, want %q", selector, test.wantSelector)
				}
				return test.response, nil
			})
			records, err := fetch("gofer://" + address + test.path)
			if err != nil {
				t.Fatalf("fetch() error = %v", err)
			}
			if err := wait(); err != nil {
				t.Fatalf("ServeConn() error = %v", err)
			}
			if !reflect.DeepEqual(records, test.response) {
				t.Fatalf("fetch() = %#v, want %#v", records, test.response)
			}
		})
	}
}

func TestFetchRejectsInvalidURI(t *testing.T) {
	for _, rawURI := range []string{
		"gofer://127.0.0.1/",
		"http://127.0.0.1:7070/",
		"gofer://127.0.0.1:7070/%ZZ",
	} {
		if _, err := fetch(rawURI); !errors.Is(err, protocol.ErrInvalidURI) {
			t.Errorf("fetch(%q) error = %v, want ErrInvalidURI", rawURI, err)
		}
	}
}

func TestFetchRejectsIncompleteResponse(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	serverErr := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverErr <- err
			return
		}
		defer conn.Close()
		_, err = protocol.ReadSelector(conn)
		if err == nil {
			_, err = io.WriteString(conn, "FACT\tpet:name\tMoko\r\n")
		}
		serverErr <- err
	}()

	_, err = fetch("gofer://" + listener.Addr().String() + "/")
	if !errors.Is(err, protocol.ErrIncompleteResponse) {
		t.Fatalf("fetch() error = %v, want ErrIncompleteResponse", err)
	}
	if err := <-serverErr; err != nil {
		t.Fatalf("test server error = %v", err)
	}
}

func TestParseResponseWithinLimit(t *testing.T) {
	wire := "FACT\tpet:name\tMoko\r\n.\r\n"
	want := []protocol.Record{{Type: protocol.TypeFact, Fields: []string{"pet:name", "Moko"}}}

	records, err := parseResponse(strings.NewReader(wire))
	if err != nil {
		t.Fatalf("parseResponse() error = %v", err)
	}
	if !reflect.DeepEqual(records, want) {
		t.Fatalf("parseResponse() = %#v, want %#v", records, want)
	}
}

func TestParseResponseWireByteLimit(t *testing.T) {
	const framingBytes = len("FACT\tk\t") + len("\r\n") + len(".\r\n")
	atLimit := "FACT\tk\t" + strings.Repeat("x", int(maxResponseSize)-framingBytes) + "\r\n.\r\n"

	if got := int64(len(atLimit)); got != maxResponseSize {
		t.Fatalf("test response size = %d, want %d", got, maxResponseSize)
	}
	if _, err := parseResponse(strings.NewReader(atLimit)); err != nil {
		t.Fatalf("parseResponse(at limit) error = %v", err)
	}

	overLimit := strings.Replace(atLimit, "\r\n.\r\n", "x\r\n.\r\n", 1)
	if _, err := parseResponse(strings.NewReader(overLimit)); !errors.Is(err, errResponseTooLarge) {
		t.Fatalf("parseResponse(over limit) error = %v, want errResponseTooLarge", err)
	}
}

func TestDialAddress(t *testing.T) {
	uri, err := protocol.ParseURI("gofer://[2001:db8::1]:7070/")
	if err != nil {
		t.Fatalf("ParseURI() error = %v", err)
	}
	if got, want := dialAddress(uri), "[2001:db8::1]:7070"; got != want {
		t.Fatalf("dialAddress() = %q, want %q", got, want)
	}
}

func TestPrintRecords(t *testing.T) {
	var output bytes.Buffer
	records := []protocol.Record{
		{Type: protocol.TypeFact, Fields: []string{"pet:name", "Moko"}},
		{Type: protocol.TypeError, Fields: []string{"NOT_FOUND"}},
	}
	if err := printRecords(&output, records); err != nil {
		t.Fatalf("printRecords() error = %v", err)
	}
	if got, want := output.String(), "FACT\tpet:name\tMoko\nERROR\tNOT_FOUND\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRunRequiresOneURI(t *testing.T) {
	if err := run(nil, io.Discard); err == nil {
		t.Fatal("run() error = nil, want argument error")
	}
	if err := run([]string{"one", "two"}, io.Discard); err == nil {
		t.Fatal("run() error = nil, want argument error")
	}
}

func serveOnce(t *testing.T, handler server.Handler) (string, func() error) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	errCh := make(chan error, 1)
	go func() {
		defer listener.Close()
		conn, err := listener.Accept()
		if err != nil {
			errCh <- err
			return
		}
		errCh <- server.ServeConn(conn, handler)
	}()

	return listener.Addr().String(), func() error { return <-errCh }
}
