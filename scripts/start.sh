#!/bin/bash

# 执行数据库迁移
echo "执行数据库迁移..."
./migrate

# 启动后端服务
echo "启动后端服务..."
./main &

# 启动 nginx
echo "启动 Nginx..."
nginx -g 'daemon off;' 