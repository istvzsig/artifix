# Artifix — Architecture Decision Records

**Project:** Artifix  
**Module:** `github.com/istvzsig/artifix`  
**Status:** Architecture design accepted; implementation in progress  
**Last updated:** 2026-11-05

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [ADR-001: Artifact Identity and Content Integrity](#adr-001-artifact-identity-and-content-integrity)
- [ADR-002: Artifact Coordinates and Uniqueness](#adr-002-artifact-coordinates-and-uniqueness)
- [ADR-003: Artifact Immutability](#adr-003-artifact-immutability)
- [ADR-004: Semantic Versioning](#adr-004-semantic-versioning)
- [ADR-005: HTTP Upload Protocol](#adr-005-http-upload-protocol)
- [ADR-006: Binary Storage Strategy](#adr-006-binary-storage-strategy)
- [ADR-007: Content-Addressed Storage and Upload Lifecycle](#adr-007-content-addressed-storage-and-upload-lifecycle)
- [ADR-008: Artifact Deletion and Content Retention](#adr-008-artifact-deletion-and-content-retention)
- [Implementation Roadmap](#implementation-roadmap)
- [Open Questions](#open-questions)

---

# Architecture Overview

Artifix is a Go-based artifact registry.

Its primary responsibility is to accept, identify, store, retrieve, and manage immutable binary artifacts.

The initial implementation will be a **modular monolith**: one executable with clearly separated responsibilities.

The architecture should allow infrastructure components to be replaced without rewriting business logic.

### Core principles

1. Artifact metadata and binary content are separate concerns.
2. Artifact records are immutable after creation.
3. Content integrity is verified server-side.
4. Artifact coordinates are unique.
5. Binary objects are identified by their SHA-256 digest.
6. Business logic must not depend on infrastructure implementations.
7. Concurrency and consistency must be handled explicitly.
8. Infrastructure complexity should be introduced only when required.

---

# ADR-001: Artifact Identity and Content Integrity

**Status:** Accepted

## Context

Artifix needs to identify artifact records independently from the actual binary content they reference.

Two artifacts may contain identical binary data while representing different releases or coordinates.

We therefore need separate identities for artifact metadata and binary content.

## Decision

Use two distinct identifiers:

- **Artifact ID:** UUID identifying an artifact record.
- **Content digest:** SHA-256 hash identifying binary content.

The server calculates the digest during upload.

Client-provided digests are never authoritative.

## Consequences

- Different artifact records may reference identical content.
- Content integrity can be independently verified.
- Content deduplication can be implemented at the storage layer.
- Artifact identity is independent of physical storage location.

---

# ADR-002: Artifact Coordinates and Uniqueness

**Status:** Accepted

## Context

Artifacts need human-readable identifiers that can be used in API requests and registry operations.

## Decision

An artifact is identified by:

`namespace/name/version`

Example:

`istvzsig/artifix/1.0.0`

The following database-level uniqueness constraint will eventually enforce coordinate uniqueness:

`UNIQUE(namespace, name, version)`

## Expected behavior

| Situation                             | Result              |
| ------------------------------------- | ------------------- |
| New coordinate                        | Create artifact     |
| Existing coordinate, identical digest | Idempotent success  |
| Existing coordinate, different digest | Conflict            |
| Same content, different coordinates   | Allowed             |
| New version                           | New artifact record |

Uniqueness must ultimately be enforced by the metadata persistence layer, not solely through application-level checks.

## Consequences

- Artifact coordinates cannot be claimed by multiple records.
- Concurrent uploads must be handled safely.
- The in-memory implementation must preserve the same uniqueness contract expected from PostgreSQL.

---

# ADR-003: Artifact Immutability

**Status:** Accepted

## Context

Overwriting an existing artifact would make its identity unreliable and could break reproducibility.

A release should always refer to the same content.

## Decision

Artifacts are immutable after creation.

An existing coordinate cannot be overwritten with different content.

Publishing a new release requires a new artifact coordinate.

## Consequences

- Existing artifacts cannot have their binary content replaced.
- Identical uploads are treated as idempotent operations.
- Different content at an existing coordinate results in a conflict.
- Deletion and retention are handled separately.

---

# ADR-004: Semantic Versioning

**Status:** Accepted

## Context

Artifact versions need a consistent and predictable format.

Custom version parsing would introduce unnecessary complexity and compatibility risks.

## Decision

Artifix follows Semantic Versioning 2.0.0 strictly.

A well-tested Go SemVer library will be used instead of implementing a custom parser.

Valid examples:

- `1.0.0`
- `1.2.3`
- `2.0.0-rc.1`
- `1.4.0-beta.2`
- `3.0.0+build.123`

The canonical stored version does not include a leading `v`.

For example, `v1.2.3` is rejected.

### Build metadata

SemVer build metadata does not affect version precedence.

However, Artifix treats the complete version string as part of the artifact coordinate.

Therefore:

- `1.0.0+build.1`
- `1.0.0+build.2`

are distinct coordinates, even though they have equal SemVer precedence.

## Consequences

- Version validation follows a standard.
- Version ordering can use a dedicated library.
- Coordinate uniqueness is based on the exact canonical version string.
- No custom SemVer implementation is required.

---

# ADR-005: HTTP Upload Protocol

**Status:** Accepted

## Context

Artifix needs an upload protocol suitable for large binary artifacts.

Encoding binary content as JSON or Base64 would introduce unnecessary memory and processing overhead.

## Decision

Use raw binary HTTP PUT requests.

Endpoint:

`PUT /v1/artifacts/{namespace}/{name}/{version}`

The request body contains the raw binary content.

The `Content-Type` header specifies the media type.

Example:

`Content-Type: application/gzip`

The server streams the request body and calculates SHA-256 while receiving the content.

The implementation must not load the entire artifact into memory.

A configurable maximum upload size must be enforced.

## Expected HTTP responses

| Status                      | Meaning                                    |
| --------------------------- | ------------------------------------------ |
| `201 Created`               | New artifact created                       |
| `200 OK`                    | Existing artifact with identical content   |
| `400 Bad Request`           | Invalid coordinates or version             |
| `409 Conflict`              | Existing coordinate with different content |
| `413 Content Too Large`     | Upload exceeds configured limit            |
| `500 Internal Server Error` | Unexpected server failure                  |

## Consequences

- Large uploads can be streamed.
- Server-side integrity verification is possible.
- Upload memory consumption remains bounded.
- Repeated identical PUT requests are idempotent.
- Existing coordinates cannot be overwritten.

---

# ADR-006: Binary Storage Strategy

**Status:** Accepted

## Context

Artifact metadata and binary content have different storage requirements.

Metadata requires querying, indexing, and transactional consistency.

Binary storage requires efficient handling of potentially large immutable objects.

Combining both concerns would make the architecture harder to evolve.

## Decision

Separate metadata persistence from binary content storage.

Application logic depends on abstractions rather than concrete infrastructure implementations.

Storage implementation will evolve in stages.

### Planned stages

**Stage 1 — In-memory adapters**

- In-memory metadata repository.
- In-memory binary storage.
- No persistence across application restarts.

**Stage 2 — Persistent local storage**

- PostgreSQL for artifact metadata.
- Filesystem for binary objects.
- Temporary files and safe publication.

**Stage 3 — Object storage**

- S3-compatible object storage.
- PostgreSQL remains the metadata source of truth.

## Consequences

- Business logic remains independent of storage technology.
- Infrastructure can be replaced without rewriting application behavior.
- The first implementation remains simple and testable.
- Persistence is introduced only after the core behavior is established.

---

# ADR-007: Content-Addressed Storage and Upload Lifecycle

**Status:** Accepted

## Context

Binary content must have a stable identity independent of the artifact record that references it.

The upload process also spans metadata persistence and binary storage.

These operations cannot be assumed to participate in one atomic transaction.

A process crash or storage failure could otherwise leave incomplete artifacts or orphaned binary objects.

## Decision

Binary objects are stored using their SHA-256 digest as their identity.

Artifact records reference content objects.

Multiple artifact records may reference the same content object.

### Example storage layout

```text
objects/
└── sha256/
    ├── ab/
    │   └── ab34...f92
    └── cd/
        └── cd12...a81

staging/
└── upload-tmp-...
```

### Upload lifecycle

The logical artifact lifecycle is:

`PENDING → AVAILABLE`

The upload process follows these steps:

1. Receive the binary into temporary staging storage.
2. Calculate SHA-256 while streaming.
3. Validate the artifact coordinates and upload constraints.
4. Reserve the artifact coordinate with a `PENDING` metadata record.
5. Publish the binary under its content digest.
6. Transition the artifact metadata to `AVAILABLE`.

Artifacts are downloadable only when their state is `AVAILABLE`.

### Failure handling

| Failure                        | Expected behavior                                  |
| ------------------------------ | -------------------------------------------------- |
| Interrupted upload             | Temporary content can be cleaned up                |
| Coordinate reservation fails   | Staged content is cleaned up                       |
| Binary publication fails       | Artifact remains pending until recovery or cleanup |
| Crash after publication        | Pending record can be reconciled                   |
| Metadata finalization succeeds | Artifact becomes available                         |

A reconciliation mechanism will eventually inspect incomplete uploads and recover or abandon them.

### Concurrent publication

Publishing the same digest concurrently must be safe.

An existing content object must never be overwritten with different bytes.

If the object already exists, its integrity must be verified before reuse.

## Consequences

- Binary content has a stable identity.
- Deduplication is possible at the storage layer.
- Multiple artifact records can reference identical content.
- Partial failures are explicitly represented.
- Recovery does not depend on pretending that PostgreSQL and binary storage share an atomic transaction.

---

# ADR-008: Artifact Deletion and Content Retention

**Status:** Accepted

## Context

Multiple artifact records may reference the same binary content.

Immediately deleting a binary when one artifact is deleted could break other artifacts that still reference it.

Artifact deletion and physical content deletion therefore need separate lifecycles.

## Decision

Artifact deletion uses soft deletion.

The artifact record is retained, with a `deleted_at` timestamp.

Normal API queries and downloads exclude soft-deleted artifacts.

Binary objects are not immediately deleted.

A garbage collection process may physically delete content only when it is no longer referenced by any eligible artifact.

Physical deletion must follow a configurable retention or grace period.

### Example

Before deletion:

```text
Artifact A ──┐
             ├── SHA256: ab34...f92
Artifact B ──┘
```

After deleting Artifact A:

```text
Artifact A (deleted) ──┐
                       ├── SHA256: ab34...f92
Artifact B (available) ┘

Binary object: retained
```

### Restoration

Soft-deleted artifacts may be restored by clearing `deleted_at`, provided their content object still exists and passes integrity verification.

## Consequences

- Artifact records remain available for auditing.
- Shared binary content is protected.
- Logical deletion is independent of physical storage deletion.
- Garbage collection can run independently.
- Restoration is possible while the underlying content remains available.

---

# Implementation Roadmap

The implementation will begin with in-memory adapters.

We will establish the domain rules and application behavior before introducing persistent infrastructure.

## Phase 1 — Domain model

Implement:

- Artifact entity.
- Artifact ID.
- Artifact coordinates.
- SemVer validation.
- SHA-256 digest representation.
- Artifact lifecycle states.
- Domain errors and invariants.

The domain must not depend on HTTP, PostgreSQL, filesystem storage, or S3.

## Phase 2 — Application service

Implement the business operations:

- Upload artifact.
- Enforce coordinate uniqueness.
- Enforce immutability.
- Handle idempotent uploads.
- Handle content deduplication.
- Retrieve artifact metadata.
- Soft-delete and restore artifacts.

Define interfaces where the application service needs to interact with external responsibilities.

## Phase 3 — In-memory adapters

Implement:

- In-memory artifact repository.
- In-memory content storage.
- Safe concurrent coordinate reservation.
- Content-addressed binary publication.

The adapters must preserve the contracts expected from future persistent implementations.

## Phase 4 — HTTP API

Implement the HTTP transport layer:

- Raw binary PUT endpoint.
- Artifact retrieval.
- Artifact deletion.
- Appropriate HTTP status codes.
- Request validation.
- Upload size limits.
- Error mapping.

HTTP handlers should delegate business operations to the application service.

## Phase 5 — Testing

Implement:

- Domain unit tests.
- Table-driven tests.
- Application service tests.
- HTTP handler tests using `httptest`.
- Concurrent upload tests.
- Idempotency tests.
- Failure scenario tests.
- Race detector checks.

Run:

```bash
go test ./...
go test -race ./...
go vet ./...
```

## Phase 6 — Persistent infrastructure

Only after the in-memory implementation is working:

- Introduce PostgreSQL metadata persistence.
- Add versioned SQL migrations.
- Introduce filesystem-backed binary storage.
- Implement crash recovery and reconciliation.
- Add garbage collection.

S3-compatible object storage can follow when required.

---

# Open Questions

The following decisions remain intentionally unresolved.

They should be addressed when the relevant implementation stage requires them.

1. **Authentication and authorization:** How will users and namespaces be authenticated and authorized?
2. **Namespace ownership:** Who can create, publish, delete, and restore artifacts within a namespace?
3. **Upload limits:** What should the default maximum artifact size be?
4. **Pending upload expiration:** How long should incomplete uploads remain recoverable?
5. **Garbage collection:** What retention period and safety checks should apply before physical deletion?
6. **Audit history:** Should artifact deletion and restoration events be recorded separately?
7. **Observability:** Which logs, metrics, and traces are required for production operation?

---

# ADR Change Management

Architecture decisions should be recorded when they materially affect the system's structure, behavior, or constraints.

Each ADR should contain:

- Number and title.
- Status.
- Context.
- Decision.
- Consequences.

Accepted decisions should not be silently rewritten.

If an accepted decision changes, its status and reasoning should be updated explicitly, including the reason for superseding the original decision.

Small implementation details do not require a new ADR unless they introduce a meaningful architectural trade-off.
