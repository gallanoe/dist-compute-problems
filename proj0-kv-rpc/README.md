Love it. Here’s a fully fleshed-out, **step-by-step implementation plan** for **Project 0: RPC Key–Value Service**—kept deliberately high-level (no design or code), but specific enough so nothing important gets skipped.

---

# Project 0 — RPC KV Service (Learning Sprint)

## Objective
Build a tiny key–value microservice and client that exercise the core building blocks you’ll reuse later: RPC, deadlines/timeouts, retries with jitter, idempotency tokens, observability (traces/metrics/logs), and basic load testing.

---

## Deliverables (end-state checklist)
- RPC service exposing create/update, read, delete operations.
- CLI or small client lib that calls the service with deadlines/retries/jitter and optional idempotency key.
- Configurable storage backend (in-memory first; file/embedded-KV optional).
- Structured logs, metrics (counters/histograms), distributed traces spanning client → server.
- Containerized service and client; simple local run with one command.
- Basic load test script + a short README capturing results and gotchas.

---

## Scope Guardrails
- Single process per service instance; single binary for server; separate client binary or CLI subcommand.
- No auth beyond a simple token header toggle.
- One region / one node (no replication yet).
- Bounded key/value size; reject oversized inputs early.

---

## Step-By-Step Plan

### 1) Project Skeleton & Interfaces
- Set up repo layout for server, client, proto/IDL, configs, and tests.
- Define a minimal API surface (key ops, health, version) in an IDL file.
- Establish a versioning scheme for the API and the service (include a version endpoint).
- Add basic config loading strategy (env vars + config file precedence).
- Decide and document operational defaults (ports, timeouts, limits).

### 2) Build & Run Loop
- Add build tasks (format, lint, compile) and pre-commit hooks.
- Create a single “dev up” command to start server with sensible defaults.
- Add a smoke test script that calls health/version and validates a happy path call.

### 3) Server Core (Happy Path Only)
- Parse config; initialize storage; register RPC handlers; start serving.
- Implement input validation and uniform error mapping to RPC errors.
- Add graceful shutdown hooks (drain listeners, flush metrics/traces, close storage).
- Implement health checks (liveness vs readiness).

### 4) Client Core (Happy Path Only)
- Initialize RPC client with per-call deadline.
- Implement thin CLI wrapper or library functions for each operation.
- Add a human-readable output mode and a machine-readable mode (e.g., JSON) for scripts.

### 5) Timeouts & Deadlines
- Introduce per-method default deadlines server-side; reject calls that exceed them.
- Enforce per-call client deadlines; ensure propagation via request metadata.
- Add an integration test that verifies deadline expiration behavior end-to-end.

### 6) Retries with Exponential Backoff + Jitter
- Define retryable vs non-retryable error categories.
- Add client-side retry policy: max attempts, base backoff, cap, jitter.
- Ensure retries are visible in logs/metrics (attempt count label).
- Create a fault-injection mode (server can simulate transient failures) for testing retries.

### 7) Idempotency & Safe Writes
- Introduce an **idempotency key** for write operations (optional header/metadata).
- Add server-side store of recent idempotency keys with expiration.
- Ensure duplicate write attempts (same key) do not produce duplicated side effects.
- Add tests that fire concurrent duplicate requests and verify a single effect.

### 8) Observability Foundation
- **Structured logging**: consistent fields (trace id, span id, request id, method, status, latency).
- **Metrics**: request counter by method/status, latency histogram per method, in-flight gauge.
- **Tracing**: client starts a span, propagates context; server creates child spans; annotate retries.
- Provide a simple dashboard JSON (latency percentile, error rate, RPS).

### 9) Input Validation, Limits, and Errors
- Enforce max key length, max value size, request size limit.
- Normalize error responses; avoid leaking internals; include stable error codes.
- Validate content type and required headers.
- Add tests for each validation path and limit breach.

### 10) Storage Layer Options
- Start with in-memory map guarded by a concurrency-safe access pattern.
- Add a flag to switch to a persistent embedded store (optional).
- Introduce basic compaction/cleanup schedule if using a log-structured store (optional).

### 11) Concurrency Controls
- Establish per-request goroutine/thread pattern and limits on concurrency.
- Add a server-side queue/bounded worker pool (or simple limiter) to prevent overload.
- Verify behavior under parallel PUT/GET/DELETE with contention on the same key.

### 12) Graceful Shutdown & Draining
- Implement signal handling; stop accepting new requests; wait for in-flight to finish with a deadline.
- Flush logs/metrics/traces; close storage; exit with correct status.
- Add a test that sends SIGTERM during load and verifies clean shutdown.

### 13) Security (Minimal, Dev-Friendly)
- Add an optional static bearer token check; allow “disabled” mode for local dev.
- Redact sensitive fields in logs.
- Document how to rotate the token without restarting (if you expose dynamic config reload).

### 14) Packaging & Local Orchestration
- Write a Dockerfile for server and client (multi-stage builds).
- Provide a docker-compose file that starts the service + metrics stack (collector, Prometheus, Grafana).
- Include sample dashboards and a quickstart script to bring everything up and curl a few calls.

### 15) Load Test & SLO Sketch
- Provide a basic load generator script (separate process).
- Run a small matrix of request mixes (read-heavy, write-heavy, mixed) at increasing RPS.
- Capture latency percentiles, error rates, and CPU/memory profiles.
- Record preliminary SLOs (e.g., p99 latency targets) and doc conditions.

### 16) Failure & Chaos Exercises
- Introduce server-side artificial delay/failure toggles (e.g., env var or admin endpoint).
- Test client retries against injected transient errors; verify no duplicate effects with idempotency.
- Test resource exhaustion scenarios (value size near limit, too many concurrent clients).
- Verify timeouts, backoff, and limiter interactions keep the system stable.

### 17) Operational Runbook (Mini)
- Document common errors, how to interpret metrics, and likely mitigations.
- Include steps to rotate tokens, bump limits, and change timeouts safely.
- Add a “first responder checklist” for high latency and elevated error rate.

### 18) Documentation & Examples
- Add a concise README that shows: building, running locally, calling the API, flipping on retries/idempotency, running load tests, opening dashboards.
- Include a “design notes” doc capturing what you learned (trade-offs, surprises, next steps).

---

## Test Plan (what to verify before calling it done)
- **Functional**: CRUD works; keys persist across process restarts (if persistent mode enabled).
- **Resilience**: Client retries only on retryable errors; idempotency prevents duplicates; deadlines enforced.
- **Performance**: Meets a small, documented target RPS with stable p95/p99 under your dev box constraints.
- **Observability**: Traces show client→server linkage; metrics populate dashboards; logs are structured & greppable.
- **Ops**: Service starts with one command; shuts down gracefully; config changes are straightforward.

---

## Extensions (optional if time allows)
- Pluggable middlewares (auth, rate-limit, logging) with an explicit order of execution.
- mTLS between client/server in local dev using a one-command CA bootstrap.
- Request batching on the client for small keys to explore throughput vs latency.
- Basic blue/green or canary procedure in your compose stack.

---

## Learning Outcomes You’ll Lock In
- Setting and propagating **deadlines**; designing **retry** policies; **idempotency** for safe writes.
- Reading **latency histograms** and traces to reason about tail latency.
- Building a minimal **runbook** and **dashboards** so you can operate what you build.

---
