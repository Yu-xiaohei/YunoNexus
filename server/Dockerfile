# Dockerfile (API 服务)
FROM golang:1.22-alpine AS builder

WORKDIR /app

# 安装依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# 运行阶段
FROM alpine:3.19

# 安装必要包
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1001 -S yunonexus && \
    adduser -u 1001 -S yunonexus -G yunonexus

# 复制编译结果
COPY --from=builder --chown=yunonexus:yunonexus /app/server /app/server
COPY --from=builder --chown=yunonexus:yunonexus /app/migrations /app/migrations

# 切换用户
USER yunonexus

# 暴露端口
EXPOSE 8080

# 启动命令
ENTRYPOINT ["/app/server"]
