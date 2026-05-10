.DEFAULT_GOAL := all

APP_NAME = backend
BUILD_DIR = bin
MAIN_PATH = ./cmd/server

GO = go
GOBUILD = $(GO) build
GOCLEAN = $(GO) clean
GOMOD = $(GO) mod
GOFMT = gofmt

# 绿色文字输出（使用 PowerShell Write-Host，不改变全局 SHELL）
define info_msg
	@powershell.exe -NoProfile -Command "Write-Host '==> $(1)' -ForegroundColor Green"
endef

# 清理 bin 目录（兼容 CMD，忽略目录不存在的错误）
define CLEAN_BIN
	@if exist $(BUILD_DIR) rmdir /s /q $(BUILD_DIR) 2>nul
endef

.PHONY: tidy
tidy:
	$(call info_msg,Tidying go.mod)
	$(GOMOD) tidy

.PHONY: fmt
fmt:
	$(call info_msg,Formatting code)
	$(GOFMT) -s -w .

.PHONY: build
build: tidy
	$(call info_msg,Cleaning bin directory)
	$(CLEAN_BIN)
	$(call info_msg,Building $(APP_NAME) (debug))
	$(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME).exe $(MAIN_PATH)

.PHONY: check
check:
	$(call info_msg,Running security vulnerability scan)
	@govulncheck ./...

.PHONY: release
release: check
	$(call info_msg,Cleaning bin directory)
	$(CLEAN_BIN)
	$(call info_msg,Cleaning build cache)
	$(GOCLEAN) -cache -testcache
	$(call info_msg,Building $(APP_NAME) (release optimized))
	$(GOBUILD) -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME).exe $(MAIN_PATH)

.PHONY: run
run:
	$(call info_msg,Cleaning bin directory)
	$(CLEAN_BIN)
	$(call info_msg,Running application)
	$(GO) run $(MAIN_PATH)/main.go

.PHONY: dev
dev:
	$(call info_msg,Cleaning bin directory)
	$(CLEAN_BIN)
	$(call info_msg,Starting hot reload with air)
	air

.PHONY: clean
clean:
	$(call info_msg,Cleaning bin directory)
	$(CLEAN_BIN)
	$(call info_msg,Cleaning build cache)
	$(GOCLEAN) -cache -testcache
	$(call info_msg,Clean completed)

.PHONY: all
all: tidy fmt build
	$(call info_msg,All tasks completed)

.PHONY: help
help:
	@echo Available targets:
	@echo   help      : Show this help
	@echo   run       : Run the application
	@echo   dev       : Run with hot reload (air)
	@echo   check     : Run security vulnerability scan
	@echo   clean     : Remove bin directory and clean build cache
	@echo   all       : Tidy, fmt, and build
	@echo   tidy      : Tidy go.mod
	@echo   fmt       : Format code
	@echo   build     : Build debug binary
	@echo   release   : Run security check, then build release binary (optimized)