// Command gngd serves the minimal Gopher-NG Core TCP daemon.
package main

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/hisashiishihara/gopher-ng/internal/protocol"
	"github.com/hisashiishihara/gopher-ng/internal/server"
)

func main() {
	listenAddress := flag.String("listen", "127.0.0.1:7070", "TCP address to listen on")
	flag.Parse()

	listener, err := net.Listen("tcp", *listenAddress)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gngd: listen: %v\n", err)
		os.Exit(1)
	}
	if err := serve(listener, demoHandler); err != nil && !errors.Is(err, net.ErrClosed) {
		fmt.Fprintf(os.Stderr, "gngd: serve: %v\n", err)
		os.Exit(1)
	}
}

func serve(listener net.Listener, handler server.Handler) error {
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go func() {
			_ = server.ServeConn(conn, handler)
		}()
	}
}

func demoHandler(selector string) ([]protocol.Record, error) {
	if selector == "/" {
		return []protocol.Record{{
			Type:   protocol.TypeEntity,
			Fields: []string{"gopher-ng:Server", "gopher-ng:root"},
		}}, nil
	}
	return []protocol.Record{{
		Type:   protocol.TypeError,
		Fields: []string{"NOT_FOUND"},
	}}, nil
}
