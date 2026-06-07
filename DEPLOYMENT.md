# Haiwo Deployment Guide

本文档记录 Haiwo 的本地部署、生产部署、Agent 部署、Webhook 配置和常见问题。

## 1. 本地启动

本地 PostgreSQL 已启动时，默认配置文件是：

```text
configs/config.toml
```

当前开发配置使用：

```toml
[PostgresConfiguration]
Enabled = true
Host = "127.0.0.1"
Port = 18432
User = "pacman"
Password = "MEB3oZZkpxgA1q4WcA10"
DBName = "pacman"
SSLMode = "disable"
```

启动 Server：

```sh
make dev
```

等价命令：

```sh
go run ./cmd/server -c config -cPath ./,./configs/
```

打开控制台：

```text
http://localhost:8080/
```

默认登录密码来自 `[WebConfiguration].Password`，当前默认是：

```text
haiwo
```

## 2. 数据库和迁移

Haiwo 使用 GORM `AutoMigrate`，不再需要单独的 `cmd/migrate` 命令。

Server 启动时如果 `[PostgresConfiguration].Enabled = true`，会自动连接 PostgreSQL 并创建/更新表结构。

相关代码：

```text
internal/database/client.go
internal/database/migrate.go
internal/database/models.go
```

如果重启后项目、流水线、Agent 消失，先检查：

- `PostgresConfiguration.Enabled` 是否为 `true`
- Server 是否连接到了同一个 PostgreSQL 实例
- `Host`、`Port`、`User`、`DBName` 是否和期望一致
- 容器部署时配置文件是否正确挂载

## 3. 构建

构建 Server 和 Agent：

```sh
make build
```

构建 Linux x86_64 Agent：

```sh
mkdir -p bin
GOOS=linux GOARCH=amd64 go build -o bin/haiwo-agent-linux-amd64 ./cmd/agent
```

Agent 安装脚本默认下载地址：

```text
https://fileoss.hacksnews.top/haiwo-agent-linux-amd64
```

更新 Agent 时，先上传新的 `bin/haiwo-agent-linux-amd64` 到文件服务器，再在目标机器重新执行控制台复制的安装命令。

## 4. Docker Compose

构建镜像：

```sh
make img-build
```

启动服务：

```sh
make up
```

启动带本地 Agent 的开发栈：

```sh
make up-agent
```

查看日志：

```sh
make logs
```

停止：

```sh
make down
```

Compose 部署时需要确认容器内 Server 能访问 PostgreSQL。容器内不能使用宿主机的 `127.0.0.1:18432` 访问外部数据库，应该改成可达的数据库主机名或 Docker 网络服务名。

## 5. Server 公开地址

生产环境需要在控制台 Settings 填写 Server 域名，只填 `http://` 或 `https://` 开头的 base URL，例如：

```text
https://cicd.kcgi.edu.kg
```

Haiwo 会自动生成固定路径：

```text
/rpc/agent/ws
/ssh/agent
/api/agents/{id}/deploy.sh
/api/webhooks/{provider}/{project_id}
```

不需要在设置里手动填写完整 WebSocket 路径。

## 6. Agent 部署

推荐流程：

1. 打开 Haiwo 控制台。
2. 进入 Settings，配置 Server 公开地址。
3. 进入 Agent 页面，点击 `+ 添加`。
4. 输入 Agent 名称，选择标签，设置并发数。
5. 创建后点击 Agent 行里的 `脚本`。
6. 控制台会复制类似下面的命令：

```sh
curl -fsSL 'https://cicd.kcgi.edu.kg/api/agents/agent_xxx/deploy.sh?token=hwagt_xxx' | bash
```

脚本行为：

- 下载最新 Linux x86_64 Agent 二进制
- 如果旧 Agent 进程存在，先停止
- 替换本地 Agent 文件
- 写入 Agent ID、Token、标签、Server WebSocket 地址
- 启动 Agent

Agent 连接 Server 使用 JSON-RPC over WebSocket：

```text
wss://<server-domain>/rpc/agent/ws
```

Agent 会自动重连。Server 重启后，Agent 应该重新上线。

## 7. Web SSH

Agent 默认开启反向 WebSocket SSH 入口。Agent 可以部署在内网，不要求 Server 主动 SSH 到 Agent。

控制台 Agent 行里的 `SSH` 按钮会打开：

```text
/ssh/agents/{agent_id}
```

如果按钮只显示配置文本或无法进入终端，通常是以下原因：

- Agent 二进制不是最新版本
- Agent 没有连上 Server
- Server 公开地址没有配置
- 反向 WebSocket 被代理层拦截

## 8. GitHub Webhook

在 Projects 页面，每个项目行有一个 Hook 复制按钮。项目配置为 GitHub 时，只需要复制 GitHub 对应的 URL。

### 8.1 创建 GitHub 项目

在 Haiwo 的 Projects 页面点击 `+ 添加`：

- 名称：项目在 Haiwo 里的显示名，例如 `kanfan`
- 代码平台：选择 `GitHub`
- 仓库地址：GitHub clone URL，例如 `https://github.com/dollarkillerx/kanfan` 或 `git@github.com:dollarkillerx/kanfan.git`
- 默认分支：通常是 `main`

创建后，项目列表会显示 Hook 复制入口。GitHub 项目只需要复制 GitHub 对应的 Hook URL，不需要复制 GitLab/Gitea URL。

### 8.2 配置 GitHub Webhook

打开 GitHub 仓库：

```text
Settings -> Webhooks -> Add webhook
```

填写：

- Payload URL: Haiwo 复制出来的 GitHub Hook URL
- Content type: `application/json`
- Secret: MVP 可留空
- SSL verification: Enable SSL verification
- Active: 勾选

事件选择按流水线触发方式决定：

| Haiwo 流水线触发 | GitHub 事件选择 |
| --- | --- |
| Push 分支 | `Just the push event` |
| Push 分支 + Commit 信息包含 | `Just the push event` |
| Tag 名称 | 推送 tag 会通过 push payload 发送，保留 push event 即可 |
| 评论命令 | `Let me select individual events`，选择 issue comment / pull request review comment 相关事件 |

GitHub 创建 webhook 后会立即发送 `ping` 事件。Haiwo 对 `ping` 的响应类似：

```json
{
  "event": "ping",
  "ignored": true,
  "reason": "ping event only verifies webhook connectivity",
  "matched": 0,
  "runs": []
}
```

这是正常行为。`ping` 只验证连通性，不应该匹配流水线，也不会启动运行。

### 8.3 Push 触发流水线

在 Haiwo 的 Pipelines 页面创建或编辑流水线：

- 触发方式：`Push 分支`
- 分支：例如 `main`
- Commit 信息包含：可选，例如 `deploy`
- 执行步骤：选择 Agent，并填写命令

匹配逻辑：

- 分支必须匹配流水线配置里的分支，例如 `main`
- 如果配置了 `Commit 信息包含`，提交信息必须包含该文本，例如 `deploy`
- 两个条件都满足才会创建 run

示例：

```sh
git checkout main
git commit --allow-empty -m "deploy staging"
git push origin main
```

如果 commit message 是 `fix typo`，而流水线要求包含 `deploy`，则不会触发。

### 8.4 Tag 触发流水线

在 Haiwo 创建流水线：

- 触发方式：`Tag`
- Tag 名称：例如 `dev*` 或 `v1.0.0`

推送 tag：

```sh
git tag dev-001
git push origin dev-001
```

Haiwo 收到 `refs/tags/dev-001` 后，会用 tag 名称匹配流水线配置。

### 8.5 评论触发流水线

在 Haiwo 创建流水线：

- 触发方式：评论命令
- 评论内容：例如 `/deploy staging`
- 如需限制分支，配置分支条件

GitHub Webhook 事件选择 `Let me select individual events`，再选择 issue comment / pull request review comment 相关事件。

然后在 GitHub issue 或 PR 里评论：

```text
/deploy staging
```

Haiwo 会根据评论内容匹配流水线。

### 8.6 GitHub 排查

如果 GitHub 显示 webhook 调用成功，但 Haiwo 返回 `matched: 0`：

- 如果事件是 `ping`，这是正常行为
- 检查流水线触发方式是否和 GitHub 事件一致
- 检查分支是否一致，例如 GitHub 推的是 `main`，流水线也必须配置 `main`
- 检查 commit message 是否包含配置的文本，例如 `deploy`
- 检查项目 Hook URL 的 project ID 是否属于当前项目
- 检查 GitHub Content type 是否为 `application/json`

如果 GitHub 请求不到 Haiwo：

- 检查 Settings 里的 Server URL 是否是公网可访问域名
- 检查 HTTPS 证书和 SSL verification
- 检查 Cloudflare/WAF 是否拦截 `/api/webhooks/github/*`
- 检查反向代理是否正确转发 POST body 和 headers

## 9. GitLab 和 Gitea Webhook

GitLab 和 Gitea 同样从 Projects 页面复制对应项目 provider 的 Hook URL：

```text
https://<server>/api/webhooks/gitlab/{project_id}
https://<server>/api/webhooks/gitea/{project_id}
```

推荐配置：

- Push events 对应 Push 分支触发
- Tag push events 对应 Tag 触发
- Note/comment events 对应评论命令触发
- Payload 使用 JSON

## 10. 流水线

流水线在 Pipelines 页面创建和编辑，不需要手写 JSON。

当前表单支持：

- 手动触发
- Push 指定分支
- Push 指定分支，并要求 commit message 包含某段文本
- Tag 名称触发
- 顺序添加多个 Agent 执行步骤

每个步骤选择一个 Agent 和一段 shell 命令。Agent 会把命令作为同一个 shell 脚本执行，所以：

```sh
cd /opt/workspace/build/kanfan
pwd
git pull
```

会在同一个 shell 中运行，`cd` 会对后续命令生效。

## 11. 运行历史和日志

Runs 页面只负责查看历史，不再放手动执行表单。

能力：

- 按流水线筛选
- 最新运行显示在最前面
- 展示状态、来源、ref、总耗时、更新时间
- 点击详情查看错误和日志

如果运行失败但没有日志，优先检查：

- Agent 是否在线
- Agent 是否为最新版本
- 任务是否在下发到 Agent 前就失败
- Pipeline 步骤是否选择了正确 Agent
- 命令是否在 Agent 机器上可执行

## 12. Cloudflare 和代理

如果安装脚本请求返回 403，例如：

```sh
curl -fsSL 'https://cicd.kcgi.edu.kg/api/agents/.../deploy.sh?token=...' | bash
```

通常是 Cloudflare/WAF 拦截了脚本下载请求。处理方式：

- 允许 `/api/agents/*/deploy.sh`
- 允许 `curl` User-Agent 或关闭该路径的 Bot/WAF 规则
- 确认 HTTPS 代理会转发 query string
- 确认 WebSocket 路径 `/rpc/agent/ws` 和 `/ssh/agent` 支持 Upgrade

Nginx 反代 WebSocket 需要包含：

```nginx
proxy_http_version 1.1;
proxy_set_header Upgrade $http_upgrade;
proxy_set_header Connection "upgrade";
proxy_set_header Host $host;
```

## 13. 验证

运行测试：

```sh
go test ./...
```

检查前端脚本语法：

```sh
node --check internal/server/web/app.js
```

检查服务：

```sh
curl -sS http://localhost:8080/healthz
```
