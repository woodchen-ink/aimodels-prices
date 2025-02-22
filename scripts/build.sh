#!/bin/bash

# 设置Go环境变量
export CGO_ENABLED=0
export GOOS=linux

# 编译 AMD64 版本
echo "Building AMD64 version..."
export GOARCH=amd64
go build -o backend/main-amd64 backend/main.go
go build -o backend/migrate-amd64 backend/cmd/migrate/main.go

# 编译 ARM64 版本
echo "Building ARM64 version..."
export GOARCH=arm64
go build -o backend/main-arm64 backend/main.go
go build -o backend/migrate-arm64 backend/cmd/migrate/main.go

echo "Build completed!" 