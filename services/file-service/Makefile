# File Service Makefile
# 提供便捷的开发和测试命令

.PHONY: help build clean test test-unit test-integration test-benchmark test-coverage lint fmt vet run docker-build docker-run deps-check security-scan

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
GO_VERSION := 1.21
APP_NAME := file-service
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go 相关变量
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := gofmt
GOVET := $(GOCMD) vet

# 构建标志
LDFLAGS := -X main.version=$(VERSION) \
           -X main.buildTime=$(BUILD_TIME) \
           -X main.gitCommit=$(GIT_COMMIT)

# 目录
BINARY_DIR := ./bin
OUTPUT_DIR := ./test-results
COVERAGE_DIR := $(OUTPUT_DIR)/coverage

# 颜色定义
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

## help: 显示帮助信息
help:
	@echo "$(BLUE)文件服务 Makefile$(NC)"
	@echo ""
	@echo "$(YELLOW)可用命令:$(NC)"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
	@echo ""
	@echo "$(YELLOW)示例:$(NC)"
	@echo "  make build          # 构建应用"
	@echo "  make test           # 运行所有测试"
	@echo "  make test-coverage  # 生成覆盖率报告"
	@echo "  make docker-build   # 构建 Docker 镜像"
	@echo ""

## build: 构建应用程序
build: clean
	@echo "$(BLUE)构建 $(APP_NAME)...$(NC)"
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)
	@echo "$(GREEN)构建完成: $(BINARY_DIR)/$(APP_NAME)$(NC)"

## build-linux: 为 Linux 构建应用程序
build-linux: clean
	@echo "$(BLUE)为 Linux 构建 $(APP_NAME)...$(NC)"
	@mkdir -p $(BINARY_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/$(APP_NAME)-linux ./cmd/$(APP_NAME)
	@echo "$(GREEN)Linux 构建完成: $(BINARY_DIR)/$(APP_NAME)-linux$(NC)"

## clean: 清理构建文件
clean:
	@echo "$(YELLOW)清理构建文件...$(NC)"
	$(GOCLEAN)
	@rm -rf $(BINARY_DIR)
	@rm -rf $(OUTPUT_DIR)
	@echo "$(GREEN)清理完成$(NC)"

## deps: 下载和整理依赖
deps:
	@echo "$(BLUE)下载依赖...$(NC)"
	$(GOGET) -d -v ./...
	$(GOMOD) tidy
	$(GOMOD) verify
	@echo "$(GREEN)依赖更新完成$(NC)"

## deps-check: 检查依赖的漏洞
deps-check:
	@echo "$(BLUE)检查依赖漏洞...$(NC)"
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "$(YELLOW)govulncheck 未安装，跳过漏洞检查$(NC)"; \
		echo "$(YELLOW)安装命令: go install golang.org/x/vuln/cmd/govulncheck@latest$(NC)"; \
	fi

## fmt: 格式化代码
fmt:
	@echo "$(BLUE)格式化代码...$(NC)"
	$(GOFMT) -s -w .
	@echo "$(GREEN)代码格式化完成$(NC)"

## vet: 运行 go vet
vet:
	@echo "$(BLUE)运行 go vet...$(NC)"
	$(GOVET) ./...
	@echo "$(GREEN)go vet 检查完成$(NC)"

## lint: 运行代码检查工具
lint: fmt vet
	@echo "$(BLUE)运行 linting...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --config .golangci.yml; \
	else \
		echo "$(YELLOW)golangci-lint 未安装$(NC)"; \
		echo "$(YELLOW)安装文档: https://golangci-lint.run/usage/install/$(NC)"; \
	fi

## test: 运行所有测试
test:
	@echo "$(BLUE)运行所有测试...$(NC)"
	./scripts/run-tests.sh -v
	@echo "$(GREEN)测试完成$(NC)"

## test-unit: 运行单元测试
test-unit:
	@echo "$(BLUE)运行单元测试...$(NC)"
	./scripts/run-tests.sh --unit-only -v
	@echo "$(GREEN)单元测试完成$(NC)"

## test-integration: 运行集成测试
test-integration:
	@echo "$(BLUE)运行集成测试...$(NC)"
	./scripts/run-tests.sh -i -v
	@echo "$(GREEN)集成测试完成$(NC)"

## test-benchmark: 运行基准测试
test-benchmark:
	@echo "$(BLUE)运行基准测试...$(NC)"
	./scripts/run-tests.sh -b -v
	@echo "$(GREEN)基准测试完成$(NC)"

## test-stress: 运行压力测试
test-stress:
	@echo "$(BLUE)运行压力测试...$(NC)"
	./scripts/run-tests.sh -s -v
	@echo "$(GREEN)压力测试完成$(NC)"

## test-coverage: 生成测试覆盖率报告
test-coverage:
	@echo "$(BLUE)生成覆盖率报告...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	./scripts/run-tests.sh -c -v
	@echo "$(GREEN)覆盖率报告生成完成$(NC)"
	@echo "$(YELLOW)HTML 报告: $(OUTPUT_DIR)/coverage.html$(NC)"

## test-all: 运行完整测试套件
test-all: lint test-unit test-integration test-benchmark test-coverage
	@echo "$(GREEN)完整测试套件运行完成$(NC)"

## run: 运行应用程序（开发模式）
run: build
	@echo "$(BLUE)启动 $(APP_NAME)...$(NC)"
	./$(BINARY_DIR)/$(APP_NAME)

## dev: 开发模式运行（带文件监控）
dev:
	@echo "$(BLUE)开发模式启动...$(NC)"
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "$(YELLOW)air 未安装，使用普通模式运行$(NC)"; \
		echo "$(YELLOW)安装 air: go install github.com/cosmtrek/air@latest$(NC)"; \
		make run; \
	fi

## docker-build: 构建 Docker 镜像
docker-build:
	@echo "$(BLUE)构建 Docker 镜像...$(NC)"
	docker build -t $(APP_NAME):$(VERSION) .
	docker build -t $(APP_NAME):latest .
	@echo "$(GREEN)Docker 镜像构建完成$(NC)"

## docker-run: 运行 Docker 容器
docker-run: docker-build
	@echo "$(BLUE)运行 Docker 容器...$(NC)"
	docker run --rm -p 8080:8080 $(APP_NAME):latest

## docker-compose-up: 使用 docker-compose 启动服务
docker-compose-up:
	@echo "$(BLUE)使用 docker-compose 启动服务...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)服务启动完成$(NC)"

## docker-compose-down: 停止 docker-compose 服务
docker-compose-down:
	@echo "$(YELLOW)停止 docker-compose 服务...$(NC)"
	docker-compose down
	@echo "$(GREEN)服务已停止$(NC)"

## security-scan: 运行安全扫描
security-scan:
	@echo "$(BLUE)运行安全扫描...$(NC)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "$(YELLOW)gosec 未安装$(NC)"; \
		echo "$(YELLOW)安装命令: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest$(NC)"; \
	fi
	make deps-check

## mock-gen: 生成 mock 文件
mock-gen:
	@echo "$(BLUE)生成 mock 文件...$(NC)"
	@if command -v mockgen >/dev/null 2>&1; then \
		find . -name "*.go" -exec grep -l "//go:generate.*mockgen" {} \; | xargs -I {} go generate {}; \
	else \
		echo "$(YELLOW)mockgen 未安装$(NC)"; \
		echo "$(YELLOW)安装命令: go install github.com/golang/mock/mockgen@latest$(NC)"; \
	fi

## proto-gen: 生成 protobuf 文件（如果有）
proto-gen:
	@echo "$(BLUE)生成 protobuf 文件...$(NC)"
	@if find . -name "*.proto" | grep -q .; then \
		if command -v protoc >/dev/null 2>&1 && command -v protoc-gen-go >/dev/null 2>&1; then \
			find . -name "*.proto" -exec protoc --go_out=. --go_opt=paths=source_relative {} \;; \
		else \
			echo "$(YELLOW)protobuf 编译器未安装$(NC)"; \
		fi; \
	else \
		echo "$(YELLOW)未找到 .proto 文件$(NC)"; \
	fi

## install-tools: 安装开发工具
install-tools:
	@echo "$(BLUE)安装开发工具...$(NC)"
	@echo "安装 golangci-lint..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.54.2
	@echo "安装其他工具..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golang/mock/mockgen@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "$(GREEN)开发工具安装完成$(NC)"

## check-tools: 检查工具安装状态
check-tools:
	@echo "$(BLUE)检查工具安装状态...$(NC)"
	@printf "%-20s " "Go:"; if command -v go >/dev/null 2>&1; then echo "$(GREEN)✓$(NC) $(shell go version)"; else echo "$(RED)✗$(NC)"; fi
	@printf "%-20s " "golangci-lint:"; if command -v golangci-lint >/dev/null 2>&1; then echo "$(GREEN)✓$(NC) $(shell golangci-lint version)"; else echo "$(RED)✗$(NC)"; fi
	@printf "%-20s " "air:"; if command -v air >/dev/null 2>&1; then echo "$(GREEN)✓$(NC) $(shell air -v)"; else echo "$(RED)✗$(NC)"; fi
	@printf "%-20s " "gosec:"; if command -v gosec >/dev/null 2>&1; then echo "$(GREEN)✓$(NC) $(shell gosec -version)"; else echo "$(RED)✗$(NC)"; fi
	@printf "%-20s " "govulncheck:"; if command -v govulncheck >/dev/null 2>&1; then echo "$(GREEN)✓$(NC)"; else echo "$(RED)✗$(NC)"; fi
	@printf "%-20s " "mockgen:"; if command -v mockgen >/dev/null 2>&1; then echo "$(GREEN)✓$(NC)"; else echo "$(RED)✗$(NC)"; fi
	@printf "%-20s " "docker:"; if command -v docker >/dev/null 2>&1; then echo "$(GREEN)✓$(NC) $(shell docker --version)"; else echo "$(RED)✗$(NC)"; fi

## version: 显示版本信息
version:
	@echo "$(BLUE)版本信息:$(NC)"
	@echo "应用名称: $(APP_NAME)"
	@echo "版本: $(VERSION)"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Git提交: $(GIT_COMMIT)"
	@echo "Go版本: $(shell go version)"

## info: 显示项目信息
info:
	@echo "$(BLUE)项目信息:$(NC)"
	@echo "应用名称: $(APP_NAME)"
	@echo "版本: $(VERSION)"
	@echo "Go版本要求: $(GO_VERSION)+"
	@echo "构建时间: $(BUILD_TIME)"
	@echo "Git提交: $(GIT_COMMIT)"
	@echo ""
	@echo "$(YELLOW)目录结构:$(NC)"
	@echo "  $(BINARY_DIR)/ - 构建输出"
	@echo "  $(OUTPUT_DIR)/ - 测试输出"
	@echo "  cmd/ - 应用入口"
	@echo "  internal/ - 私有代码"
	@echo "  test/ - 测试文件"
	@echo "  scripts/ - 构建脚本"

## ci: CI 流水线命令
ci: deps lint test-all security-scan build
	@echo "$(GREEN)CI 流水线完成$(NC)"

## release: 准备发布
release: clean ci build-linux
	@echo "$(BLUE)准备发布版本 $(VERSION)...$(NC)"
	@mkdir -p release
	@cp $(BINARY_DIR)/$(APP_NAME)-linux release/
	@cp README.md release/ 2>/dev/null || true
	@cp CHANGELOG.md release/ 2>/dev/null || true
	@tar -czf release/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz -C release $(APP_NAME)-linux README.md CHANGELOG.md 2>/dev/null || tar -czf release/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz -C release $(APP_NAME)-linux
	@echo "$(GREEN)发布包已准备完成: release/$(APP_NAME)-$(VERSION)-linux-amd64.tar.gz$(NC)"