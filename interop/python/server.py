#!/usr/bin/env python3
"""Minimal blocking Gopher-NG Core server used for interoperability testing."""

from __future__ import annotations

import argparse
import socket

from protocol import ProtocolError, encode_records, validate_selector


ROOT_RECORDS = [
    ("ENTITY", "example:PythonServer", "example:python-root"),
    ("FACT", "example:message", "Hello from Python"),
]


def _read_selector(connection: socket.socket) -> str:
    data = bytearray()
    while not data.endswith(b"\r\n"):
        chunk = connection.recv(1)
        if not chunk:
            raise ProtocolError("connection closed before selector terminator")
        data.extend(chunk)
    try:
        selector = bytes(data[:-2]).decode("utf-8", errors="strict")
    except UnicodeDecodeError as exc:
        raise ProtocolError("selector is not valid UTF-8") from exc
    return validate_selector(selector)


def serve_connection(connection: socket.socket) -> None:
    try:
        selector = _read_selector(connection)
    except ProtocolError:
        records = [("ERROR", "BAD_SELECTOR")]
    else:
        records = ROOT_RECORDS if selector == "/" else [("ERROR", "NOT_FOUND")]
    connection.sendall(encode_records(records))


def parse_listen(value: str) -> tuple[str, int]:
    host, separator, port_text = value.rpartition(":")
    if not separator or not host:
        raise argparse.ArgumentTypeError("listen address must be HOST:PORT")
    try:
        port = int(port_text)
    except ValueError as exc:
        raise argparse.ArgumentTypeError("port must be decimal") from exc
    if not 1 <= port <= 65535:
        raise argparse.ArgumentTypeError("port must be from 1 through 65535")
    return host, port


def main() -> int:
    parser = argparse.ArgumentParser(description="serve a minimal Gopher-NG Core resource")
    parser.add_argument("--listen", type=parse_listen, default=parse_listen("127.0.0.1:7072"))
    args = parser.parse_args()
    with socket.create_server(args.listen) as listener:
        while True:
            connection, _ = listener.accept()
            with connection:
                serve_connection(connection)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
