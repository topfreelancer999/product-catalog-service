# Product Catalog Service

This repository contains a **Product Catalog Service** implemented in Go, following **Clean Architecture**, **Domain-Driven Design (DDD)**, **CQRS**, and the **Golden Mutation Pattern**.  
It is designed to closely match production-grade patterns used with **Google Cloud Spanner**, **gRPC**, and **transactional outbox**.

---

## Tech Stack

- **Go** 1.21+
- **gRPC + Protocol Buffers**
- **Google Cloud Spanner** (Emulator for local dev)
- **CommitPlan** for transactional mutations
- **math/big** for precise money calculations
- **Testify** for testing

---

## Project Structure (High Level)

```
cmd/server          -> Service entrypoint (main.go)
internal/app        -> Domain, usecases, queries
internal/services   -> Dependency injection (options.go)
internal/transport  -> gRPC handlers
internal/pkg        -> Shared infra (clock, committer)
proto/              -> gRPC API definitions
migrations/         -> Spanner DDL
tests/e2e           -> End-to-end tests
```

---

## Prerequisites

- Docker + Docker Compose
- Go 1.21+
- gcloud CLI (for Spanner emulator tooling)
- make

---

## Local Development Setup

### 1. Start Spanner Emulator

```bash
docker-compose up -d
```

Set environment variable (required):

```bash
export SPANNER_EMULATOR_HOST=localhost:9010
```

---

### 2. Configure gcloud for Emulator (one-time)

```bash
gcloud config configurations create emulator || true
gcloud config set auth/disable_credentials true
gcloud config set project test-project
gcloud config set api_endpoint_overrides/spanner http://localhost:9020/
```

---

### 3. Create Spanner Instance & Database

```bash
gcloud spanner instances create test-instance   --config=emulator-config   --description="Spanner Emulator"   --nodes=1
```

```bash
gcloud spanner databases create product_catalog   --instance=test-instance
```

---

### 4. Run Database Migrations

```bash
make migrate
```

This applies the schema from:

```
migrations/001_initial_schema.sql
```

---

## Running the Service

### Start gRPC Server

```bash
make run
```

The server will start on:

```
localhost:50051
```

Reflection is enabled, so you can use `grpcurl` or Evans.

---

## Running Tests

### Run All Tests (including E2E)

```bash
make test
```

Tests use:
- Real Spanner emulator
- Real repositories
- Real usecases (no mocks)

---

## Makefile Commands

```bash
make up        # Start Spanner emulator
make migrate   # Run DB migrations
make test      # Run all tests
make run       # Start gRPC server
```

---

## Design Decisions & Trade-offs

### Domain Purity
- Domain layer is **pure Go**
- No context, no database, no protobuf imports
- Business rules enforced at aggregate level

**Trade-off:**  
More boilerplate, but extremely testable and safe.

---

### CQRS
- Commands go through aggregates + CommitPlan
- Queries use read models and DTOs directly

**Trade-off:**  
Slight duplication of models, but much better read performance and clarity.

---

### Golden Mutation Pattern
- Repositories only **return mutations**
- Usecases apply CommitPlan
- Guarantees atomic writes and consistent outbox events

**Trade-off:**  
More explicit code, but zero hidden side effects.

---

### Transactional Outbox
- Domain emits intent events
- Usecases enrich and persist events in same transaction
- Safe for async processing and event-driven systems

---

### Spanner Emulator
- Used for local dev and E2E tests
- Same client and mutations as production

**Trade-off:**  
Slightly slower than in-memory DB, but production-realistic.

---

## Notes

- This service is intentionally verbose to demonstrate **production-level patterns**
- Optimistic locking can be added via `updated_at` or version columns if required
- Outbox poller is intentionally out-of-scope for this task

---

## Author

Built as part of a **Middle Golang Engineer** test assignment.
