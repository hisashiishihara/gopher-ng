# Design Decision 0001: `SERVICE` is not a Core record

## Status

Accepted for Gopher-NG v0.0.1.

## Context

An early Gopher-NG Core draft defined a `SERVICE` record:

```text
SERVICE<TAB>relation<TAB>uri
```

Its purpose was to advertise a service associated with the current entity without invoking that service. For example:

```text
SERVICE	pet:medical-record	https://vet.example/records/123
```

At first glance this appears useful for autonomous agents: discovery could reveal not only facts and Gopher-NG links, but also external endpoints that might later be used by another protocol.

However, Gopher-NG Core is intentionally a discovery and traversal protocol, not a general service-description or execution protocol. Under Gopherism, usefulness by itself is not sufficient reason for a feature to exist in the Core.

The relevant question is:

> Does the protocol still work cleanly without it?

For `SERVICE`, the answer is yes.

## Decision

`SERVICE` is not a Gopher-NG Core record type.

Core keeps four record types:

```text
ENTITY  what this resource is
FACT    what is known about it
LINK    where Gopher-NG can traverse next
ERROR   why the transaction failed
```

A URI belonging to another protocol may be represented as ontology-defined data in a `FACT`:

```text
FACT	pet:medical-record	https://vet.example/records/123
```

Gopher-NG Core treats that value as opaque data. It does not assign permission, traversal, invocation, authentication, method, capability, or execution semantics to it.

If the destination is itself a Gopher-NG resource and is intended to be traversed as part of the Gopher-NG information space, `LINK` is the appropriate record:

```text
LINK	pet:veterinarian	gofer://vet.example:7070/pet/123
```

## Why not keep `SERVICE` as a convenience?

Because convenience would create a new semantic category in the transport core without adding necessary expressive power.

Once `SERVICE` exists, Core must answer questions it otherwise does not need to own:

- What qualifies as a service rather than an ordinary URI-valued fact?
- Does advertising a service imply that a client may dereference or invoke it?
- Does Core need to distinguish APIs, MCP servers, HTTP resources, RPC endpoints, message queues, or other schemes?
- How are authentication and authorization associated with the advertised endpoint?
- Does a client treat `SERVICE` differently from arbitrary ontology data?
- Where is the boundary between discovery and execution?

Any normative answer increases protocol surface area. Refusing to answer them keeps those concerns in the protocol or application layer that actually owns them.

## Gopherism rationale

Removing `SERVICE` follows several Gopherism principles:

1. **Discovery before execution** — Core may describe information without acquiring execution semantics.
2. **Semantics outside the transport core** — whether an external URI denotes a medical-record API, a document, or a tool is domain knowledge.
3. **One record type should have one irreducible purpose** — `ENTITY`, `FACT`, `LINK`, and `ERROR` each remain distinct.
4. **Complexity must justify its existence** — `SERVICE` can be removed without losing the ability to describe the same information.

This is not minimalism for aesthetics alone. It is a boundary decision: Gopher-NG should know how to describe and traverse its own information space, and should avoid becoming a registry for the semantics of other protocols.

## Consequences

The Core vocabulary is smaller and the distinction between traversal and arbitrary data is sharper.

Clients MUST NOT infer that a URI contained in a `FACT` is safe or permitted to dereference or invoke. Such behavior belongs to application policy or to another protocol.

A future extension may define service or capability advertisement if a concrete interoperability requirement proves that `FACT` plus `LINK` is insufficient. Such an extension should justify why the semantics must be standardized at the Gopher-NG layer rather than in an ontology or execution protocol.

Until then, omission is intentional.
