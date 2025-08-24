This is a project-based path that takes you from “I know how to code” to “I can reason about and build real distributed systems.” The plan is sequenced so you can jump in where it feels right. Each project lists goals, why it matters, a recommended stack (favoring widely used tech), milestones, testing/validation ideas, and stretch goals.

---

# Toolbelt (used across most projects)

- **Languages**: Go (primary for systems + infra), Java (JDK 21+, Spring Boot where useful), Rust (optional where you want safety/perf), Python (for load gen + quick scripts).
- **Protocols**: gRPC + Protobuf, HTTP/2, basic TCP/UDP.
- **Infra**: Docker + Docker Compose; Kubernetes (kind or minikube) as you advance.
- **Observability**: OpenTelemetry (traces, metrics, logs), Prometheus, Grafana.
- **Testing**: Integration tests (Testcontainers), Jepsen-like fault tests (you’ll write focused versions), k6/Vegeta for load, chaos tooling (chaos-mesh or simple kill/iptables scripts).
- **Persistence Options**: BadgerDB/RocksDB (via bindings), Postgres, Redis.
- **Code quality**: static checks (golangci-lint, SpotBugs/Checkstyle), fuzz testing (Go 1.18+ built-in, Jazzer for Java).

---

# Project 0 — Systems & Networking Warm-Up (2–5 days)
**Goal:** Rebuild the muscles you’ll use later: serialization, RPC, timeouts, backoff, basic tracing.

**Deliverable:** A tiny **key–value microservice** with:
- `PUT/GET/DELETE` RPCs over gRPC.
- Client with retries (exponential backoff + jitter), deadlines, and idempotency tokens.
- Basic request tracing and metrics.

**Stack:** Go + gRPC + Protobuf, OpenTelemetry SDK, Prometheus.

**Milestones**
1. Protobuf IDL + server/client scaffolding.
2. Deadlines, retries, and idempotent semantics for `PUT`.
3. OTel traces (client → server) and Prometheus counters/histograms.

**Validation**
- Unit tests for client retry behavior (fake server).
- Latency/throughput chart under k6.

**Stretch**
- Pluggable storage (in-mem vs BadgerDB).
- mTLS between client and server.

---

# Project 1 — **Consistent-Hashing Distributed Cache** with Gossip Membership
**Why:** Teaches **partitioning/sharding**, **membership**, and **failure detection**.

**Deliverable:** A memcached-like cache horizontally scaled across N nodes with:
- Virtual nodes + consistent hashing ring.
- Gossip-based membership (SWIM-style or simplified) to detect joins/leaves/failures.
- Replication factor (e.g., 3) and hinted handoff on node loss.
- Simple admin page/endpoint showing ring state.

**Stack:** Go (net/http + gRPC), Hash ring lib or your own, Serf-like gossip (write your own light version).

**Milestones**
1. Single node cache + gRPC API.
2. Consistent hashing with virtual nodes.
3. Gossip membership + ring rebalancing.
4. Replication + hinted handoff.

**Validation**
- Kill nodes under load; verify read/write availability and key movement bounds.
- Metrics: cache hit rate, ring churn, replication lag.

**Stretch**
- Add **Rendezvous hashing** mode and compare movement under scaling events.
- Write a small **proxy/sidecar client** that can be dropped next to apps.

**What you learn:** Sharding, CAP trade-offs, eventual consistency basics, membership & failure detection.

---

# Project 2 — **Raft-Backed Replicated KV Store** (Single Shard, Strong Consistency)
**Why:** Consensus is table stakes. Raft is the most approachable.

**Deliverable:** A multi-node KV store with:
- Raft for leader election & log replication (either implement simplified Raft or embed a battle-tested library and focus on integration).
- Linearizable `PUT/GET`, snapshots, log compaction.
- Read-index/lease-read optimization for fast reads.

**Stack:** Go + an embeddable Raft (e.g., `etcd/raft`) or implement a learning Raft; BadgerDB for state; gRPC.

**Milestones**
1. Raft node + persistent log, leader election.
2. Replicate `PUT` commands; apply to state machine.
3. Snapshots + compaction.
4. Linearizable reads with read index.

**Validation**
- Deterministic tests for split-brain, leader failover, partition and heal.
- Latency SLOs before/after snapshotting.

**Stretch**
- **Joint consensus** for membership changes.
- **Lease reads** and benchmark vs read-index.

**What you learn:** Consensus, replicated logs, strong consistency, durability & snapshots.

---

# Project 3 — **Kafka-Lite**: Commit Log + Replication + Consumer Groups
**Why:** Logs are the backbone of many distributed systems. You’ll internalize **append-only logs**, **segmenting**, **replication**, **acks**, **consumer offsets**, and **backpressure**.

**Deliverable:** A “just enough” Kafka:
- Topics/partitions, segment files with index + sparse index, fsync policy.
- Producer acks (0/1/all with quorum), idempotent producer (sequence numbers).
- ISR (in-sync replicas) tracking and leadership.
- Consumer groups with committed offsets and rebalancing.

**Stack:** Go or Java (Go is fine; Java lets you borrow from the Kafka mental model), Raft or your own leader election for controller.

**Milestones**
1. Single broker append-only log w/ segment roll + index.
2. Partition replication + acks.
3. Controller to assign partitions and track ISR.
4. Consumer group coordination + rebalances.

**Validation**
- Produce at high throughput; verify ordering per partition.
- Kill leader mid-produce; ensure durability based on ack mode.
- Observe consumer lag metrics & rebalance times.

**Stretch**
- Tiered storage (older segments to object storage).
- Exactly-once semantics (EOS) with idempotent producers + transactional writes.

**What you learn:** Log-centric design, replication quorums, durability/latency trade-offs, group coordination.

---

# Project 4 — **MapReduce-Mini** (Distributed Batch Compute)
**Why:** Teaches **task scheduling**, **checkpointing**, **fault tolerance**, **data locality**, and **stragglers**.

**Deliverable:** A master/worker system that:
- Splits input into shards, assigns map tasks, shuffles to reduce tasks.
- Heartbeats, retries, speculative execution for slow workers.
- Pluggable compute functions (user supplies map/reduce in a sandboxed plugin).

**Stack:** Go (plugins via RPC/grpc + WASM sandbox optional), local disk for shuffle, Postgres/Redis for state if needed.

**Milestones**
1. Map only; success/failure + retry.
2. Shuffle + reduce; deterministic partitioning by key.
3. Heartbeats + speculative execution.
4. Checkpointing and resume.

**Validation**
- Inject random worker crashes; job must finish.
- Metrics: task runtime distribution, speculative save.

**Stretch**
- Data locality (prefer sending tasks to nodes with input data).
- K8s deployment with horizontal autoscaling by queue depth.

**What you learn:** Distributed coordination, idempotence, scheduling, backpressure, and failure handling.

---

# Project 5 — **Distributed Transactions Two Ways**: 2PC and Saga
**Why:** Almost every real app needs cross-service consistency at some point.

**Deliverable:** A sample “banking” domain (accounts, ledger, notifications) implemented twice:
1. **2PC** coordinator with XA-like prepare/commit across microservices.
2. **Saga** orchestration with compensating actions, outbox pattern, and idempotent handlers.

**Stack:** Java + Spring Boot (widely used in industry for service boundaries), Postgres per service, gRPC between services, Kafka-Lite from Project 3 (or Kafka) for the outbox/event bus.

**Milestones**
- 2PC: coordinator, participants, prepare/commit, failure cases.
- Saga: orchestrator, compensation, timeouts, dead-letter queue.
- Outbox pattern to ensure “write + publish” atomicity.

**Validation**
- Bank transfer chaos tests: drop messages, kill nodes, observe invariants.
- Verify no double-debits and eventual balance correctness.

**Stretch**
- Anti-entropy reconciliation job that detects/repairs drift.
- Compare latency/availability between 2PC and Saga under failure.

**What you learn:** ACID vs BASE, coordination costs, idempotence, exactly-once delivery *in practice*.

---

# Project 6 — **CRDT-Backed Collaborative Doc Service**
**Why:** Explore **conflict-free replicated data types** for **AP-side** systems and offline/merge semantics.

**Deliverable:** Real-time collaborative doc (or todo list) with:
- Client-side CRDT (e.g., sequence CRDT/WOOT/RGA or a library).
- Server as relay + durable store; multi-region simulated with 2+ servers.
- Causal metadata (vector clocks or dotted version vectors).
- Offline edits that merge without central coordination.

**Stack:** TypeScript (client) + Go or Rust (server), WebSockets for realtime, RocksDB/BadgerDB for server persistence.

**Milestones**
1. Single server CRDT edits + persistence.
2. Multi-server replication w/ causal broadcast.
3. Offline editing + conflict resolution demo.

**Validation**
- Fuzz random concurrent edits; verify convergence property.
- Time travel/debug views to inspect causal history.

**Stretch**
- Delta-state CRDTs to cut bandwidth.
- Attach access-control lists and per-doc encryption.

**What you learn:** Availability-first design, causality, eventual consistency, correctness properties like convergence/commutativity.

---

# Project 7 — **HDFS/S3-Lite**: Chunked Distributed File Store
**Why:** Practice **metadata vs data plane separation**, **chunk replication**, **rebalancing**, **recovery**, and **client-side striping**.

**Deliverable:**  
- **NameNode** (metadata): file tree, chunk mapping, lease/locks for writers.  
- **DataNodes**: store fixed-size chunks, replicate, checksum, periodic reports.  
- Client library that splits/merges files, retries reads across replicas.

**Stack:** Go (great for IO + concurrency), Raft for active/standby NameNode, DataNodes with gRPC.

**Milestones**
1. Write path: client splits → chunk servers store + checksum.
2. Read path with replica selection + retries.
3. Heartbeats, replication factor, rebalancer on disk pressure.
4. NameNode HA (leader + follower) + edit logs and checkpoints.

**Validation**
- Corrupt random chunks; verify repair.
- Lose a DataNode; verify re-replication and availability.

**Stretch**
- Erasure coding for cold data.
- S3-compatible API shim.

**What you learn:** Metadata scaling, durability and repair, large object IO patterns.

---

# Project 8 — **Streaming Analytics Pipeline** with Windows/State/Exactly-Once
**Why:** Ties together log ingestion, stateful stream processing, and sinks with end-to-end semantics.

**Deliverable:**  
- Ingestion: your Kafka-Lite (or Kafka) → **Flink** or **Kafka Streams** job.  
- Stateful operators with keyed windows, timers, and exactly-once sinks (to Postgres or S3-Lite).  
- Rolling deploys with state savepoints.

**Stack:**  
- **Option A (Java):** Apache Flink (very common in industry).  
- **Option B (Java):** Kafka Streams/Spring Cloud Stream.  
- **Option C (Go):** Benthos/Goka (smaller ecosystem but viable).  

**Milestones**
1. Define a streaming job (e.g., fraud detection with session windows).
2. State snapshots + restore on upgrade.
3. EOS sink via two-phase commit sink or transactional producer.

**Validation**
- Kill/restart job during processing; verify zero duplicates and no loss.
- Backpressure test: slow sink, monitor watermarks/latency.

**Stretch**
- Side outputs for late data; DLQ.
- Multi-job pipeline with schema evolution (Protobuf versioning).

**What you learn:** Stateful streaming, time semantics (event vs processing time), exactly-once.

---

# Cross-Cutting “Professional” Layers (add to any project)

- **Security**: mTLS everywhere, rotating certs, RBAC/ABAC enforcement at the gateway.
- **Resilience**: Circuit breakers (resilience4j for Java, gobreaker for Go), bulkheads, rate limiting (token bucket in Redis with Lua).
- **Config & Discovery**: Service discovery (DNS SRV or Consul), dynamic config via etcd/Consul.
- **CI & Deploys**: GitHub Actions; Canary deploys on kind/minikube; Helm charts for each service.
- **SRE Hygiene**: SLOs, golden signals, error budgets; runbooks for common incidents.

---

# Suggested Sequencing (pick your entry point)

1. **0 → 1 → 2** gives you partitioning + consensus foundation.  
2. From there, branch to **3 (logs)** or **4 (batch compute)** based on interest.  
3. Then **5 (transactions)** to understand cross-service consistency.  
4. Add **6 (CRDTs)** to balance your CAP instincts.  
5. Capstone with **7 (DFS)** and **8 (streaming)** as production-scale patterns.

---

# Repo & Project Structure (template)

```
/dist-projects
  /proj0-kv-rpc
    /proto
    /cmd/server
    /cmd/client
    /pkg/kv
    /internal/otel
    docker-compose.yml
  /proj1-hash-cache
  /proj2-raft-kv
  /proj3-kafka-lite
  /proj4-mapreduce
  /proj5-tx-2pc-saga
  /proj6-crdt-collab
  /proj7-fs-lite
  /proj8-streaming
  /infra
    /k8s
    /grafana-dashboards
    /prometheus
  /scripts (loadgen, chaos, netem)
```

---

# How to choose languages per project (short rationale)

- **Go**: Dominant in infra/distributed tooling (K8s, etcd, Consul, NATS). Great stdlib networking, easy deployment.
- **Java**: Ubiquitous in data infra and stream processing (Kafka, Flink, big shops). Rich ecosystem for transactions and frameworks.
- **Rust** (optional): Where zero-copy IO and memory safety matter (high-perf components, protocol libs).
- **Python**: Keep for **testing**, load generation, and orchestration scripts—not for the core systems.

---

# Exit Criteria (what “done” feels like)

- You can explain and defend design trade-offs across **consistency, availability, latency, and cost**.
- You’ve implemented both **CP** (Raft KV) and **AP** (CRDT) systems and can reason when to use each.
- You can trace requests end-to-end, read metrics to debug tail latency, and design **retries** and **idempotency** correctly.
- You can discuss **quorums, snapshots, compaction, backpressure, rebalancing, partition tolerance,** and **schema evolution** comfortably.

---

