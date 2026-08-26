# Gopher-NG

> The Web evolved for humans. Gopher may have been waiting for machines.

**[GOFER's Ambition](docs/gofers-ambition.md)**

Gopher-NG is a minimal federated semantic discovery protocol for autonomous agents, inspired by [RFC 1436](https://www.rfc-editor.org/rfc/rfc1436).

> **Status: experimental.** This repository contains a minimal Gopher-NG v0.0.1 Core reference implementation.

It is not an attempt to recreate the 1990s Internet for nostalgia's sake. Gopher-NG starts from a different observation: modern software agents often need a small, traversable information space, while the modern Web frequently requires them to cross layers of presentation, application state, scripting, and framework machinery before reaching the information itself.

Gopher already had an interesting property: a server exposed a navigable information space through a deliberately small protocol. Gopher-NG asks what that idea looks like when the primary consumer may be an autonomous agent.

## Gopherism

Gopher-NG is guided by **Gopherism**: the design philosophy that a network protocol should expose information with the minimum machinery necessary for a human or program to retrieve, traverse, and understand it.

In practice:

- information comes before presentation;
- navigation comes before application behavior;
- one request should mean one thing;
- discovery is separate from execution;
- federation happens through traversable links rather than a mandatory central registry;
- semantic meaning belongs outside the transport core;
- protocol complexity must justify its existence.

A feature is not accepted merely because it is useful. The question is stricter:

> Does the protocol still work cleanly without it?

If the answer is yes, removal is preferred.

See [docs/gopherism.md](docs/gopherism.md) for the longer design rationale.

## Why now?

Agent protocols such as MCP have made an old problem visible in a new form: machines need structured ways to discover information and capabilities.

Gopher-NG does **not** attempt to replace MCP, HTTP, databases, ontologies, or tool-execution protocols. Its scope is intentionally narrower. It provides a small discovery and traversal layer that other systems may build on or consume.

The goal is not to make Gopher capable of everything the Web can do.

The goal is to preserve what made Gopher interesting in the first place: a client asks for a location, receives a finite description of what is there and where it can go next, and does not need an application framework to understand the answer.

## Core scope

Gopher-NG uses the `gofer://` URI scheme and intentionally defines discovery only.

The Core protocol does not define HTTP-style methods or headers, sessions, cookies, authentication, ontology schemas, databases, mutation, subscriptions, cross-protocol service advertisement, MCP execution, tool invocation, or UDP discovery.

Its transaction model is deliberately small:

```text
one connection
one selector
one complete response
explicit response completion
connection close
```

Protocol details are in [docs/protocol.md](docs/protocol.md).

Go implementation entry points live in `cmd/gng` and `cmd/gngd`, with shared protocol code in `internal/protocol`.

## Quick start

Clone the repository:

```text
git clone https://github.com/hisashiishihara/gopher-ng.git
cd gopher-ng
```

In one terminal, start the reference daemon:

```text
go run ./cmd/gngd -listen 127.0.0.1:7070
```

In another terminal, retrieve the root resource:

```text
go run ./cmd/gng gofer://127.0.0.1:7070/
```

```text
ENTITY	gopher-ng:Server	gopher-ng:root
```

An unknown valid selector returns a Core error response:

```text
go run ./cmd/gng gofer://127.0.0.1:7070/missing
```

```text
ERROR	NOT_FOUND
```

`127.0.0.1:7070` is a reference-implementation/development default. Gopher-NG v0.0.1 defines no protocol default port.

The reference implementation supports the minimal v0.0.1 transaction: an absolute `gofer://` URI with explicit host and port, one selector, one Core response with an explicit terminator, and connection close. It uses plain TCP. It does not implement TLS, authentication, automatic traversal, recursive LINK following, service execution, MCP, ontology interpretation, persistence, or discovery registries. The current reference client does not automatically follow LINK records. As a client safety policy, it limits a complete Core response to 1 MiB (1,048,576 wire bytes), including record framing and the completion marker; Gopher-NG v0.0.1 itself defines no protocol response-size limit.

For an optional private-network deployment example, see [Gopher-NG over Tailscale](docs/tailscale.md).

## License

Licensed under the [BSD 3-Clause License](LICENSE).

## Design direction

Gopher-NG should remain recognizable as Gopher in spirit, not merely in name. Compatibility with the historical protocol is less important than preserving its essential character: small requests, finite responses, traversable information, independent servers, and very little machinery between a client and the thing it wants to know.

Gopher is not interesting because it is old.

It is interesting because the Web became complicated enough for its old constraints to become useful again.
