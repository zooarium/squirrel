# Deployment Guide

This document describes how to deploy the Squirrel API as a standalone service on a Linux server.

## 1. Build the Binary

The project includes a `Makefile` target to build a statically linked production binary inside a Docker container. This ensures the binary has all its dependencies bundled and can run on any compatible Linux system.

Run the following command on your build machine (requires Docker):

```bash
make build-prod
```

The resulting binary will be located at `bin/squirrel`.

## 2. Prepare the Server

### Create a Dedicated User

For security, it's recommended to run the service under a dedicated non-root user.

```bash
sudo useradd -r -s /bin/false squirrel
```

### Setup Directory Structure

Ensure the application directory exists and has the necessary subdirectories.

```bash
sudo mkdir -p /var/www/zoo/squirrel/bin
sudo mkdir -p /var/www/zoo/squirrel/config
sudo mkdir -p /var/www/zoo/squirrel/data
sudo mkdir -p /var/www/zoo/squirrel/log
```

### Deploy Files

Copy the binary and configuration files to the deployment directory.

```bash
sudo cp bin/squirrel /var/www/zoo/squirrel/bin/
sudo cp config/config.yaml /var/www/zoo/squirrel/config/
```

### Set Permissions

Change the ownership of the application directory to the `squirrel` user.

```bash
sudo chown -R squirrel:squirrel /var/www/zoo/squirrel
sudo chmod +x /var/www/zoo/squirrel/bin/squirrel
```

## 3. Configure the Service

### Systemd Unit File

Create a new systemd unit file at `/etc/systemd/system/squirrel.service`:

```ini
[Unit]
Description=Squirrel Expense Management API
After=network.target

[Service]
User=squirrel
Group=squirrel
WorkingDirectory=/var/www/zoo/squirrel

# Environment Variables
Environment="SQUIRREL_ENVIRONMENT=production"
Environment="SQUIRREL_SERVER_ADDR=:8081"
Environment="SQUIRREL_SERVER_HOST=localhost:8081"
Environment="SQUIRREL_SERVER_READ_TIMEOUT=5s"
Environment="SQUIRREL_SERVER_WRITE_TIMEOUT=10s"
Environment="SQUIRREL_SERVER_IDLE_TIMEOUT=120s"
Environment="SQUIRREL_DATABASE_PATH=/var/www/zoo/squirrel/data/squirrel.db"
Environment="SQUIRREL_LOG_DIR=/var/www/zoo/squirrel/log"
Environment="SQUIRREL_LOG_LEVEL=info"
Environment="SQUIRREL_AUTH_JWT_SECRET=change-me-to-a-secure-random-string"
Environment="SQUIRREL_AUTH_JWT_EXPIRY=24h"
Environment="SQUIRREL_CORS_ALLOWED_ORIGINS=*"

ExecStart=/var/www/zoo/squirrel/bin/squirrel
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### Environment Variables Reference

The application uses Viper for configuration, which maps nested keys to environment variables using underscores and a `SQUIRREL_` prefix (e.g., `SERVER.ADDR` becomes `SQUIRREL_SERVER_ADDR`).

| Variable | Description | Default |
|----------|-------------|---------|
| `SQUIRREL_ENVIRONMENT` | Deployment environment (production, development) | `production` |
| `SQUIRREL_SERVER_ADDR` | Port/Address to listen on | `:8081` |
| `SQUIRREL_SERVER_HOST` | Public host name for Swagger docs | `localhost:8081` |
| `SQUIRREL_SERVER_READ_TIMEOUT` | Max duration for reading the entire request | `5s` |
| `SQUIRREL_SERVER_WRITE_TIMEOUT` | Max duration before timing out writes of the response | `10s` |
| `SQUIRREL_SERVER_IDLE_TIMEOUT` | Max amount of time to wait for the next request | `120s` |
| `SQUIRREL_DATABASE_PATH` | Path to the SQLite database file | `data/squirrel.db` |
| `SQUIRREL_LOG_DIR` | Directory where logs will be stored | `log` |
| `SQUIRREL_LOG_LEVEL` | Logging level (debug, info, warn, error) | `info` |
| `SQUIRREL_AUTH_JWT_SECRET` | Secret key for signing JWT tokens | (Required) |
| `SQUIRREL_AUTH_JWT_EXPIRY` | Duration until JWT tokens expire | `24h` |
| `SQUIRREL_CORS_ALLOWED_ORIGINS` | Allowed origins for CORS (comma-separated) | `*` |

## 4. Start and Enable the Service

Reload systemd to recognize the new service, then start and enable it to run at boot.

```bash
sudo systemctl daemon-reload
sudo systemctl start squirrel
sudo systemctl enable squirrel
```

## 5. Verification

### Check Service Status

```bash
sudo systemctl status squirrel
```

### Check Logs

```bash
tail -f /var/www/zoo/squirrel/log/api.log
# OR
journalctl -u squirrel -f
```

### Test the API

```bash
curl http://localhost:8081/health
```

## 6. Updating the Application

To update the application to a new version:

1. Build the new binary: `make build-prod`.
2. Stop the service: `sudo systemctl stop squirrel`.
3. Replace the binary: `sudo cp bin/squirrel /var/www/zoo/squirrel/bin/`.
4. Start the service: `sudo systemctl start squirrel`.

The application will automatically handle database migrations on startup.
