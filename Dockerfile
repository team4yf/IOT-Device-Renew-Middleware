# Compile stage
FROM golang:1.14.4-alpine AS builder

# The latest alpine images don't have some tools like (`git` and `bash`).
# Adding git, bash and openssh to the image
# RUN apk add --no-cache git=2.24.3-r0 \
#     --repository http://mirrors.aliyun.com/alpine/v3.11/community \
#     --repository http://mirrors.aliyun.com/alpine/v3.11/main

# 为我们的镜像设置必要的环境变量
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY="https://goproxy.io,direct"

# 移动到工作目录：/app
WORKDIR /app

# 复制项目中的 go.mod 和 go.sum文件并下载依赖信息
COPY go.mod .
COPY go.sum .
RUN go mod download

# 将代码复制到容器中
COPY . .

# Build the Go app
RUN go build -o main .


FROM alpine:latest

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
  echo http://dl-cdn.alpinelinux.org/alpine/edge/testing >> /etc/apk/repositories && \
  apk --no-cache add ca-certificates

WORKDIR /

COPY --from=builder /app/main /
# Copy the Pre-built binary file from the previous stage

# Command to run the executable
ENTRYPOINT ["/main"]
CMD []