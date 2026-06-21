<h1 align="center">Go Async Job Queue</h1>

<p align="center">
  <a href="LICENSE">
    <img src="https://img.shields.io/badge/License-MIT-green" />
  </a>
  <img src="https://img.shields.io/badge/status-educational-blue" />
  <img src="https://img.shields.io/badge/Backend-Go-00ADD8" />
  <img src="https://img.shields.io/badge/Database-PostgreSQL-4169E1" />
  <img src="https://img.shields.io/badge/Migrations-Goose-2F3136" />
  <img src="https://img.shields.io/badge/Queries-sqlc-2F3136" />
  <img src="https://img.shields.io/badge/Runtime-Docker_Compose-2496ED" />
  <img src="https://img.shields.io/badge/Metrics-Prometheus-E6522C" />
</p>

A production-inspired background job queue written in Go, backed by PostgreSQL.

The project demonstrates durable job submission, scheduled execution, idempotency, concurrent workers, retry with backoff, dead-letter handling, graceful shutdown, and Prometheus metrics.

## Features

- PostgreSQL-backed durable job queue
- REST API for submitting, reading, and cancelling jobs
- Idempotent job submission with `idempotency_key`
- Scheduled/delayed jobs via `run_at`
- Concurrent worker pool
- Safe job claiming with `FOR UPDATE SKIP LOCKED`
- Mock job execution for multiple job types
- Retry with backoff
- Dead-letter handling
- Graceful shutdown for HTTP server and workers
- Prometheus metrics exposed at `/metrics`
- Docker Compose setup for PostgreSQL and Prometheus

## Architecture

```text
Client
  |
  v
REST API
  |
  v
PostgreSQL jobs table
  |
  v
Worker pool
  |
  v
Mock job handler
  |
  +--> succeeded
  +--> pending retry
  +--> dead
```

Workers claim jobs with a PostgreSQL row-locking pattern:

```sql
FOR UPDATE SKIP LOCKED
```

This allows multiple workers, or multiple application instances, to safely process jobs from the same table without claiming the same job twice.

## Tech Stack

- Go `net/http`
- PostgreSQL
- Docker Compose
- Goose migrations
- sqlc
- pgx
- Prometheus Go client

## Project Structure

```text
cmd/server          application entrypoint
internal/api        HTTP router, handlers, and response helpers
internal/config     environment configuration
internal/db         PostgreSQL pool setup
internal/metrics    Prometheus metrics
internal/store      data access wrapper around sqlc
internal/worker     worker pool and job execution logic
migrations          Goose database migrations
prometheus.yml      Prometheus scrape configuration
```

## Getting Started

### Prerequisites

- Go
- Docker and Docker Compose
- Goose CLI
- sqlc CLI

### Environment

Copy the example environment file:

```bash
cp .env.example .env
```

Default local values:

```env
PORT=8080
DATABASE_URL=postgres://jobqueue:jobqueue@localhost:5433/jobqueue?sslmode=disable
WORKER_COUNT=3
WORKER_BATCH_SIZE=10
WORKER_POLL_INTERVAL=2s
WORKER_SHUTDOWN_TIMEOUT=30s
```

### Start Infrastructure

```bash
docker compose up -d
```

This starts PostgreSQL on `localhost:5433` and Prometheus on `localhost:9090`.

### Run Migrations

```bash
goose -dir migrations postgres "postgres://jobqueue:jobqueue@localhost:5433/jobqueue?sslmode=disable" up
```

Check migration status:

```bash
goose -dir migrations postgres "postgres://jobqueue:jobqueue@localhost:5433/jobqueue?sslmode=disable" status
```

### Run the Server

```bash
go run ./cmd/server
```

The API listens on:

```text
http://localhost:8080
```

## API

### Health Check

```bash
curl http://localhost:8080/healthz
```

### Submit a Job

```bash
curl -i -X POST http://localhost:8080/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "type": "send_email",
    "payload": {
      "to": "user@example.com",
      "subject": "Hello"
    },
    "idempotency_key": "email-user-001"
  }'
```

Example response:

```json
{
  "id": "0aaf2945-f996-4757-aaa1-cd857ce2f6dc",
  "type": "send_email",
  "payload": {
    "to": "user@example.com",
    "subject": "Hello"
  },
  "status": "pending",
  "attempt": 0,
  "max_attempts": 5,
  "run_at": "2026-06-21T08:12:56.085053Z",
  "idempotency_key": "email-user-001",
  "created_at": "2026-06-21T08:12:56.084469Z",
  "updated_at": "2026-06-21T08:12:56.084469Z"
}
```

### Submit a Scheduled Job

```bash
curl -i -X POST http://localhost:8080/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "type": "send_email",
    "payload": {
      "to": "scheduled@example.com",
      "subject": "Scheduled job"
    },
    "run_at": "2026-06-21T10:41:00+07:00",
    "idempotency_key": "scheduled-email-001"
  }'
```

Workers only claim jobs whose `run_at <= NOW()`.

### Get Job Status

```bash
curl http://localhost:8080/v1/jobs/<job_id>
```

### Cancel a Job

```bash
curl -i -X POST http://localhost:8080/v1/jobs/<job_id>/cancel
```

Only `pending` jobs can be cancelled.

If the job exists but is no longer pending, the API returns:

```text
409 Conflict
```

## Job Types

The worker executes mock handlers for these job types:

- `send_email`
- `generate_report`
- `webhook_delivery`
- `image_resize`
- `always_fail`
- `slow_job`

`always_fail` is useful for testing retry and dead-letter behavior.

`slow_job` is useful for testing graceful shutdown and lock release behavior.

## Job Statuses

```text
pending     waiting to be processed
processing  claimed by a worker
succeeded   completed successfully
cancelled   cancelled before processing
dead        exhausted retries and moved to dead-letter state
```

## Retry and Dead-Letter

When a job fails, the worker increments `attempt`, records `last_error`, clears the lock, and either schedules a retry or marks the job as `dead`.

Backoff schedule:

| Failed attempt | Retry delay |
| --- | --- |
| 1 | 5 seconds |
| 2 | 30 seconds |
| 3 | 2 minutes |
| 4+ | 10 minutes |

When `attempt >= max_attempts`, the job is marked as:

```text
dead
```

## Idempotency

Clients must provide an `idempotency_key` when submitting jobs.

If the same key is submitted again, the API returns the existing job instead of creating a duplicate.

```text
First request  -> 201 Created
Repeated key   -> 200 OK with existing job
```

## Worker Pool

Worker settings are configurable through environment variables:

```env
WORKER_COUNT=3
WORKER_BATCH_SIZE=10
WORKER_POLL_INTERVAL=2s
WORKER_SHUTDOWN_TIMEOUT=30s
```

Total worker concurrency is:

```text
number of app instances * WORKER_COUNT
```

All workers share the same PostgreSQL-backed queue. `FOR UPDATE SKIP LOCKED` prevents multiple workers from claiming the same job.

## Graceful Shutdown

On `SIGINT` or `SIGTERM`, the app:

1. Stops accepting new HTTP requests.
2. Stops workers from polling and claiming new jobs.
3. Waits for in-flight workers to finish.
4. Releases processing locks if worker shutdown exceeds `WORKER_SHUTDOWN_TIMEOUT`.
5. Closes the database pool.
6. Exits cleanly.

## Metrics

Metrics are exposed at:

```text
GET /metrics
```

Custom metrics:

```text
jobs_submitted_total
jobs_succeeded_total
jobs_failed_total
jobs_dead_total
jobs_processing
job_execution_duration_seconds
```

Prometheus is available at:

```text
http://localhost:9090
```

Check scrape targets:

```text
http://localhost:9090/targets
```

Example Prometheus queries:

```text
jobs_submitted_total
jobs_succeeded_total
job_execution_duration_seconds_count
```

## Useful Commands

Run tests:

```bash
go test ./...
```

Regenerate sqlc code:

```bash
sqlc generate
```

Open psql:

```bash
docker compose exec postgres psql -U jobqueue -d jobqueue
```

Stop Docker services while keeping data:

```bash
docker compose down
```

Stop Docker services and remove volumes:

```bash
docker compose down -v
```

## Resume Bullet

Built a PostgreSQL-backed background job queue in Go with concurrent worker pool, delayed jobs, idempotent submission, retry with exponential backoff, dead-letter handling, graceful shutdown, and Prometheus metrics.

## Disclaimer

This code is for educational purposes only, has not been audited, and is provided without any warranties or guarantees.

## License

This project is licensed under the MIT License.
