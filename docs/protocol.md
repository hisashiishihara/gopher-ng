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

Core does not define HTTP compatibility or HTTP-like methods, headers, sessions, cookies, content negotiation, chunked encoding, or keep-alive. It also does not define authentication, ontology schemas, databases, query or search syntax, mutation operations, subscriptions, cross-protocol service advertisement, MCP integration, tool invocation, or UDP discovery.

## Transport

The v0.0.1 transport stack is:

```text
Gopher-NG
TCP
IP
```

For each transaction, a client:

1. connects to the server;
2. sends one selector;
3. receives one complete response and its explicit terminator; and
4. closes the connection.

The server closes the connection after sending the complete response. Neither side reuses a connection for another request. v0.0.1 Core does not define TLS; a future version may define a TLS transport and security profile.

## URIs and selectors

Gopher-NG locations use the `gofer` scheme. URI scheme comparison follows normal URI scheme comparison semantics.

```text
gofer://host:port/path
```

The host and explicit port identify the server. The host MUST be present. The port MUST be explicitly present and MUST be a decimal TCP port in the range 1 through 65535. Core does not define a default port. IPv6 literals use the normal bracketed URI form:

```text
gofer://[2001:db8::1]:7070/pet/123
```

The URI path represents the selector sent to the server. A `gofer://` URI MUST contain an explicit path beginning with `/`; an absent path is not implicitly converted to `/`. The path `/` represents the root selector.

URI percent-encoded octets in the path are decoded before the selector is sent. The percent-decoded bytes MUST form valid UTF-8, and the resulting selector MUST satisfy the Core selector rules below. Percent-decoded CR, LF, TAB, and other prohibited control characters remain invalid because the resulting selector fails validation. Unicode normalization is unspecified.

For example, resolving:

```text
gofer://example.org:7070/pet/123
```

conceptually sends the following request after the TCP connection is established:

```text
/pet/123\r\n
```

Likewise:

```text
gofer://example.org:7070/pet/Moko%20Chan
```

maps to the selector:

```text
/pet/Moko Chan
```

Gopher-NG v0.0.1 Core URIs MUST NOT contain userinfo, a query, or a fragment. The following URIs are invalid:

```text
gofer://user@example.org:7070/
gofer://example.org:7070/pet/123?x=1
gofer://example.org:7070/pet/123#section
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
FACT	pet:medical-record	https://vet.example/records/123
```

Predicates and values are opaque to Gopher-NG Core. A value MAY therefore identify a resource belonging to another protocol, but Core assigns no execution or traversal semantics to such a value.

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

Client connects to `example.org:7070` using TCP.

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

- v0.0.1 uses plain TCP and therefore provides no transport confidentiality or peer authentication.
- Discovered data is untrusted input.
- `FACT` data MUST NOT automatically be treated as agent instructions.
- `LINK` targets are untrusted.
- A URI appearing in a `FACT` value does not imply permission to dereference or invoke it.
- Clients SHOULD impose limits on traversal depth and response size.
- Credentials or authorization state belonging to other protocols MUST NOT automatically propagate to a different origin during traversal.
- Future work may define a TLS transport and security profile.

## Future work / deferred types

The following record types are intentionally not defined in v0.0.1: `SERVICE`, `ACTION`, `CREDENTIAL`, `CAPABILITY`, and `EVENT`.

`SERVICE` was considered for Core and deliberately deferred. A service URI can already be represented as ontology-defined data in a `FACT`, while a traversable Gopher-NG relation is represented by `LINK`. Giving service advertisement a dedicated Core type would therefore add cross-protocol semantics without adding necessary discovery power. The full rationale is recorded in [Design Decision 0001: `SERVICE` is not a Core record](decisions/0001-no-service-in-core.md).

Keeping discovery separate from execution avoids prematurely defining semantics unnecessary for the first implementation.

## TBD

- Whether a default TCP port will be assigned in a future version remains TBD; v0.0.1 requires every `gofer://` URI to contain an explicit port.
- Unicode normalization requirements for selectors and record fields remain TBD.
- A maximum selector length and response-size limits remain TBD; clients SHOULD nevertheless enforce local limits.
- A TLS transport and security profile, including minimum TLS version, cipher-suite policy, and certificate-validation profile, remains future work.
