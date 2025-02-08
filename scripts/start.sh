#!/bin/sh

# 启动后端服务
./main &

# 启动前端服务
cd /app/frontend && node server.js &

# 启动 nginx
nginx -g 'daemon off;' 