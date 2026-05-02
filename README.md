# 🟥 Mini Redis (Go)

A production-grade Redis clone built from scratch in Go.

This project demonstrates systems-level thinking, streaming TCP parsing, protocol correctness (RESP2), clean layered architecture, and secure containerization practices.

---

## 🚀 Highlights

- 🔹 Streaming-safe RESP2 parser (handles partial TCP reads)
- 🔹 Recursive array parsing with exact byte accounting
- 🔹 Clean layered architecture (Parser → Decoder → Executor → Encoder)
- 🔹 Command routing with handler injection
- 🔹 Protocol-accurate RESP encoding
- 🔹 Multi-stage Docker build (distroless runtime)
- 🔹 SBOM & provenance ready (supply-chain aware)

---

## 🧠 System Architecture

```
                    ┌────────────────────┐
                    │     TCP Socket     │
                    └─────────┬──────────┘
                              │
                              ▼
                ┌──────────────────────────┐
                │   Streaming RESP Parser  │
                │  (bytes → RespValue)     │
                └─────────┬────────────────┘
                          │
                          ▼
                ┌──────────────────────────┐
                │        Decoder           │
                │ (RespValue → Command)    │
                └─────────┬────────────────┘
                          │
                          ▼
                ┌──────────────────────────┐
                │        Executor          │
                │   (Command Router)       │
                └─────────┬────────────────┘
                          │
                          ▼
                ┌──────────────────────────┐
                │        Handlers          │
                │  (Business Logic Layer)  │
                └─────────┬────────────────┘
                          │
                          ▼
                ┌──────────────────────────┐
                │        Datastore         │
                │    map[string]string     │
                └─────────┬────────────────┘
                          │
                          ▼
                ┌──────────────────────────┐
                │         Encoder          │
                │ (RespValue → bytes)      │
                └─────────┬────────────────┘
                          │
                          ▼
                    ┌────────────────────┐
                    │     TCP Write      │
                    └────────────────────┘
```

Each layer has a single responsibility and is independently testable.

---

## 📡 Protocol Correctness (RESP2)

Implemented according to the official Redis specification:

https://redis.io/docs/latest/develop/reference/protocol-spec/

### Supported RESP Types

| Type            | Example                         |
|-----------------|---------------------------------|
| Simple String  | `+OK\r\n`                       |
| Error          | `-ERR message\r\n`              |
| Integer        | `:1000\r\n`                     |
| Bulk String    | `$5\r\nhello\r\n`               |
| Null Bulk      | `$-1\r\n`                       |
| Array          | `*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n` |

Correct distinction is maintained between:

- Simple strings (`+OK`)
- Bulk strings (`$len\r\nvalue\r\n`)
- Null bulk (`$-1`)
- Empty bulk (`$0\r\n\r\n`)

---

## 🔁 Streaming Parser Design

TCP is treated as a continuous byte stream:

- No assumptions about message boundaries
- Handles partial reads
- Returns:
    - `Success`
    - `NeedMoreData`
    - `ProtocolError`
- Maintains exact `bytesConsumed`

Example (bulk string availability check):

```go
payloadStart := index + 2
required := payloadStart + length + 2

if len(bufferValue) < required {
    return getParseNeedMoreDataResp()
}
```

References:
- https://www.openmymind.net/2012/1/12/Reading-From-TCP-Streams/
- https://stackoverflow.com/questions/9563563/what-is-a-message-boundary

---

## 🧩 Internal Representation (AST)

All protocol data is represented using a structured internal AST:

```go
type RespValue struct {
    Type   RespType
    Str    string
    Int    int64
    Array  []*RespValue
    Error  error
    IsNull bool
}
```

### Design Notes

- `Type` differentiates:
    - Simple String
    - Bulk String
    - Integer
    - Error
    - Array
- `Str` stores both simple and bulk string values.
- `IsNull` differentiates:
    - `$-1` (null bulk)
    - `$0\r\n\r\n` (empty bulk)
- `Array` enables recursive RESP parsing.
- `Error` carries execution or protocol errors.

Flow:

- Parser → produces `RespValue`
- Executor → returns `RespValue`
- Encoder → serializes `RespValue`

This mirrors the protocol exactly.

---

## 📦 Implemented Commands

| Command | Behavior |
|----------|----------|
| `PING` | `+PONG` |
| `SET key value` | `+OK` |
| `GET key` | Bulk string |
| `GET missing` | Null bulk |
| `INCR key` | Integer reply |
| `DEL key` | Integer reply |
| `ECHO value` | Bulk string |

Example execution routing:

```go
func (e *Executor) Execute(cmd Command) RespValue {
    handler := CommandHandler(cmd.Name)
    if handler == nil {
        return ErrorResp("unknown command")
    }
    return handler(cmd, e.dataStore)
}
```

Handlers operate directly on the datastore (`map[string]string`).

---

## 🧱 Clean Command Routing

The architecture separates:

- Protocol parsing
- Command decoding
- Business logic
- Data storage
- Response encoding

This allows:

- Independent unit testing
- Easy extensibility (TTL, persistence)
- Clear separation of concerns
- Maintainable codebase

---

## 🐳 Containerization

Multi-stage production build:

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/server .
USER nonroot:nonroot
ENTRYPOINT ["./server"]
```

### Why Distroless?

- Minimal attack surface
- No shell in runtime image
- Reduced CVEs
- Static Go binary

Reference:
https://github.com/GoogleContainerTools/distroless

---

## 🔐 Supply-Chain Awareness

Supports:

- SBOM generation (Syft)
- Docker provenance attestations
- Minimal runtime footprint

References:
- https://github.com/anchore/syft
- https://docs.docker.com/build/attestations/

---

## 📚 Learning References

- RESP Specification  
  https://redis.io/docs/latest/develop/reference/protocol-spec/

- TCP Streams & Framing  
  https://www.openmymind.net/2012/1/12/Reading-From-TCP-Streams/

- Redis Internals  
  https://redis.io/docs/latest/

- Distroless Containers  
  https://github.com/GoogleContainerTools/distroless

---

## 🎯 Future Work

- TTL support
- RDB persistence
- Append-only file (AOF)
- Event-loop I/O multiplexing
- Concurrency safety
- Benchmarking
- Replication support

---

## 🏁 Status

✔ Streaming RESP parser  
✔ Recursive array parsing  
✔ Protocol-accurate encoder  
✔ Clean layered architecture  
✔ Command execution layer  
✔ Containerized (multi-stage + distroless)  
✔ Security-aware design

---

Built as a systems engineering exercise focused on correctness, architecture, and production-grade backend design.