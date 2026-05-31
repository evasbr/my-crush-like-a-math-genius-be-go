# Go Fiber Clean Architecture Boilerplate

A production-ready, highly-optimized REST API boilerplate built in **Go** using the **Go Fiber** web framework, following **Clean Architecture** and **Domain-Driven Design (DDD)** principles.

---

## Technical Stack

*   **Language:** Go (v1.25+)
*   **Web Framework:** Go Fiber v2
*   **Database ORM:** GORM (PostgreSQL driver)
*   **Cache Store:** Redis
*   **Logging:** Logrus (structured JSON, hourly log file rotation, and context-aware Request ID tracing)
*   **Hot Reload:** Air
*   **Unit Testing:** Testify (database-free mocking for service layers)
*   **API Documentation:** Swagger / OpenAPI 3.0 (Swaggo)
*   **Database Migration:** golang-migrate

---

## Directory Structure

```text
├── client/          # Outbound API integrations (Gateways/Adapters)
│   ├── restclient/  # Concrete HTTP/REST implementations for clients
│   └── README.md    # Guide for adding external client integrations
├── common/          # Cross-cutting utility helpers (JWT, Logger, Validator, HTTP)
├── configuration/   # Infrastructure setup (DB connections, Redis, Fiber, Env reader)
├── controller/      # HTTP Presentation layer (Route handlers, Body parsing)
├── db/migrations/   # SQL Schema migrations for database version control
├── db/seed/         # SQL Seed scripts and programmatic seeder runner
├── entity/          # Core Domain entities (Database table mapping)
├── exception/       # Panic recovery and custom HTTP error handlers
├── middleware/      # Fiber middleware (JWT, Request ID, Logging)
├── model/           # DTOs (Data Transfer Objects for HTTP Request/Response payloads)
├── repository/      # Data access layer contracts (Interfaces & GORM implementations)
├── service/         # Domain Business logic layer (Interfaces & implementations)
├── main.go          # Dependency injection container & application entrypoint
└── .air.toml        # Air Hot Reload configuration
```

---

## Getting Started

### 1. Prerequisites
Ensure you have the following installed on your machine:
*   [Go](https://go.dev/doc/install) (v1.25 or later)
*   [Docker](https://www.docker.com/products/docker-desktop/) (for running database & Redis containers)
*   [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate) (for running migrations)

### 2. Environment Configuration
Copy the template file `.env.example` into `.env` (which is git-ignored for security):
```bash
cp .env.example .env
```
Open `.env` and fill in your local credentials for PostgreSQL and Redis.

### 3. Running Infrastructure
Start the database and Redis services using Docker Compose:
```bash
docker compose up -d
```

### 4. Running Database Migrations
To apply the database schema, run the migrations:
```bash
migrate -database "postgres://postgres:postgres@localhost:5432/gofiber_clean_architecture?sslmode=disable" -path db/migrations up
```

### 5. Running Database Seeders
To populate your database with initial seed data (e.g., default administrator account):
```bash
go run main.go -seed
```
To roll back the last applied seeder:
```bash
go run main.go -seed-rollback
```
To generate a new empty seeder file with a timestamp prefix:
```bash
go run main.go -seed-create="name_of_seeder"
```

### 6. Running the Application (Development Mode with Hot Reload)
Run the application with hot reloading enabled via **Air**:
```bash
air
```
The server will start listening at `http://localhost:9999`. Any changes in the `.go` or `.env` files will automatically trigger a rebuild and hot-restart.

---

## Testing

The project is split into service-layer unit tests (with mock dependencies) and controller-layer integration tests.

### 1. Running Unit Tests (Fast & Database-Free)
To test service logic independently without connecting to PostgreSQL or Redis:
```bash
go test -v ./service/impl/...
```

### 2. Running Integration Tests
To test HTTP routes and integration with a running database:
```bash
go test -v ./controller/...
```

---

## Development Workflow: Adding a New Feature

To add a new feature (e.g., a "Customer" module), follow this systematic Clean Architecture workflow:

### Step 1: Create Database Entity
Define the table schema as a Go struct in `entity/customer.go`.
```go
package entity

type Customer struct {
	ID    string `gorm:"primaryKey;column:id"`
	Name  string `gorm:"column:name"`
	Email string `gorm:"column:email"`
}
```

### Step 2: Create SQL Migration Script
Generate the migration files using `golang-migrate`:
```bash
migrate create -ext sql -dir db/migrations -seq create_table_customer
```
Write the SQL DDL statements in the created `.up.sql` and `.down.sql` files, then run `migrate ... up`.

### Step 3: Create Repository Contract & GORM Implementation
1.  Define the database operations interface in `repository/customer_repository.go`.
2.  Implement the GORM database queries inside `repository/impl/customer_repository_impl.go`.

### Step 4: Define Request/Response Models (DTOs)
Create JSON DTOs in `model/customer_model.go` with validation tags:
```go
package model

type CustomerCreateRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}
```

### Step 5: Define & Implement Business Service Logic
1.  Define the business contract interface in `service/customer_service.go`.
2.  Implement the business service in `service/impl/customer_service_impl.go`. Call `common.Validate(request)` to enforce constraints, run business logic, and call repositories.

### Step 6: Create Controller & Register Routes
Create `controller/customer_controller.go`. Set up route paths (e.g., `app.Post(...)`), bind the JSON request using `c.BodyParser`, call the service, and return a standardized JSON format.
```go
return c.Status(fiber.StatusCreated).JSON(model.GeneralResponse{
    Code:    201,
    Message: "Success",
    Data:    response,
})
```

### Step 7: Perform Dependency Injection
Register and wire all the newly created components in `main.go` inside the `main()` function:
```go
// Repository
customerRepository := repository.NewCustomerRepositoryImpl(database)

// Service
customerService := service.NewCustomerServiceImpl(&customerRepository)

// Controller
customerController := controller.NewCustomerController(&customerService)
customerController.Route(app)
```