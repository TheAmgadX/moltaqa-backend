# Testing Guide

This document explains how to run, filter, and analyze tests and code coverage for the `moltaqa-backend` workspace.

---

## Prerequisites & Architecture

The testing architecture in this project is designed to be **fast, isolated, and deterministic**.

### 1. Database Mocking (No Database Required)
All service-level and transport-level tests use an in-memory repository mock (`fakeRepo` and `grpcFakeRepo`). 
*   No connection to PostgreSQL is opened during unit tests.
*   You can run these tests without having PostgreSQL running.
*   You can simulate errors (e.g., `not found`, `duplicate key`, or generic database errors) by directly setting error flags on the mock repositories inside the test cases (e.g., `repo.getErr = errors.New("not found")`).

### 2. Kafka Integration
The test suites use a real Kafka client targeting `127.0.0.1:9092`.
*   **Prerequisite**: You must start the Kafka container before running the tests.
*   The tests use a short delivery timeout configuration to ensure tests fail quickly if the broker is unavailable:
    ```go
    kgo.RecordDeliveryTimeout(1*time.Second)
    kgo.RequestTimeoutOverhead(100*time.Millisecond)
    ```
*   To start the Kafka broker using Docker Compose:
    ```bash
    docker-compose up -d
    ```

---

## Running Tests

You can run Go tests directly from the project's root directory.

### 1. Run the Entire Test Suite
To run all tests in the workspace recursively:
```bash
go test -v ./...
```

### 2. Run Specific Sub-Packages
If you only want to test a single component or package:
*   **User Service Business Logic**:
    ```bash
    go test -v ./services/user-service/testing/service/...
    ```
*   **gRPC API handlers & translations**:
    ```bash
    go test -v ./services/user-service/testing/grpc/...
    ```
*   **API Gateway HTTP Routes & CORS**:
    ```bash
    go test -v ./services/api-gateway/cmd/...
    ```
*   **Domain logic validations**:
    ```bash
    go test -v ./services/user-service/internal/domain/...
    ```
*   **Environment variable helpers**:
    ```bash
    go test -v ./shared/env/...
    ```
*   **Postgres error parser utilities**:
    ```bash
    go test -v ./shared/utils/postgres/...
    ```


---

## Code Coverage

Since tests reside in separate test packages (e.g., `grpc_test`, `service_test`), we use the `-coverpkg` flag to instrument the actual implementation packages.

### 1. Generate Coverage Profile
Run the tests and output the coverage profile to `coverage.out`:
```bash
go test -coverpkg="./..." -coverprofile="coverage.out" ./...
```

### 2. View Coverage in the Terminal
To see a function-by-function statement coverage report:
```bash
go tool cover -func="coverage.out"
```

To filter out auto-generated Protobuf files (`.pb.go`) and focus only on written code:
- **PowerShell (Windows)**:
  ```powershell
  go tool cover -func="coverage.out" | Select-String -NotMatch "(\.pb\.go|\.pb\.gw\.go|_grpc\.pb\.go|tools/)"
  ```
- **Bash / macOS / Linux**:
  ```bash
  go tool cover -func="coverage.out" | grep -vE "(\.pb\.go|\.pb\.gw\.go|_grpc\.pb\.go|tools/)"
  ```

### 3. View Coverage in the Browser (HTML)
To generate an interactive HTML report highlighting exact lines covered (green) vs. uncovered (red):
```bash
go tool cover -html="coverage.out"
```
This will automatically launch your default web browser displaying the interactive UI.

---

## Testing File Directory & Inventory

Here is an overview of each test file and what it is responsible for validating:

### 1. Service Layer Tests
*   **[`services/user-service/testing/service/user_service_test.go`](./services/user-service/testing/service/user_service_test.go)**
    Tests the core business logic of `UserService`. Verifies happy paths, input validations, error propagation from the repository, assigned defaults (like UUIDs), and event production to Kafka.
*   **[`services/user-service/testing/service/kafka_mock_demo_test.go`](./services/user-service/testing/service/kafka_mock_demo_test.go)**
    Demonstrates how to decouple the service layer from concrete `kgo.Client` dependencies using interfaces. Provides an offline, 0.00s execution mock producer pattern.

### 2. Transport & Mapping Tests
*   **[`services/user-service/testing/grpc/server_test.go`](./services/user-service/testing/grpc/server_test.go)**
    Tests the gRPC handler functions on `UserGRPCServer`. Ensures correct response structures and validates that domain errors are translated into standard gRPC status codes.
*   **[`services/user-service/testing/grpc/mapper_test.go`](./services/user-service/testing/grpc/mapper_test.go)**
    Specifically tests the conversion logic (`mapper.go`) between Protobuf message payloads and Domain structs (covers lookups, badges, visibilities, and times).
*   **[`services/user-service/testing/grpc/helpers_test.go`](./services/user-service/testing/grpc/helpers_test.go)**
    Utility file containing setup test helpers (like `mustNewGRPCServerWithRepo`) to configure gRPC servers with custom repository mock implementations.

### 3. Domain Logic Tests
*   **[`services/user-service/internal/domain/errors_test.go`](./services/user-service/internal/domain/errors_test.go)**
    Validates translation of generic PostgreSQL database errors to domain-specific service errors and tests format parameters inside validation error structs.

### 4. Shared Utilities Tests
*   **[`shared/utils/postgres/errors_test.go`](./shared/utils/postgres/errors_test.go)**
    Tests the pure mapping function that translates driver-specific PostgreSQL error codes (such as duplicate key, foreign key, lock timeout) into shared SQL sentinels.
*   **[`shared/env/env_test.go`](./shared/env/env_test.go)**
    Tests configuration settings helpers (`GetString`, `GetInt`, `GetBool`) under different virtual environment states using isolated variable sets.
*   **[`shared/utils/assets/validation_test.go`](./shared/utils/assets/validation_test.go)**
    Tests image file properties validation (`ValidateProfileImagePath`) including path traversal attacks (`..`), absolute prefixes (`/`), max length constraints, and file extension checks.

### 5. API Gateway Tests
*   **[`services/api-gateway/cmd/main_test.go`](./services/api-gateway/cmd/main_test.go)**
    Tests API Gateway route routing (e.g. root `"/"` route returning `"Hello, World!"`) and CORS preflight handling via `middlewares.CORSMiddleware`.




