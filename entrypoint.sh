#/bin/sh

export RCLONE_DIR_CACHE_TIME=${RCLONE_DIR_CACHE_TIME:-30s}
export AUTH_CONFIG=${AUTH_CONFIG:-/etc/auth.json}

if [ -n "$PORT_WEBDAV" ]; then
  rclone serve webdav --auth-proxy /usr/bin/auth_proxy --dir-cache-time $RCLONE_DIR_CACHE_TIME --addr :$PORT_WEBDAV &
fi

if [ -n "$PORT_SFTP" ]; then
  rclone serve sftp --auth-proxy /usr/bin/auth_proxy --dir-cache-time $RCLONE_DIR_CACHE_TIME --addr :$PORT_SFTP &
fi

wait -n

exit $?
