# Haiwo

![Haiwo logo](internal/server/web/logo.png)

[English README](README.md) | [部署文档](DEPLOYMENT.md)

Haiwo 是一个使用 Go + PostgreSQL 构建的轻量 CI/CD MVP。系统由 Server 和 Agent 组成：Server 负责项目、流水线、触发器、运行历史和 Web 控制台；Agent 主动连接 Server，负责在目标机器上执行命令、拉取代码、回传日志和状态。

当前版本使用 GORM `AutoMigrate` 自动维护数据库表结构。开启 PostgreSQL 后，项目、流水线、触发器、Agent、运行记录和系统设置会持久化保存。

## 功能概览

- Server + Agent 架构
- JSON-RPC 2.0 over WebSocket 通信
- Agent 主动连接 Server，适合内网机器
- GitHub / GitLab / Gitea Webhook 触发
- 手动触发、Push 分支触发、Push commit 文本触发、Tag 触发、评论命令触发
- 可视化创建和编辑流水线，不需要手写 JSON
- 多 Agent 顺序执行步骤
- 运行历史、运行详情、错误和日志查看
- Agent 在线状态、负载、CPU、内存信息
- 反向 WebSocket WebSSH 入口
- 中 / 英 / 日 多语言 Web 控制台
- 密码保护的 Web 控制台

## 架构

```text
GitHub/GitLab/Gitea
        |
        | webhook / push / tag / comment
        v
Haiwo Server  <---- JSON-RPC 2.0 over WebSocket ----  Haiwo Agent
        |                                                |
        | REST API + Web 控制台                           | shell/git
        v                                                v
PostgreSQL + GORM AutoMigrate                    目标机器 / 内网机器
```

Server 负责：

- 提供 Web 控制台
- 管理项目、流水线、触发器、Agent、运行历史、系统设置
- 接收 GitHub/GitLab/Gitea Webhook
- 匹配触发规则并创建运行
- 通过 JSON-RPC 下发任务到 Agent
- 接收 Agent 心跳、日志、进度和完成状态

Agent 负责：

- 主动连接 Server 的 `/rpc/agent/ws`
- 注册 Agent 身份、标签、容量
- 定期上报在线状态、CPU、内存、负载
- 执行流水线步骤中的 shell 命令
- 回传 stdout/stderr、错误和完成状态
- 提供反向 WebSocket SSH 能力

## 快速启动

本地启动 Server：

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

默认密码来自 `configs/config.toml`：

```toml
[WebConfiguration]
Password = "haiwo"
```

## PostgreSQL 和 AutoMigrate

配置文件：

```text
configs/config.toml
```

示例：

```toml
[PostgresConfiguration]
Enabled = true
Host = "127.0.0.1"
Port = 18432
User = "pacman"
Password = "MEB3oZZkpxgA1q4WcA10"
DBName = "pacman"
SSLMode = "disable"
TimeZone = "UTC"
LogMode = "none"
```

Server 启动时如果 `Enabled = true`，会自动连接 PostgreSQL 并执行 GORM `AutoMigrate`。不需要单独运行迁移命令。

如果重启后项目或流水线消失，优先检查：

- `PostgresConfiguration.Enabled` 是否为 `true`
- 是否连接到同一个 PostgreSQL
- `Host`、`Port`、`User`、`DBName` 是否正确
- 容器部署时配置文件是否正确挂载

## Web 控制台流程

1. 启动 Server。
2. 打开 `http://localhost:8080/`。
3. 输入 Web 密码登录。
4. 在 Settings 配置 Server 公开地址，例如 `https://cicd.kcgi.edu.kg`。
5. 在 Agent 页面添加 Agent。
6. 点击 Agent 行里的 `脚本`，复制部署命令。
7. 在目标机器执行部署命令，让 Agent 上线。
8. 在 Projects 页面创建项目。
9. 复制项目 Hook URL 到 GitHub/GitLab/Gitea。
10. 在 Pipelines 页面创建或编辑流水线。
11. 点击流水线列表中的 `Run` 手动运行，或等待 Git 事件触发。
12. 在 Runs 页面查看历史、详情、日志和错误。

## Agent 部署

在 Agent 页面创建 Agent 后，点击 `脚本` 会复制类似命令：

```sh
curl -fsSL 'https://cicd.kcgi.edu.kg/api/agents/agent_xxx/deploy.sh?token=hwagt_xxx' | bash
```

部署脚本会：

- 下载 Linux x86_64 Agent 二进制
- 停止旧 Agent 进程
- 替换本地 Agent 文件
- 写入 Agent ID、Token、标签、Server WebSocket 地址
- 启动 Agent

默认 Agent 二进制下载地址：

```text
https://fileoss.hacksnews.top/haiwo-agent-linux-amd64
```

构建 Linux x86_64 Agent：

```sh
mkdir -p bin
GOOS=linux GOARCH=amd64 go build -o bin/haiwo-agent-linux-amd64 ./cmd/agent
```

更新 Agent 时，把新的二进制上传到文件服务器，然后重新执行部署命令。

## GitHub 使用说明

### 创建项目

在 Projects 页面点击 `+ 添加`：

- 名称：例如 `kanfan`
- 代码平台：选择 `GitHub`
- 仓库地址：例如 `https://github.com/dollarkillerx/kanfan` 或 `git@github.com:dollarkillerx/kanfan.git`
- 默认分支：例如 `main`

创建后，在项目列表点击 Hook 复制按钮。GitHub 项目只需要复制 GitHub Hook URL。

### 配置 GitHub Webhook

打开 GitHub 仓库：

```text
Settings -> Webhooks -> Add webhook
```

填写：

- Payload URL：Haiwo 复制出来的 GitHub Hook URL
- Content type：选择 `application/json`
- Secret：MVP 可留空
- SSL verification：保持启用
- Active：勾选

事件选择：

| Haiwo 触发方式 | GitHub 事件 |
| --- | --- |
| Push 分支 | `Just the push event` |
| Push 分支 + Commit 信息包含 | `Just the push event` |
| Tag 名称 | tag push 会通过 push payload 发送，保留 push event |
| 评论命令 | `Let me select individual events`，选择 issue comment / pull request review comment |

GitHub 创建 webhook 后会发送 `ping` 事件。Haiwo 会返回类似：

```json
{
  "event": "ping",
  "ignored": true,
  "matched": 0,
  "runs": []
}
```

这是正常的连通性检查，不会触发流水线。

### Push 触发示例

Haiwo 流水线配置：

- 触发方式：`Push 分支`
- 分支：`main`
- Commit 信息包含：`deploy`

会触发：

```sh
git checkout main
git commit --allow-empty -m "deploy staging"
git push origin main
```

不会触发：

```sh
git commit --allow-empty -m "fix typo"
git push origin main
```

## 流水线

Pipelines 页面支持选项式创建和编辑流水线。

当前支持：

- 手动触发
- Push 指定分支
- Push 指定分支，并要求 commit message 包含某段文本
- Tag 名称触发
- 顺序添加多个 Agent 命令步骤

每个步骤选择一个 Agent 和 shell 命令。Agent 会把命令作为同一个 shell 脚本执行，所以 `cd` 和 `export` 会对后续命令生效。

示例：

```sh
cd /opt/workspace/build/kanfan
pwd
git pull
```

## 运行历史

Runs 页面只用于查看历史，不放手动执行表单。

支持：

- 按流水线筛选
- 最新运行排在最前
- 展示状态、来源、ref、总耗时、更新时间
- 点击详情查看错误和日志

如果失败但没有日志，检查：

- Agent 是否在线
- Agent 是否为最新版本
- 任务是否在下发到 Agent 前失败
- 流水线步骤是否选中了正确 Agent
- 命令是否能在 Agent 机器上执行

## WebSSH

Agent 默认支持反向 WebSocket SSH。Agent 可以部署在内网，Server 不需要直接 SSH 到 Agent。

Agent 页面点击 `SSH` 会打开：

```text
/ssh/agents/{agent_id}
```

如果无法进入终端，检查：

- Agent 是否在线
- Agent 二进制是否最新
- Settings 是否配置 Server 公开地址
- 代理层是否允许 WebSocket Upgrade

## Docker Compose

构建镜像：

```sh
make img-build
```

启动：

```sh
make up
```

启动可选本地 Agent：

```sh
make up-agent
```

停止：

```sh
make down
```

容器部署时需要确认 Server 容器能访问 PostgreSQL。容器内的 `127.0.0.1` 是容器自身，不是宿主机。

## 开发检查

运行测试：

```sh
go test ./...
```

检查前端脚本：

```sh
node --check internal/server/web/app.js
```

构建：

```sh
make build
```

## 更多文档

- [部署文档](DEPLOYMENT.md)
- [英文 README](README.md)
