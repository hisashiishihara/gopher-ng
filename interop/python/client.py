#!/usr/bin/env python3
"""Minimal Gopher-NG Core client used for interoperability testing."""

from __future__ import annotations

import argparse
import socket
import sys

from protocol import ProtocolError, parse_location, parse_response


def fetch(uri: str, timeout: float = 5.0) -> list[tuple[str, ...]]:
    location = parse_location(uri)
    with socket.create_connection((location.host, location.port), timeout=timeout) as connection:
        connection.sendall(location.selector.encode("utf-8") + b"\r\n")
        chunks: list[bytes] = []
        while True:
            chunk = connection.recv(4096)
            if not chunk:
                break
            chunks.append(chunk)
    return parse_response(b"".join(chunks))


def main() -> int:
    parser = argparse.ArgumentParser(description="query a Gopher-NG Core server")
    parser.add_argument("uri")
    args = parser.parse_args()
    try:
        records = fetch(args.uri)
    except (OSError, ProtocolError) as exc:
        print(f"gng-python: {exc}", file=sys.stderr)
        return 1
    for record in records:
        print("\t".join(record))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
