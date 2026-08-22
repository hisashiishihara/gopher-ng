package protocol

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestParseRecord(t *testing.T) {
	tests := []struct {
		name string
		line string
		want Record
	}{
		{"ENTITY", "ENTITY\tpet:Pet\tpet:123\r\n", Record{Type: TypeEntity, Fields: []string{"pet:Pet", "pet:123"}}},
		{"FACT", "FACT\tpet:name\tMoko\r\n", Record{Type: TypeFact, Fields: []string{"pet:name", "Moko"}}},
		{"LINK", "LINK\tpet:veterinarian\tgofer://vet.example:7070/pet/123\r\n", Record{Type: TypeLink, Fields: []string{"pet:veterinarian", "gofer://vet.example:7070/pet/123"}}},
		{"ERROR", "ERROR\tNOT_FOUND\r\n", Record{Type: TypeError, Fields: []string{"NOT_FOUND"}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ParseRecord(test.line)
			if err != nil {
				t.Fatalf("ParseRecord() error = %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("ParseRecord() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestParseRecordRejectsInvalidInput(t *testing.T) {
	tests := []string{
		"ENTITY\tpet:Pet\r\n",
		"FACT\tpet:name\tMoko\textra\r\n",
		"ERROR\tNOT_FOUND\textra\r\n",
		"entity\tpet:Pet\tpet:123\r\n",
		"ENTITY-\tpet:Pet\tpet:123\r\n",
		"ENTITY\t\tpet:123\r\n",
		"ENTITY\tpet:Pet\tpet:123\n",
	}

	for _, line := range tests {
		if _, err := ParseRecord(line); !errors.Is(err, ErrInvalidRecord) {
			t.Errorf("ParseRecord(%q) error = %v, want ErrInvalidRecord", line, err)
		}
	}

	for _, line := range []string{
		"UNKNOWN\tx\ty\r\n",
		"SERVICE\tpet:medical-record\thttps://vet.example/records/123\r\n",
	} {
		if _, err := ParseRecord(line); !errors.Is(err, ErrUnknownRecordType) {
			t.Errorf("ParseRecord(%q) error = %v, want ErrUnknownRecordType", line, err)
		}
	}
}

func TestValidateSelector(t *testing.T) {
	if err := ValidateSelector("/"); err != nil {
		t.Fatalf("ValidateSelector(root) error = %v", err)
	}

	for _, selector := range []string{"pet/123", "/pet\t123", "/pet\n123", "/pet\x00123"} {
		if err := ValidateSelector(selector); !errors.Is(err, ErrInvalidSelector) {
			t.Errorf("ValidateSelector(%q) error = %v, want ErrInvalidSelector", selector, err)
		}
	}
}

func TestWriteSelector(t *testing.T) {
	var buffer bytes.Buffer
	if err := WriteSelector(&buffer, "/pet/123"); err != nil {
		t.Fatalf("WriteSelector() error = %v", err)
	}
	if got, want := buffer.String(), "/pet/123\r\n"; got != want {
		t.Fatalf("WriteSelector() = %q, want %q", got, want)
	}
}

func TestParseResponse(t *testing.T) {
	t.Run("complete", func(t *testing.T) {
		response := "ENTITY\tpet:Pet\tpet:123\r\nFACT\tpet:name\tMoko\r\n.\r\n"
		records, err := ParseResponse(strings.NewReader(response))
		if err != nil {
			t.Fatalf("ParseResponse() error = %v", err)
		}
		if got, want := len(records), 2; got != want {
			t.Fatalf("len(records) = %d, want %d", got, want)
		}
	})

	t.Run("empty", func(t *testing.T) {
		records, err := ParseResponse(strings.NewReader(".\r\n"))
		if err != nil {
			t.Fatalf("ParseResponse() error = %v", err)
		}
		if len(records) != 0 {
			t.Fatalf("len(records) = %d, want 0", len(records))
		}
	})

	t.Run("ERROR only", func(t *testing.T) {
		records, err := ParseResponse(strings.NewReader("ERROR\tNOT_FOUND\r\n.\r\n"))
		if err != nil {
			t.Fatalf("ParseResponse() error = %v", err)
		}
		if want := []Record{{Type: TypeError, Fields: []string{"NOT_FOUND"}}}; !reflect.DeepEqual(records, want) {
			t.Fatalf("ParseResponse() = %#v, want %#v", records, want)
		}
	})

	for _, test := range []struct {
		name     string
		response string
	}{
		{"ERROR followed by FACT", "ERROR\tNOT_FOUND\r\nFACT\tpet:name\tMoko\r\n.\r\n"},
		{"FACT followed by ERROR", "FACT\tpet:name\tMoko\r\nERROR\tNOT_FOUND\r\n.\r\n"},
		{"multiple ERROR records", "ERROR\tNOT_FOUND\r\nERROR\tTEMPORARY_FAILURE\r\n.\r\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseResponse(strings.NewReader(test.response))
			if !errors.Is(err, ErrInvalidResponse) {
				t.Fatalf("ParseResponse() error = %v, want ErrInvalidResponse", err)
			}
		})
	}

	t.Run("incomplete", func(t *testing.T) {
		_, err := ParseResponse(strings.NewReader("ENTITY\tpet:Pet\tpet:123\r\n"))
		if !errors.Is(err, ErrIncompleteResponse) {
			t.Fatalf("ParseResponse() error = %v, want ErrIncompleteResponse", err)
		}
	})

	t.Run("LF only", func(t *testing.T) {
		_, err := ParseResponse(strings.NewReader("ENTITY\tpet:Pet\tpet:123\n.\n"))
		if !errors.Is(err, ErrInvalidResponse) {
			t.Fatalf("ParseResponse() error = %v, want ErrInvalidResponse", err)
		}
	})

	t.Run("preserves record error", func(t *testing.T) {
		_, err := ParseResponse(strings.NewReader("UNKNOWN\tx\ty\r\n.\r\n"))
		if !errors.Is(err, ErrInvalidResponse) {
			t.Fatalf("ParseResponse() error = %v, want ErrInvalidResponse", err)
		}
		if !errors.Is(err, ErrUnknownRecordType) {
			t.Fatalf("ParseResponse() error = %v, want ErrUnknownRecordType", err)
		}
	})
}

func TestEncodeParseRoundTrip(t *testing.T) {
	want := Record{Type: TypeLink, Fields: []string{"pet:veterinarian", "gofer://vet.example:7070/pet/123"}}
	line, err := EncodeRecord(want)
	if err != nil {
		t.Fatalf("EncodeRecord() error = %v", err)
	}
	got, err := ParseRecord(line)
	if err != nil {
		t.Fatalf("ParseRecord() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip = %#v, want %#v", got, want)
	}
}

func TestWriteResponse(t *testing.T) {
	var buffer bytes.Buffer
	records := []Record{{Type: TypeError, Fields: []string{"NOT_FOUND"}}}
	if err := WriteResponse(&buffer, records); err != nil {
		t.Fatalf("WriteResponse() error = %v", err)
	}
	if got, want := buffer.String(), "ERROR\tNOT_FOUND\r\n.\r\n"; got != want {
		t.Fatalf("WriteResponse() = %q, want %q", got, want)
	}

	for _, records := range [][]Record{
		{{Type: TypeError, Fields: []string{"NOT_FOUND"}}, {Type: TypeFact, Fields: []string{"pet:name", "Moko"}}},
		{{Type: TypeFact, Fields: []string{"pet:name", "Moko"}}, {Type: TypeError, Fields: []string{"NOT_FOUND"}}},
	} {
		buffer.Reset()
		if err := WriteResponse(&buffer, records); !errors.Is(err, ErrInvalidResponse) {
			t.Errorf("WriteResponse() error = %v, want ErrInvalidResponse", err)
		}
	}
}
