# 使用多阶段构建
FROM golang:1.21-alpine AS builder

WORKDIR /build

# 复制后端代码
COPY backend/ .

# 构建后端（禁用 CGO，使用纯 Go 构建）
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 最终镜像
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

# 从构建阶段复制后端二进制文件
COPY --from=builder /build/main ./

# 复制 nginx 配置
COPY backend/config/nginx.conf /etc/nginx/nginx.conf

# 复制前端构建产物
COPY frontend/dist /app/frontend

# 复制启动脚本
COPY scripts/start.sh ./
RUN chmod +x start.sh

EXPOSE 80

# 启动服务
CMD ["./start.sh"] 