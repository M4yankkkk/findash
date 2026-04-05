# Finance Dashboard API

A backend REST API for a role-based finance dashboard system. Built with Go, Gin, and PostgreSQL.

---

## Tech Stack

| Layer      | Technology          |
|------------|---------------------|
| Language   | Go 1.21             |
| Framework  | Gin                 |
| Database   | PostgreSQL          |
| Auth       | JWT (HS256)         |
| Docs       | Swagger (swaggo)    |

---

## Features

- **Role-Based Access Control** — three roles: `admin`, `manager`, `viewer`
- **Financial Entry Management** — create, read, update, soft-delete income/expense records
- **Analytics** — summary totals, category breakdown, monthly trend
- **Audit Logging** — every significant action is recorded
- **Pagination + Search** — all list endpoints support page, page_size, and filters
- **Soft Deletes** — financial records are never hard-deleted
- **JWT Authentication** — stateless, role-aware tokens

---

## Project Structure

```
finance-dashboard/
├── cmd/
│   └── main.go                        # Entry point
├── internal/
│   ├── config/config.go               # Environment config
│   ├── database/postgres.go           # DB connection + migration runner
│   ├── models/                        # Data structs (user, entry, audit_log)
│   ├── repository/                    # DB queries (user, entry, audit, analytics)
│   ├── services/                      # Business logic
│   ├── handlers/                      # HTTP layer
│   ├── middleware/                    # Auth + role middleware
│   └── routes/routes.go              # All route registrations
├── pkg/utils/                         # Shared helpers (response, jwt, validator)
├── migrations/                        # SQL migration files
├── docs/                              # Auto-generated Swagger docs
├── .env.example
└── README.md
```

---

## Setup

### Prerequisites
- Go 1.21+
- PostgreSQL 14+

### 1. Clone and install dependencies

```bash
git clone https://github.com/mayank/finance-dashboard
cd finance-dashboard
go mod tidy
```

### 2. Configure environment

```bash
cp .env.example .env
```

Edit `.env` with your values:

```env
PORT=8080
GIN_MODE=debug

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=yourpassword
DB_NAME=finance_dashboard
DB_SSLMODE=disable

JWT_SECRET=your_super_secret_key_min_32_chars
JWT_EXPIRY_HOURS=24
```

### 3. Create the database

```bash
psql -U postgres -c "CREATE DATABASE finance_dashboard;"
```

### 4. Run the server

```bash
go run cmd/main.go
```

Migrations run automatically on startup. You should see:

```
✅ Database connected
✅ Migration applied: 001_create_users.sql
✅ Migration applied: 002_create_entries.sql
✅ Migration applied: 003_create_audit_logs.sql
🚀 Server running on port 8080
```

---

## Roles and Permissions

| Action                  | Viewer | Manager | Admin |
|-------------------------|--------|---------|-------|
| Register / Login        | ✅     | ✅      | ✅    |
| View own entries        | ✅     | ✅      | ✅    |
| View all entries        | ❌     | ❌      | ✅    |
| Create entry            | ❌     | ✅      | ✅    |
| Update own entry        | ❌     | ✅      | ✅    |
| Delete own entry        | ❌     | ✅      | ✅    |
| Update/delete any entry | ❌     | ❌      | ✅    |
| View analytics          | ❌     | ✅      | ✅    |
| List all users          | ❌     | ❌      | ✅    |
| Change user roles       | ❌     | ❌      | ✅    |

> The first registered user is automatically assigned the `admin` role.

---

## API Reference

All endpoints are prefixed with `/api/v1`.  
Protected endpoints require: `Authorization: Bearer <token>`

### Auth

| Method | Endpoint         | Auth | Description          |
|--------|------------------|------|----------------------|
| POST   | /auth/register   | ❌   | Register a new user  |
| POST   | /auth/login      | ❌   | Login, receive JWT   |

**Register**
```json
POST /api/v1/auth/register
{
  "name": "Mayank",
  "email": "mayank@example.com",
  "password": "password123"
}
```

**Login**
```json
POST /api/v1/auth/login
{
  "email": "mayank@example.com",
  "password": "password123"
}
```

---

### Users (Admin only)

| Method | Endpoint            | Description         |
|--------|---------------------|---------------------|
| GET    | /users              | List all users      |
| GET    | /users/:id          | Get user by ID      |
| PATCH  | /users/:id/role     | Update user's role  |

**Query params for GET /users:**
- `page` (default: 1)
- `page_size` (default: 20, max: 100)

**Update role body:**
```json
{ "role": "manager" }
```

---

### Financial Entries

| Method | Endpoint       | Roles           | Description              |
|--------|----------------|-----------------|--------------------------|
| POST   | /entries       | manager, admin  | Create a new entry       |
| GET    | /entries       | all             | List entries (paginated) |
| GET    | /entries/:id   | all             | Get entry by ID          |
| PUT    | /entries/:id   | manager, admin  | Update an entry          |
| DELETE | /entries/:id   | manager, admin  | Soft-delete an entry     |

**Create entry body:**
```json
{
  "title": "Monthly Salary",
  "amount": 75000,
  "type": "income",
  "category": "salary",
  "description": "April salary",
  "date": "2026-04-01T00:00:00Z"
}
```

**Query params for GET /entries:**
- `page` (default: 1)
- `page_size` (default: 20)
- `type` — `income` or `expense`
- `category` — filter by category name
- `date_from` — format: `YYYY-MM-DD`
- `date_to` — format: `YYYY-MM-DD`

---

### Analytics (Manager + Admin)

| Method | Endpoint               | Description                        |
|--------|------------------------|------------------------------------|
| GET    | /analytics/summary     | Total income, expenses, net balance|
| GET    | /analytics/by-category | Breakdown by category              |
| GET    | /analytics/trend       | Monthly income vs expense trend    |

**Query params for GET /analytics/trend:**
- `months` — number of months to include (default: 6, max: 24)

**Sample summary response:**
```json
{
  "success": true,
  "data": {
    "total_income": 150000.00,
    "total_expenses": 42000.00,
    "net_balance": 108000.00,
    "entry_count": 12
  }
}
```

---

## Standard Response Format

All responses follow a consistent envelope:

```json
{
  "success": true,
  "message": "operation description",
  "data": { }
}
```

Error responses:
```json
{
  "success": false,
  "error": "human readable error message"
}
```

Paginated responses:
```json
{
  "success": true,
  "data": {
    "items": [],
    "total": 42,
    "page": 1,
    "page_size": 20,
    "has_more": true
  }
}
```

---

## Status Codes

| Code | Meaning                              |
|------|--------------------------------------|
| 200  | Success                              |
| 201  | Resource created                     |
| 400  | Bad request / validation error       |
| 401  | Unauthenticated (missing/invalid JWT)|
| 403  | Forbidden (insufficient role)        |
| 404  | Resource not found                   |
| 409  | Conflict (e.g. email already exists) |
| 500  | Internal server error                |

---

## Design Decisions and Tradeoffs

**Soft deletes over hard deletes**  
Financial records are never permanently deleted. `deleted_at` is set instead of removing the row. This preserves audit integrity and allows potential recovery.

**Generic auth error messages**  
Login always returns `"invalid email or password"` regardless of whether the email exists. This prevents user enumeration attacks.

**First user is admin**  
The first registered user automatically receives the `admin` role. This avoids needing a manual seeding step and is documented clearly.

**Self-role-change prevention**  
An admin cannot change their own role, preventing accidental lockout.

**Non-fatal audit logging**  
Audit log failures use `_ =` (ignored errors) so a logging hiccup never blocks the main operation.

**Admin data scoping**  
Admins see all data in analytics and entries. Other roles are scoped to their own records at the service layer, not the route layer.

**No unit tests**  
Given the assignment timeline, unit tests were omitted in favour of a complete, well-structured implementation. The handler → service → repository separation makes the codebase straightforward to test with standard Go testing patterns.

---

## Swagger Docs

After installing swaggo:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
swag init -g cmd/main.go
```

Then visit: `http://localhost:8080/swagger/index.html`