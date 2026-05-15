# Mnemo — In-Memory Data Store in Go

A production-grade Redis-compatible in-memory data store, built from scratch in Go.

Mnemo implements the RESP2 protocol over raw TCP, handles concurrent clients through a single-threaded executor model, and is designed with correctness, performance, and observability as first-class concerns.

---

## Architecture

```
  Client A ─┐
  Client B ─┤──► goroutine per client
  Client C ─┘         │
                       │  reads from TCP stream
                       ▼
             ┌──────────────────┐
             │  RESP2 Parser    │  TCP byte stream → RespValue
             └────────┬─────────┘
                      │
                      ▼
             ┌──────────────────┐
             │    Decoder       │  RespValue → Command
             └────────┬─────────┘
                      │
                      ▼
             ┌──────────────────────────────────┐
             │     Executor Channel             │  bounded, size 1024
             └────────────────┬─────────────────┘
                              │  single goroutine
                              ▼
             ┌──────────────────────────────────┐
             │     Command Executor             │  routes to handlers
             └────────────────┬─────────────────┘
                              │
                              ▼
             ┌──────────────────────────────────┐
             │     Handlers + Datastore         │  business logic
             └────────────────┬─────────────────┘
                              │
                              ▼
             ┌──────────────────────────────────┐
             │     Encoder                      │  RespValue → bytes
             └────────────────┬─────────────────┘
                              │
                              ▼
                   TCP Write (per client)
```

Each layer has a single responsibility and is independently testable. The single-threaded executor eliminates lock contention on the datastore entirely — concurrency is handled at the I/O layer, not the execution layer.

---

## Protocol Support (RESP2)

Implemented against the official Redis protocol specification.

| Type          | Wire Format                        |
| ------------- | ---------------------------------- |
| Simple String | `+OK\r\n`                          |
| Error         | `-ERR message\r\n`                 |
| Integer       | `:1000\r\n`                        |
| Bulk String   | `$5\r\nhello\r\n`                  |
| Null Bulk     | `$-1\r\n`                          |
| Array         | `*2\r\n$3\r\nGET\r\n$3\r\nkey\r\n` |

The parser treats TCP as a continuous byte stream with no assumptions about message boundaries. Every parse attempt returns one of three states: `Success`, `NeedMoreData`, or `ProtocolError`, with exact byte accounting for safe buffer advancement.

---

## Implemented Commands

| Command         | Response    |
| --------------- | ----------- |
| `PING`          | `+PONG`     |
| `SET key value` | `+OK`       |
| `GET key`       | Bulk string |
| `GET missing`   | Null bulk   |
| `INCR key`      | Integer     |
| `DEL key`       | Integer     |
| `ECHO value`    | Bulk string |

---

## Performance

Benchmarked using `redis-benchmark` against a local instance. Numbers reflect a development machine and will vary by hardware. The table below documents the optimization progression, not an absolute performance claim.

```bash
# Baseline — no pipeline
redis-benchmark -p 6379 -c 50 -n 100000 -t set,get --dbnum 0 -q

# Pipeline P16
redis-benchmark -p 6379 -c 50 -n 100000 -t set,get --dbnum 0 -q -P 16

# Pipeline P32
redis-benchmark -p 6379 -c 50 -n 100000 -t set,get --dbnum 0 -q -P 32

# High concurrency P8
redis-benchmark -p 6379 -c 100 -n 100000 -t set,get --dbnum 0 -q -P 8
```

| Optimization                 | What Changed                                                                                                                                                           |
| ---------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Baseline                     | Naive implementation, flush on every write                                                                                                                             |
| Adaptive Flush               | Flush only when the response channel is empty, batching writes under load                                                                                              |
| Buffer Pool (scratch space)  | Replaced `fmt.Sprintf` in encoder with append-based formatting; added `sync.Pool` for scratch buffers, reducing GC pressure while keeping exact-size final allocations |
| Parser Buffer Reset          | `buf[:0]` slice reset to reuse underlying array across reads, avoiding fresh allocations                                                                               |
| Executor Channel Size Tuning | Tested 512, 1024, 2048, 4096; settled on 1024 as the optimal balance                                                                                                   |

**Results at pipeline depth P32 (SET / GET):**

| Run | Pipeline Depth   | SET ops/sec | GET ops/sec | SET p50  | GET p50  |
| --- | ---------------- | ----------- | ----------- | -------- | -------- |
| 1   | P1 (no pipeline) | ~70,000     | ~70,000     | ~0.40 ms | ~0.40 ms |
| 2   | P8               | ~800,000    | ~885,000    | ~0.50 ms | ~0.48 ms |
| 3   | P32              | ~1,176,000  | ~1,315,000  | ~0.69 ms | ~0.66 ms |
| 4   | P8 (c100)        | ~534,000    | ~529,000    | ~0.78 ms | ~0.79 ms |

---

## Testing

The project follows a tests-first approach. The suite covers:

- Unit tests — parser, encoder, decoder, handlers in isolation
- Integration tests — full command round-trips over a real TCP connection
- Fuzz tests — malformed and partial RESP input to the parser
- Race detector — all tests run with `-race` to surface concurrency issues

```bash
go test ./... -race
go test ./... -fuzz=FuzzParser
```

---

## Containerization

Multi-stage build targeting a distroless runtime image.

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

The runtime image contains no shell, no package manager, and no toolchain — only the compiled binary. Container integration testing is planned.

---

## Roadmap

| Phase | Focus                                             | Status      |
| ----- | ------------------------------------------------- | ----------- |
| 1     | Correctness and test suite                        | Complete    |
| 2     | Benchmarking and optimization                     | Complete    |
| 3     | Backpressure and overload protection              | Complete    |
| 4     | Mnemo-CLI compatibility                           | In Progress |
| 5     | Data structures (Lists, Hashes, Sets)             | Planned     |
| 6     | Persistence (AOF + RDB)                           | Planned     |
| 7     | Observability (structured logging, INFO, metrics) | Planned     |
| 8     | Transactions (MULTI/EXEC/WATCH)                   | Planned     |
| 9     | Memory management (LRU, maxmemory, TTL expiry)    | Planned     |

---

## References

**RESP Protocol & Redis Internals**

- Redis Serialization Protocol (RESP) — https://redis.io/docs/latest/develop/reference/protocol-spec/
- Redis Documentation — https://redis.io/docs/latest/
- Redis Persistence (RDB & AOF) — https://redis.io/docs/latest/operate/oss_and_stack/management/persistence/
- Redis Performance Optimization — https://redis.io/docs/latest/operate/oss_and_stack/management/optimization/
- Redis Event Library Internals — https://redis.io/docs/latest/operate/oss_and_stack/reference/internals/internals-rediseventlib/

**TCP Streams & Message Framing**

- OpenMyMind — Reading from TCP Streams — https://www.openmymind.net/2012/1/12/Reading-From-TCP-Streams/
- StackOverflow — What is a message boundary? — https://stackoverflow.com/questions/9563563/what-is-a-message-boundary
- Stephen Cleary — Message Framing — https://blog.stephencleary.com/2009/04/message-framing.html
- Wikipedia — Transmission Control Protocol — https://en.wikipedia.org/wiki/Transmission_Control_Protocol
- Wikipedia — Message Framing — https://en.wikipedia.org/wiki/Message_framing

**Networking & I/O Multiplexing**

- Beej's Guide to Network Programming — https://beej.us/guide/bgnet/html/
- Wikipedia — I/O Multiplexing — https://en.wikipedia.org/wiki/I/O_multiplexing
- Linux man pages — epoll — https://man7.org/linux/man-pages/man7/epoll.7.html

**Redis Architecture & Systems Design**

- StackOverflow — Why Redis is single-threaded — https://stackoverflow.com/questions/45364256/why-redis-is-single-threaded-event-driven
- ByteByteGo — A Crash Course in Redis — https://blog.bytebytego.com/p/a-crash-course-in-redis
- Medium — Redis Internals by Rebuilding It — https://skshmgpt.medium.com/redis-internals-understanding-it-by-rebuilding-it-e16d6dd102e2
- Medium — What is Redis and how does it work internally — https://medium.com/@ayushsaxena823/what-is-redis-and-how-does-it-work-cfe2853eb9a9

**Containerization**

- Docker Docs — Multi-stage Builds — https://docs.docker.com/build/building/multi-stage/
- Docker Docs — Dockerfile Reference — https://docs.docker.com/reference/dockerfile/
- Google Distroless Containers — https://github.com/GoogleContainerTools/distroless
- Docker Blog — Is Your Container Image Really Distroless? — https://www.docker.com/blog/is-your-container-image-really-distroless/
- Medium — Alpine vs Distroless vs Scratch — https://medium.com/google-cloud/alpine-distroless-or-scratch-caac35250e0b

**Testing & Correctness**

- Go testing package — https://pkg.go.dev/testing
- Go — Table Driven Tests — https://go.dev/wiki/TableDrivenTests
- Go — Introducing the Race Detector — https://go.dev/blog/race-detector
- Go — Data Race Detector reference — https://go.dev/doc/articles/race_detector
- Go — Fuzzing documentation — https://go.dev/doc/fuzz/
- Go — Getting started with fuzzing (tutorial) — https://go.dev/doc/tutorial/fuzz
- Testify assertion library — https://github.com/stretchr/testify
- GopherCon 2017 — Advanced Testing in Go (Mitchell Hashimoto) — https://www.youtube.com/watch?v=8hQG7QlcLBk
