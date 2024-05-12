#/bin/sh

if [ -n "$PORT_WEBDAV" ]; then
  rclone serve webdav --auth-proxy /usr/bin/auth_proxy --dir-cache-time 30s --addr :$PORT_WEBDAV &
fi

if [ -n "$PORT_SFTP" ]; then
  rclone serve sftp --auth-proxy /usr/bin/auth_proxy --dir-cache-time 30s --addr :$PORT_SFTP &
fi

wait -n

exit $?
