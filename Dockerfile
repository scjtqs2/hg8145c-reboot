FROM golang:1.24-alpine AS builder

RUN  sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk add --no-cache git \
  && go env -w GO111MODULE=auto \
  && go env -w CGO_ENABLED=0 \
  && go env -w GOPROXY=https://goproxy.cn,direct

WORKDIR /build

COPY ./ .

RUN set -ex \
    && BUILD=`date +%FT%T%z` \
    && COMMIT_SHA1=`git rev-parse HEAD` \
    && go build -ldflags "-s -w -extldflags '-static' -X main.Version=${COMMIT_SHA1}|${BUILD}" -o job_app

FROM alpine AS production

RUN  sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk add --no-cache tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /data

COPY --from=builder /build/job_app /data/job_app

RUN chmod +x /data/job_app

ENV SELENIUM_ADDR="http://chrome:4444"
ENV LOGIN_URL="http://192.168.1.1"
ENV LOGIN_USERNAME="useradmin"
ENV LOGIN_PASSWORD="12345"
ENV CRONTAB="30 0 * * *"

ENTRYPOINT [ "/data/job_app" ]