package protocol

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseURI(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want URI
	}{
		{"hostname", "gofer://example.org:7070/pet/123", URI{Host: "example.org", Port: 7070, Selector: "/pet/123"}},
		{"root", "gofer://example.org:7070/", URI{Host: "example.org", Port: 7070, Selector: "/"}},
		{"decoded space", "gofer://example.org:7070/pet/Moko%20Chan", URI{Host: "example.org", Port: 7070, Selector: "/pet/Moko Chan"}},
		{"case insensitive scheme", "GOFER://example.org:7070/", URI{Host: "example.org", Port: 7070, Selector: "/"}},
		{"IPv6", "gofer://[2001:db8::1]:7070/pet/123", URI{Host: "2001:db8::1", Port: 7070, Selector: "/pet/123"}},
		{"lowest port", "gofer://example.org:1/", URI{Host: "example.org", Port: 1, Selector: "/"}},
		{"highest port", "gofer://example.org:65535/", URI{Host: "example.org", Port: 65535, Selector: "/"}},
		{"UTF-8 selector", "gofer://example.org:7070/pet/猫", URI{Host: "example.org", Port: 7070, Selector: "/pet/猫"}},
		{"encoded UTF-8 selector", "gofer://example.org:7070/pet/%E7%8C%AB", URI{Host: "example.org", Port: 7070, Selector: "/pet/猫"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParseURI(test.raw)
			if err != nil {
				t.Fatalf("ParseURI() error = %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("ParseURI() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestParseURIRejectsInvalidInput(t *testing.T) {
	tests := []string{
		"http://example.org:7070/",
		"gofer://:7070/",
		"gofer://example.org/",
		"gofer://example.org:0/",
		"gofer://example.org:65536/",
		"gofer://example.org:http/",
		"gofer://example.org:+7070/",
		"gofer://example.org:-7070/",
		"gofer://example.org:7070",
		"gofer://user@example.org:7070/",
		"gofer://example.org:7070/pet/123?x=1",
		"gofer://example.org:7070/pet/123?",
		"gofer://example.org:7070/pet/123#section",
		"gofer://example.org:7070/pet/123#",
		"gofer://example.org:7070/pet/%ZZ",
		"gofer://example.org:7070/pet/%",
		"gofer://example.org:7070/pet/%FF",
		"gofer://example.org:7070/pet/%0Dname",
		"gofer://example.org:7070/pet/%0Aname",
		"gofer://example.org:7070/pet/%09name",
		"gofer://example.org:7070/pet/%00name",
		"gofer://example.org:7070/pet\x1fname",
		"gofer://2001:db8::1:7070/pet/123",
		"gofer://[example.org]:7070/",
		"gofer://[example:org]:7070/",
	}

	for _, raw := range tests {
		if _, err := ParseURI(raw); !errors.Is(err, ErrInvalidURI) {
			t.Errorf("ParseURI(%q) error = %v, want ErrInvalidURI", raw, err)
		}
	}
}
