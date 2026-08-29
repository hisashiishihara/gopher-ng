package protocol

import (
	"bytes"
	"errors"
	"io"
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
		{"FACT HTTP URI", "FACT\tpet:medical-record\thttps://vet.example/records/123\r\n", Record{Type: TypeFact, Fields: []string{"pet:medical-record", "https://vet.example/records/123"}}},
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
		"LINK\tpet:veterinarian\thttp://vet.example:7070/pet/123\r\n",
		"LINK\tpet:veterinarian\tgofer://vet.example/pet/123\r\n",
		"LINK\tpet:veterinarian\tgofer://vet.example:7070/pet/%ZZ\r\n",
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

func TestEncodeRecordValidatesLinkURI(t *testing.T) {
	invalid := Record{Type: TypeLink, Fields: []string{"pet:veterinarian", "http://vet.example:7070/pet/123"}}
	if _, err := EncodeRecord(invalid); !errors.Is(err, ErrInvalidRecord) || !errors.Is(err, ErrInvalidURI) {
		t.Fatalf("EncodeRecord() error = %v, want ErrInvalidRecord and ErrInvalidURI", err)
	}

	fact := Record{Type: TypeFact, Fields: []string{"pet:medical-record", "https://vet.example/records/123"}}
	if _, err := EncodeRecord(fact); err != nil {
		t.Fatalf("EncodeRecord(FACT) error = %v", err)
	}
}

func TestParseRecordInvalidLinkErrorChain(t *testing.T) {
	_, err := ParseRecord("LINK\tpet:veterinarian\thttp://vet.example:7070/pet/123\r\n")
	if !errors.Is(err, ErrInvalidRecord) || !errors.Is(err, ErrInvalidURI) {
		t.Fatalf("ParseRecord() error = %v, want ErrInvalidRecord and ErrInvalidURI", err)
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

func TestReadSelector(t *testing.T) {
	tests := []struct {
		name string
		wire string
		want string
	}{
		{"root", "/\r\n", "/"},
		{"path", "/pet/123\r\n", "/pet/123"},
		{"UTF-8", "/pet/猫\r\n", "/pet/猫"},
		{"spaces", "/pet/Moko Chan\r\n", "/pet/Moko Chan"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ReadSelector(strings.NewReader(test.wire))
			if err != nil {
				t.Fatalf("ReadSelector() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("ReadSelector() = %q, want %q", got, test.want)
			}
		})
	}
}

func TestReadSelectorRejectsInvalidInput(t *testing.T) {
	tests := []string{
		"pet/123\r\n",
		"/pet/123\n",
		"/pet/123",
		"/pet\t123\r\n",
		"/pet\r123\r\n",
		"/pet\n123\r\n",
		"/pet\x00123\r\n",
		"/pet/\xff\r\n",
	}

	for _, wire := range tests {
		if _, err := ReadSelector(strings.NewReader(wire)); !errors.Is(err, ErrInvalidSelector) {
			t.Errorf("ReadSelector(%q) error = %v, want ErrInvalidSelector", wire, err)
		}
	}
}

func TestReadSelectorLimit(t *testing.T) {
	exact := "/" + strings.Repeat("a", int(DefaultMaxSelectorBytes)-3) + "\r\n"
	selector, err := ReadSelector(strings.NewReader(exact))
	if err != nil || selector != strings.TrimSuffix(exact, "\r\n") {
		t.Fatalf("ReadSelector(exact) = %q, %v", selector, err)
	}

	for _, wire := range []string{
		"/" + strings.Repeat("a", int(DefaultMaxSelectorBytes)-2) + "\r\n",
		"/" + strings.Repeat("a", int(DefaultMaxSelectorBytes)),
	} {
		_, err := ReadSelector(strings.NewReader(wire))
		if !errors.Is(err, ErrSelectorTooLarge) || !errors.Is(err, ErrInvalidSelector) {
			t.Fatalf("ReadSelector(oversize) error = %v", err)
		}
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

func TestParseResponseWithLimitBoundaries(t *testing.T) {
	record := "FACT\tpet:name\tMoko\r\n"
	response := record + ".\r\n"

	records, err := ParseResponseWithLimit(strings.NewReader(response), int64(len(response)))
	if err != nil || len(records) != 1 {
		t.Fatalf("exact limit = %#v, %v", records, err)
	}
	if records, err := ParseResponseWithLimit(strings.NewReader(".\r\n"), 3); err != nil || len(records) != 0 {
		t.Fatalf("empty exact minimum = %#v, %v", records, err)
	}

	for _, test := range []struct {
		name  string
		wire  string
		limit int64
		want  error
	}{
		{"one byte over", response, int64(len(response) - 1), ErrResponseTooLarge},
		{"terminator crossing boundary", response, int64(len(response) - 2), ErrResponseTooLarge},
		{"missing below limit", record, int64(len(record) + 1), ErrIncompleteResponse},
		{"missing exactly at limit", record, int64(len(record)), ErrIncompleteResponse},
		{"max plus one unterminated", strings.Repeat("x", 9), 8, ErrResponseTooLarge},
	} {
		t.Run(test.name, func(t *testing.T) {
			records, err := ParseResponseWithLimit(strings.NewReader(test.wire), test.limit)
			if !errors.Is(err, test.want) || records != nil {
				t.Fatalf("got %#v, %v; want nil, %v", records, err, test.want)
			}
		})
	}

	for _, limit := range []int64{0, -1} {
		if records, err := ParseResponseWithLimit(strings.NewReader(".\r\n"), limit); err == nil || records != nil {
			t.Fatalf("limit %d = %#v, %v; want programmer error", limit, records, err)
		}
	}
}

func TestParseResponseWithLimitAggregateAndNoPartialRecords(t *testing.T) {
	line := "FACT\tk\tv\r\n"
	wire := strings.Repeat(line, 64) + ".\r\n"
	if records, err := ParseResponseWithLimit(strings.NewReader(wire), int64(len(wire))); err != nil || len(records) != 64 {
		t.Fatalf("aggregate exact = %d records, %v", len(records), err)
	}
	if records, err := ParseResponseWithLimit(strings.NewReader(wire), int64(len(wire)-1)); !errors.Is(err, ErrResponseTooLarge) || records != nil {
		t.Fatalf("aggregate over = %#v, %v", records, err)
	}

	malformed := line + "UNKNOWN\tx\ty\r\n" + strings.Repeat("z", 100)
	records, err := ParseResponseWithLimit(strings.NewReader(malformed), 32)
	if records != nil || !errors.Is(err, ErrInvalidResponse) || !errors.Is(err, ErrUnknownRecordType) {
		t.Fatalf("malformed prefix = %#v, %v", records, err)
	}
}

func TestParseResponseWithLimitLongLineReadsAtMostMaxPlusOne(t *testing.T) {
	r := &countingReader{reader: strings.NewReader(strings.Repeat("x", 1<<20))}
	records, err := ParseResponseWithLimit(r, 128)
	if records != nil || !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("long line = %#v, %v", records, err)
	}
	if r.bytesRead != 129 {
		t.Fatalf("bytes read = %d, want 129", r.bytesRead)
	}
}

type countingReader struct {
	reader    io.Reader
	bytesRead int
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.bytesRead += n
	return n, err
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
