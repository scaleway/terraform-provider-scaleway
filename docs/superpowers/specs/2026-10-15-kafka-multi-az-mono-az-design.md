# Design: Kafka multi_az / mono_az on `scaleway_kafka_cluster`

**Issue:** [#4491](https://github.com/scaleway/terraform-provider-scaleway/issues/4491)  
**Date:** 2026-10-15  
**Status:** Approved (brainstorming)

## Goal

Expose Scaleway Kafka Multi-AZ / Mono-AZ placement on the existing SDKv2 resource `scaleway_kafka_cluster`, matching the API oneof and CLI flags.

## API / SDK facts

- Package: `github.com/scaleway/scaleway-sdk-go/api/kafka/v1alpha1`
- `CreateClusterRequest` and `Cluster` both carry:
    - `MultiAz *ClusterMultiAZDetails` (empty struct)
    - `MonoAz *ClusterMonoAZDetails` with `Zone *scw.Zone`
    - Comment: *Precisely one of MultiAz, MonoAz must be set*
- Absent from `UpdateClusterRequest` → create-only → Terraform **ForceNew**
- Resource is **regional** (`region`); mono mode pins a **zone** inside that region
- No equivalent oneof exists on other Scaleway TF resources; CLI exposes `multi-az` / `mono-az.zone`
- AWS/GCP/Azure TF Kafka surfaces differ (subnets / always-multi / deprecated `zone_redundant`) — not reused

## Schema

Flat attributes (approach aligned with the issue):

| Attribute | Type | Flags | Notes |
|---|---|---|---|
| `multi_az` | bool | Optional, Computed, ForceNew | `ConflictsWith: ["mono_az_zone"]` |
| `mono_az_zone` | string | Optional, Computed, ForceNew | Zone name e.g. `fr-par-1`; `ConflictsWith: ["multi_az"]` |

### Create expand rules

1. If `multi_az` is set **and true** → `req.MultiAz = &kafkaapi.ClusterMultiAZDetails{}`
2. Else if `mono_az_zone` is set → `req.MonoAz = &kafkaapi.ClusterMonoAZDetails{Zone: &zone}`
3. Else (neither set) → **default mono-AZ** with zone from `meta.ExtractZone(d, m)` (provider default zone)
4. If `multi_az` is explicitly `false` and `mono_az_zone` is unset → return a clear create error (do not treat false as multi or silent mono)

### Read flatten rules

In `setClusterState` (shared by resource + data source):

- If `cluster.MultiAz != nil` → set `multi_az = true` and `mono_az_zone = ""`
- If `cluster.MonoAz != nil` → set `multi_az = false` and `mono_az_zone` from `MonoAz.Zone`
- If API returns neither (legacy) → do not invent values (leave unset)

### Update / Delete / Import

- Update: no AZ fields (ForceNew handles replace)
- Delete: unchanged
- Import: passthrough; Read fills Computed attrs from API

## Helpers

- Prefer inline expand in create + flatten in `setClusterState` / `types.go` only if the branch is non-trivial
- Reuse `meta.ExtractZone` for the default path; no new generic locality helper
- HARD GATE (`scaleway-helpers`): grep before any new `func`; product-specific expand/flatten stays under `internal/services/kafka/`

## Tests

Acceptance tests + VCR cassettes under `internal/services/kafka/testdata/`:

1. **Basic** — neither attr set → assert mono default (`mono_az_zone` set after apply from API/provider zone)
2. **MultiAZ** — `multi_az = true` → assert `multi_az = true`
3. **MonoAZ** — `mono_az_zone = "fr-par-1"` → assert zone

Re-record cassettes (`TF_UPDATE_CASSETTES=true`) then compress. Update data-source cluster checks if state now exposes the new attrs.

## Docs

- Resource description / schema descriptions
- Example HCL for multi and mono
- Note: omitting both defaults to mono-AZ on the provider default zone

## Out of scope

- Nested blocks (`multi_az {}` / `mono_az { zone }`)
- Framework migration of kafka
- Making AZ mode updatable
- Changing the default to multi-AZ later
- Product-catalog `is_multi_az` metadata

## Risks

- Old cassettes omit AZ fields; live API now may require oneof → Basic must send mono default and cassettes must be re-recorded
- `GetOk("multi_az")` cannot distinguish unset vs `false` → explicit error on false without zone
- Provider default zone must belong to the cluster `region` (same constraint as other zonal defaults)
