FROM golang AS builder
WORKDIR /app
COPY auth_proxy.go /app
RUN go build auth_proxy.go

FROM rclone/rclone
ENV PORT_WEBDAV=80 PORT_SFTP=22
COPY --from=builder /app/auth_proxy /usr/bin/auth_proxy
COPY entrypoint.sh /entrypoint.sh
ENTRYPOINT ["sh", "-c", "/entrypoint.sh"]
