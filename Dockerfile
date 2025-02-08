# 第一阶段：构建后端
FROM golang:1.21-alpine AS backend-builder

WORKDIR /app/backend

# 安装依赖
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# 复制后端源代码
COPY backend/ .

# 编译后端
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 第二阶段：构建前端
FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend

# 安装依赖
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

# 复制前端源代码
COPY frontend/ .

# 构建前端
RUN npm run build

# 第三阶段：最终镜像
FROM alpine:3.18

WORKDIR /app

# 安装 nginx
RUN apk add --no-cache nginx

# 创建数据目录
RUN mkdir -p /app/data

# 复制后端二进制文件
COPY --from=backend-builder /app/backend/main ./
COPY backend/config/nginx.conf /etc/nginx/nginx.conf

# 复制前端构建产物
COPY --from=frontend-builder /app/frontend/.next/static /app/frontend/static
COPY --from=frontend-builder /app/frontend/public /app/frontend/public
COPY --from=frontend-builder /app/frontend/.next/standalone /app/frontend

# 复制启动脚本
COPY scripts/start.sh ./
RUN chmod +x start.sh

EXPOSE 80

# 启动服务
CMD ["./start.sh"] 