# rclone-server

Create a file server with `rclone serve <protocol> --auth-proxy`.

See [the documentation](https://rclone.org/commands/rclone_serve/) for more details.

## Usage

Create `./data/auth_data.json`:

```json
{
  "user1": {
    "auth": {
      "pass": "login_pass"
    },
    "config": {
      "type": "local",
      "_root": "/data/files"
    }
  },
  "user2": {
    "auth": {
      "pass": "login_pass",
      "public_keys": ["my_public_key"]
    },
    "config": {
      "type": "webdav",
      "_root": "/Documents/files",
      "url": "https://nextcloud.example.com/remote.php/dav/files/gerald/",
      "vendor": "nextcloud",
      "user": "gerald",
      "pass": "override-pass"
    }
  }
}
```

Note:

- `auth` can contain two fields: `pass`, which enables password login, and `public_keys`, which enables public key login. If absent, the login method is disabled.
- Once authenticated, `config` will be passed to `rclone` to create a backend. The current authentication info (`user`, `pass`, `public_key`) will also be passed to `rclone` if not provided in `config`.

Create `compose.yml`:

```yaml
services:
  rclone:
    image: ghcr.io/gera2ld/rclone-server
    restart: unless-stopped
    volumes:
      - ./data/auth_data.json:/auth_data.json:ro
      - ./path/to/my/files:/data/files
    environment:
      # - PORT_WEBDAV=80
      # - PORT_SFTP=22
    ports:
      - 8080:80
      - 8022:22
```

Start service with `docker compose up -d`.
