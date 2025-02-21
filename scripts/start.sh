#!/bin/bash

# 执行数据库迁移
echo "执行数据库迁移..."
./migrate

# 启动后端服务
./main &

# 启动 nginx
nginx -g 'daemon off;' 