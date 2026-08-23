// Package server provides the Gopher-NG Core server transaction primitive.
package server

import (
	"errors"
	"net"

	"github.com/hisashiishihara/gopher-ng/internal/protocol"
)

// Handler produces the Core response records for one selector.
type Handler func(selector string) ([]protocol.Record, error)

// ServeConn serves one Gopher-NG Core transaction and closes conn before returning.
func ServeConn(conn net.Conn, handler Handler) error {
	defer conn.Close()

	selector, err := protocol.ReadSelector(conn)
	if err != nil {
		if errors.Is(err, protocol.ErrInvalidSelector) {
			_ = protocol.WriteResponse(conn, []protocol.Record{{
				Type:   protocol.TypeError,
				Fields: []string{"BAD_SELECTOR"},
			}})
		}
		return err
	}
	if handler == nil {
		err := errors.New("nil handler")
		_ = writeTemporaryFailure(conn)
		return err
	}

	records, err := handler(selector)
	if err != nil {
		_ = writeTemporaryFailure(conn)
		return err
	}
	return protocol.WriteResponse(conn, records)
}

func writeTemporaryFailure(conn net.Conn) error {
	return protocol.WriteResponse(conn, []protocol.Record{{
		Type:   protocol.TypeError,
		Fields: []string{"TEMPORARY_FAILURE"},
	}})
}
