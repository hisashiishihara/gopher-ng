# Gopher-NG Protocol

## Status

This is the initial, deliberately minimal design note for Gopher-NG. The protocol is not implemented yet.

## Purpose

Gopher-NG lets autonomous agents discover semantic resources exposed by independently operated servers. Discovery is federated: each server publishes its own resource descriptions and may link to other `gofer://` locations.

## URI scheme

Gopher-NG locations use the `gofer` scheme:

```
gofer://authority/path
```

The authority identifies a server. The path identifies a discoverable location on that server. The precise wire representation and response format remain intentionally unspecified at this stage.

## Minimal boundary

Gopher-NG defines discovery and links between discoverable locations. It does not define:

- HTTP-like methods, headers, status codes, or content negotiation
- Authentication or authorization
- Ontology or schema standards
- Databases or storage requirements
- MCP execution or tool invocation
- UDP discovery

Servers and clients may evolve within those boundaries without treating any of the excluded concerns as part of this protocol.
