# Design: `scaleway_redis_cluster_renew_certificate` Action

**Issue:** [#4465](https://github.com/scaleway/terraform-provider-scaleway/issues/4465)  
**Date:** 2026-10-09  
**Status:** Approved for implementation planning

## Problem

Redis TLS certificate renewal is an imperative day-2 API (`RenewClusterCertificate`). It cannot be expressed as a desired-state attribute on `scaleway_redis_cluster`. After renewal, clients must download and use the new certificate; the previous certificate no longer works.

This matches how Terraform Actions are used elsewhere (RDB renew certificate, instance power, flush, etc.). A migrate Action was rejected (#3917) because migration is already covered by resource Update and would risk state drift.

## Goals

- Expose `RenewClusterCertificate` as a Framework Action.
- Mirror `scaleway_rdb_instance_renew_certificate` (naming and schema shape), adapted to Redis zonality.
- Reuse existing Redis helpers; no new shared helpers.

## Non-goals

- Redis migrate / scale / version upgrade Actions.
- Updating Terraform state of `scaleway_redis_cluster` from the Action.
- OpenTofu support (skip tests when running OpenTofu, same as other Actions).

## Approach

Clone the RDB renew-certificate Action pattern into the `redis` service package (dedicated files). Reject a mega Redis `action` enum and reject a shared RDB/Redis abstraction (region vs zone).

## Behavior

| Item | Detail |
|---|---|
| Type name | `scaleway_redis_cluster_renew_certificate` |
| API | `redis.API.RenewClusterCertificate` |
| `cluster_id` | Required. Plain UUID or zonal ID (`{zone}/{id}`). |
| `zone` | Optional. `zonal.SchemaAttribute`. Resolution order: attribute → parse from `cluster_id` → provider default zone. |
| `wait` | Optional bool. If true, call existing `waitForCluster` with `defaultRedisClusterTimeout`. |
| Errors | Diagnostics on missing/invalid inputs, API failure, wait failure. |
| Docs | Warn that old certificates stop working and clients must be updated (SDK note). |

## Files

| Path | Role |
|---|---|
| `internal/services/redis/cluster_renew_certificate_action.go` | Action implementation |
| `internal/services/redis/cluster_renew_certificate_action_test.go` | Acceptance test |
| `internal/services/redis/descriptions/cluster_renew_certificate_action.md` | Embedded description |
| `provider/framework.go` | Register `redis.NewClusterRenewCertificateAction` |
| `internal/services/redis/testdata/action-redis-cluster-renew-certificate-basic.cassette.yaml` | VCR cassette (record separately) |

## Helpers (reuse only)

- `newAPI`
- `waitForCluster`
- `locality.ExpandID`
- `zonal.ParseID` / `zonal.SchemaAttribute`
- `meta.Meta` for default zone

## Testing

- Pattern: `TestAccActionRedisClusterRenewCertificate_Basic`
- Skip if `acctest.IsRunningOpenTofu()`
- Create `scaleway_redis_cluster`, `lifecycle.action_trigger` `after_create` → renew action with `wait = true`
- Check: cluster present / status ready (same spirit as RDB renew check)
- `CheckDestroy` for the cluster
- Cassette name derived from test name per `TESTING.md` / `scaleway-testing`

## Success criteria

- Action registered and invocable via `action_trigger` and `-invoke`
- Acc test passes against cassette
- Docs generated / description present
- No new reinvented helpers
