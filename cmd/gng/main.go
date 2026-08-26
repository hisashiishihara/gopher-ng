// Command gng retrieves one Gopher-NG Core resource over TCP.
package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/hisashiishihara/gopher-ng/internal/protocol"
)

const maxResponseSize int64 = 1 << 20

var errResponseTooLarge = errors.New("response too large")

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "gng: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: gng gofer://host:port/path")
	}

	records, err := fetch(args[0])
	if err != nil {
		return err
	}
	return printRecords(stdout, records)
}

func fetch(rawURI string) ([]protocol.Record, error) {
	uri, err := protocol.ParseURI(rawURI)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", dialAddress(uri))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := protocol.WriteSelector(conn, uri.Selector); err != nil {
		return nil, err
	}
	return parseResponse(conn)
}

func parseResponse(r io.Reader) ([]protocol.Record, error) {
	return protocol.ParseResponse(&boundedResponseReader{r: r, remaining: maxResponseSize})
}

type boundedResponseReader struct {
	r         io.Reader
	remaining int64
}

func (r *boundedResponseReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.remaining == 0 {
		var extra [1]byte
		n, err := r.r.Read(extra[:])
		if n > 0 {
			return 0, errResponseTooLarge
		}
		return 0, err
	}
	if int64(len(p)) > r.remaining {
		p = p[:r.remaining]
	}
	n, err := r.r.Read(p)
	r.remaining -= int64(n)
	return n, err
}

func dialAddress(uri protocol.URI) string {
	return net.JoinHostPort(uri.Host, strconv.FormatUint(uint64(uri.Port), 10))
}

func printRecords(w io.Writer, records []protocol.Record) error {
	for _, record := range records {
		line, err := protocol.EncodeRecord(record)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintln(w, strings.TrimSuffix(line, "\r\n")); err != nil {
			return err
		}
	}
	return nil
}
