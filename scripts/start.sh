#!/bin/bash

# 执行数据库迁移（如果存在 migrate 文件）
if [ -f "./migrate" ]; then
    echo "执行数据库迁移..."
    ./migrate
fi

# 启动 Go 服务（包含 API 和静态文件服务）
echo "启动服务..."
exec ./main 