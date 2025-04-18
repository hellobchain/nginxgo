FROM golang:1.19 as build

ENV GOPROXY=https://goproxy.cn,direct
# 移动到工作目录：/dm-rwa-api
WORKDIR /nginxgo
# 将代码复制到容器中
COPY . .
# 编译成二进制可执行文件app
RUN rm -rf go.sum && make
# 移动到用于存放生成的二进制文件的 /build 目录
WORKDIR /build
RUN cp -r /nginxgo/bin .
RUN cp -r /nginxgo/configs .
# FROM ubuntu:22.04
FROM alpine:3.14
# 切换软件源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories \
    && apk update \
    && apk add tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone
WORKDIR /nginxgo
# 将二进制文件从 /build 目录复制到这里
COPY --from=build /build/configs /nginxgo/configs
COPY --from=build /build/bin/nginxgo.bin /usr/local/bin

# 启动容器时运行的命令
WORKDIR /nginxgo
CMD nginxgo.bin





