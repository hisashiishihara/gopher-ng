// Command server-b serves the second half of the Gopher-NG federation example.
package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/hisashiishihara/gopher-ng/internal/protocol"
	"github.com/hisashiishihara/gopher-ng/internal/server"
)

func main() {
	listenAddress := flag.String("listen", "127.0.0.1:7071", "TCP address to listen on")
	flag.Parse()

	if err := run(*listenAddress); err != nil {
		fmt.Fprintf(os.Stderr, "server-b: %v\n", err)
		os.Exit(1)
	}
}

func run(listenAddress string) error {
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return err
	}
	return serve(listener, handler)
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

func handler(selector string) ([]protocol.Record, error) {
	if selector == "/resource" {
		return []protocol.Record{
			{Type: protocol.TypeEntity, Fields: []string{"example:Resource", "example:federation-target"}},
			{Type: protocol.TypeFact, Fields: []string{"example:message", "Hello from server B"}},
		}, nil
	}
	return []protocol.Record{{Type: protocol.TypeError, Fields: []string{"NOT_FOUND"}}}, nil
}
