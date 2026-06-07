.PHONY: help fmt tidy vet test build dev dev-agent \
	img-build up up-agent down logs ps clean

IMG_REGISTRY ?= haiwo
IMG_TAG ?= latest
COMPOSE ?= docker compose

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

# 编译 server 和 agent 二进制
build:
	mkdir -p bin
	go build -o bin/haiwo-server ./cmd/server
	go build -o bin/haiwo-agent ./cmd/agent

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
	@echo "  build           编译 server 和 agent"
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
