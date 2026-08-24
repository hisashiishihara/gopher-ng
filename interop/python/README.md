# Python interoperability proof

This directory is a small, independent interoperability proof, not a second
official reference implementation. It was implemented from
[`docs/protocol.md`](../../docs/protocol.md) and does not depend on Go internal
packages.

Run the Python server on localhost:

```sh
python3 interop/python/server.py --listen 127.0.0.1:7072
```

Query it with the Go client:

```sh
go run ./cmd/gng gofer://127.0.0.1:7072/
```

In another direction, run the Go daemon:

```sh
go run ./cmd/gngd -listen 127.0.0.1:7070
```

Query it with the Python client:

```sh
python3 interop/python/client.py gofer://127.0.0.1:7070/
```

Run all Python protocol and interoperability tests with:

```sh
python3 -m unittest discover -s interop/python -p 'test_*.py' -v
```

The success criterion is interoperability in both directions:

```text
Go client -> Python server
Python client -> Go server
```

Automatic `LINK` traversal is intentionally not implemented. No Core protocol
changes are required unless this experiment actually discovers a specification
defect.
