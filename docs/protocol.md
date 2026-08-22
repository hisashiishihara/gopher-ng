# Gopher-NG v0.0.1 Core Protocol

## Status and scope

This document specifies the Gopher-NG v0.0.1 Core Protocol. Gopher-NG is a minimal federated semantic discovery and traversal protocol for autonomous agents, inspired by RFC 1436.

The protocol is guided by the non-normative design philosophy described in [Gopherism](gopherism.md). That document explains why the Core intentionally keeps discovery, traversal, execution, ontology semantics, and application behavior separate.

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in [RFC 2119](https://www.rfc-editor.org/rfc/rfc2119) and [RFC 8174](https://www.rfc-editor.org/rfc/rfc8174) when, and only when, they appear in all capitals.

Its transaction model is deliberately small:

```text
one connection
one selector
one complete response
explicit response completion
connection close
```

Transactions are stateless. Servers are independently operated and federation occurs through traversable links. Core defines discovery and traversal only; it does not define execution.

Core does not define HTTP compatibility or HTTP-like methods, headers, sessions, cookies, content negotiation, chunked encoding, or keep-alive. It also does not define authentication, ontology schemas, databases, query or search syntax, mutation operations, subscriptions, MCP integration, tool invocation, or UDP discovery.

## Transport

The initial transport stack is:

```text
Gopher-NG
TLS
TCP
IP
```

For each transaction, a client:

1. connects to the server;
2. completes a TLS handshake;
3. sends one selector;
4. receives one complete response and its explicit terminator; and
5. closes the connection.

The server closes the connection after sending the complete response. Neither side reuses a connection for another request. TLS is part of the protocol design; Gopher-NG defines no application-layer encryption.

## URIs and selectors

Gopher-NG locations use the `gofer` scheme:

```text
gofer://authority:port/path
```

The authority and explicit port identify the server, and the path corresponds to the selector sent to that server. Core does not define a default port; a `gofer://` URI MUST contain an explicit port. For example, resolving:

```text
gofer://example.org:7070/pet/123
```

conceptually sends the following request after TLS is established:

```text
/pet/123\r\n
```

A selector is UTF-8 text that MUST begin with `/` and MUST NOT contain CR, LF, TAB, or other C0 control characters. The root selector is `/`. A selector is sent exactly as:

```text
<selector>\r\n
```

There are no request headers, methods, request bodies, or parameters defined by Core.

## Response framing

A response is a sequence of zero or more record lines followed by exactly one completion marker:

```text
record\r\n
record\r\n
.\r\n
```

Each record is UTF-8 text with this general form:

```text
TYPE<TAB>field<TAB>field...\r\n
```

`TYPE` is an uppercase ASCII token. Fields are non-empty UTF-8 text and MUST NOT contain raw TAB, CR, or LF characters. Core defines no general-purpose escaping language.

The completion marker is the line consisting of a single period followed by CRLF:

```text
.\r\n
```

It is semantically required, not merely a framing convenience. A client that observes connection closure before receiving this marker MUST treat the response as incomplete and MUST NOT treat any received records as a complete authoritative response.

## Core record types

All Core records below have exactly the shown fields. A successful response has one or more Core records, or may be empty, followed by the completion marker. It has no HTTP-style success status line.

### ENTITY

`ENTITY` describes the semantic entity represented by the current selector.

```text
ENTITY<TAB>type<TAB>identifier
```

Example:

```text
ENTITY	pet:Pet	pet:123
```

### FACT

`FACT` states a fact about the current entity.

```text
FACT<TAB>predicate<TAB>value
```

Examples:

```text
FACT	pet:name	Moko
FACT	pet:species	pet:Dog
```

Predicates and values are opaque to Gopher-NG Core.

### LINK

`LINK` defines a traversable semantic relation to another Gopher-NG resource.

```text
LINK<TAB>relation<TAB>gofer-uri
```

Example:

```text
LINK	pet:veterinarian	gofer://vet.example:7070/pet/123
```

The `gofer-uri` field MUST be an absolute `gofer://` URI with an explicit port. `LINK` is the primary federation and traversal mechanism.

### SERVICE

`SERVICE` advertises a service related to the entity without invoking it.

```text
SERVICE<TAB>relation<TAB>uri
```

Example:

```text
SERVICE	pet:medical-record	https://vet.example/records/123
```

A `SERVICE` record is descriptive only. Receiving one MUST NOT cause automatic service invocation.

## Errors

Core uses a minimal symbolic error response instead of numeric status codes. If a server cannot produce a successful response, it sends exactly one error record followed by the completion marker, and no other records:

```text
ERROR<TAB>code\r\n
.\r\n
```

The initial error codes are:

- `NOT_FOUND` — the selector does not identify a resource on this server.
- `BAD_SELECTOR` — the selector does not satisfy Core selector syntax.
- `TEMPORARY_FAILURE` — the server cannot complete the transaction at this time.

An `ERROR` response is complete only after its completion marker. Successful parsing of records followed by the marker is sufficient to identify a successful response; Core has no `200 OK` equivalent.

## Example transaction

Client connects to `example.org:7070` using TLS.

```text
C: /pet/123\r\n

S: ENTITY	pet:Pet	pet:123\r\n
S: FACT	pet:name	Moko\r\n
S: FACT	pet:species	pet:Dog\r\n
S: LINK	pet:veterinarian	gofer://vet.example:7070/pet/123\r\n
S: .\r\n
```

The following is incomplete, even though it contains valid record lines, because the completion marker was not received before the connection closed:

```text
S: ENTITY	pet:Pet	pet:123\r\n
S: FACT	pet:name	Moko\r\n
S: [connection closes]
```

The client MUST NOT treat this as a complete authoritative response.

## Ontology boundary

Gopher-NG defines syntax and traversal semantics. Ontologies define meaning.

Terms such as `pet:Pet`, `pet:name`, `construction:Room`, and `nightlife:Customer` are opaque identifiers from the perspective of Gopher-NG Core. Gopher-NG MUST NOT require knowledge of any specific ontology.

## Security considerations

- Discovered data is untrusted input.
- `FACT` data MUST NOT automatically be treated as agent instructions.
- `LINK` targets are untrusted.
- `SERVICE` advertisement does not imply permission to invoke a service.
- Clients SHOULD impose limits on traversal depth and response size.
- Credentials or authorization state belonging to other protocols MUST NOT automatically propagate to a different origin during traversal.

## Future work / deferred types

The following record types are intentionally not defined in v0.0.1: `ACTION`, `CREDENTIAL`, `CAPABILITY`, and `EVENT`.

Keeping discovery separate from execution avoids prematurely defining semantics unnecessary for the first implementation.

## TBD

- Whether a default TCP port will be assigned in a future version remains TBD; v0.0.1 requires every `gofer://` URI to contain an explicit port.
- URI percent-encoding rules and the precise mapping between an encoded URI path and the UTF-8 selector remain TBD.
- Unicode normalization requirements for selectors and record fields remain TBD.
- A maximum selector length and response-size limits remain TBD; clients SHOULD nevertheless enforce local limits.
- Minimum TLS version, cipher-suite policy, and certificate-validation profile remain TBD.
