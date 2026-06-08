# Haiwo

![Haiwo logo](internal/server/web/logo.png)

[English README](README.md) | [部署文档](DEPLOYMENT.md)

Haiwo 是一个用 Go 编写的轻量、自托管 CI/CD 平台。一个带内嵌 Web 控制台的中心 Server，驱动主动连接的 Agent，在远程机器（包括 NAT 后 / 内网机器）上执行你的构建和部署命令。

## 功能

- **Server + Agent**：Agent 主动外连 Server（WebSocket），适合内网 / NAT 后的机器。
- **多种触发**：手动、Cron 定时、GitHub / GitLab / Gitea Webhook（Push 分支、commit 文本匹配、Tag、评论命令）。
- **可视化流水线**：界面创建和编辑，不用手写 JSON；支持多 Agent 顺序步骤。
- **运行历史**：每次运行的状态、来源、耗时、实时日志和错误。
- **Agent 面板**：在线状态、负载、CPU、内存。
- **Web SSH**：反向 WebSocket 终端进入 Agent，无需对 Agent 开放 SSH 入站。
- **内嵌控制台**：单个自包含二进制（前端已打包）；中 / 英 / 日；密码保护登录。
- **存储**：默认 file 后端（无需数据库），也可用 PostgreSQL。

## 部署

### 1. 二进制部署（Linux）

一键交互式安装脚本：先选语言，首次安装会让你设置**绑定端口**和**控制台密码**；之后再次运行可选择 **重启 / 停止 / 更新 / 修改密码**。

```sh
curl -fsSL https://raw.githubusercontent.com/dollarkillerx/Haiwo/refs/heads/main/scripts/install-server.sh | bash
```

访问控制台：`http://<服务器IP>:<端口>/`（默认端口 `8080`）。数据保存在 `~/.haiwo/server/data`（file 后端，无需数据库）。

### 2. Docker 部署

```sh
docker compose up -d
```

访问控制台：`http://<服务器IP>:8181/`。数据持久化到 `./data`（file 后端）。若要改用 PostgreSQL，编辑 `configs/config.toml`，详见[部署文档](DEPLOYMENT.md)。

---

Server 启动后，登录控制台，在 **Agent** 页面添加 Agent，把它生成的一行部署命令在目标机器上执行即可让 Agent 上线。

## 文档

- [部署文档](DEPLOYMENT.md)：完整配置、存储后端、Webhook、Web SSH、构建与交叉编译。
- [English README](README.md)

## 许可证

见 [LICENSE](LICENSE)。
