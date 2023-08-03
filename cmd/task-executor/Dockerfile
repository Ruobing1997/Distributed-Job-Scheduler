# 使用官方的Golang镜像作为构建环境
FROM golang:1.20 AS builder

# 设置工作目录
WORKDIR /go/src/MoreFun_SuperNova

# 将go.mod和go.sum文件复制到工作目录
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 将源代码复制到工作目录
COPY . .

# 构建worker程序
RUN CGO_ENABLED=0 GOOS=linux go build -o executor cmd/task-executor/main.go

# 使用alpine作为基础镜像
FROM alpine:latest
RUN apk add --no-cache bash
RUN apk add --no-cache bash python3
RUN mkdir -p /logs/workers

# 从builder镜像复制编译好的程序到当前镜像
COPY --from=builder /go/src/MoreFun_SuperNova/executor /app/executor

# 设置启动命令
CMD ["/app/executor", "start"]