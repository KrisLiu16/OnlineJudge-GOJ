FROM node:20-alpine AS builder

WORKDIR /app

# 复制 package.json 和 package-lock.json
COPY package*.json ./

# 安装依赖
RUN npm install

# 复制源代码和环境配置
COPY . .

# 确保环境变量正确设置
RUN echo "VITE_BASE_URL=/" > .env.production && \
    echo "VITE_API_URL=/api" >> .env.production

# 构建应用
RUN npm run build

# 生产环境
FROM nginx:alpine

# 复制构建产物到 nginx 目录
COPY --from=builder /app/dist /usr/share/nginx/html/

# 复制 favicon.ico 到正确位置
COPY --from=builder /app/public/favicon.ico /usr/share/nginx/html/

# 复制 nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 确保目录权限正确
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chmod -R 755 /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"] 