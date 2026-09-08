"""Adapter for the shared, language-neutral Core v0.0.1 corpus."""
from __future__ import annotations

import io
import json
from pathlib import Path
import unittest
from unittest.mock import patch

from protocol import ProtocolError, parse_location, parse_response
from server import _read_selector


CORPUS = Path(__file__).resolve().parents[2] / "conformance" / "v0.0.1"


class MemoryConnection:
    """Socket-shaped input using the production request reader, without TCP."""

    def __init__(self, data):
        self.input = io.BytesIO(data)

    def recv(self, size):
        return self.input.read(size)


class ConformanceTests(unittest.TestCase):
    def test_vectors(self):
        for name in ("normative", "policies"):
            corpus = json.loads((CORPUS / (name + ".json")).read_text(encoding="utf-8"))
            self.assertEqual(corpus["core_version"], "v0.0.1")
            self.assertTrue(corpus["vectors"])
            for vector in corpus["vectors"]:
                with self.subTest(corpus=name, id=vector["id"]):
                    if vector.get("applies_to", "python") != "python":
                        self.skipTest("implementation policy for " + vector["applies_to"])
                    wire = bytes.fromhex(vector.get("wire_hex", ""))
                    wire += vector.get("repeat_ascii", "").encode("ascii") * vector.get("repeat_count", 0)
                    wire += bytes.fromhex(vector.get("suffix_hex", ""))
                    # Parsing LINK/FACT data must never initiate traversal.
                    with patch("socket.create_connection", side_effect=AssertionError("automatic traversal")):
                        if vector["valid"]:
                            self.assertEqual(self.parse(vector, wire), self.expected(vector))
                        else:
                            with self.assertRaises(ProtocolError):
                                self.parse(vector, wire)

    def parse(self, vector, wire):
        kind = vector["kind"]
        if kind == "request":
            return _read_selector(MemoryConnection(wire))
        if kind == "response":
            return [list(record) for record in parse_response(wire)]
        if kind == "uri":
            location = parse_location(vector["uri"])
            return [location.host, location.port, location.selector]
        self.fail("unknown vector kind: " + kind)

    def expected(self, vector):
        if vector["kind"] == "request":
            return vector["selector"]
        if vector["kind"] == "response":
            return vector["records"]
        return [vector["host"], vector["port"], vector["selector"]]


if __name__ == "__main__":
    unittest.main()
