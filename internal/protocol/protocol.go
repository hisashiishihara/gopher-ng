// Package protocol implements the Gopher-NG v0.0.1 Core wire format.
package protocol

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

var (
	// ErrInvalidSelector indicates a selector that does not meet Core syntax.
	ErrInvalidSelector = errors.New("invalid selector")
	// ErrInvalidRecord indicates a malformed Core record.
	ErrInvalidRecord = errors.New("invalid record")
	// ErrUnknownRecordType indicates a record type outside the Core vocabulary.
	ErrUnknownRecordType = errors.New("unknown record type")
	// ErrInvalidResponse indicates invalid response framing or record content.
	ErrInvalidResponse = errors.New("invalid response")
	// ErrIncompleteResponse indicates EOF before the response completion marker.
	ErrIncompleteResponse = errors.New("incomplete response")
	// ErrResponseTooLarge indicates a response exceeded the configured local limit.
	ErrResponseTooLarge = errors.New("response too large")
	// ErrSelectorTooLarge indicates a selector exceeded the server's local limit.
	ErrSelectorTooLarge = errors.New("selector too large")
)

const (
	// DefaultMaxResponseBytes is the reference client's inbound wire-byte budget.
	DefaultMaxResponseBytes int64 = 1 << 20
	// DefaultMaxSelectorBytes is the Go server's inbound selector wire-byte budget.
	DefaultMaxSelectorBytes int64 = 4 << 10
)

// RecordType identifies a Gopher-NG Core record type.
type RecordType string

const (
	TypeEntity RecordType = "ENTITY"
	TypeFact   RecordType = "FACT"
	TypeLink   RecordType = "LINK"
	TypeError  RecordType = "ERROR"
)

// Record is a Gopher-NG Core record. Fields excludes the record type.
type Record struct {
	Type   RecordType
	Fields []string
}

// ValidateSelector reports whether selector meets the Core selector syntax.
func ValidateSelector(selector string) error {
	if !utf8.ValidString(selector) || !strings.HasPrefix(selector, "/") {
		return ErrInvalidSelector
	}
	for _, r := range selector {
		if r <= 0x1f {
			return ErrInvalidSelector
		}
	}
	return nil
}

// WriteSelector validates selector and writes it followed by CRLF.
func WriteSelector(w io.Writer, selector string) error {
	if err := ValidateSelector(selector); err != nil {
		return err
	}
	return writeString(w, selector+"\r\n")
}

// ReadSelector reads and validates one CRLF-terminated selector line.
func ReadSelector(r io.Reader) (string, error) {
	reader := newBoundedLineReader(r, DefaultMaxSelectorBytes)
	line, err := reader.readLine()
	if err != nil {
		if errors.Is(err, errBoundedReadTooLarge) {
			return "", fmt.Errorf("%w: %w", ErrInvalidSelector, ErrSelectorTooLarge)
		}
		if errors.Is(err, io.EOF) {
			return "", ErrInvalidSelector
		}
		return "", err
	}
	if !strings.HasSuffix(line, "\r\n") {
		return "", ErrInvalidSelector
	}

	selector := strings.TrimSuffix(line, "\r\n")
	if err := ValidateSelector(selector); err != nil {
		return "", err
	}
	return selector, nil
}

// ParseRecord parses one CRLF-terminated Core record line.
func ParseRecord(line string) (Record, error) {
	if !strings.HasSuffix(line, "\r\n") || !utf8.ValidString(line) {
		return Record{}, ErrInvalidRecord
	}

	body := strings.TrimSuffix(line, "\r\n")
	if body == "" || strings.ContainsAny(body, "\r\n") {
		return Record{}, ErrInvalidRecord
	}

	parts := strings.Split(body, "\t")
	if !validRecordType(parts[0]) {
		return Record{}, ErrInvalidRecord
	}

	recordType := RecordType(parts[0])
	fieldCount, ok := requiredFieldCount(recordType)
	if !ok {
		return Record{}, fmt.Errorf("%w: %s", ErrUnknownRecordType, recordType)
	}
	if len(parts)-1 != fieldCount {
		return Record{}, ErrInvalidRecord
	}
	for _, field := range parts[1:] {
		if field == "" || !utf8.ValidString(field) || strings.ContainsAny(field, "\t\r\n") {
			return Record{}, ErrInvalidRecord
		}
	}
	if recordType == TypeLink {
		if _, err := ParseURI(parts[2]); err != nil {
			return Record{}, fmt.Errorf("%w: %w", ErrInvalidRecord, err)
		}
	}

	return Record{Type: recordType, Fields: append([]string(nil), parts[1:]...)}, nil
}

// EncodeRecord validates record and returns its CRLF-terminated wire form.
func EncodeRecord(record Record) (string, error) {
	if !validRecordType(string(record.Type)) {
		return "", ErrInvalidRecord
	}
	fieldCount, ok := requiredFieldCount(record.Type)
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrUnknownRecordType, record.Type)
	}
	if len(record.Fields) != fieldCount {
		return "", ErrInvalidRecord
	}
	for _, field := range record.Fields {
		if field == "" || !utf8.ValidString(field) || strings.ContainsAny(field, "\t\r\n") {
			return "", ErrInvalidRecord
		}
	}
	if record.Type == TypeLink {
		if _, err := ParseURI(record.Fields[1]); err != nil {
			return "", fmt.Errorf("%w: %w", ErrInvalidRecord, err)
		}
	}

	return string(record.Type) + "\t" + strings.Join(record.Fields, "\t") + "\r\n", nil
}

// ParseResponse parses a complete response ending in the required completion marker.
func ParseResponse(r io.Reader) ([]Record, error) {
	return parseResponse(&bufioLineReader{reader: bufio.NewReader(r)})
}

// ParseResponseWithLimit parses a complete response while enforcing a positive
// total wire-byte limit. The completion marker and all CRLF bytes count toward
// the limit.
func ParseResponseWithLimit(r io.Reader, maxBytes int64) ([]Record, error) {
	if maxBytes <= 0 {
		return nil, fmt.Errorf("response byte limit must be positive: %d", maxBytes)
	}
	return parseResponse(newBoundedLineReader(r, maxBytes))
}

type lineReader interface {
	readLine() (string, error)
}

type bufioLineReader struct {
	reader *bufio.Reader
}

func (r *bufioLineReader) readLine() (string, error) {
	return r.reader.ReadString('\n')
}

func parseResponse(reader lineReader) ([]Record, error) {
	var records []Record

	for {
		line, err := reader.readLine()
		if err != nil {
			if errors.Is(err, errBoundedReadTooLarge) {
				return nil, ErrResponseTooLarge
			}
			if errors.Is(err, io.EOF) {
				return nil, ErrIncompleteResponse
			}
			return nil, err
		}
		if !strings.HasSuffix(line, "\r\n") {
			return nil, ErrInvalidResponse
		}
		if line == ".\r\n" {
			return records, nil
		}

		record, err := ParseRecord(line)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidResponse, err)
		}
		if record.Type == TypeError && len(records) > 0 {
			return nil, ErrInvalidResponse
		}
		if record.Type != TypeError && len(records) == 1 && records[0].Type == TypeError {
			return nil, ErrInvalidResponse
		}
		records = append(records, record)
	}
}

var errBoundedReadTooLarge = errors.New("bounded read too large")

type boundedLineReader struct {
	reader   *bufio.Reader
	maxBytes int64
	consumed int64
}

func newBoundedLineReader(r io.Reader, maxBytes int64) *boundedLineReader {
	// Exposing max+1 bytes distinguishes overflow from EOF at the exact limit.
	limited := io.LimitReader(r, maxBytes+1)
	return &boundedLineReader{reader: bufio.NewReader(limited), maxBytes: maxBytes}
}

func (r *boundedLineReader) readLine() (string, error) {
	line, err := r.reader.ReadString('\n')
	r.consumed += int64(len(line))
	if r.consumed > r.maxBytes {
		return "", errBoundedReadTooLarge
	}
	return line, err
}

// WriteResponse writes records followed by the required completion marker.
func WriteResponse(w io.Writer, records []Record) error {
	if len(records) > 1 {
		for _, record := range records {
			if record.Type == TypeError {
				return ErrInvalidResponse
			}
		}
	}
	for _, record := range records {
		line, err := EncodeRecord(record)
		if err != nil {
			return err
		}
		if err := writeString(w, line); err != nil {
			return err
		}
	}
	return writeString(w, ".\r\n")
}

func requiredFieldCount(recordType RecordType) (int, bool) {
	switch recordType {
	case TypeEntity, TypeFact, TypeLink:
		return 2, true
	case TypeError:
		return 1, true
	default:
		return 0, false
	}
}

func validRecordType(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func writeString(w io.Writer, value string) error {
	n, err := io.WriteString(w, value)
	if err != nil {
		return err
	}
	if n != len(value) {
		return io.ErrShortWrite
	}
	return nil
}
