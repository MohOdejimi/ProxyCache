# ProxyCache

A in-memory caching reverse proxy built in Go. ProxyCache sits between clients and an origin server, forwarding requests and caching responses to reduce latency and origin load on repeated requests.

---

## Features

- **Reverse proxy** — transparently forwards client requests to a configurable origin server
- **In-memory caching** — caches GET responses and serves them directly on subsequent identical requests, bypassing the origin entirely
- **TTL-based expiry** — respects the origin's `Cache-Control: max-age` directive, falling back to a default TTL of 5 minutes
- **Background cache cleanup** — expired entries are automatically purged from memory every 10 minutes
- **Smart cache exclusion** — responses are not cached if they contain any of the following:
  - `Cache-Control: no-cache`, `must-revalidate`, or `proxy-revalidate`
  - `Pragma: no-cache`
  - `Set-Cookie` header (session-specific responses)
  - `Authorization` header on the request (user-specific responses)
  - `Expires` header indicating the response has already expired
  - Body exceeding 10MB
- **`X-Cache` header** — every response includes an `X-Cache: HIT` or `X-Cache: MISS` header so clients can observe cache behaviour
- **Cache clearing** — all cached entries can be flushed using the `--clear` flag

---

## Architecture & Design

## ProxyCache

![ProxyCache Architecture](assets/proxycache.png)

A in‑memory caching reverse proxy built in Go…

### How the flows work

**Cache Miss**

When a request arrives and no valid cached entry exists, ProxyCache forwards the request to the origin server. If the response is cacheable, it is stored in memory with a TTL before being returned to the client with `X-Cache: MISS`.

```
Client → Proxy → [Cache Lookup: MISS] → Origin Server
                                      ← Response Data
       ← [Store Response + TTL] ← 
       ← Response (X-Cache: MISS)
```

**Cache Hit**

When a valid cached entry exists, ProxyCache serves the response directly from memory. The origin server is never contacted.

```
Client → Proxy → [Cache Lookup: HIT]
       ← Response (X-Cache: HIT)
```

**Clear Cache**

Clients can flush all cached entries by passing the `--clear` flag at the CLI.

```
Client → Proxy (--clear) → [Purge All Entries]
       ← Cache Cleared
```

### Concurrency

The cache store uses a `sync.RWMutex` — allowing multiple concurrent reads while ensuring writes are exclusive. This means cache hits under high concurrency are non-blocking.

---

## Installation & Usage

### Prerequisites

- Go 1.21 or higher

### Clone and build

```bash
git clone https://github.com/MohOdejimi/ProxyCache.git
cd ProxyCache
go build -o proxycache ./cmd
```

### Start the proxy server

```bash
./proxycache --port 8080 --origin https://dummyjson.com
```

The proxy will start on port `8080` and forward all requests to `https://dummyjson.com`.

### Send a request

```bash
curl -v http://localhost:8080/products
```

On the first request you will see:

```
X-Cache: MISS
```

On subsequent identical requests:

```
X-Cache: HIT
```

### Clear the cache

```bash
./proxycache --clear
```

### CLI flags

| Flag | Description | Default |
|------|-------------|---------|
| `--port` | Port for the proxy server to listen on | `8080` |
| `--origin` | URL of the origin server to forward requests to | required |
| `--clear` | Flush all entries from the cache | — |