# GOFER's Ambition

## Why Gopher?

Gopher was once a widely used way to publish and navigate information
on the Internet.

The Web eventually proved better suited to richer documents,
presentation, interaction, and applications.

Machine-oriented clients may have somewhat different requirements.

## The hypothesis

An autonomous agent often does not require the presentation machinery
expected by a human-facing Web application.

In many cases, it may be enough to expose:

```text
location
    ↓
semantic description
    ↓
traversable links
```

Gopher-NG explores whether this smaller model is useful for machine-oriented
information discovery.

## Scope

Gopher-NG is not intended to replace the Web, HTTP, MCP, or application
protocols.

Its scope is deliberately narrower:

> Make distributed information simple to expose,
> simple to discover,
> and simple for machines to traverse.

## Influences

Gopher-NG borrows a few ideas from earlier systems.

From Gopher:

- simple transactions
- navigable information spaces
- independent servers
- federation through links

From Unix:

- small mechanisms
- explicit composition
- avoiding unnecessary machinery

For modern machine clients:

- structured semantic records
- machine-oriented traversal
- separation of discovery from execution

## Differences from historical Gopher

Gopher-NG is inspired by Gopher, but is not intended to be wire-compatible
with the historical protocol.

It uses the `gofer://` URI scheme and defines its own small set of Core
record types.

Independent implementations are encouraged.

The reference implementation is released under the BSD 3-Clause License.

## A small success criterion

Gopher-NG does not need broad adoption to test the idea.

If two independently operated servers can expose semantic resources,
and a client can traverse between them without requiring a Web application
stack, the basic model is already worth evaluating.

**Status: demonstrated.** The [two-server federation proof](../examples/federation/README.md)
shows independently running servers exposing semantic resources and a client
manually following an ordinary `LINK` between them. The [separate-host
validation](tailscale.md) demonstrates the same traversal between independently
operated Linux hosts. Together, this evidence satisfies the small technical
criterion without a Web application stack or a Core extension.

This is evidence for discovery and manual traversal, not execution. It does not
demonstrate automatic or recursive traversal, service execution, TLS or
authentication in Core, a central registry, or protocol changes. With the small
technical criterion met, the remaining question is how useful this model is in
real-world evaluation.
