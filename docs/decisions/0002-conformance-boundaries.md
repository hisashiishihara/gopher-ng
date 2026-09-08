# 0002: Conformance receiver boundaries awaiting review

Status: open questions, not a Core amendment. Core remains v0.0.1.

The conformance corpus deliberately excludes the following cases from normative
pass/fail expectations. No parser semantics are changed by this patch.

## Bytes after completion

For bytes `2e0d0a2e0d0a` (`.\r\n.\r\n`), Go's streaming parser returns at the
first marker; Python's whole-buffer parser rejects the extra marker. Core says
exactly one completion marker and connection close, but does not explicitly say
whether a receiver must wait for EOF and inspect trailing bytes before accepting
an otherwise complete response. This also affects trailing invalid UTF-8.

Maintainer question: must receivers reject trailing bytes, or may they complete
at the first marker? Sender framing is already specified; any receiver rule
should address streaming and when results may become authoritative. Until then,
neither receiver's behavior is promoted to a shared normative expectation.

## URI lexical edge cases

Go rejects `gofer://example.org:7070/?` and `gofer://example.org:7070/#`;
Python currently accepts them because it checks component contents. The ban on
queries/fragments suggests rejecting even empty components, but examples do not
spell out presence versus content. Maintainers should confirm this boundary.

Go rejects `gofer://example.org:7070/%ZZ`; Python's percent decoder preserves
it. Generic URI syntax suggests rejection, but Core does not explicitly select
a URI grammar or malformed-escape handling. Python's URL splitter also removes
some raw controls before validation, unlike Go. A future clarification should
identify the lexical URI grammar and reject prohibited input before URL-library
normalization. These cases need a bounded URI-focused follow-up, not new Core
features, and are excluded rather than silently changing either implementation.

## Local limits

Selector and response maxima remain unspecified by Core. The separate policy
vectors exercise existing Go budgets only; Python has no equivalent limit API.
These differences are implementation policies, not normative disagreements.
