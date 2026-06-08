# Haiwo

![Haiwo logo](internal/server/web/logo.png)

[中文说明](README_CN.md) | [Deployment Guide](DEPLOYMENT.md)

Haiwo is a lightweight, self-hosted CI/CD platform written in Go. A central server with an embedded web console drives pull-based agents that run your build and deploy commands on remote machines — including machines behind NAT or inside a private network.

## Features

- **Server + Agent** — agents dial out to the server over WebSocket, so they work behind NAT / in private networks.
- **Triggers** — manual, cron schedule, and GitHub / GitLab / Gitea webhooks (push branch, commit-message match, tag, comment command).
- **Visual pipelines** — create and edit pipelines in the UI, no hand-written JSON; ordered multi-agent steps.
- **Run history** — per-run status, source, duration, live logs and errors.
- **Agents dashboard** — online status, load, CPU and memory.
- **Web SSH** — reverse-WebSocket terminal into agents, no inbound SSH required.
- **Embedded console** — single self-contained binary (frontend included); Chinese / English / Japanese; password-protected login.
- **Storage** — file backend by default (no database required), or PostgreSQL.

## Deployment

### 1. Binary (Linux)

One-line interactive installer. It lets you pick a language, then set the bind port and console password on first install; on later runs it offers **restart / stop / update / change password**.

```sh
curl -fsSL https://raw.githubusercontent.com/dollarkillerx/Haiwo/refs/heads/main/scripts/install-server.sh | bash
```

Open the console at `http://<server-ip>:<port>/` (default port `8080`). Data is stored under `~/.haiwo/server/data` (file backend, no database needed).

### 2. Docker

```sh
docker compose up -d
```

Open the console at `http://<server-ip>:8181/`. State is persisted to `./data` (file backend). To switch to PostgreSQL, edit `configs/config.toml`. See the [Deployment Guide](DEPLOYMENT.md).

---

After the server is running, log in, add an **Agent** in the console, and copy its one-line deploy command onto the target machine to bring the agent online.

## Documentation

- [Deployment Guide](DEPLOYMENT.md) — full configuration, storage backends, webhooks, Web SSH, building and cross-compiling.
- [中文说明](README_CN.md)

## License

See [LICENSE](LICENSE).
