# Gopher-NG

> The Web evolved for humans. Gopher may have been waiting for machines.

Gopher-NG is a minimal federated semantic discovery protocol for autonomous agents, inspired by [RFC 1436](https://www.rfc-editor.org/rfc/rfc1436).

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

The Core protocol does not define HTTP-style methods or headers, sessions, cookies, authentication, ontology schemas, databases, mutation, subscriptions, MCP execution, tool invocation, or UDP discovery.

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

## Design direction

Gopher-NG should remain recognizable as Gopher in spirit, not merely in name. Compatibility with the historical protocol is less important than preserving its essential character: small requests, finite responses, traversable information, independent servers, and very little machinery between a client and the thing it wants to know.

Gopher is not interesting because it is old.

It is interesting because the Web became complicated enough for its old constraints to become useful again.
