from __future__ import annotations

import unittest

from protocol import ProtocolError, parse_location, parse_record, parse_response, validate_selector


class SelectorTests(unittest.TestCase):
    def test_valid_root_selector(self) -> None:
        self.assertEqual(validate_selector("/"), "/")

    def test_selector_requires_leading_slash(self) -> None:
        with self.assertRaises(ProtocolError):
            validate_selector("missing")

    def test_selector_rejects_cr(self) -> None:
        with self.assertRaises(ProtocolError):
            validate_selector("/bad\rvalue")

    def test_selector_rejects_lf(self) -> None:
        with self.assertRaises(ProtocolError):
            validate_selector("/bad\nvalue")

    def test_selector_rejects_tab(self) -> None:
        with self.assertRaises(ProtocolError):
            validate_selector("/bad\tvalue")

    def test_selector_rejects_other_c0_control(self) -> None:
        with self.assertRaises(ProtocolError):
            validate_selector("/bad\x00value")


class RecordTests(unittest.TestCase):
    def test_valid_entity(self) -> None:
        self.assertEqual(parse_record("ENTITY\texample:Type\texample:id"),
                         ("ENTITY", "example:Type", "example:id"))

    def test_valid_fact(self) -> None:
        self.assertEqual(parse_record("FACT\texample:name\tMoko"),
                         ("FACT", "example:name", "Moko"))

    def test_valid_link(self) -> None:
        self.assertEqual(parse_record("LINK\texample:next\tgofer://localhost:7070/next"),
                         ("LINK", "example:next", "gofer://localhost:7070/next"))

    def test_valid_single_error(self) -> None:
        self.assertEqual(parse_response(b"ERROR\tNOT_FOUND\r\n.\r\n"),
                         [("ERROR", "NOT_FOUND")])

    def test_wrong_field_counts(self) -> None:
        for line in ("ENTITY\tonly-one", "FACT\ta\tb\textra", "LINK\trelation", "ERROR\ta\tb"):
            with self.subTest(line=line), self.assertRaises(ProtocolError):
                parse_record(line)

    def test_error_mixed_with_success(self) -> None:
        with self.assertRaises(ProtocolError):
            parse_response(b"ENTITY\ta\tb\r\nERROR\tNOT_FOUND\r\n.\r\n")

    def test_multiple_errors(self) -> None:
        with self.assertRaises(ProtocolError):
            parse_response(b"ERROR\tNOT_FOUND\r\nERROR\tTEMPORARY_FAILURE\r\n.\r\n")

    def test_missing_response_terminator(self) -> None:
        with self.assertRaises(ProtocolError):
            parse_response(b"ENTITY\ta\tb\r\n")


class LocationTests(unittest.TestCase):
    def test_explicit_uri_port_required(self) -> None:
        with self.assertRaises(ProtocolError):
            parse_location("gofer://localhost/")

    def test_absent_uri_path_rejected(self) -> None:
        with self.assertRaises(ProtocolError):
            parse_location("gofer://localhost:7070")

    def test_query_rejected(self) -> None:
        with self.assertRaises(ProtocolError):
            parse_location("gofer://localhost:7070/?query=yes")

    def test_fragment_rejected(self) -> None:
        with self.assertRaises(ProtocolError):
            parse_location("gofer://localhost:7070/#fragment")

    def test_userinfo_rejected(self) -> None:
        with self.assertRaises(ProtocolError):
            parse_location("gofer://user@localhost:7070/")

    def test_percent_decoded_selector(self) -> None:
        self.assertEqual(parse_location("gofer://localhost:7070/Moko%20Chan").selector,
                         "/Moko Chan")

    def test_percent_decoded_control_rejected(self) -> None:
        with self.assertRaises(ProtocolError):
            parse_location("gofer://localhost:7070/bad%09selector")


if __name__ == "__main__":
    unittest.main()
