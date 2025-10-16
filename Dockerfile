# 使用 Alpine 作为基础镜像
FROM alpine:3.18

WORKDIR /app

# 安装必要的包
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    bash

# 创建必要的目录
RUN mkdir -p /app/data /app/frontend

# 复制构建产物
COPY backend/main-* ./
RUN if [ "$(uname -m)" = "aarch64" ]; then \
      cp main-arm64 main; \
    else \
      cp main-amd64 main; \
    fi && \
    rm main-* && \
    chmod +x main

# 复制前端静态文件
COPY frontend/dist /app/frontend

# 复制启动脚本
COPY scripts/start.sh ./
RUN chmod +x start.sh

# 暴露端口 (默认 8080)
EXPOSE 8080

# 启动服务
CMD ["./start.sh"] 