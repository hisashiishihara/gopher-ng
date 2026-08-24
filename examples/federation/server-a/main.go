// Command server-a serves the first half of the Gopher-NG federation example.
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
	listenAddress := flag.String("listen", "127.0.0.1:7070", "TCP address to listen on")
	target := flag.String("target", "gofer://127.0.0.1:7071/resource", "Gopher-NG URI for the related resource")
	flag.Parse()

	if err := run(*listenAddress, *target); err != nil {
		fmt.Fprintf(os.Stderr, "server-a: %v\n", err)
		os.Exit(1)
	}
}

func run(listenAddress, target string) error {
	if _, err := protocol.ParseURI(target); err != nil {
		return err
	}
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return err
	}
	return serve(listener, handler(target))
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

func handler(target string) server.Handler {
	return func(selector string) ([]protocol.Record, error) {
		if selector == "/" {
			return []protocol.Record{
				{Type: protocol.TypeEntity, Fields: []string{"example:Directory", "example:federation-root"}},
				{Type: protocol.TypeLink, Fields: []string{"related", target}},
			}, nil
		}
		return []protocol.Record{{Type: protocol.TypeError, Fields: []string{"NOT_FOUND"}}}, nil
	}
}
