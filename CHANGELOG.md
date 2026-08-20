<!--
This changelog should always be read on `master` branch. Its contents on other branches
does not necessarily reflect the changes.
-->

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [0.0.7] - 2026-08-20

### Added
- Multi-API-Key support for clusters (`Keys` + `KeyPolicy`) with integration tests.
- Provider/Model prefix routing support.
- Model pricing (model-price) menu and endpoints.
- Route rule `Cond` expression syntax validation.
- OpenAPI/InnerAPI schema integration tests.
- Centralized Redis quota operations in `model/quotacache`.
- Export `key_id` field in API-Key tokens.
- Added `AGENTS.md` for AI agent collaboration guidelines.

### Changed
- Rebranded organization migration: `infinity-ai-gateway` → `rainway-ai-gateway`.
- Unified RMB quota conversion using `github.com/bfenetworks/go-lib/quota`.
- Removed redundant `Currency` from exported `QuotaPlan` and updated InnerAPI docs.
- Normalized route rule field names to `snake_case`.
- Switched `cluster_conf` to native BFE `cluster_conf.AIConf`.
- Restructured `model` and `storage/rdb` package layouts to align with each other.

### Fixed
- `/api-keys/{id}/quota-plan` now returns the real quota balance.
- Prevent deletion of clusters/models referenced by route rules.
- Scan all route rules when checking cluster/model deletion references (#57).
- Validate model-pricing `limits` values as non-negative integers (#66).
- Various fixes for quota, Redis sync, and OpenAPI response details.

## [0.0.6] - 2026-08-05

### Added
- OpenAPI v0.3.0 upgrade: converged OpenAPI interfaces to API design v0.3.0.
- New endpoint `GET /route-tables` for route table listing.
- OpenAPI request parameter validation across endpoints.
- Auto-creation of the global route table during system initialization.
- Unit tests for `stateful`, `storage`, and all `model` packages with CI coverage gate.
- Integration test fixtures for `model_provider_type` and tool endpoints.

### Changed
- Refactored cluster model: aligned `instance_pool` model and related endpoints.
- Restructured test directory layout and updated integration test infrastructure.
- Split OpenAPI and InnerAPI design documents by module.
- Updated Chinese API design documents to v0.3.0.

### Fixed
- Fixed OpenAPI authorization flags: use `iauth.FA` instead of `iauth.FAP` for `/api-keys` and `/clusters` endpoints.
- Fixed concurrent map write in APIKey export.
- Fixed InnerAPI export issues for quota, routing, and rate-limiting policies.
- Fixed cluster deletion to check references from route rules.
- Fixed `sticky_sessions` default value to `CLIENT_IP_ONLY` and added validation for `hash_header` when `enabled=true`.
- Fixed API-Key, cluster, and AI-route issues (#42, #43, #44).
- Fixed integration tests: return JSON 404 for unmatched API paths and repaired failing test infrastructure.

## [0.0.4] - 2026-07-07

### Added
- Inner-API export endpoint `/inner-api/v1/configs/mod-body-process` for mod_body_process configuration with version management.
- Redis key existence check and creation logic in Entity update (PATCH) and full update (PUT) endpoints, ensuring quota keys are properly initialized.
- Redis key initialization in Entity create endpoint: when quota plan is unlimited, Redis key is initialized or reset to default value 100000000.

### Changed
- Optimized `allow_models` logic in exporter: token file Models is set to intersection of non-empty, non-* model lists from API-Key and all associated entities; if intersection is empty, Models is empty and Enabled is set to 2.
- Entity list interface uses BindForm for query parameters instead of requiring body parameters.
- Entity update interface supports PATCH (partial update) and PUT (full update) methods.

### Fixed
- API-Key create/update/full_update endpoints now return the created/updated API-Key data instead of nil.
- API-Key list endpoint returns proper format with `list` and `pagination`, supports filtering by `enabled`, `entity_id`, `unlimited_quota`.
- API-Key detail endpoint returns 404 when API-Key does not exist.
- API-Key `Enable` field JSON tag changed from `enable` to `enabled`, `UpdatedTime` from `updated_time` to `update_time`.
- Entity create endpoint adds required validation for `name` and existence check for `type`.
- Entity detail endpoint returns 404 when Entity does not exist.
- Entity delete endpoint checks for child entities and associated API-Keys before deletion.
- Entity list endpoint returns complete fields including `allow_models`, `block_models`, `quota_plan`, `rate_limit_policy`, `create_time`, `update_time`.
- Entity-Type create/update/list/one/delete endpoints fixed for proper return values and error handling.
- Reset quota endpoints return consistent format with `id`, `previous_quota`, `new_quota`, `balance`.
- HTTP status code handling: business errors (400-599) now return HTTP 200 with `ErrNum` field for business error code.
- Nil pointer dereference fixes in Entity and Entity-Type managers.
- Duplicate name/type_name detection with proper error codes.

## [0.0.3] - 2026-05-20

### Added
- **Quota Control & Rate Limiting**:
  - API-Key quota management with configurable quota amount, reset periods, and balance tracking.
  - API-Key rate limiting: TPM (Token Per Minute), RPM (Request Per Minute), and maximum concurrency limits.
  - Entity management with hierarchical structures (department/team/person), quota and rate limiting policies.
  - Entity-Type management with 1-5 level hierarchy relationships.
  - Quota balance synchronization: Redis-to-database synchronization with scheduled and manual reset support.
  - Unlimited quota mode for both API-Keys and Entities.

- **OpenAPI v1 Endpoints**:
  - `/api-keys`: Full CRUD with quota plan and rate limiting policy configuration.
  - `/api-keys/{id}/quota-plan`: Query API-Key quota plan with real-time balance.
  - `/api-keys/{id}/quota-plan/reset`: Reset API-Key quota balance.
  - `/entity-types`: Entity type CRUD endpoints.
  - `/entities`: Entity CRUD endpoints with hierarchical structure support.
  - `/entities/{id}/quota-plan`: Query Entity quota plan with real-time balance.
  - `/entities/{id}/quota-plan/reset`: Reset Entity quota balance.
  - `/models`: Get supported AI service model list.

- **Inner-API Endpoints**:
  - `/configs/mod-api-key`: Export API-Key configuration for BFE data-plane nodes.
  - `/configs/rate-limit-policy`: Export rate limiting policies for BFE enforcement.

- **Observability & Stability**:
  - Exception logger capturing panic information from main program, scheduled tasks, and HTTP handlers.
  - Global panic recovery mechanism.

- **Data Models**:
  - `QuotaPlan`: quota plan with unlimited, pass_when_no_enough_quota, quota, unit, reset_period fields.
  - `RateLimitPolicy`: rate limiting policy with enabled, rules (tpm, rpm, max_concurrency) fields.
  - `Entity`: entity with name, type, parent_id, allow_models, block_models fields.
  - `EntityType`: entity type with type_name, description, level fields.

## [0.0.1] - 2026-02-13

### Added
- First public release of AI Gateway API (control-plane component).
- OpenAPI v1 for managing gateway policies/configurations (products, clusters/subclusters, pools, domains, certificates, routes/forward rules, traffic scheduling, auth).
- AI route rules management for multi-model routing.
- API key management (enable/disable, quota, expiry, allowed models/subnets).
- Export endpoints for data-plane/Conf Agent configuration distribution.
- Built-in logging and Prometheus metrics endpoint on monitor port.
