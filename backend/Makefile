# 容器编排管理平台 - 后端Makefile

.PHONY: help build run test clean lint fmt migrate-up migrate-down docker-build docker-run

# 默认目标
help:
	@echo "容器编排管理平台 - 可用命令:"
	@echo "  build      - 构建应用程序"
	@echo "  run        - 运行开发服务器"
	@echo "  test       - 运行测试"
	@echo "  clean      - 清理构建文件"
	@echo "  lint       - 运行代码质量检查"
	@echo "  fmt        - 格式化代码"
	@echo "  migrate-up - 运行数据库迁移"
	@echo "  migrate-down - 回滚数据库迁移"
	@echo "  docker-build - 构建Docker镜像"
	@echo "  docker-run  - 运行Docker容器"

# 变量
APP_NAME := container-platform-backend
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR := bin
MAIN_FILE := cmd/server/main.go

# 构建信息
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)"

# 构建
build:
	@echo "构建 $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)

# 运行开发服务器
run:
	@echo "启动开发服务器..."
	go run cmd/server/main.go

# 测试
test:
	@echo "运行测试..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 运行基准测试
benchmark:
	@echo "运行基准测试..."
	go test -bench=. -benchmem ./...

# 清理
clean:
	@echo "清理构建文件..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	go clean -cache -modcache -testcache

# 代码质量检查
lint:
	@echo "运行代码质量检查..."
	golangci-lint run

# 格式化代码
fmt:
	@echo "格式化代码..."
	go fmt ./...
	goimports -w .

# 依赖管理
deps:
	@echo "下载依赖..."
	go mod download
	go mod tidy

# 数据库迁移
migrate-up:
	@echo "运行数据库迁移..."
	go run cmd/migrate/main.go up

migrate-down:
	@echo "回滚数据库迁移..."
	go run cmd/migrate/main.go down

migrate-create:
	@echo "创建新的迁移文件..."
	@if [ -z "$(name)" ]; then echo "使用方法: make migrate-create name=migration_name"; exit 1; fi
	migrate create -ext sql -dir migrations -seq $(name)

# Docker相关
docker-build:
	@echo "构建Docker镜像..."
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

docker-run:
	@echo "运行Docker容器..."
	docker run -p 8080:8080 --name $(APP_NAME) $(APP_NAME):latest

# 开发环境设置
dev-setup:
	@echo "设置开发环境..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 生成Swagger文档
swagger:
	@echo "生成Swagger文档..."
	swag init -g cmd/server/main.go -o docs

# 安装预提交钩子
install-hooks:
	@echo "安装Git预提交钩子..."
	cp scripts/pre-commit .git/hooks/
	chmod +x .git/hooks/pre-commit

# 检查更新
deps-update:
	@echo "更新依赖..."
	go list -u -m all | grep '\['
	go get -u ./...
	go mod tidy

# 安全扫描
security:
	@echo "运行安全扫描..."
	gosec ./...

# 生成mock文件
mock:
	@echo "生成mock文件..."
	go generate ./...

# 完整的CI流程
ci: deps fmt lint security test build
	@echo "CI流程完成成功!"