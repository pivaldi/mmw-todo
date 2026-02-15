# Todo API - DDD & Hexagonal Architecture Design

**Date:** 2026-02-03
**Status:** Approved
**Architecture:** Domain-Driven Design (DDD) with Hexagonal Architecture (Ports & Adapters)

## Overview

This document describes the design of a Todo List CRUD API that demonstrates both the golang-standards/project-layout structure and DDD/Hexagonal Architecture patterns. The application serves as a reference implementation showing how to apply these patterns in a Go project.

### Purpose

- Demonstrate DDD principles in Go
- Show Hexagonal Architecture in practice
- Provide a working example using the standard Go project layout
- Create a production-ready CRUD API with dual HTTP/gRPC support

### Technology Stack

- **API Protocol:** Buf Connect (dual HTTP/gRPC from single protobuf definitions)
- **Database:** PostgreSQL with pgx driver
- **Code Generation:** Buf with protoc-gen-go and protoc-gen-connect-go
- **Architecture:** Hexagonal (Ports & Adapters)
- **Domain Modeling:** Domain-Driven Design (DDD)

## Architecture Overview

### Core Philosophy

The application follows **Hexagonal Architecture** (also known as Ports and Adapters):

- **Domain Layer** (center) - Pure business logic with no external dependencies
- **Application Layer** - Orchestrates use cases, coordinates domain objects
- **Ports** - Interface definitions declaring what the core needs
- **Adapters** (outer) - Implementations of technical concerns (HTTP/gRPC, PostgreSQL)

**Dependency Rule:** All dependencies point inward. The domain knows nothing about databases, frameworks, or protocols.

### Project Structure Mapping

```
/cmd/todo          → Application entry point, dependency wiring
/internal/domain/todo      → Domain entities, value objects, events, business rules
/internal/application → Use cases, application services, DTOs
/internal/ports       → Interface definitions (repositories, event dispatchers)
/internal/adapters    → Implementations (PostgreSQL, Connect handlers)
/api                  → Buf/protobuf definitions
/configs              → Configuration files
/scripts              → Database migrations
/test                 → Integration and API tests
```

### Key DDD Patterns

- **Aggregate Root:** Todo entity protects invariants
- **Value Objects:** Type-safe domain concepts (TaskTitle, Priority, DueDate)
- **Repository Pattern:** Abstracted via ports
- **Domain Events:** For cross-aggregate communication and side effects
- **Rich Domain Model:** Behavior in entities, not anemic data structures

## Domain Model (DDD Layer)

### Todo Aggregate Root

The `Todo` entity is the aggregate root, responsible for maintaining consistency and enforcing business rules.

```go
type Todo struct {
    id          TodoID
    title       TaskTitle
    description string
    status      TaskStatus
    priority    Priority
    dueDate     *DueDate      // optional
    createdAt   time.Time
    updatedAt   time.Time
    events      []DomainEvent // unpublished domain events
}
```

**Aggregate Responsibilities:**
- Enforce business invariants
- Emit domain events when state changes
- Provide factory methods for creation
- Control all mutations through methods

### Value Objects

Value objects are immutable, self-validating types that represent domain concepts:

- **TodoID** - UUID wrapper ensuring valid identifiers
- **TaskTitle** - String constrained to 1-200 characters, trimmed
- **TaskStatus** - Enum: Pending, InProgress, Completed, Cancelled
- **Priority** - Enum: Low, Medium, High, Urgent
- **DueDate** - Timestamp that must be in the future when set

**Value Object Characteristics:**
- Immutable once created
- Validation in constructor
- Equality by value, not reference
- No identity

### Business Rules

The domain enforces these invariants:

1. **Status Transitions:**
   - Cannot complete a cancelled task
   - Cannot modify a completed task (only reopen)
   - Can reopen a completed task back to pending

2. **Validation Rules:**
   - Title is required (1-200 characters)
   - Description is optional
   - Due date must be in the future when set
   - Priority defaults to Medium if not specified

3. **Timestamps:**
   - Created timestamp set on creation
   - Updated timestamp modified on any change
   - Completed tasks record completion time

### Domain Events

Events emitted when aggregate state changes:

- **TodoCreated** - When a new todo is created
- **TodoCompleted** - When status changes to completed
- **TodoReopened** - When a completed todo returns to pending
- **TodoUpdated** - When todo properties change
- **TodoDeleted** - When a todo is removed

**Event Purpose:**
- Enable loose coupling between aggregates
- Trigger side effects (notifications, logging)
- Support eventual consistency
- Enable event sourcing (future enhancement)

### Domain Services

For this initial implementation, no domain services are needed. All behavior naturally fits within the Todo aggregate. If we discover operations that:
- Span multiple aggregates
- Don't naturally belong to a single entity
- Require complex coordination

...then we'll extract them into domain services.

## Application Layer (Use Cases)

### Purpose

The application layer orchestrates domain objects to fulfill use cases. It:
- Coordinates domain operations
- Manages transactions
- Handles DTO mapping
- Publishes domain events
- **Does NOT contain business logic** (that's in the domain)

### TodoApplicationService

The main application service providing use cases:

```go
type TodoApplicationService struct {
    repo       ports.TodoRepository
    dispatcher ports.EventDispatcher
    uow        ports.UnitOfWork
}
```

**Operations:**
- `CreateTodo(ctx, req)` - Create new todo with validation
- `UpdateTodo(ctx, id, req)` - Update todo properties
- `CompleteTodo(ctx, id)` - Mark todo as completed
- `ReopenTodo(ctx, id)` - Reopen completed todo
- `DeleteTodo(ctx, id)` - Remove todo
- `GetTodo(ctx, id)` - Retrieve single todo
- `ListTodos(ctx, filters)` - Query todos with filtering

### Use Case Flow (CreateTodo Example)

1. Receive `CreateTodoRequest` DTO from handler
2. Validate basic structure (non-empty fields)
3. Create value objects (TaskTitle, Priority, DueDate)
4. Call domain factory method `Todo.NewTodo()`
5. Domain validates business rules, emits `TodoCreated` event
6. Begin database transaction
7. Repository saves aggregate
8. Commit transaction
9. Event dispatcher publishes domain events
10. Map domain entity to `TodoResponse` DTO
11. Return result to handler

### Transaction Boundaries

**One use case = one transaction:**
- Begin transaction at use case start
- Save aggregate(s)
- Store events in outbox table
- Commit transaction
- Publish events from outbox (separate process)

This ensures **transactional consistency** - either everything succeeds or everything rolls back.

### Data Transfer Objects (DTOs)

DTOs cross architectural boundaries without exposing domain internals:

**Inbound (Requests):**
```go
type CreateTodoRequest struct {
    Title       string
    Description string
    Priority    string
    DueDate     *time.Time
}

type UpdateTodoRequest struct {
    Title       *string
    Description *string
    Priority    *string
    DueDate     *time.Time
}
```

**Outbound (Responses):**
```go
type TodoResponse struct {
    ID          string
    Title       string
    Description string
    Status      string
    Priority    string
    DueDate     *time.Time
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### Error Handling

Application layer wraps domain errors with context:

```go
func (s *TodoApplicationService) CompleteTodo(ctx context.Context, id string) error {
    todo, err := s.repo.FindByID(ctx, domain.TodoID(id))
    if err != nil {
        return fmt.Errorf("finding todo: %w", err)
    }

    if err := todo.Complete(); err != nil {
        // Domain error (business rule violation)
        return fmt.Errorf("completing todo %s: %w", id, err)
    }

    if err := s.repo.Update(ctx, todo); err != nil {
        // Infrastructure error
        return fmt.Errorf("updating todo: %w", err)
    }

    return s.dispatcher.Dispatch(ctx, todo.Events())
}
```

## Ports (Interface Definitions)

### Purpose

Ports define contracts between layers. The inner layers (domain/application) declare what they need; outer layers (adapters) provide implementations.

**Location:** `/internal/ports/`

### Primary Ports (Driving - Input)

Implemented by application layer, called by adapters:

```go
// TodoService is the main application interface
type TodoService interface {
    CreateTodo(ctx context.Context, req CreateTodoRequest) (*TodoResponse, error)
    UpdateTodo(ctx context.Context, id string, req UpdateTodoRequest) (*TodoResponse, error)
    CompleteTodo(ctx context.Context, id string) error
    ReopenTodo(ctx context.Context, id string) error
    DeleteTodo(ctx context.Context, id string) error
    GetTodo(ctx context.Context, id string) (*TodoResponse, error)
    ListTodos(ctx context.Context, filters ListFilters) ([]*TodoResponse, error)
}
```

### Secondary Ports (Driven - Output)

Needed by application, implemented by infrastructure:

```go
// TodoRepository - persistence abstraction
type TodoRepository interface {
    Save(ctx context.Context, todo *domain.Todo) error
    FindByID(ctx context.Context, id domain.TodoID) (*domain.Todo, error)
    FindAll(ctx context.Context, filters Filters) ([]*domain.Todo, error)
    Update(ctx context.Context, todo *domain.Todo) error
    Delete(ctx context.Context, id domain.TodoID) error
}

// EventDispatcher - domain event publishing
type EventDispatcher interface {
    Dispatch(ctx context.Context, events []domain.DomainEvent) error
}

// UnitOfWork - transaction management
type UnitOfWork interface {
    Begin(ctx context.Context) (Transaction, error)
}

type Transaction interface {
    Commit() error
    Rollback() error
    TodoRepository() TodoRepository
}
```

### Port Benefits

- **Testability:** Mock implementations for testing
- **Flexibility:** Swap implementations (PostgreSQL → MongoDB)
- **Dependency Inversion:** Inner layers don't depend on outer layers
- **Clear Contracts:** Explicit about what each layer needs

## Adapters (Implementations)

### Input Adapter: Connect Handler

**Location:** `/internal/adapters/handler/connect/`

Implements the Buf Connect service interface, handling both HTTP and gRPC requests:

```go
type TodoServiceHandler struct {
    todoService ports.TodoService
}

func (h *TodoServiceHandler) CreateTodo(
    ctx context.Context,
    req *connect.Request[todov1.CreateTodoRequest],
) (*connect.Response[todov1.CreateTodoResponse], error) {
    // 1. Extract proto message
    protoReq := req.Msg

    // 2. Map proto → application DTO
    appReq := mapProtoToCreateRequest(protoReq)

    // 3. Call application service
    result, err := h.todoService.CreateTodo(ctx, appReq)
    if err != nil {
        // 4. Map error to Connect status code
        return nil, mapErrorToConnectError(err)
    }

    // 5. Map DTO → proto response
    protoResp := mapTodoResponseToProto(result)

    return connect.NewResponse(protoResp), nil
}
```

**Responsibilities:**
- Protocol handling (HTTP/gRPC via Connect)
- Request/response mapping (proto ↔ DTO)
- Error translation (domain errors → status codes)
- Basic input validation
- Logging and observability

### Output Adapter: PostgreSQL Repository

**Location:** `/internal/adapters/repository/postgres/`

Implements `TodoRepository` port using PostgreSQL:

```go
type PostgresTodoRepository struct {
    db *pgxpool.Pool
}

func (r *PostgresTodoRepository) Save(ctx context.Context, todo *domain.Todo) error {
    query := `
        INSERT INTO todos (id, title, description, status, priority, due_date, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

    _, err := r.db.Exec(ctx, query,
        todo.ID().String(),
        todo.Title().String(),
        todo.Description(),
        todo.Status().String(),
        todo.Priority().String(),
        todo.DueDate(),
        todo.CreatedAt(),
        todo.UpdatedAt(),
    )

    return err
}

func (r *PostgresTodoRepository) FindByID(ctx context.Context, id domain.TodoID) (*domain.Todo, error) {
    // Query database
    // Map row → domain entity
    // Reconstitute aggregate
    return todo, nil
}
```

**Responsibilities:**
- SQL query construction
- Transaction management
- Domain entity ↔ database row mapping
- Connection pooling
- Error handling (wrap SQL errors)

### Output Adapter: Event Dispatcher

**Location:** `/internal/adapters/events/`

Initial implementation: simple in-memory dispatcher that logs events.

```go
type InMemoryEventDispatcher struct {
    logger *slog.Logger
}

func (d *InMemoryEventDispatcher) Dispatch(ctx context.Context, events []domain.DomainEvent) error {
    for _, event := range events {
        d.logger.Info("domain event",
            "type", event.Type(),
            "aggregate_id", event.AggregateID(),
            "data", event.Data(),
        )
    }
    return nil
}
```

**Future Enhancement:** Integrate with message broker (RabbitMQ, Kafka) for durable event publishing.

### Adapter Principles

- Adapters depend on ports (interfaces), never directly on domain
- Multiple adapters can implement the same port
- Adapters handle all technical concerns
- Easy to swap implementations (mock for tests, different DB, etc.)

## API Definitions & Data Flow

### Protobuf Schema

**Location:** `/api/todo/v1/todo.proto`

```protobuf
syntax = "proto3";

package todo.v1;

import "google/protobuf/timestamp.proto";

service TodoService {
  rpc CreateTodo(CreateTodoRequest) returns (CreateTodoResponse);
  rpc GetTodo(GetTodoRequest) returns (GetTodoResponse);
  rpc UpdateTodo(UpdateTodoRequest) returns (UpdateTodoResponse);
  rpc CompleteTodo(CompleteTodoRequest) returns (CompleteTodoResponse);
  rpc ReopenTodo(ReopenTodoRequest) returns (ReopenTodoResponse);
  rpc DeleteTodo(DeleteTodoRequest) returns (DeleteTodoResponse);
  rpc ListTodos(ListTodosRequest) returns (ListTodosResponse);
}

message Todo {
  string id = 1;
  string title = 2;
  string description = 3;
  TaskStatus status = 4;
  Priority priority = 5;
  google.protobuf.Timestamp due_date = 6;
  google.protobuf.Timestamp created_at = 7;
  google.protobuf.Timestamp updated_at = 8;
}

enum TaskStatus {
  TASK_STATUS_UNSPECIFIED = 0;
  TASK_STATUS_PENDING = 1;
  TASK_STATUS_IN_PROGRESS = 2;
  TASK_STATUS_COMPLETED = 3;
  TASK_STATUS_CANCELLED = 4;
}

enum Priority {
  PRIORITY_UNSPECIFIED = 0;
  PRIORITY_LOW = 1;
  PRIORITY_MEDIUM = 2;
  PRIORITY_HIGH = 3;
  PRIORITY_URGENT = 4;
}

message CreateTodoRequest {
  string title = 1;
  string description = 2;
  Priority priority = 3;
  google.protobuf.Timestamp due_date = 4;
}

message CreateTodoResponse {
  Todo todo = 1;
}

// ... additional request/response messages
```

### Buf Configuration

**buf.yaml** - Module configuration:
```yaml
version: v1
name: buf.build/yourorg/todo
deps:
  - buf.build/googleapis/googleapis
breaking:
  use:
    - FILE
lint:
  use:
    - DEFAULT
```

**buf.gen.yaml** - Code generation:
```yaml
version: v1
plugins:
  - plugin: go
    out: gen/go
    opt: paths=source_relative
  - plugin: connect-go
    out: gen/go
    opt: paths=source_relative
```

### Request Flow

**Complete flow for CreateTodo operation:**

1. **Client Request:**
   - HTTP: `POST /todo.v1.TodoService/CreateTodo` with JSON body
   - gRPC: `CreateTodo` RPC call with protobuf message

2. **Connect Runtime:**
   - Deserializes request (JSON → proto or binary proto)
   - Routes to handler method

3. **Connect Handler (Adapter):**
   - Validates proto message structure
   - Maps proto → `CreateTodoRequest` DTO
   - Calls `todoService.CreateTodo(ctx, req)`

4. **Application Service:**
   - Creates value objects from DTO
   - Calls domain factory `Todo.NewTodo()`
   - Begins transaction

5. **Domain Layer:**
   - Validates business rules
   - Creates Todo aggregate
   - Emits `TodoCreated` event

6. **Repository (Adapter):**
   - Saves aggregate to PostgreSQL
   - Stores events in outbox table

7. **Application Service:**
   - Commits transaction
   - Dispatches domain events
   - Maps domain → `TodoResponse` DTO
   - Returns result

8. **Connect Handler:**
   - Maps DTO → proto response
   - Returns to client

9. **Connect Runtime:**
   - Serializes response (proto → JSON or binary)
   - Sends HTTP/gRPC response

### Error Flow

Errors bubble up and are translated at each boundary:

- **Domain:** `ErrInvalidTitle`
- **Application:** Wraps with context
- **Handler:** Maps to Connect `InvalidArgument` code
- **Client:** Receives 400 with error details

## Database Schema & Persistence

### PostgreSQL Schema

**Todos Table:**
```sql
CREATE TABLE todos (
    id UUID PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL,
    priority VARCHAR(20) NOT NULL,
    due_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,

    CONSTRAINT valid_status CHECK (status IN ('pending', 'in_progress', 'completed', 'cancelled')),
    CONSTRAINT valid_priority CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    CONSTRAINT future_due_date CHECK (due_date IS NULL OR due_date > created_at)
);

CREATE INDEX idx_todos_status ON todos(status);
CREATE INDEX idx_todos_due_date ON todos(due_date) WHERE due_date IS NOT NULL;
CREATE INDEX idx_todos_created_at ON todos(created_at DESC);
```

**Event Outbox Table:**
```sql
CREATE TABLE domain_events (
    id BIGSERIAL PRIMARY KEY,
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_data JSONB NOT NULL,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL,
    published_at TIMESTAMP WITH TIME ZONE,

    INDEX idx_unpublished_events (published_at) WHERE published_at IS NULL
);
```

### Transactional Outbox Pattern

**Problem:** We need to save aggregates AND publish events atomically. If we save to DB and then publish to message broker, the publish might fail, leaving inconsistent state.

**Solution:** Store events in database within same transaction:

1. Begin transaction
2. Save aggregate to `todos` table
3. Insert events into `domain_events` table
4. Commit transaction (both succeed or both fail)
5. Separate process reads `domain_events` WHERE `published_at IS NULL`
6. Publishes events to message broker
7. Updates `published_at` timestamp

This ensures **at-least-once delivery** of domain events.

### Migration Strategy

**Tool:** golang-migrate or goose
**Location:** `/scripts/migrations/`

**Migration Files:**
```
001_create_todos.up.sql
001_create_todos.down.sql
002_create_events.up.sql
002_create_events.down.sql
003_add_indexes.up.sql
003_add_indexes.down.sql
```

**Applied:** During application startup or via CLI command:
```bash
migrate -path ./scripts/migrations -database postgres://... up
```

### Repository Mapping Strategy

**Domain → Database:**
- `Todo` aggregate → `todos` table row
- `TodoID` value object → UUID column
- `TaskTitle` value object → VARCHAR column
- `Priority` enum → VARCHAR with CHECK constraint
- Domain events → `domain_events` table

**Reconstitution (Database → Domain):**
Repository queries database and rebuilds domain objects:
```go
func (r *PostgresTodoRepository) FindByID(ctx, id) (*domain.Todo, error) {
    var row struct {
        ID          string
        Title       string
        Description string
        Status      string
        Priority    string
        // ... other fields
    }

    // Query database
    err := r.db.QueryRow(ctx, "SELECT ... FROM todos WHERE id = $1", id).Scan(...)

    // Reconstitute domain aggregate
    todoID := domain.TodoID(row.ID)
    title, _ := domain.NewTaskTitle(row.Title)
    status := domain.TaskStatus(row.Status)
    // ... build value objects

    // Use factory or constructor
    todo := domain.ReconstituteTodo(todoID, title, description, status, ...)

    return todo, nil
}
```

### Connection Management

**Configuration:** `/configs/database.yaml`
```yaml
database:
  host: localhost
  port: 5435
  name: todo
  user: todo
  password: ${DB_PASSWORD}
  pool:
    max_conns: 25
    min_conns: 5
    max_conn_lifetime: 1h
    max_conn_idle_time: 30m
```

**Connection Pooling:**
- Use pgxpool for automatic connection pooling
- Configure pool size based on workload
- Set timeouts for query execution
- Health check endpoint verifies DB connectivity

## Error Handling Strategy

### Error Categories

**1. Domain Errors (Business Rules):**
Expected errors representing business rule violations:

```go
// /internal/domain/todo/errors.go
package domain

var (
    ErrInvalidTitle              = errors.New("title must be between 1-200 characters")
    ErrInvalidDueDate            = errors.New("due date must be in the future")
    ErrCannotCompleteCancelled   = errors.New("cannot complete a cancelled task")
    ErrCannotModifyCompleted     = errors.New("cannot modify completed tasks")
    ErrTodoNotFound              = errors.New("todo not found")
)

type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type BusinessRuleError struct {
    Rule    string
    Message string
}

func (e BusinessRuleError) Error() string {
    return fmt.Sprintf("business rule violated: %s", e.Message)
}
```

**2. Application Errors:**
Orchestration and coordination errors:

```go
// /internal/application/errors.go
package application

func (s *TodoApplicationService) CompleteTodo(ctx, id) error {
    todo, err := s.repo.FindByID(ctx, id)
    if err != nil {
        // Wrap with context
        return fmt.Errorf("finding todo %s: %w", id, err)
    }

    if err := todo.Complete(); err != nil {
        // Domain error - preserve it
        return fmt.Errorf("completing todo: %w", err)
    }

    return nil
}
```

**3. Infrastructure Errors:**
Technical failures (database, network):

```go
// Adapter error handling
func (r *PostgresTodoRepository) Save(ctx, todo) error {
    _, err := r.db.Exec(ctx, query, args...)
    if err != nil {
        // Translate DB errors
        if isUniqueViolation(err) {
            return fmt.Errorf("todo already exists: %w", domain.ErrTodoAlreadyExists)
        }
        return fmt.Errorf("saving todo to database: %w", err)
    }
    return nil
}
```

### Error Translation (Handler → Client)

Connect handler maps errors to appropriate status codes:

```go
func mapErrorToConnectError(err error) error {
    switch {
    case errors.Is(err, domain.ErrTodoNotFound):
        return connect.NewError(connect.CodeNotFound, err)

    case errors.Is(err, domain.ErrInvalidTitle),
         errors.Is(err, domain.ErrInvalidDueDate):
        return connect.NewError(connect.CodeInvalidArgument, err)

    case errors.Is(err, domain.ErrCannotCompleteCancelled),
         errors.Is(err, domain.ErrCannotModifyCompleted):
        return connect.NewError(connect.CodeFailedPrecondition, err)

    case isPostgresError(err):
        return connect.NewError(connect.CodeUnavailable,
            fmt.Errorf("database unavailable"))

    default:
        return connect.NewError(connect.CodeInternal,
            fmt.Errorf("internal server error"))
    }
}
```

**Status Code Mapping:**
- `InvalidArgument` (400) - Validation failures
- `NotFound` (404) - Resource not found
- `FailedPrecondition` (400/409) - Business rule violations
- `Internal` (500) - Unexpected errors
- `Unavailable` (503) - Infrastructure failures

### Error Response Format

**JSON Error Response:**
```json
{
  "code": "invalid_argument",
  "message": "title must be between 1-200 characters",
  "details": [
    {
      "@type": "type.googleapis.com/todo.v1.ValidationError",
      "field": "title",
      "constraint": "length",
      "actual": 0,
      "expected": "1-200"
    }
  ]
}
```

### Logging Strategy

**Domain Layer:** No logging (pure business logic)

**Application Layer:** Log use case entry/exit and errors
```go
func (s *TodoApplicationService) CreateTodo(ctx, req) (*TodoResponse, error) {
    s.logger.Info("creating todo", "title", req.Title)

    result, err := s.createTodo(ctx, req)
    if err != nil {
        s.logger.Error("failed to create todo", "error", err, "title", req.Title)
        return nil, err
    }

    s.logger.Info("todo created", "id", result.ID)
    return result, nil
}
```

**Adapter Layer:** Log technical operations
```go
func (r *PostgresTodoRepository) Save(ctx, todo) error {
    r.logger.Debug("saving todo", "id", todo.ID())

    err := r.db.Exec(...)
    if err != nil {
        r.logger.Error("database save failed", "error", err, "id", todo.ID())
        return err
    }

    return nil
}
```

**Structured Logging:**
- Use `log/slog` for structured logging
- Include correlation IDs from context
- Log levels: Debug, Info, Warn, Error
- Never log sensitive data (passwords, tokens)

## Testing Strategy

### Testing Pyramid

```
          /\
         /  \
        / E2E \         ← Few, slow, expensive
       /______\
      /        \
     / Integr.  \       ← Some, medium speed
    /__________\
   /            \
  /  Unit Tests  \     ← Many, fast, cheap
 /________________\
```

### 1. Domain Layer Tests (Unit)

**Location:** `/internal/domain/todo/*_test.go`

**Characteristics:**
- Pure unit tests with zero dependencies
- Fast execution (milliseconds)
- High coverage target (90%+)
- Test business rules and invariants

**Examples:**

```go
// todo_test.go
func TestTodo_Complete_WithPendingStatus_Succeeds(t *testing.T) {
    todo := createPendingTodo()

    err := todo.Complete()

    assert.NoError(t, err)
    assert.Equal(t, domain.Completed, todo.Status())
    assert.NotNil(t, todo.CompletedAt())
}

func TestTodo_Complete_WithCancelledStatus_ReturnsError(t *testing.T) {
    todo := createCancelledTodo()

    err := todo.Complete()

    assert.ErrorIs(t, err, domain.ErrCannotCompleteCancelled)
}

func TestTodo_Complete_EmitsTodoCompletedEvent(t *testing.T) {
    todo := createPendingTodo()

    err := todo.Complete()

    assert.NoError(t, err)
    events := todo.Events()
    assert.Len(t, events, 1)
    assert.IsType(t, domain.TodoCompleted{}, events[0])
}

// value_objects_test.go
func TestTaskTitle_Validate_EmptyString_ReturnsError(t *testing.T) {
    _, err := domain.NewTaskTitle("")

    assert.ErrorIs(t, err, domain.ErrInvalidTitle)
}

func TestTaskTitle_Validate_TooLong_ReturnsError(t *testing.T) {
    longTitle := strings.Repeat("a", 201)

    _, err := domain.NewTaskTitle(longTitle)

    assert.ErrorIs(t, err, domain.ErrInvalidTitle)
}

func TestTaskTitle_Validate_TrimsWhitespace(t *testing.T) {
    title, err := domain.NewTaskTitle("  Valid Title  ")

    assert.NoError(t, err)
    assert.Equal(t, "Valid Title", title.String())
}
```

**Table-Driven Tests:**
```go
func TestTodo_StatusTransitions(t *testing.T) {
    tests := []struct {
        name        string
        initialStatus domain.TaskStatus
        operation   func(*domain.Todo) error
        expectError error
    }{
        {
            name:        "complete pending todo succeeds",
            initialStatus: domain.Pending,
            operation:   (*domain.Todo).Complete,
            expectError: nil,
        },
        {
            name:        "complete cancelled todo fails",
            initialStatus: domain.Cancelled,
            operation:   (*domain.Todo).Complete,
            expectError: domain.ErrCannotCompleteCancelled,
        },
        // ... more cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            todo := createTodoWithStatus(tt.initialStatus)
            err := tt.operation(todo)
            if tt.expectError != nil {
                assert.ErrorIs(t, err, tt.expectError)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 2. Application Layer Tests (Unit with Mocks)

**Location:** `/internal/application/*_test.go`

**Characteristics:**
- Mock repositories and ports
- Test use case orchestration
- Verify correct port interactions
- Test error propagation

**Examples:**

```go
// todo_service_test.go
func TestTodoService_CreateTodo_ValidInput_SavesAndPublishesEvent(t *testing.T) {
    // Arrange
    mockRepo := &MockTodoRepository{}
    mockEvents := &MockEventDispatcher{}
    service := application.NewTodoApplicationService(mockRepo, mockEvents)

    req := application.CreateTodoRequest{
        Title:       "Test Todo",
        Description: "Description",
        Priority:    "medium",
    }

    // Act
    result, err := service.CreateTodo(context.Background(), req)

    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, result.ID)
    assert.Equal(t, "Test Todo", result.Title)

    // Verify repository was called
    mockRepo.AssertCalled(t, "Save", mock.Anything, mock.Anything)

    // Verify events were dispatched
    mockEvents.AssertCalled(t, "Dispatch", mock.Anything, mock.MatchedBy(func(events []domain.DomainEvent) bool {
        return len(events) == 1 && events[0].Type() == "TodoCreated"
    }))
}

func TestTodoService_CompleteTodo_TodoNotFound_ReturnsError(t *testing.T) {
    mockRepo := &MockTodoRepository{}
    mockRepo.On("FindByID", mock.Anything, mock.Anything).
        Return(nil, domain.ErrTodoNotFound)

    service := application.NewTodoApplicationService(mockRepo, nil)

    err := service.CompleteTodo(context.Background(), "nonexistent-id")

    assert.ErrorIs(t, err, domain.ErrTodoNotFound)
}
```

**Mock Implementations:**
```go
type MockTodoRepository struct {
    mock.Mock
}

func (m *MockTodoRepository) Save(ctx context.Context, todo *domain.Todo) error {
    args := m.Called(ctx, todo)
    return args.Error(0)
}

func (m *MockTodoRepository) FindByID(ctx context.Context, id domain.TodoID) (*domain.Todo, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.Todo), args.Error(1)
}
```

### 3. Integration Tests

**Location:** `/test/integration/*_test.go`

**Characteristics:**
- Test with real PostgreSQL (testcontainers)
- Verify adapter implementations
- Test database migrations
- Medium execution speed (seconds)

**Examples:**

```go
// repository_test.go
func TestPostgresRepository_SaveAndFind_RoundTrip(t *testing.T) {
    // Setup real PostgreSQL with testcontainers
    db := setupTestDatabase(t)
    defer db.Close()

    repo := postgres.NewTodoRepository(db)

    // Create domain aggregate
    title, _ := domain.NewTaskTitle("Integration Test Todo")
    priority := domain.PriorityMedium
    todo := domain.NewTodo(title, "Description", priority, nil)

    // Save to database
    err := repo.Save(context.Background(), todo)
    require.NoError(t, err)

    // Retrieve from database
    found, err := repo.FindByID(context.Background(), todo.ID())
    require.NoError(t, err)

    // Verify round-trip
    assert.Equal(t, todo.ID(), found.ID())
    assert.Equal(t, todo.Title().String(), found.Title().String())
    assert.Equal(t, todo.Priority(), found.Priority())
}

func TestPostgresRepository_Update_ExistingTodo_UpdatesFields(t *testing.T) {
    db := setupTestDatabase(t)
    defer db.Close()

    repo := postgres.NewTodoRepository(db)

    // Create and save initial todo
    todo := createAndSaveTodo(t, repo)

    // Modify todo
    newTitle, _ := domain.NewTaskTitle("Updated Title")
    todo.UpdateTitle(newTitle)

    // Update in database
    err := repo.Update(context.Background(), todo)
    require.NoError(t, err)

    // Verify update
    found, err := repo.FindByID(context.Background(), todo.ID())
    require.NoError(t, err)
    assert.Equal(t, "Updated Title", found.Title().String())
}

// Test setup helper
func setupTestDatabase(t *testing.T) *pgxpool.Pool {
    ctx := context.Background()

    // Start PostgreSQL container
    container, err := postgres.RunContainer(ctx,
        testcontainers.WithImage("postgres:15"),
        postgres.WithDatabase("testdb"),
    )
    require.NoError(t, err)

    t.Cleanup(func() {
        container.Terminate(ctx)
    })

    // Get connection string
    connString, err := container.ConnectionString(ctx)
    require.NoError(t, err)

    // Run migrations
    runMigrations(t, connString)

    // Create pool
    pool, err := pgxpool.New(ctx, connString)
    require.NoError(t, err)

    return pool
}
```

### 4. API Tests (End-to-End)

**Location:** `/test/api/*_test.go`

**Characteristics:**
- Test full HTTP/gRPC flow
- Use generated Connect client
- Test both protocols
- Verify error responses and status codes

**Examples:**

```go
// todo_api_test.go
func TestTodo_CreateTodo_ValidRequest_ReturnsCreated(t *testing.T) {
    // Start test server
    server := startTestServer(t)
    defer server.Stop()

    // Create Connect client
    client := todov1connect.NewTodoServiceClient(
        http.DefaultClient,
        server.URL,
    )

    // Make request
    req := connect.NewRequest(&todov1.CreateTodoRequest{
        Title:       "API Test Todo",
        Description: "Testing the API",
        Priority:    todov1.Priority_PRIORITY_HIGH,
    })

    resp, err := client.CreateTodo(context.Background(), req)

    // Assert success
    require.NoError(t, err)
    assert.NotEmpty(t, resp.Msg.Todo.Id)
    assert.Equal(t, "API Test Todo", resp.Msg.Todo.Title)
    assert.Equal(t, todov1.TaskStatus_TASK_STATUS_PENDING, resp.Msg.Todo.Status)
}

func TestTodo_CreateTodo_InvalidTitle_Returns400(t *testing.T) {
    server := startTestServer(t)
    defer server.Stop()

    client := todov1connect.NewTodoServiceClient(http.DefaultClient, server.URL)

    req := connect.NewRequest(&todov1.CreateTodoRequest{
        Title: "", // Invalid: empty title
    })

    _, err := client.CreateTodo(context.Background(), req)

    // Assert error
    require.Error(t, err)

    var connectErr *connect.Error
    require.True(t, errors.As(err, &connectErr))
    assert.Equal(t, connect.CodeInvalidArgument, connectErr.Code())
}

func TestTodo_GetTodo_NotFound_Returns404(t *testing.T) {
    server := startTestServer(t)
    defer server.Stop()

    client := todov1connect.NewTodoServiceClient(http.DefaultClient, server.URL)

    req := connect.NewRequest(&todov1.GetTodoRequest{
        Id: "nonexistent-id",
    })

    _, err := client.GetTodo(context.Background(), req)

    var connectErr *connect.Error
    require.True(t, errors.As(err, &connectErr))
    assert.Equal(t, connect.CodeNotFound, connectErr.Code())
}
```

### Test Coverage Goals

- **Domain Layer:** 90%+ (critical business logic)
- **Application Layer:** 80%+
- **Adapters:** 70%+ (via integration tests)
- **Overall Project:** 80%+

### Testing Tools & Libraries

- **testing** - Standard library
- **testify/assert** - Assertions
- **testify/mock** - Mocking interfaces
- **testify/require** - Fail-fast assertions
- **testcontainers-go** - Real PostgreSQL for integration tests
- **sqlmock** - Mock database driver (for repository unit tests)
- **connectrpc/connect-go** - Generated client for API tests

### Test Organization

```
project-layout/
├── internal/
│   ├── domain/
│   │   ├── todo.go
│   │   └── todo_test.go          ← Unit tests alongside code
│   ├── application/
│   │   ├── todo_service.go
│   │   └── todo_service_test.go  ← Unit tests with mocks
│   └── adapters/
│       └── repository/
│           └── postgres/
│               ├── todo_repository.go
│               └── todo_repository_test.go  ← Can use sqlmock
│
└── test/
    ├── integration/
    │   ├── repository_test.go     ← Real DB tests
    │   └── setup.go               ← Test helpers
    ├── api/
    │   └── todo_api_test.go       ← End-to-end API tests
    └── testdata/
        └── fixtures.json          ← Test data
```

## Detailed Project Structure

### Complete Directory Layout

```
project-layout/
├── cmd/
│   └── todo/                           # Application entry point
│       └── main.go                        # Dependency injection, server startup
│
├── internal/                              # Private application code
│   ├── domain/                            # Domain layer (pure business logic)
│   │   ├── todo.go                        # Todo aggregate root
│   │   ├── todo_test.go
│   │   ├── value_objects.go               # TaskTitle, Priority, DueDate, etc.
│   │   ├── value_objects_test.go
│   │   ├── events.go                      # Domain events (TodoCreated, etc.)
│   │   └── errors.go                      # Domain-specific errors
│   │
│   ├── application/                       # Application/use case layer
│   │   ├── todo_service.go                # TodoApplicationService
│   │   ├── todo_service_test.go
│   │   └── dto.go                         # Request/Response DTOs
│   │
│   ├── ports/                             # Interface definitions (hexagonal ports)
│   │   ├── repository.go                  # TodoRepository interface
│   │   ├── events.go                      # EventDispatcher interface
│   │   └── service.go                     # TodoService interface
│   │
│   └── adapters/                          # Adapter implementations
│       ├── handler/
│       │   └── connect/
│       │       ├── todo_handler.go        # Connect service implementation
│       │       ├── todo_handler_test.go
│       │       ├── mapper.go              # Proto ↔ DTO mapping
│       │       └── errors.go              # Error translation
│       │
│       ├── repository/
│       │   └── postgres/
│       │       ├── todo_repository.go     # PostgreSQL implementation
│       │       ├── todo_repository_test.go
│       │       └── mapper.go              # Domain ↔ DB row mapping
│       │
│       └── events/
│           ├── dispatcher.go              # In-memory event dispatcher
│           └── dispatcher_test.go
│
├── api/                                   # API definitions (protobuf)
│   └── todo/
│       └── v1/
│           └── todo.proto                 # Service & message definitions
│
├── gen/                                   # Generated code (gitignored)
│   └── go/
│       └── todo/
│           └── v1/
│               ├── todo.pb.go             # Generated protobuf code
│               └── todoconnect/
│                   └── todo.connect.go    # Generated Connect service
│
├── configs/                               # Configuration files
│   ├── config.yaml                        # Application config template
│   └── database.yaml                      # Database connection config
│
├── scripts/                               # Build and operational scripts
│   ├── migrations/                        # Database migrations
│   │   ├── 000001_create_todos_table.up.sql
│   │   ├── 000001_create_todos_table.down.sql
│   │   ├── 000002_create_events_table.up.sql
│   │   └── 000002_create_events_table.down.sql
│   ├── seed_data.sql                      # Test/demo data
│   └── run_migrations.sh                  # Migration helper script
│
├── test/                                  # Additional external tests
│   ├── integration/                       # Integration tests (real DB)
│   │   ├── repository_test.go
│   │   └── setup.go                       # testcontainers setup
│   ├── api/                               # End-to-end API tests
│   │   └── todo_api_test.go
│   └── testdata/                          # Test fixtures
│       └── todos.json
│
├── docs/                                  # Documentation
│   ├── architecture.md                    # Architecture decisions
│   ├── api.md                             # API documentation
│   └── plans/
│       └── 2026-02-03-todo-api-design.md  # This document
│
├── build/                                 # Build configurations
│   ├── package/
│   │   └── Dockerfile                     # Container image
│   └── ci/
│       └── .github/
│           └── workflows/
│               └── ci.yml                 # CI pipeline
│
├── deployments/                           # Deployment configurations
│   ├── docker-compose.yml                 # Local development
│   └── kubernetes/                        # K8s manifests
│       ├── deployment.yaml
│       └── service.yaml
│
├── .gitignore
├── buf.yaml                               # Buf module configuration
├── buf.gen.yaml                           # Buf code generation config
├── go.mod
├── go.sum
├── Makefile                               # Build targets
└── README.md                              # Updated with DDD/Hex example
```

### Key Files & Responsibilities

**`cmd/todo/main.go`**
- Application entry point
- Dependency injection (wire up layers)
- Configuration loading
- Server startup (HTTP/gRPC)
- Graceful shutdown handling

**`internal/domain/todo/`**
- Pure business logic
- No external dependencies
- Aggregate roots, entities, value objects
- Domain events
- Business rule enforcement

**`internal/application/`**
- Use case orchestration
- Transaction management
- Domain coordination
- DTO mapping
- No business logic

**`internal/ports/`**
- Interface definitions only
- Contracts between layers
- Enable dependency inversion

**`internal/adapters/`**
- All technical implementations
- HTTP/gRPC handling (Connect)
- Database access (PostgreSQL)
- Event publishing
- External service integration

**`api/todo/v1/`**
- Protocol buffer definitions
- Service contracts
- Message schemas
- Versioned API definitions

### Dependency Flow

```
Outer → Inner (dependencies point inward)

[Connect Handler]  ─depends on→  [TodoService interface]  ─implemented by→  [TodoApplicationService]
                                         ↓
[PostgreSQL Repo]  ─depends on→  [TodoRepository interface]  ─used by→  [TodoApplicationService]
                                         ↓
                                   [Domain Layer]
                                   (no dependencies)
```

**Critical Rule:** Inner layers NEVER depend on outer layers. Domain doesn't know about databases, APIs, or frameworks.

## Implementation Checklist

### Phase 1: Foundation
- [ ] Setup project structure (directories)
- [ ] Configure Go modules
- [ ] Setup Buf configuration (buf.yaml, buf.gen.yaml)
- [ ] Define protobuf schema (api/todo/v1/todo.proto)
- [ ] Generate code with Buf
- [ ] Setup database migrations
- [ ] Create PostgreSQL schema

### Phase 2: Domain Layer
- [ ] Implement value objects (TaskTitle, Priority, DueDate, etc.)
- [ ] Implement Todo aggregate root
- [ ] Implement domain events
- [ ] Write domain layer unit tests
- [ ] Verify 90%+ test coverage on domain

### Phase 3: Ports
- [ ] Define TodoService interface
- [ ] Define TodoRepository interface
- [ ] Define EventDispatcher interface
- [ ] Define UnitOfWork interface

### Phase 4: Application Layer
- [ ] Implement TodoApplicationService
- [ ] Implement DTOs
- [ ] Write application layer tests with mocks
- [ ] Verify transaction boundaries

### Phase 5: Adapters - Persistence
- [ ] Implement PostgreSQL repository
- [ ] Implement domain-to-DB mapping
- [ ] Write integration tests with testcontainers
- [ ] Test migrations

### Phase 6: Adapters - API
- [ ] Implement Connect handlers
- [ ] Implement proto-to-DTO mapping
- [ ] Implement error translation
- [ ] Write API tests (HTTP and gRPC)

### Phase 7: Adapters - Events
- [ ] Implement in-memory event dispatcher
- [ ] Add structured logging
- [ ] Future: Integrate message broker

### Phase 8: Main Application
- [ ] Implement dependency injection in main.go
- [ ] Configure server startup
- [ ] Add configuration loading
- [ ] Implement graceful shutdown
- [ ] Add health check endpoints

### Phase 9: DevOps
- [ ] Create Dockerfile
- [ ] Create docker-compose.yml for local dev
- [ ] Setup CI pipeline
- [ ] Add linting and formatting
- [ ] Configure code coverage reporting

### Phase 10: Documentation
- [ ] Update main README with DDD/Hex example
- [ ] Document API endpoints
- [ ] Add architecture decision records
- [ ] Create usage examples
- [ ] Add deployment guide

## Future Enhancements

### Short Term
- Add filtering and sorting to ListTodos
- Implement pagination for large result sets
- Add authentication and authorization
- Implement rate limiting
- Add OpenTelemetry for observability

### Medium Term
- Integrate message broker for domain events (RabbitMQ/Kafka)
- Add CQRS pattern (separate read/write models)
- Implement event sourcing for audit trail
- Add caching layer (Redis)
- Support bulk operations

### Long Term
- Multi-tenancy support
- Advanced querying with GraphQL
- Real-time updates with WebSockets
- Saga pattern for distributed transactions
- Event replay and temporal queries

## Conclusion

This design demonstrates how to apply DDD and Hexagonal Architecture principles to a Go project using the standard project layout. The key benefits:

**DDD Benefits:**
- Business logic is explicit and testable
- Rich domain model with behavior
- Clear separation of concerns
- Domain events enable loose coupling

**Hexagonal Benefits:**
- Framework independence
- Easy to test (mock ports)
- Flexible (swap implementations)
- Technology decisions deferred to adapters

**Project Layout Benefits:**
- Familiar structure for Go developers
- Clear organization and navigation
- Scalable as project grows
- Community-accepted patterns

The combination creates a maintainable, testable, and evolvable codebase that clearly expresses business intent while maintaining technical flexibility.
