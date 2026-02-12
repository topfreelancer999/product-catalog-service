# Product Catalog Service

This is a simplified **Product Catalog Service** implemented as a test task for a Middle Golang Engineer. The service follows **Domain-Driven Design (DDD)**, **Clean Architecture**, **CQRS**, and **Transactional Outbox** patterns.

It provides product management, pricing rules with discount support, and gRPC APIs for commands and queries.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Running Instructions](#running-instructions)
3. [Project Structure](#project-structure)
4. [Design Decisions & Trade-offs](#design-decisions--trade-offs)
5. [Testing](#testing)

---

## Prerequisites

* Go 1.21+
* Docker & Docker Compose
* Make (optional, for `make` commands)
* Git

---

## Running Instructions

### 1. Start Spanner Emulator

```bash
docker-compose up -d
```

This will start a local Spanner emulator for development and testing. By default, it runs on `localhost:9010`.

### 2. Run Migrations

```bash
make migrate
```

This will create the required `products` and `outbox_events` tables in the Spanner emulator.

### 3. Run Tests

```bash
make test
```

This executes **unit** and **E2E tests**, covering:

* Product creation and update flows
* Discount application and effective price calculation
* Activation/deactivation
* Business rule validations
* Outbox event generation

### 4. Start gRPC Server

```bash
make run
```

The gRPC server exposes the `ProductService` API on `localhost:50051`. You can test it using gRPC clients such as `grpcurl` or Postman.

---

## Project Structure

```
product-catalog-service/
├── cmd/                # Service entry point
├── internal/           # Application code
│   ├── app/            # Domain & usecases
│   ├── models/         # DB models
│   ├── transport/      # gRPC handlers & mappers
│   └── services/       # DI container
├── proto/              # gRPC service definitions
├── migrations/         # Spanner DDL scripts
├── tests/              # E2E tests
├── docker-compose.yml  # Spanner emulator
├── go.mod
└── README.md
```

---

## Design Decisions & Trade-offs

### 1. **Domain Layer Purity**

* All business logic is isolated in `internal/app/product/domain`.
* Domain layer does not depend on database, gRPC, or external frameworks.
* Uses `*big.Rat` for money calculations to prevent floating-point inaccuracies.

### 2. **CQRS & Golden Mutation**

* Commands go through domain aggregates with **CommitPlan** transactions.
* Queries bypass domain when performance matters.
* This separation ensures business rules are always enforced for writes.

### 3. **Transactional Outbox**

* All domain events are stored in `outbox_events` table.
* Ensures reliable event publishing without external dependencies.

### 4. **Pricing Rules**

* Discounts are percentage-based and time-bound.
* Only one active discount per product is allowed.
* Domain service `PricingCalculator` ensures correct price calculations.

### 5. **Trade-offs**

* Used **Spanner emulator** for local development; production deployment may need additional configurations.
* Simplified read models and pagination for brevity.
* Concurrency control is optional; optimistic locking can be added in future iterations.

---

## Testing

* **Unit Tests:** Test domain logic in isolation (money, discount, pricing calculations).
* **E2E Tests:** Validate flows against real Spanner emulator with commitplan transactions.
* **Test Commands:**

```bash
# Run unit + e2e tests
make test
```

---

## Makefile Commands

```bash
# Start Spanner emulator
make spanner-up

# Run migrations
make migrate

# Run all tests
make test

# Start gRPC server
make run
```

---

## References

* [CommitPlan Docs](https://github.com/Vektor-AI/commitplan)
* [Google Cloud Spanner Emulator](https://cloud.google.com/spanner/docs/emulator)
* [Big Rational Numbers in Go (`math/big`)](https://pkg.go.dev/math/big)
