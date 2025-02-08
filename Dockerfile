# 使用 Alpine 作为基础镜像
FROM alpine:3.18

WORKDIR /app

# 安装 nginx
RUN apk add --no-cache nginx

# 创建必要的目录
RUN mkdir -p /app/data /app/frontend

# 复制后端二进制文件
COPY backend/main ./

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