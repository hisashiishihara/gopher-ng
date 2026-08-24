"""Small, standard-library implementation of Gopher-NG Core framing."""

from __future__ import annotations

from dataclasses import dataclass
from urllib.parse import unquote_to_bytes, urlsplit


class ProtocolError(ValueError):
    """Raised when Core syntax or framing is invalid."""


@dataclass(frozen=True)
class Location:
    host: str
    port: int
    selector: str


FIELD_COUNTS = {"ENTITY": 2, "FACT": 2, "LINK": 2, "ERROR": 1}


def validate_selector(selector: str) -> str:
    if not selector.startswith("/"):
        raise ProtocolError("selector must begin with /")
    if any(ord(char) < 0x20 for char in selector):
        raise ProtocolError("selector contains a C0 control character")
    return selector


def _validate_field(field: str) -> None:
    if not field:
        raise ProtocolError("record fields must be non-empty")
    if "\t" in field or "\r" in field or "\n" in field:
        raise ProtocolError("record field contains a prohibited delimiter")


def parse_location(uri: str) -> Location:
    try:
        parsed = urlsplit(uri)
    except ValueError as exc:
        raise ProtocolError(f"invalid Gopher-NG URI: {exc}") from exc
    if parsed.scheme.lower() != "gofer" or not parsed.netloc:
        raise ProtocolError("URI must be an absolute gofer:// URI")
    if parsed.username is not None or parsed.password is not None:
        raise ProtocolError("userinfo is prohibited")
    if parsed.query:
        raise ProtocolError("query is prohibited")
    if parsed.fragment:
        raise ProtocolError("fragment is prohibited")
    if not parsed.hostname:
        raise ProtocolError("host is required")
    try:
        port = parsed.port
    except ValueError as exc:
        raise ProtocolError(f"invalid explicit TCP port: {exc}") from exc
    if port is None or not 1 <= port <= 65535:
        raise ProtocolError("an explicit TCP port from 1 through 65535 is required")
    if parsed.path == "":
        raise ProtocolError("an explicit path beginning with / is required")
    try:
        selector = unquote_to_bytes(parsed.path).decode("utf-8", errors="strict")
    except UnicodeDecodeError as exc:
        raise ProtocolError("percent-decoded path is not valid UTF-8") from exc
    return Location(parsed.hostname, port, validate_selector(selector))


def parse_record(line: str) -> tuple[str, ...]:
    parts = tuple(line.split("\t"))
    record_type = parts[0]
    expected = FIELD_COUNTS.get(record_type)
    if expected is None:
        raise ProtocolError(f"unsupported Core record type: {record_type}")
    if len(parts) - 1 != expected:
        raise ProtocolError(f"{record_type} requires exactly {expected} fields")
    for field in parts[1:]:
        _validate_field(field)
    if record_type == "LINK":
        parse_location(parts[2])
    return parts


def validate_records(records: list[tuple[str, ...]]) -> list[tuple[str, ...]]:
    errors = sum(record[0] == "ERROR" for record in records)
    if errors and (errors != 1 or len(records) != 1):
        raise ProtocolError("an ERROR response must contain exactly one record")
    return records


def parse_response(data: bytes) -> list[tuple[str, ...]]:
    try:
        text = data.decode("utf-8", errors="strict")
    except UnicodeDecodeError as exc:
        raise ProtocolError("response is not valid UTF-8") from exc
    lines = text.split("\r\n")
    if len(lines) < 2 or lines[-2:] != [".", ""]:
        raise ProtocolError("response is missing its completion marker")
    records = [parse_record(line) for line in lines[:-2]]
    return validate_records(records)


def encode_records(records: list[tuple[str, ...]]) -> bytes:
    validate_records([parse_record("\t".join(record)) for record in records])
    body = b"".join("\t".join(record).encode("utf-8") + b"\r\n" for record in records)
    return body + b".\r\n"
