package main

import (
	"errors"
	"reflect"
	"testing"

	"github.com/hisashiishihara/gopher-ng/internal/protocol"
)

func TestHandler(t *testing.T) {
	target := "gofer://127.0.0.1:7071/resource"
	h := handler(target)

	t.Run("root", func(t *testing.T) {
		records, err := h("/")
		if err != nil {
			t.Fatalf("handler() error = %v", err)
		}
		want := []protocol.Record{
			{Type: protocol.TypeEntity, Fields: []string{"example:Directory", "example:federation-root"}},
			{Type: protocol.TypeLink, Fields: []string{"related", target}},
		}
		if !reflect.DeepEqual(records, want) {
			t.Fatalf("handler() = %#v, want %#v", records, want)
		}
		if _, err := protocol.ParseURI(records[1].Fields[1]); err != nil {
			t.Fatalf("LINK target ParseURI() error = %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		records, err := h("/missing")
		if err != nil {
			t.Fatalf("handler() error = %v", err)
		}
		want := []protocol.Record{{Type: protocol.TypeError, Fields: []string{"NOT_FOUND"}}}
		if !reflect.DeepEqual(records, want) {
			t.Fatalf("handler() = %#v, want %#v", records, want)
		}
	})
}

func TestRunRejectsInvalidTarget(t *testing.T) {
	if err := run("127.0.0.1:0", "http://example.org:7071/resource"); !errors.Is(err, protocol.ErrInvalidURI) {
		t.Fatalf("run() error = %v, want ErrInvalidURI", err)
	}
}
