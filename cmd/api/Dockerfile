# 使用官方Golang镜像作为构建环境
FROM golang:1.20 AS builder

# 设置工作目录
WORKDIR /go/src/MoreFun_SuperNova

# 复制Go模块和构建缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=linux go build -o api-server cmd/api/main.go

# 使用alpine作为基础镜像
FROM alpine:latest

RUN mkdir /app

# 从构建环境复制构建的应用程序到当前容器
COPY --from=builder /go/src/MoreFun_SuperNova/api-server /app/
COPY --from=builder /go/src/MoreFun_SuperNova/frontend /app/frontend

# Copy frontend folder
COPY ./frontend ./frontend

# 设置容器启动应用程序时的命令
ENTRYPOINT ["/app/api-server"]