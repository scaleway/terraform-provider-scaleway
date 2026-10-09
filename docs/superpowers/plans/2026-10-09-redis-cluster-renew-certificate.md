# Redis Cluster Renew Certificate Action — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans (or subagent-driven-development). Steps use checkbox (`- [ ]`) syntax.

**Goal:** Add `scaleway_redis_cluster_renew_certificate` Framework Action wrapping `RenewClusterCertificate`.

**Architecture:** Clone RDB renew-certificate Action into `internal/services/redis`, zonal locality, reuse `newAPI` + `waitForCluster`.

**Tech Stack:** Go, terraform-plugin-framework actions, scaleway-sdk-go redis/v1, VCR cassettes.

**Spec:** `docs/superpowers/specs/2026-10-09-redis-cluster-renew-certificate-design.md`

## Global Constraints

- ExactlyOneOf N/A — only `cluster_id` (+ optional `zone`, `wait`)
- No new helpers; reuse existing redis/locality/meta helpers
- Skip OpenTofu in acc tests
- Signed English one-line commits; no Cursor attribution
- Cassette recorded separately (`TF_UPDATE_CASSETTES=true`) — ship test without cassette if recording not available in this session; note for user

---

### Task 1: Action implementation + registration

**Files:**
- Create: `internal/services/redis/descriptions/cluster_renew_certificate_action.md`
- Create: `internal/services/redis/cluster_renew_certificate_action.go`
- Modify: `provider/framework.go` (add `redis.NewClusterRenewCertificateAction` in Actions list, alphabetically near other redis if any / after rdb block before s2svpn)

- [ ] **Step 1:** Write description markdown (mirror RDB + Redis docs warning about client certs)
- [ ] **Step 2:** Implement action (Configure/Metadata/Schema/Invoke) mirroring vpcgw zonal + RDB renew
- [ ] **Step 3:** Register in `framework.go`
- [ ] **Step 4:** `go build ./...` (or `make build`)

### Task 2: Acceptance test

**Files:**
- Create: `internal/services/redis/cluster_renew_certificate_action_test.go`

- [ ] **Step 1:** TestAccActionRedisClusterRenewCertificate_Basic with after_create trigger, wait=true, CheckDestroy, ready check via GetCluster
- [ ] **Step 2:** Compile tests: `go test -c -o /dev/null ./internal/services/redis`
- [ ] **Step 3:** Commit

### Task 3: Docs/cassette follow-up (if env allows)

- [ ] Record cassette with `TF_UPDATE_CASSETTES=true` when credentials available
- [ ] `make docs` if required by repo for new actions
