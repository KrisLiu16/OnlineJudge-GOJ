FROM golang:1.22-alpine

WORKDIR /app

# 安装必要的系统依赖
RUN apk add --no-cache gcc musl-dev

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN go build -o main cmd/main.go

# 暴露端口
EXPOSE 3000

# 运行应用
CMD ["./main"] 