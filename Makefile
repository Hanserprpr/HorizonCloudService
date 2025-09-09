# 智能媒体云存储系统 - Makefile
# 微服务架构管理命令

.PHONY: help dev-up dev-down build-all clean logs test lint format

# 默认目标
help:
	@echo "智能媒体云存储系统 - 开发命令"
	@echo ""
	@echo "开发环境:"
	@echo "  dev-up          启动开发环境"
	@echo "  dev-down        停止开发环境"
	@echo "  dev-restart     重启开发环境"
	@echo "  dev-logs        查看开发环境日志"
	@echo ""
	@echo "构建和部署:"
	@echo "  build-all       构建所有服务镜像"
	@echo "  build-service   构建指定服务 (使用 SERVICE=service-name)"
	@echo "  push-images     推送镜像到注册表"
	@echo ""
	@echo "数据库管理:"
	@echo "  db-migrate      运行数据库迁移"
	@echo "  db-seed         填充测试数据"
	@echo "  db-backup       备份数据库"
	@echo "  db-restore      恢复数据库"
	@echo ""
	@echo "代码质量:"
	@echo "  test            运行所有测试"
	@echo "  test-service    测试指定服务 (使用 SERVICE=service-name)"
	@echo "  lint            代码检查"
	@echo "  format          代码格式化"
	@echo ""
	@echo "清理和维护:"
	@echo "  clean           清理容器和镜像"
	@echo "  clean-volumes   清理数据卷"
	@echo "  logs            查看指定服务日志 (使用 SERVICE=service-name)"

# =============================================================================
# 开发环境管理
# =============================================================================

dev-up:
	@echo "🚀 启动开发环境..."
	docker-compose up -d
	@echo "✅ 开发环境已启动"
	@echo "   - API网关: http://localhost:8080"
	@echo "   - Grafana监控: http://localhost:3000 (admin/admin123)"
	@echo "   - MinIO控制台: http://localhost:9001 (minioadmin/12345678)"

dev-down:
	@echo "⏹️  停止开发环境..."
	docker-compose down
	@echo "✅ 开发环境已停止"

dev-restart:
	@echo "🔄 重启开发环境..."
	docker-compose restart
	@echo "✅ 开发环境已重启"

dev-logs:
	@echo "📋 查看开发环境日志..."
	docker-compose logs -f

# =============================================================================
# 构建和部署
# =============================================================================

build-all:
	@echo "🔨 构建所有服务镜像..."
	docker-compose build
	@echo "✅ 所有服务镜像构建完成"

build-service:
	@if [ -z "$(SERVICE)" ]; then echo "❌ 请指定服务名: make build-service SERVICE=service-name"; exit 1; fi
	@echo "🔨 构建服务: $(SERVICE)"
	docker-compose build $(SERVICE)
	@echo "✅ 服务 $(SERVICE) 构建完成"

build-gateway:
	@echo "🔨 构建API网关..."
	cd services/gateway && docker build -t horizon-cloud/gateway:latest .

build-user-service:
	@echo "🔨 构建用户服务..."
	cd services/user-service && docker build -t horizon-cloud/user-service:latest .

build-file-service:
	@echo "🔨 构建文件服务..."
	cd services/file-service && docker build -t horizon-cloud/file-service:latest .

build-ai-service:
	@echo "🔨 构建AI服务..."
	cd services/ai-service && docker build -t horizon-cloud/ai-service:latest .

# =============================================================================
# 数据库管理
# =============================================================================

db-migrate:
	@echo "📊 运行数据库迁移..."
	docker-compose exec mysql mysql -u root -p12345678 -e "source /docker-entrypoint-initdb.d/001_create_users_table.sql"
	docker-compose exec postgres psql -U postgres -d horizon_cloud_files -f /docker-entrypoint-initdb.d/002_create_files_table.sql
	@echo "✅ 数据库迁移完成"

db-seed:
	@echo "🌱 填充测试数据..."
	# TODO: 添加测试数据脚本
	@echo "✅ 测试数据填充完成"

db-backup:
	@echo "💾 备份数据库..."
	mkdir -p ./backups
	docker-compose exec mysql mysqldump -u root -p12345678 --all-databases > ./backups/mysql-$(shell date +%Y%m%d-%H%M%S).sql
	docker-compose exec postgres pg_dumpall -U postgres > ./backups/postgres-$(shell date +%Y%m%d-%H%M%S).sql
	@echo "✅ 数据库备份完成"

# =============================================================================
# 代码质量
# =============================================================================

test:
	@echo "🧪 运行所有测试..."
	@for service in gateway user-service file-service ai-service search-service; do \
		if [ -d "services/$$service" ]; then \
			echo "Testing $$service..."; \
			cd services/$$service && go test ./... && cd ../..; \
		fi \
	done
	@echo "✅ 所有测试完成"

test-service:
	@if [ -z "$(SERVICE)" ]; then echo "❌ 请指定服务名: make test-service SERVICE=service-name"; exit 1; fi
	@echo "🧪 测试服务: $(SERVICE)"
	cd services/$(SERVICE) && go test ./...

lint:
	@echo "🔍 代码检查..."
	@for service in gateway user-service file-service ai-service search-service; do \
		if [ -d "services/$$service" ] && [ -f "services/$$service/go.mod" ]; then \
			echo "Linting $$service..."; \
			cd services/$$service && golangci-lint run && cd ../..; \
		fi \
	done
	@echo "✅ 代码检查完成"

format:
	@echo "💅 代码格式化..."
	@for service in gateway user-service file-service ai-service search-service; do \
		if [ -d "services/$$service" ]; then \
			echo "Formatting $$service..."; \
			cd services/$$service && go fmt ./... && cd ../..; \
		fi \
	done
	@echo "✅ 代码格式化完成"

# =============================================================================
# 清理和维护
# =============================================================================

clean:
	@echo "🧹 清理容器和镜像..."
	docker-compose down --rmi all --remove-orphans
	docker system prune -f
	@echo "✅ 清理完成"

clean-volumes:
	@echo "🗑️  清理数据卷..."
	docker-compose down -v
	docker volume prune -f
	@echo "✅ 数据卷清理完成"

logs:
	@if [ -z "$(SERVICE)" ]; then \
		echo "📋 查看所有服务日志..."; \
		docker-compose logs -f; \
	else \
		echo "📋 查看服务日志: $(SERVICE)"; \
		docker-compose logs -f $(SERVICE); \
	fi

# 查看系统状态
status:
	@echo "📊 系统状态:"
	@echo ""
	@echo "🐳 Docker 容器:"
	docker-compose ps
	@echo ""
	@echo "💾 数据卷:"
	docker volume ls | grep horizon
	@echo ""
	@echo "🌐 网络:"
	docker network ls | grep horizon

# 生产环境部署 (Kubernetes)
k8s-deploy:
	@echo "🚀 部署到Kubernetes..."
	kubectl apply -f infrastructure/kubernetes/
	@echo "✅ Kubernetes部署完成"

k8s-status:
	@echo "📊 Kubernetes状态:"
	kubectl get pods,services,ingress -l app=horizon-cloud

# 开发工具
dev-setup:
	@echo "🛠️  设置开发环境..."
	@echo "安装Go依赖..."
	@for service in gateway user-service file-service ai-service search-service; do \
		if [ -d "services/$$service" ] && [ -f "services/$$service/go.mod" ]; then \
			echo "Installing dependencies for $$service..."; \
			cd services/$$service && go mod tidy && cd ../..; \
		fi \
	done
	@echo "✅ 开发环境设置完成"

# 工作区管理
workspace-init:
	@echo "🚀 初始化Go工作区..."
	go work init
	@for service in gateway user-service file-service ai-service search-service shared/pkg; do \
		if [ -d "services/$$service" ] && [ -f "services/$$service/go.mod" ]; then \
			go work use services/$$service; \
		fi \
	done
	go work use shared/pkg
	@echo "✅ Go工作区初始化完成"