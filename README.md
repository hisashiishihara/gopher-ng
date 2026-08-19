# Gopher-NG

Gopher-NG is a minimal federated semantic discovery protocol for autonomous agents, inspired by [RFC 1436](https://www.rfc-editor.org/rfc/rfc1436).

The protocol uses the `gofer://` URI scheme and intentionally stays small: it defines discovery only. It does not define HTTP-style methods or headers, authentication, ontology schemas, databases, MCP execution, or UDP discovery.

Protocol notes are in [docs/protocol.md](docs/protocol.md). Go implementation entry points will live in `cmd/gng` and `cmd/gngd`, with shared protocol code in `internal/protocol`.
