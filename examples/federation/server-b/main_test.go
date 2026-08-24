package main

import (
	"reflect"
	"testing"

	"github.com/hisashiishihara/gopher-ng/internal/protocol"
)

func TestHandler(t *testing.T) {
	t.Run("resource", func(t *testing.T) {
		records, err := handler("/resource")
		if err != nil {
			t.Fatalf("handler() error = %v", err)
		}
		want := []protocol.Record{
			{Type: protocol.TypeEntity, Fields: []string{"example:Resource", "example:federation-target"}},
			{Type: protocol.TypeFact, Fields: []string{"example:message", "Hello from server B"}},
		}
		if !reflect.DeepEqual(records, want) {
			t.Fatalf("handler() = %#v, want %#v", records, want)
		}
	})

	t.Run("not found", func(t *testing.T) {
		records, err := handler("/missing")
		if err != nil {
			t.Fatalf("handler() error = %v", err)
		}
		want := []protocol.Record{{Type: protocol.TypeError, Fields: []string{"NOT_FOUND"}}}
		if !reflect.DeepEqual(records, want) {
			t.Fatalf("handler() = %#v, want %#v", records, want)
		}
	})
}
