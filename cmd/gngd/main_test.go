package main

import (
	"errors"
	"io"
	"net"
	"reflect"
	"strings"
	"testing"

	"github.com/hisashiishihara/gopher-ng/internal/protocol"
	"github.com/hisashiishihara/gopher-ng/internal/server"
)

func TestDemoHandler(t *testing.T) {
	tests := []struct {
		name     string
		selector string
		want     []protocol.Record
	}{
		{
			name:     "root",
			selector: "/",
			want: []protocol.Record{{
				Type:   protocol.TypeEntity,
				Fields: []string{"gopher-ng:Server", "gopher-ng:root"},
			}},
		},
		{
			name:     "not found",
			selector: "/missing",
			want:     []protocol.Record{{Type: protocol.TypeError, Fields: []string{"NOT_FOUND"}}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := demoHandler(test.selector)
			if err != nil {
				t.Fatalf("demoHandler() error = %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("demoHandler() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestServeTransactions(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	errCh := make(chan error, 1)
	go func() { errCh <- serve(listener, demoHandler) }()

	for _, test := range []struct {
		selector string
		want     []protocol.Record
	}{
		{"/", []protocol.Record{{Type: protocol.TypeEntity, Fields: []string{"gopher-ng:Server", "gopher-ng:root"}}}},
		{"/missing", []protocol.Record{{Type: protocol.TypeError, Fields: []string{"NOT_FOUND"}}}},
	} {
		got := request(t, listener.Addr().String(), test.selector)
		if !reflect.DeepEqual(got, test.want) {
			t.Fatalf("response for %q = %#v, want %#v", test.selector, got, test.want)
		}
	}

	if err := listener.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}
	if err := <-errCh; !errors.Is(err, net.ErrClosed) {
		t.Fatalf("serve() error = %v, want net.ErrClosed", err)
	}
}

func TestServeContinuesAfterInvalidSelector(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	errCh := make(chan error, 1)
	go func() { errCh <- serve(listener, demoHandler) }()

	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	if _, err := io.WriteString(conn, "invalid\r\n"); err != nil {
		t.Fatalf("write invalid selector: %v", err)
	}
	response, err := io.ReadAll(conn)
	if err != nil {
		t.Fatalf("read invalid selector response: %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatalf("close invalid selector connection: %v", err)
	}
	if got, want := string(response), "ERROR\tBAD_SELECTOR\r\n.\r\n"; got != want {
		t.Fatalf("invalid selector response = %q, want %q", got, want)
	}

	got := request(t, listener.Addr().String(), "/")
	want := []protocol.Record{{Type: protocol.TypeEntity, Fields: []string{"gopher-ng:Server", "gopher-ng:root"}}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("later response = %#v, want %#v", got, want)
	}

	if err := listener.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}
	if err := <-errCh; !errors.Is(err, net.ErrClosed) {
		t.Fatalf("serve() error = %v, want net.ErrClosed", err)
	}
}

func request(t *testing.T, address, selector string) []protocol.Record {
	t.Helper()

	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	if err := protocol.WriteSelector(conn, selector); err != nil {
		t.Fatalf("write selector: %v", err)
	}
	response, err := io.ReadAll(conn)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	records, err := protocol.ParseResponse(strings.NewReader(string(response)))
	if err != nil {
		t.Fatalf("parse response: %v", err)
	}
	return records
}

var _ server.Handler = demoHandler
