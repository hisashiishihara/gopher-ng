# Two-server federation proof

This example demonstrates the smallest useful federation claim in Gopher-NG v0.0.1:

```text
independent Server A
        |
        | LINK
        v
independent Server B
```

No Core extension is involved. Server A returns an ordinary `LINK` record containing an absolute `gofer://` URI, and the client can use that URI to retrieve a semantic resource from Server B.

The example deliberately keeps traversal manual so the protocol boundary remains visible.

## Run it

In Terminal 1, start Server B:

```sh
go run ./examples/federation/server-b \
  -listen 127.0.0.1:7071
```

In Terminal 2, start Server A and configure its target URI:

```sh
go run ./examples/federation/server-a \
  -listen 127.0.0.1:7070 \
  -target gofer://127.0.0.1:7071/resource
```

In Terminal 3, retrieve Server A:

```sh
go run ./cmd/gng gofer://127.0.0.1:7070/
```

```text
ENTITY	example:Directory	example:federation-root
LINK	related	gofer://127.0.0.1:7071/resource
```

The response describes a local entity and exposes a semantic relation to a resource on another independently running server.

Follow the returned URI:

```sh
go run ./cmd/gng gofer://127.0.0.1:7071/resource
```

```text
ENTITY	example:Resource	example:federation-target
FACT	example:message	Hello from server B
```

That is the complete proof: one server can expose a semantic link to another server, and a client can traverse the distributed information space using only Gopher-NG Core records.

## What this proves

- federation does not require a central registry;
- `LINK` already represents cross-server traversal in v0.0.1;
- independently operated servers can form one traversable semantic space;
- no HTTP server or Web application stack is needed for this exchange;
- discovery remains separate from execution.

## What this does not add

This example intentionally does **not** add:

- automatic or recursive `LINK` following;
- a crawler or graph walker;
- service execution;
- MCP integration;
- authentication or TLS;
- new Core record types.

Those would be separate concerns. The point of this example is to show that basic federation already follows from the minimal Core protocol.

Whitespace on the wire is TAB-delimited according to [`docs/protocol.md`](../../docs/protocol.md). The `127.0.0.1:7070` and `127.0.0.1:7071` values are development addresses only; Gopher-NG v0.0.1 defines no default port.
