# Federation example

This example runs two independent Gopher-NG servers. Server A exposes a
`LINK` to a resource on Server B.

In Terminal 1, start Server B:

```text
go run ./examples/federation/server-b \
    -listen 127.0.0.1:7071
```

In Terminal 2, start Server A with its target URI:

```text
go run ./examples/federation/server-a \
    -listen 127.0.0.1:7070 \
    -target gofer://127.0.0.1:7071/resource
```

In Terminal 3, retrieve Server A:

```text
go run ./cmd/gng gofer://127.0.0.1:7070/
```

```text
ENTITY	example:Directory	example:federation-root
LINK	related	gofer://127.0.0.1:7071/resource
```

Then manually follow the LINK:

```text
go run ./cmd/gng gofer://127.0.0.1:7071/resource
```

```text
ENTITY	example:Resource	example:federation-target
FACT	example:message	Hello from server B
```

Whitespace on the wire is TAB-delimited according to the existing protocol.
This example deliberately does not add automatic LINK traversal. Federation is
already represented by the existing LINK record.

No HTTP server, Web application stack, central registry, or service execution
protocol is involved. The `127.0.0.1:7070` and `127.0.0.1:7071` values are
example development addresses, not Gopher-NG protocol default ports.
