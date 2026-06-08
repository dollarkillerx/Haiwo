.PHONY: help fmt tidy vet test frontend build build-linux build-mac build-all dev dev-agent \
	img-build up up-agent down logs ps clean

IMG_REGISTRY ?= haiwo
IMG_TAG ?= latest
COMPOSE ?= docker compose

# 前端静态资源：手写文件，无需编译，由 internal/server/static.go 的 //go:embed 嵌入二进制。
WEB_DIR ?= internal/server/web
WEB_ASSETS := index.html app.js styles.css logo.png

# 交叉编译参数：纯 Go（无 cgo）静态二进制，去符号表减小体积。
LDFLAGS ?= -s -w
GOBUILD_RELEASE := CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)"

## ——— 代码质量 ———

# 格式化代码
fmt:
	go fmt ./...

# 整理依赖
tidy:
	go mod tidy

# 静态检查
vet:
	go vet ./...

# 运行单元测试
test:
	go test ./...

## ——— 构建 ———

# 编译前端：当前前端是手写静态文件，无需打包，这里校验将被 //go:embed 嵌入的资源是否齐全。
# 若以后接入前端工程（npm/vite 等），把构建命令放在这里并输出到 $(WEB_DIR)。
frontend:
	@echo "Checking frontend assets in $(WEB_DIR)..."
	@for f in $(WEB_ASSETS); do \
		test -f "$(WEB_DIR)/$$f" || { echo "  MISSING: $(WEB_DIR)/$$f"; exit 1; }; \
	done
	@echo "frontend assets OK"

# 编译当前平台的 server 和 agent 二进制（前端已嵌入）
build: frontend
	mkdir -p bin
	go build -o bin/haiwo-server ./cmd/server
	go build -o bin/haiwo-agent ./cmd/agent

# 交叉编译 Linux（amd64），前端已嵌入，静态二进制
build-linux: frontend
	mkdir -p bin
	GOOS=linux GOARCH=amd64 $(GOBUILD_RELEASE) -o bin/haiwo-server-linux-amd64 ./cmd/server
	GOOS=linux GOARCH=amd64 $(GOBUILD_RELEASE) -o bin/haiwo-agent-linux-amd64 ./cmd/agent

# 交叉编译 macOS（arm64），前端已嵌入
build-mac: frontend
	mkdir -p bin
	GOOS=darwin GOARCH=arm64 $(GOBUILD_RELEASE) -o bin/haiwo-server-darwin-arm64 ./cmd/server
	GOOS=darwin GOARCH=arm64 $(GOBUILD_RELEASE) -o bin/haiwo-agent-darwin-arm64 ./cmd/agent

# 交叉编译所有平台
build-all: build-linux build-mac

## ——— 本地运行 ———

# 本地开发运行 server
dev:
	go run ./cmd/server -c config -cPath ./,./configs/

# 本地开发运行 agent
dev-agent:
	HAIWO_SERVER_URL=ws://localhost:8080/rpc/agent/ws \
	HAIWO_AGENT_TOKEN=dev-agent-token \
	HAIWO_AGENT_ID=local-agent \
	HAIWO_AGENT_NAME=local-agent \
	HAIWO_AGENT_LABELS=build,deploy,staging \
	HAIWO_AGENT_SSH_ENABLED=true \
	go run ./cmd/agent

## ——— Docker 镜像 ———

# 构建 Docker 镜像
img-build:
	@echo "Building Haiwo image..."
	docker build -t $(IMG_REGISTRY)/haiwo:$(IMG_TAG) -f Dockerfile .

## ——— Docker Compose ———

# 启动 Postgres 和 server
up:
	$(COMPOSE) up -d postgres server

# 启动 Postgres、server 和可选 agent
up-agent:
	$(COMPOSE) --profile agent up -d

# 停止所有服务
down:
	$(COMPOSE) down

# 查看日志
logs:
	$(COMPOSE) logs -f

# 查看服务状态
ps:
	$(COMPOSE) ps

## ——— 清理 ———

# 删除本地构建产物
clean:
	rm -rf bin

## ——— 帮助 ———

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "代码质量:"
	@echo "  fmt             格式化代码"
	@echo "  tidy            整理依赖"
	@echo "  vet             静态检查"
	@echo "  test            运行单元测试"
	@echo ""
	@echo "构建:"
	@echo "  frontend        校验前端静态资源（将被嵌入二进制）"
	@echo "  build           编译当前平台 server 和 agent（含前端）"
	@echo "  build-linux     交叉编译 Linux amd64（含前端）"
	@echo "  build-mac       交叉编译 macOS arm64（含前端）"
	@echo "  build-all       交叉编译所有平台"
	@echo ""
	@echo "本地运行:"
	@echo "  dev             本地开发运行 server"
	@echo "  dev-agent       本地开发运行 agent"
	@echo ""
	@echo "Docker 镜像:"
	@echo "  img-build       构建 Docker 镜像"
	@echo ""
	@echo "Docker Compose:"
	@echo "  up              启动 Postgres 和 server"
	@echo "  up-agent        启动 Postgres、server 和 agent"
	@echo "  down            停止所有服务"
	@echo "  logs            查看日志"
	@echo "  ps              查看服务状态"
	@echo ""
	@echo "清理:"
	@echo "  clean           删除本地构建产物"
