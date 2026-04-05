.PHONY: build build-standalone build-minimal run run-standalone test clean build-all

# 变量
STANDALONE_NAME=claude-insight
GO_FILES=$(shell find . -name '*.go' -type f)
LDFLAGS=-ldflags="-s -w"  # 去掉调试信息和符号表

# 构建一体化版本
build-standalone:
	go build $(LDFLAGS) -o bin/$(STANDALONE_NAME) ./cmd/standalone

# 构建最小化版本（尝试 UPX 压缩）
build-minimal:
	go build $(LDFLAGS) -o bin/$(STANDALONE_NAME)-minimal ./cmd/standalone
	@if command -v upx >/dev/null 2>&1; then \
		echo "使用 UPX 压缩..."; \
		upx --best --lzma bin/$(STANDALONE_NAME)-minimal; \
	else \
		echo "UPX 未安装，跳过压缩（安装 UPX 可进一步减小体积）"; \
	fi
	@ls -lh bin/$(STANDALONE_NAME)-minimal

# 默认构建
build: build-standalone

# 运行一体化版本
run: build-standalone
	./bin/$(STANDALONE_NAME)

run-standalone: build-standalone
	./bin/$(STANDALONE_NAME)

# 测试
test:
	go test -v ./...

# 清理
clean:
	rm -rf bin/

# 多平台构建
build-all:
	@echo "构建 Linux amd64..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(STANDALONE_NAME)-linux-amd64 ./cmd/standalone
	@echo "构建 macOS amd64..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(STANDALONE_NAME)-darwin-amd64 ./cmd/standalone
	@echo "构建 macOS arm64..."
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(STANDALONE_NAME)-darwin-arm64 ./cmd/standalone
	@echo "构建 Windows amd64..."
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(STANDALONE_NAME)-windows-amd64.exe ./cmd/standalone
	@echo "所有平台构建完成！"
	@ls -lh bin/
