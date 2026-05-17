# Haiwo

Haiwo is a Go-based CI/CD MVP with a central server, pull-based agents, JSON-RPC communication, webhook/schedule/manual triggers, pipeline orchestration, and MySQL/PostgreSQL backup and restore jobs.

The current implementation is an MVP scaffold. Runtime state is still stored in memory, but the database schema is now defined as GORM models and created with `AutoMigrate`.

## 1. Design Document

### 1.1 Goals

Haiwo is designed to provide a lightweight CI/CD control plane that can:

- run command, checkout, backup, restore, and rollback jobs on remote agents
- trigger pipelines manually, by schedule, or from GitHub/GitLab/Gitea webhook events
- match triggers by branch, comment, or branch + comment
- execute jobs on selected agents by ID or label
- support single-agent, parallel multi-agent, and chained multi-agent execution
- record run history, deployment metadata, and database backup metadata
- support MySQL and PostgreSQL backup/restore through agent-side database clients

### 1.2 Architecture

```text
GitHub/GitLab/Gitea
        |
        | webhook/comment/push/tag
        v
Haiwo Server  <---- JSON-RPC 2.0 over WebSocket ----  Haiwo Agent
        |                                                |
        | REST API + embedded Web UI                     | shell/git/db tools
        v                                                v
GORM AutoMigrate schema                          target host / database
```

Server responsibilities:

- serve the embedded web console at `/`
- expose REST APIs for projects, pipelines, runs, triggers, agents, databases, and backups
- accept GitHub/GitLab/Gitea webhook events
- match triggers and create pipeline runs
- orchestrate stages and jobs
- dispatch tasks to online agents through JSON-RPC
- receive agent heartbeat, logs, progress, completion, and artifact events

Agent responsibilities:

- connect to the server as a JSON-RPC WebSocket client
- register itself with labels and capacity
- keep heartbeat status updated
- execute server-dispatched tasks
- stream logs and completion status back to the server
- run database backup/restore commands locally on the agent machine

### 1.3 Communication Model

Server and agent use JSON-RPC 2.0 over WebSocket.

Agent connection endpoint:

```text
GET /rpc/agent/ws
```

Authentication:

- Agent sends `Authorization: Bearer <token>`.
- On the server, the token is configured by `[AgentConfiguration].Token` in `configs/config.toml`.
- On the agent, the same token is configured by `HAIWO_AGENT_TOKEN`.
- The server binds the WebSocket connection to the registered agent identity.

Agent-to-server JSON-RPC methods:

- `agent.register`: register or re-register the agent.
- `agent.heartbeat`: report labels, status, version, and current load.
- `task.log`: stream stdout/stderr lines.
- `task.progress`: report task progress.
- `task.complete`: report final job status, exit code, artifacts, and backup ID.
- `artifact.uploadComplete`: report artifact upload completion.

Server-to-agent JSON-RPC methods:

- `task.run`: execute a command, checkout, backup, restore, or rollback task.
- `task.cancel`: cancel a running task.
- `agent.ping`: check connection health.
- `agent.updateConfig`: update labels, capacity, or work directory.

### 1.4 Domain Model

Project:

- source repository metadata
- provider: `github`, `gitlab`, or `gitea`
- default branch

Pipeline:

- belongs to a project
- contains ordered stages
- each stage contains jobs

Stage:

- stages run serially
- jobs inside one stage run concurrently
- MVP uses fail-fast behavior for required jobs

Job:

- `command`
- `git_checkout`
- `db_backup`
- `db_restore`
- `rollback_code`
- `rollback_db`

Agent selection:

- by explicit `agent_ids`
- by required `agent_labels`

Agent execution mode:

- `single`: choose one matching agent
- `parallel`: execute on all matching agents concurrently
- `chain`: execute on matching agents one by one

Run:

- immutable execution history record
- tracks source, ref, comment, status, timestamps, and metadata
- rollback creates a new run instead of modifying old history

Backup:

- records backup ID, database type, database name, environment, path, checksum, agent, run, job, and creation time

### 1.5 Trigger Design

Supported trigger types:

- `manual`: user starts a pipeline through API or web UI
- `schedule`: cron expression starts a pipeline
- `webhook`: GitHub/GitLab/Gitea events start a pipeline

Webhook activation supports:

- push
- tag
- pull request / merge request comment
- issue comment style command

Trigger matching supports:

- branch pattern, for example `main` or `release/*`
- comment pattern, for example `/deploy prod`
- branch + comment together, for example only `main` branch can run `/deploy prod`

### 1.6 Database Backup And Restore

MVP supports:

- MySQL
- PostgreSQL

Database target fields:

- `type`: `mysql` or `postgresql`
- `host`
- `port`
- `database`
- `username`
- `password`
- `password_secret_id`
- `extra_options`
- `environment`
- `allowed_agent_labels`
- `backup_format`
- `confirm`

MySQL backup:

```text
mysqldump --single-transaction --routines --triggers --events
```

MySQL restore:

```text
mysql < backup.sql.gz
```

PostgreSQL backup:

```text
pg_dump -Fc
```

PostgreSQL restore:

```text
pg_restore --clean --if-exists
```

Production restore protection:

- if `database_target.environment == "prod"`, restore requires `confirm=true`

### 1.7 Frontend Design

The web console is embedded into the server binary and served from `/`.

Implemented screens:

- Overview dashboard
- Chinese and English language switch
- password-protected login page
- Projects
- Pipelines
- Runs
- Agents
- Backups

The frontend is intentionally dependency-free for the MVP:

- `internal/server/web/index.html`
- `internal/server/web/styles.css`
- `internal/server/web/app.js`

### 1.8 Current MVP Limits

- Runtime state is in memory.
- GORM schema exists, but the store is not yet wired to Postgres.
- Migrations use GORM `AutoMigrate`; SQL migration files are not used.
- Secrets are modeled but not yet encrypted or persisted.
- Web UI supports common MVP operations, not full trigger/database/rollback editing yet.
- Backup files are stored on the agent local filesystem.
- Git provider API status updates are not implemented yet.

## 2. Usage Document

### 2.1 Requirements

Required:

- Go 1.26+
- PostgreSQL, when `PostgresConfiguration.Enabled = true`

Required on agent machines for database jobs:

- MySQL: `mysqldump` and `mysql`
- PostgreSQL: `pg_dump` and `pg_restore`

### 2.2 Run Server

```sh
go run ./cmd/server -c config -cPath ./,./configs/
```

Default server address:

```text
http://localhost:8080
```

Open the web console:

```text
http://localhost:8080/
```

### 2.3 Run Agent

```sh
HAIWO_AGENT_ID=local-agent \
HAIWO_AGENT_LABELS=build,deploy,db,staging \
go run ./cmd/agent
```

The agent connects to:

```text
ws://localhost:8080/rpc/agent/ws
```

### 2.4 Configuration

Haiwo uses TOML configuration. The default file is:

```text
configs/config.toml
```

The server and migrate command follow the same flag style:

```sh
-c config -cPath ./,./configs/
```

Server environment variables:

```toml
[ServiceConfiguration]
Addr = ":8080"
Debug = true

[AgentConfiguration]
Token = "dev-agent-token"

[WebConfiguration]
Password = "haiwo"

[PostgresConfiguration]
Enabled = false
Host = "127.0.0.1"
Port = 5432
User = "haiwo"
Password = "haiwo"
DBName = "haiwo"
SSLMode = "disable"
TimeZone = "UTC"
LogMode = "none"
```

Postgres settings:

| Name | Description |
| --- | --- |
| `Enabled` | When true, server startup connects to Postgres and runs GORM `AutoMigrate` |
| `Host` | PostgreSQL host |
| `Port` | PostgreSQL port |
| `User` | PostgreSQL username |
| `Password` | PostgreSQL password |
| `DBName` | PostgreSQL database name |
| `SSLMode` | PostgreSQL sslmode, usually `disable` locally |
| `TimeZone` | Session timezone |
| `LogMode` | `none`, `console`, or `slow_query` |

Web settings:

| Name | Description |
| --- | --- |
| `Password` | Password required to open the web console. Set an empty value only for local development without login. |

Agent environment variables:

| Name | Default | Description |
| --- | --- | --- |
| `HAIWO_SERVER_URL` | `ws://localhost:8080/rpc/agent/ws` | Server JSON-RPC WebSocket endpoint |
| `HAIWO_AGENT_TOKEN` | `dev-agent-token` | Shared token sent to the server |
| `HAIWO_AGENT_ID` | `local-agent` | Agent ID |
| `HAIWO_AGENT_NAME` | `local-agent` | Human-readable agent name |
| `HAIWO_AGENT_LABELS` | `build,deploy,db,staging` | Comma-separated labels for scheduling |
| `HAIWO_AGENT_WORKDIR` | `.haiwo-agent` | Agent task and backup workspace |

### 2.5 Web Console Workflow

1. Start the server.
2. Open `http://localhost:8080/`.
3. Log in with `[WebConfiguration].Password`.
4. Switch language with `EN` / `中` in the top bar when needed.
5. Start at least one agent.
6. Create a project in the Projects screen.
7. Create a pipeline in the Pipelines screen using JSON.
8. Start a run in the Runs screen.
9. Watch run status, agent status, and backup records refresh automatically.

### 2.6 Database Migration

Enable Postgres in `configs/config.toml`:

```toml
[PostgresConfiguration]
Enabled = true
Host = "127.0.0.1"
Port = 5432
User = "haiwo"
Password = "haiwo"
DBName = "haiwo"
SSLMode = "disable"
TimeZone = "UTC"
LogMode = "console"
```

Run GORM AutoMigrate without starting the server:

```sh
go run ./cmd/migrate -c config -cPath ./,./configs/
```

Or start the server with database enabled; it will run AutoMigrate during startup:

```sh
go run ./cmd/server -c config -cPath ./,./configs/
```

AutoMigrate models are defined in:

```text
internal/database/models.go
```

The migration entrypoint is:

```text
internal/database/migrate.go
```

### 2.7 API Workflow

Create a project:

```sh
curl -sS -X POST http://localhost:8080/api/projects \
  -H 'content-type: application/json' \
  -d '{"name":"demo","provider":"github","repo_url":"https://github.com/example/app.git","default_branch":"main"}'
```

Create a pipeline:

```sh
curl -sS -X POST http://localhost:8080/api/projects/PROJECT_ID/pipelines \
  -H 'content-type: application/json' \
  -d @examples/pipeline.json
```

Run a pipeline manually:

```sh
curl -sS -X POST http://localhost:8080/api/pipelines/PIPELINE_ID/runs \
  -H 'content-type: application/json' \
  -d '{"type":"manual","ref":"main"}'
```

Check agents:

```sh
curl -sS http://localhost:8080/api/agents
```

Check runs:

```sh
curl -sS http://localhost:8080/api/runs
```

Check one run:

```sh
curl -sS http://localhost:8080/api/runs/RUN_ID
```

Check backups:

```sh
curl -sS http://localhost:8080/api/backups
```

### 2.8 Pipeline JSON Example

```json
{
  "name": "staging deploy",
  "stages": [
    {
      "name": "deploy",
      "jobs": [
        {
          "id": "deploy",
          "name": "Deploy service",
          "job_type": "command",
          "agent_mode": "single",
          "agent_labels": ["deploy", "staging"],
          "commands": ["echo deploy staging"],
          "timeout_seconds": 300,
          "required": true
        }
      ]
    }
  ]
}
```

### 2.9 Database Backup Job Example

```json
{
  "id": "backup-prod",
  "name": "Backup production database",
  "job_type": "db_backup",
  "agent_mode": "single",
  "agent_labels": ["db", "prod"],
  "database_target": {
    "type": "mysql",
    "host": "127.0.0.1",
    "port": 3306,
    "database": "app",
    "username": "root",
    "password": "secret",
    "environment": "prod"
  },
  "required": true
}
```

### 2.10 Database Restore Job Example

```json
{
  "id": "restore-prod",
  "name": "Restore production database",
  "job_type": "db_restore",
  "agent_mode": "single",
  "agent_labels": ["db", "prod"],
  "backup_id": "bak_20260510120000",
  "database_target": {
    "type": "postgresql",
    "host": "127.0.0.1",
    "port": 5432,
    "database": "app",
    "username": "postgres",
    "password": "secret",
    "environment": "prod",
    "confirm": true
  },
  "required": true
}
```

### 2.11 Development Checks

Run tests:

```sh
go test ./...
```

Build server and agent:

```sh
go build ./cmd/server ./cmd/agent
```

Build migrate command:

```sh
go build ./cmd/migrate
```

Format Go code:

```sh
gofmt -w cmd internal
```
