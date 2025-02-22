#!/bin/bash

# 启动后端服务
echo "启动后端服务..."
./main &

# 启动 nginx
echo "启动 Nginx..."
nginx -g 'daemon off;' 