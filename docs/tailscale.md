# Gopher-NG over Tailscale

Gopher-NG does not require Tailscale.

However, a private overlay network such as Tailscale is a useful deployment
option for Gopher-NG's plain TCP transport. It can provide private reachability
between independently operated hosts without exposing the Gopher-NG listening
port to the public Internet.

The roles remain separate:

```text
Gopher-NG
    discovery and traversal

TCP
    transport

Tailscale
    private network reachability and encrypted transport between hosts
```

Tailscale is not part of the Gopher-NG Core protocol.

## Validation

On 2026-08-24, Gopher-NG federation was validated between two independent
Linux hosts on the same tailnet:

```text
Host A
    |
    | LINK
    v
Host B
```

Server B was started on Host B:

```text
go run ./examples/federation/server-b \
    -listen 0.0.0.0:7071
```

Host A could retrieve Server B directly through the Tailscale network using a
MagicDNS hostname:

```text
go run ./cmd/gng \
    gofer://host-b:7071/resource
```

The response was:

```text
ENTITY	example:Resource	example:federation-target
FACT	example:message	Hello from server B
```

MagicDNS name resolution was sufficient; no public Gopher-NG port exposure,
HTTP server, reverse proxy, central registry, or service-execution protocol was
required.

The same resource was then exposed as a `LINK` by Server A on Host A:

```text
ENTITY	example:Directory	example:federation-root
LINK	related	gofer://host-b:7071/resource
```

The client manually followed that LINK and retrieved Server B successfully.

This demonstrated federation across separate hosts using only the existing
Gopher-NG `LINK` semantics. No Core protocol change was necessary.

## Security boundary

Gopher-NG v0.0.1 uses plain TCP and does not define transport encryption or
peer authentication.

When deployed over Tailscale, those network-security concerns are handled by
the overlay network rather than by the Gopher-NG protocol itself. This is an
example deployment architecture, not a requirement of Gopher-NG.
