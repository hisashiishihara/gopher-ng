from __future__ import annotations

import os
import signal
import socket
import subprocess
import sys
import time
import unittest
from pathlib import Path


REPOSITORY = Path(__file__).resolve().parents[2]
PYTHON_DIR = Path(__file__).resolve().parent


def unused_port() -> int:
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        return sock.getsockname()[1]


def wait_for_server(process: subprocess.Popen[str], port: int) -> None:
    deadline = time.monotonic() + 10
    while time.monotonic() < deadline:
        if process.poll() is not None:
            stdout, stderr = process.communicate()
            raise AssertionError(f"server exited early\nstdout: {stdout}\nstderr: {stderr}")
        try:
            with socket.create_connection(("127.0.0.1", port), timeout=0.1):
                return
        except OSError:
            time.sleep(0.05)
    raise AssertionError("server did not begin listening within 10 seconds")


class InteroperabilityTests(unittest.TestCase):
    def run_server(self, command: list[str], port: int) -> subprocess.Popen[str]:
        process = subprocess.Popen(
            command,
            cwd=REPOSITORY,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            start_new_session=True,
        )
        self.addCleanup(self.stop_server, process)
        wait_for_server(process, port)
        return process

    @staticmethod
    def stop_server(process: subprocess.Popen[str]) -> None:
        if process.poll() is None:
            os.killpg(process.pid, signal.SIGTERM)
            try:
                process.communicate(timeout=5)
            except subprocess.TimeoutExpired:
                os.killpg(process.pid, signal.SIGKILL)
                process.communicate(timeout=5)
        else:
            process.communicate()

    def test_go_client_to_python_server(self) -> None:
        port = unused_port()
        self.run_server(
            [sys.executable, str(PYTHON_DIR / "server.py"), "--listen", f"127.0.0.1:{port}"],
            port,
        )
        root = subprocess.run(
            ["go", "run", "./cmd/gng", f"gofer://127.0.0.1:{port}/"],
            cwd=REPOSITORY, text=True, capture_output=True, timeout=30,
        )
        self.assertEqual(root.returncode, 0, root.stderr)
        self.assertIn("ENTITY\texample:PythonServer\texample:python-root", root.stdout)
        self.assertIn("FACT\texample:message\tHello from Python", root.stdout)

        missing = subprocess.run(
            ["go", "run", "./cmd/gng", f"gofer://127.0.0.1:{port}/missing"],
            cwd=REPOSITORY, text=True, capture_output=True, timeout=30,
        )
        self.assertEqual(missing.returncode, 0, missing.stderr)
        self.assertIn("ERROR\tNOT_FOUND", missing.stdout)

    def test_python_client_to_go_server(self) -> None:
        port = unused_port()
        self.run_server(["go", "run", "./cmd/gngd", "-listen", f"127.0.0.1:{port}"], port)
        root = subprocess.run(
            [sys.executable, str(PYTHON_DIR / "client.py"), f"gofer://127.0.0.1:{port}/"],
            cwd=REPOSITORY, text=True, capture_output=True, timeout=30,
        )
        self.assertEqual(root.returncode, 0, root.stderr)
        self.assertRegex(root.stdout, r"(?m)^ENTITY\t[^\t\r\n]+\t[^\t\r\n]+$")

        missing = subprocess.run(
            [sys.executable, str(PYTHON_DIR / "client.py"), f"gofer://127.0.0.1:{port}/missing"],
            cwd=REPOSITORY, text=True, capture_output=True, timeout=30,
        )
        self.assertEqual(missing.returncode, 0, missing.stderr)
        self.assertEqual(missing.stdout, "ERROR\tNOT_FOUND\n")


if __name__ == "__main__":
    unittest.main()
