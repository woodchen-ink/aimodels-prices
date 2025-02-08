# 使用 Alpine 作为基础镜像
FROM alpine:3.18

WORKDIR /app

# 安装必要的包
RUN apk add --no-cache \
    nginx \
    ca-certificates \
    tzdata \
    bash \
    wget

# 创建必要的目录
RUN mkdir -p /app/data /app/frontend

# 复制构建产物
COPY backend/main ./
COPY frontend/dist /app/frontend
COPY backend/config/nginx.conf /etc/nginx/nginx.conf
COPY scripts/start.sh ./

RUN chmod +x start.sh main

EXPOSE 80

# 启动服务
CMD ["./start.sh"] 