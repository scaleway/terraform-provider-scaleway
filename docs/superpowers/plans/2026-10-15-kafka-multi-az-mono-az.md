# Kafka multi_az / mono_az Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (inline; user requested implement). Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expose Multi-AZ / Mono-AZ on `scaleway_kafka_cluster` per `docs/superpowers/specs/2026-10-15-kafka-multi-az-mono-az-design.md` and issue #4491.

**Architecture:** Flat Optional+Computed+ForceNew attrs `multi_az` and `mono_az_zone` with ConflictsWith; create expand (default mono via `meta.ExtractZone`); flatten in `setClusterState`.

**Tech Stack:** Terraform SDKv2, scaleway-sdk-go kafka/v1alpha1, VCR acceptance tests.

## Global Constraints

- Stay on SDKv2 for kafka (existing service)
- ForceNew for both AZ attrs (not on UpdateClusterRequest)
- Default when neither set: MonoAz with provider default zone
- `multi_az = false` without zone → create error
- Reuse helpers (`meta.ExtractZone`, `scw.ParseZone`); no new generic helpers
- One English sentence per signed commit (`git commit -S`)
- Never `git add -A`

## File map

- Modify: `internal/services/kafka/cluster.go` — schema, create expand, flatten
- Modify: `internal/services/kafka/types.go` — flatten AZ helpers if needed
- Modify: `internal/services/kafka/cluster_test.go` — Basic assert + MultiAZ + MonoAZ tests
- Modify: `docs/resources/kafka_cluster.md` + `examples/resources/scaleway_kafka_cluster/`
- Re-record: `internal/services/kafka/testdata/*.cassette.yaml`

---

### Task 1: Schema + create expand + read flatten

**Files:**

- Modify: `internal/services/kafka/cluster.go`
- Modify: `internal/services/kafka/types.go` (optional flatten helpers)
- [ ] **Step 1:** Add schema attrs `multi_az` / `mono_az_zone`
- [ ] **Step 2:** Add `expandClusterAvailability` used in create
- [ ] **Step 3:** Flatten in `setClusterState`
- [ ] **Step 4:** `go test ./internal/services/kafka -count=1` (compile)
- [ ] **Step 5:** Commit

### Task 2: Acceptance tests

**Files:**

- Modify: `internal/services/kafka/cluster_test.go`
- [ ] **Step 1:** Assert AZ on Basic; add `TestAccCluster_MultiAZ` and `TestAccCluster_MonoAZ`
- [ ] **Step 2:** Replay if cassettes exist / note re-record needed
- [ ] **Step 3:** Commit

### Task 3: Docs + examples

**Files:**

- Modify: `docs/resources/kafka_cluster.md`
- Create: `examples/resources/scaleway_kafka_cluster/multi_az.tf`
- Create: `examples/resources/scaleway_kafka_cluster/mono_az.tf`
- [ ] **Step 1:** Document attrs + examples
- [ ] **Step 2:** Commit

### Task 4: Cassettes (when credentials available)

- [ ] Re-record Basic/MultiAZ/MonoAZ/data-source with `TF_UPDATE_CASSETTES=true`
- [ ] Compress with `go run ./cmd/vcr-compressor …`
- [ ] Commit cassettes
