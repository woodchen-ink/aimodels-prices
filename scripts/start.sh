#!/bin/bash

# 启动后端服务
./main &

# 启动 nginx
nginx -g 'daemon off;' 