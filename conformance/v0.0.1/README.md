# Core v0.0.1 conformance corpus

`normative.json` is a language-neutral JSON object with `core_version` and a
finite `vectors` array. It requires no Go packages. The normative source remains
[the Core specification](../../docs/protocol.md); this corpus adds examples,
not semantics. IDs are stable within this version.

Each vector has `id`, `kind` (`request`, `response`, or `uri`), and `valid`.
Wire inputs use `wire_hex`: decode pairs of hexadecimal digits to bytes, without
newline conversion or Unicode replacement. URI inputs use `uri` as JSON text.
For accepted requests, compare `selector`. For accepted responses, compare
`records`, an ordered array of arrays containing the type followed by its fields.
For accepted URIs, compare `host`, `port` (integer), and decoded `selector`.
Rejected inputs must report failure; error wording and error classes are local.
A rejected/incomplete response must not yield authoritative partial records.

`unknown-selector-error` describes the response for a syntactically valid selector
absent from a server's resource map; existence is application state, not selector
syntax. The fixture's `request_selector` supplies that context, not a universal
resource lookup rule. These adapters test parsing, not server routing.

The federation example has source authority `origin.example:7070` and a LINK to
`independent.example:7071`. Parsing must preserve that URI as data and perform
no network requests. Neither LINK nor an HTTPS-valued FACT authorizes automatic
traversal or execution. These finite parser vectors do not certify all client
network behavior.

`policies.json` uses the same envelope and wire/result conventions, but is NOT
normative. `repeat_ascii` optionally appends its ASCII text `repeat_count` times
to `wire_hex` before appending decoded `suffix_hex`. `applies_to` identifies the
adapter; `limit_bytes` is the configured response budget. All wire bytes,
including CRLF and the completion marker, count. Go's selector reader has a
4096-byte local limit; Python currently has no corresponding bounded reader.
Python explicitly skips Go-only policy cases rather than claiming agreement.

Run the adapters from the repository root:

```
go test ./conformance -v
python3 -m unittest discover -s interop/python -p 'test_conformance.py' -v
```

They also run under the ordinary Go and Python test discovery commands. A third
party can implement the three operations above directly from these files.
See [open questions](../../docs/decisions/0002-conformance-boundaries.md) for
excluded cases whose receiver behavior needs maintainer clarification.
