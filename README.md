# Haiwo

![Haiwo logo](internal/server/web/logo.png)

Haiwo is a Go-based CI/CD MVP with a central server, pull-based agents, JSON-RPC communication, webhook/schedule/manual triggers, pipeline orchestration, and code rollback history.

The current implementation is an MVP scaffold. When Postgres is enabled, projects, pipelines, triggers, agents, runs, and settings are persisted through GORM models created with `AutoMigrate`.

## 1. Design Document

### 1.1 Goals

Haiwo is designed to provide a lightweight CI/CD control plane that can:

- run command, checkout, deploy, and rollback jobs on remote agents
- trigger pipelines manually, by schedule, or from GitHub/GitLab/Gitea webhook events
- match triggers by branch, comment, or branch + comment
- execute jobs on selected agents by ID or label
- support single-agent, parallel multi-agent, and chained multi-agent execution
- record run history and deployment metadata

### 1.2 Architecture

```text
GitHub/GitLab/Gitea
        |
        | webhook/comment/push/tag
        v
Haiwo Server  <---- JSON-RPC 2.0 over WebSocket ----  Haiwo Agent
        |                                                |
        | REST API + embedded Web UI                     | shell/git tools
        v                                                v
GORM AutoMigrate schema                          target host
```

Server responsibilities:

- serve the embedded web console at `/`
- expose REST APIs for projects, pipelines, runs, triggers, agents, and settings
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

Agent management:

- agents can be created from the web console with a name, labels, and concurrency; SSH entry metadata is enabled by default
- the console generates a deployment script containing `HAIWO_AGENT_ID`, `HAIWO_AGENT_TOKEN`, server WebSocket URL, labels, and SSH-related environment variables
- direct SSH metadata can produce an `ssh` command for reachable hosts
- reverse SSH over WebSocket is represented as configuration in this MVP; Settings stores only the public `http://` or `https://` base URL, and Haiwo generates the fixed `/ssh/agent` WebSocket path

### 1.3 Communication Model

Server and agent use JSON-RPC 2.0 over WebSocket.

Agent connection endpoint:

```text
GET /rpc/agent/ws
```

Authentication:

- Agent sends `Authorization: Bearer <token>`.
- MVP keeps `[AgentConfiguration].Token` in `configs/config.toml` as a development fallback token.
- Production-style agents should be created in the web console. Each created agent gets its own token.
- On the agent, the token is configured by `HAIWO_AGENT_TOKEN`.
- The server binds the WebSocket connection to the registered agent identity.
- If a per-agent token is used, the server only accepts registration for that token's agent ID.

Agent-to-server JSON-RPC methods:

- `agent.register`: register or re-register the agent.
- `agent.heartbeat`: report labels, status, version, and current load.
- `task.log`: stream stdout/stderr lines.
- `task.progress`: report task progress.
- `task.complete`: report final job status, exit code, and artifacts.
- `artifact.uploadComplete`: report artifact upload completion.

Server-to-agent JSON-RPC methods:

- `task.run`: execute a command, checkout, deploy, or rollback task.
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
- `rollback_code`

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

### 1.6 Operational Boundary

Haiwo focuses on code delivery, agent execution, trigger matching, run history, and code rollback. Data-layer operational work is handled manually outside Haiwo.

### 1.7 Frontend Design

The web console is embedded into the server binary and served from `/`.

The UI uses a Microsoft Fluent / Azure Portal inspired visual style: Segoe UI typography, neutral surfaces, Microsoft blue accents, compact command buttons, and thin bordered cards.

Implemented screens:

- Overview dashboard
- Chinese, English, and Japanese language switch
- password-protected login page
- Projects
- Pipelines with option-based creation for a simple stage/job
- Runs
- Agents
- Settings

The frontend is intentionally dependency-free for the MVP:

- `internal/server/web/index.html`
- `internal/server/web/styles.css`
- `internal/server/web/app.js`
- `internal/server/web/logo.png`

Brand asset:

- The Haiwo logo is served as `/logo.png`.
- The same logo is used by the password login page and the main console sidebar.

### 1.8 Current MVP Limits

- Runtime state falls back to memory when Postgres is disabled.
- When Postgres is enabled, core control-plane records are persisted and reloaded on server startup.
- Migrations use GORM `AutoMigrate`; SQL migration files are not used.
- Secrets are modeled but not yet encrypted or persisted.
- Web UI supports common MVP operations, not full rollback editing yet.
- Pipeline creation supports manual, push branch, push commit-message, and tag triggers plus ordered Agent command steps.
- Git provider API status updates are not implemented yet.

## 2. Usage Document

### 2.1 Requirements

Required:

- Go 1.26+
- PostgreSQL, when `PostgresConfiguration.Enabled = true`

### 2.2 Run Server

```sh
go run ./cmd/server -c config -cPath ./,./configs/
```

Or use Make:

```sh
make dev
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

Recommended workflow:

1. Open the web console.
2. Go to Agents.
3. If reverse SSH is needed, go to Settings and configure the public `http://` or `https://` Server URL.
4. Add an agent with a name, label multi-select, and concurrency. SSH entry metadata is enabled by default.
5. Copy the generated deployment script.
6. Put the built `haiwo-agent` binary on the target machine and run the script.

Local development fallback:

```sh
HAIWO_AGENT_ID=local-agent \
HAIWO_AGENT_LABELS=build,deploy,staging \
go run ./cmd/agent
```

The agent connects to:

```text
ws://localhost:8080/rpc/agent/ws
```

Or use Make:

```sh
make dev-agent
```

### 2.4 Docker Compose

Build the image:

```sh
make img-build
```

Start Postgres and the Haiwo server:

```sh
make up
```

Start Postgres, server, and the optional local compose agent:

```sh
make up-agent
```

Stop services:

```sh
make down
```

### 2.5 Configuration

Haiwo uses TOML configuration. The default file is:

```text
configs/config.toml
```

The server uses this flag style:

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
| `HAIWO_AGENT_LABELS` | `build,deploy,staging` | Comma-separated labels for scheduling |
| `HAIWO_AGENT_WORKDIR` | `.haiwo-agent` | Agent task workspace |
| `HAIWO_AGENT_MAX_RUNNING` | `1` | Maximum concurrent tasks accepted by the agent |
| `HAIWO_AGENT_SSH_ENABLED` | `false` | Marks the agent as SSH-enabled in the control plane |
| `HAIWO_AGENT_REVERSE_SSH_URL` | empty | Optional generated reverse SSH endpoint, for example `wss://haiwo.example.com/ssh/agent` |

### 2.6 Web Console Workflow

1. Start the server.
2. Open `http://localhost:8080/`.
3. Log in with `[WebConfiguration].Password`.
4. Switch language with `EN` / `中` / `日` in the top bar when needed.
5. Configure the public Server URL in Settings if agents need reverse SSH metadata. Haiwo automatically generates `/ssh/agent`.
6. Add an agent in the Agents screen and copy its deployment script.
7. Start at least one agent with that script.
8. Create a project in the Projects screen.
9. Create a pipeline in the Pipelines screen using the option-based form.
   Choose a trigger: manual only, push branch, push branch with commit-message text, or tag name.
   Add one or more ordered steps. Each step selects one Agent and command content; steps run sequentially.
10. Start a run in the Runs screen, or click `Run` in the Pipelines list to trigger the pipeline with the project's default branch.
11. Watch run status and agent status refresh automatically.

### 2.7 Database AutoMigrate

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

Start the server with database enabled; it runs GORM `AutoMigrate` during startup:

```sh
go run ./cmd/server -c config -cPath ./,./configs/
```

AutoMigrate models are defined in:

```text
internal/database/models.go
```

The AutoMigrate entrypoint is:

```text
internal/database/migrate.go
```

### 2.8 API Workflow

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

### 2.9 Pipeline JSON Example

```json
{
  "name": "staging deploy",
  "stages": [
    {
      "name": "step-1",
      "jobs": [
        {
          "id": "step-1-agent-a",
          "name": "1. agent-a",
          "job_type": "command",
          "agent_mode": "single",
          "agent_ids": ["agent-a"],
          "commands": ["echo deploy staging"],
          "timeout_seconds": 300,
          "required": true
        }
      ]
    },
    {
      "name": "step-2",
      "jobs": [
        {
          "id": "step-2-agent-b",
          "name": "2. agent-b",
          "job_type": "command",
          "agent_mode": "single",
          "agent_ids": ["agent-b"],
          "commands": ["systemctl restart app"],
          "timeout_seconds": 300,
          "required": true
        }
      ]
    }
  ]
}
```

Create a push trigger for a pipeline:

```sh
curl -sS -X POST http://localhost:8080/api/projects/PROJECT_ID/triggers \
  -H 'content-type: application/json' \
  -d '{"pipeline_id":"PIPELINE_ID","type":"push","branch_pattern":"main","commit_pattern":"deploy"}'
```

Create a tag trigger:

```sh
curl -sS -X POST http://localhost:8080/api/projects/PROJECT_ID/triggers \
  -H 'content-type: application/json' \
  -d '{"pipeline_id":"PIPELINE_ID","type":"tag","tag_pattern":"dev*"}'
```

### 2.10 Development Checks

Run tests:

```sh
go test ./...
```

Build server and agent:

```sh
make build
```

Format Go code:

```sh
make fmt
```
