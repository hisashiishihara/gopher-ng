# Gopherism

## Definition

**Gopherism** is the design philosophy that a network protocol should expose information with the minimum machinery necessary for a human or program to retrieve, traverse, and understand it.

Gopherism is inspired by the original Gopher protocol, but it is not nostalgia and it is not strict historical compatibility. It is an attempt to preserve a useful design instinct: information should be reachable without forcing the client to participate in unnecessary application machinery.

A Gopherist design therefore treats complexity as something that must earn its place.

## Principles

### 1. Information before presentation

A protocol should describe what information exists and how to reach it before prescribing how it should look.

Presentation belongs at the edge. The transport should not require a browser engine, a document object model, client-side application state, or a rendering framework in order to expose the underlying information.

### 2. Navigation before application

A server should be able to expose an information space without first becoming an application platform.

Gopherism prefers traversable resources and links over remote application behavior. A client should be able to discover what exists, inspect it, and decide what to do next.

### 3. One request should mean one thing

A transaction should be easy to describe.

For Gopher-NG Core, one connection carries one selector and produces one complete response. The protocol does not multiplex unrelated operations into the same request or grow a general-purpose method system merely because it could.

### 4. Discovery before execution

Knowing that something exists is different from invoking it.

Gopher-NG Core defines discovery and traversal. It intentionally does not define tool execution, mutation, credentials, subscriptions, workflows, or agent actions.

Other protocols may perform those jobs. Gopher-NG should not absorb them merely to become more capable.

### 5. Semantics outside the transport core

The protocol should carry enough structure to traverse information without becoming the owner of every domain model.

Gopher-NG therefore treats ontology terms as opaque identifiers. A veterinary ontology, a construction ontology, and a music ontology may all coexist without requiring changes to the transport protocol.

### 6. Federation by links

Independent servers should be able to participate without a mandatory central registry.

A link is sufficient to extend the information space across administrative boundaries. Central indexes may exist, but they are conveniences rather than architectural requirements.

### 7. Human-readable enough, machine-readable by design

The protocol should remain simple enough that its wire representation can be inspected without specialized tooling, while still being deterministic enough for software agents to parse reliably.

The goal is not to optimize exclusively for humans or exclusively for machines. It is to avoid unnecessary layers between either kind of client and the information being requested.

### 8. Complexity must justify its existence

The default answer to a proposed feature is not "yes, that could be useful."

The relevant question is:

> Does the protocol still work cleanly without it?

If the answer is yes, omission is preferred.

This is the central discipline of Gopherism.

## Why this matters again

The original Gopher protocol lost to the Web for understandable reasons. The Web became a richer platform for human-facing documents and applications.

That success also accumulated machinery: scripting, application state, frameworks, client-side rendering, authentication flows, content negotiation, APIs behind APIs, and presentation layers that often obscure the small piece of information a machine actually needs.

Autonomous agents have made that mismatch more visible. An agent frequently wants something much simpler than a web application: enumerate resources, retrieve a description, follow a relation, and repeat.

This does not make Gopher superior to the Web. It makes some of Gopher's old constraints newly interesting.

Gopher is not interesting because it is old.

It is interesting because the Web became complicated enough for its old constraints to become useful again.

## Relationship to MCP

MCP and Gopher-NG solve different problems.

MCP connects models to external context, resources, prompts, and tools. Gopher-NG Core is deliberately narrower: it exposes a federated semantic information space and defines how an agent traverses it.

A future bridge may expose Gopher-NG resources through MCP, or allow an MCP implementation to use Gopher-NG for discovery. Such integration does not belong in the Gopher-NG Core protocol.

The separation is intentional:

```text
Gopher-NG: discover and traverse
MCP/tool protocols: expose context and perform actions
applications: decide policy
```

## What Gopherism is not

Gopherism does not mean that every system must be tiny, text-only, or feature-poor.

It does not prohibit databases, authentication, rich user interfaces, search engines, ontologies, or agent tooling.

It says only that those concerns should not be pulled into a lower layer unless that layer genuinely requires them.

The aim is not minimalism for its own sake.

The aim is to keep each layer responsible for as little as it can get away with — and no less than it must.
