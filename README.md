### HTTP/1.1 Server from Scratch — Learning Project

Implementing core pieces of HTTP/1.1 directly over TCP sockets in Go. Built to practice Go, networking fundamentals, and the HTTP protocol without relying on the standard `net/http` server.

### Learning goals
- Go networking with `net`: TCP listeners, `net.Conn`, read/write semantics
- Incremental stream parsing and state machines (request line → headers → body)
- HTTP/1.1 basics: CRLF framing, request line, header rules, Content-Length
- Response writing: status line, headers, body, and basic chunked transfer encoding
- Concurrency with goroutines per connection and graceful shutdown
- Test-Driven development
- Error handling and tests for parsers

### What it does
- Parses HTTP/1.1 requests: method, target, version, headers, and optional body (Content‑Length)
- Validates request line, methods, version, and header syntax; surfaces precise errors
- Builds HTTP responses over raw `net.Conn` with sensible defaults (Content‑Length, Connection)
- Supports writing chunked bodies and trailers helpers
- Minimal TCP HTTP server

