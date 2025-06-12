# 16. Deploying Goe Applications 🚀

Deploying a Goe application involves compiling it into a binary, managing its configuration, and running it reliably in your chosen environment. Here are common strategies and considerations.

## 1. Building Optimized Binaries

Go applications compile into statically linked binaries by default (if not using Cgo). For deployment, you typically want to build an optimized, production-ready binary.

```bash
# Build for the current OS/architecture
go build -o myapp_binary ./cmd/myapp/main.go

# Build for Linux AMD64 (common for servers)
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o myapp_linux_amd64 ./cmd/myapp/main.go
```

*   **`GOOS=linux GOARCH=amd64`**: Cross-compiles for a specific target operating system and architecture.
*   **`-ldflags="-s -w"`**:
    *   `-s`: Omits the symbol table (makes the binary smaller).
    *   `-w`: Omits the DWARF debugging information (makes the binary smaller).
    These flags are generally recommended for production builds to reduce binary size, but they make debugging with tools like Delve harder.
*   **`-o myapp_linux_amd64`**: Specifies the output binary name.
*   **`./cmd/myapp/main.go`**: Path to your application's main package.

## 2. Configuration Management in Production

*   **Environment Variables**: This is the **recommended** way to configure your Goe application in production.
    *   Set environment variables directly on your host machine, through your container orchestrator (Kubernetes, Docker Swarm), PaaS configuration, or CI/CD pipeline.
    *   Goe automatically loads these, and they take precedence over `.env` files.
    *   **Never commit `.env` files with production secrets to your repository.**
*   **Configuration Files (e.g., `.prod.env`)**:
    *   If you choose to use environment-specific `.env` files (e.g., `.prod.env` by setting `GOE_ENV=prod`), ensure this file is securely managed and deployed to your server. It should contain only non-sensitive overrides or be protected if it includes secrets.
    *   Using environment variables directly is generally safer for secrets than managing secret files on disk.
*   **Secrets Management Tools**: For sensitive data (API keys, database passwords), integrate with secrets management tools like HashiCorp Vault, AWS Secrets Manager, Google Cloud Secret Manager, etc. Your application or startup script would fetch secrets from these tools and expose them as environment variables to the Goe application.

## 3. Running the Application

Once you have your binary and configuration sorted, you need to run the application.

### a) Directly on a Server (Process Managers)

If deploying to a traditional VM or bare-metal server, use a process manager to ensure your application runs reliably, restarts on failure, and starts on boot.

**Systemd (Linux):**

Create a systemd service file (e.g., `/etc/systemd/system/myapp.service`):

```ini
[Unit]
Description=My Goe Application Service
After=network.target # Ensure network is up

[Service]
Type=simple
User=myappuser       # Run as a non-root user
Group=myappgroup
WorkingDirectory=/opt/myapp # Where your binary and assets are
ExecStart=/opt/myapp/myapp_linux_amd64 # Path to your binary

# Environment variables (can also be in a file specified by EnvironmentFile)
Environment="GOE_ENV=production"
Environment="APP_PORT=8080"
Environment="DB_HOST=prod.db.example.com"
# EnvironmentFile=/etc/myapp/myapp.conf # Alternative for many vars

Restart=on-failure   # Restart if it fails
RestartSec=5s        # Delay before restarting

StandardOutput=journal # Log to systemd journal
StandardError=journal
SyslogIdentifier=myapp

# Graceful shutdown (Goe handles SIGINT/SIGTERM)
KillSignal=SIGINT
TimeoutStopSec=30s   # Time to wait for graceful shutdown

[Install]
WantedBy=multi-user.target
```

**Systemd Commands:**

```bash
sudo systemctl daemon-reload        # Reload systemd after creating/modifying the file
sudo systemctl enable myapp.service # Enable to start on boot
sudo systemctl start myapp.service  # Start the service
sudo systemctl status myapp.service # Check status
sudo journalctl -u myapp.service -f # View logs
```

**Other Process Managers**: Supervisor, PM2 (more common in Node.js but can manage Go apps).

### b) Containerization (Docker)

Docker is a very popular way to deploy Go applications.

**Create a `Dockerfile`:**

A multi-stage Dockerfile is recommended to keep the final image small and secure.

```dockerfile
# ---- Build Stage ----
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
# This layer is cached unless go.mod/go.sum changes
RUN go mod download && go mod verify

# Copy the entire source code
COPY . .

# Build the application
# Replace ./cmd/myapp/main.go with the path to your main package
ARG APP_CMD_PATH=./cmd/myapp/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/myapp_binary ${APP_CMD_PATH}


# ---- Production Stage ----
FROM alpine:latest

# Set working directory
WORKDIR /app

# (Optional) Add a non-root user for security
# RUN addgroup -S myappgroup && adduser -S myappuser -G myappgroup
# USER myappuser

# Copy the built binary from the builder stage
COPY --from=builder /app/myapp_binary /app/myapp_binary

# (Optional) Copy .env files if you use them for non-sensitive defaults,
# but prefer environment variables for containers.
# COPY configs/.production.env /app/.env

# (Optional) Copy static assets or templates if your app serves them
# COPY web/static /app/web/static
# COPY web/templates /app/web/templates

# Expose the port your application listens on (matches APP_PORT or HTTP_PORT)
EXPOSE 8080

# Command to run the application
# The binary will pick up environment variables passed to the container
CMD ["/app/myapp_binary"]
```

**Build and Run Docker Image:**

```bash
# Build the image
docker build -t myapp-image .
# To specify the main package path during build (if not default)
# docker build --build-arg APP_CMD_PATH=./cmd/myotherapp/main.go -t myotherapp-image .


# Run the container
docker run -d -p 8080:8080   -e "GOE_ENV=production"   -e "APP_PORT=8080"   -e "DB_PASSWORD=your_secure_password"   --name myapp-container myapp-image
```

### c) Orchestration (Kubernetes, Docker Swarm, Nomad)

For managing containerized applications at scale, use orchestration platforms:

*   **Kubernetes**: Define Deployments, Services, ConfigMaps (for non-sensitive config), and Secrets (for sensitive config).
*   **Docker Swarm**: Similar concepts with Services, Configs, and Secrets.
*   **HashiCorp Nomad**: Another popular orchestrator.

These platforms provide features like scaling, self-healing, rolling updates, and sophisticated configuration and secret management.

## 4. Graceful Shutdown

Goe applications are designed for graceful shutdown. When `goe.Run()` is active:

*   It listens for `SIGINT` (Ctrl+C) and `SIGTERM` signals.
*   Upon receiving such a signal, it initiates the Fx application shutdown.
*   `OnStop` hooks for all registered modules (including Goe's core modules like HTTP, DB) are called.
    *   The HTTP server stops accepting new connections and waits for existing requests to complete (up to a timeout).
    *   Database connections are closed.
    *   Custom module `OnStop` hooks are executed for cleanup.

**Considerations for Deployment:**

*   **Process Managers/Orchestrators**: Ensure your process manager or orchestrator sends `SIGINT` or `SIGTERM` (not `SIGKILL` initially) to allow for graceful shutdown. Configure appropriate stop timeouts (e.g., `TimeoutStopSec` in systemd, `stop_grace_period` in Docker Compose, `terminationGracePeriodSeconds` in Kubernetes).
*   **Health Checks**: Implement health check endpoints in your Goe application (e.g., a `/healthz` route) that your orchestrator can use to determine if the application instance is healthy and ready to serve traffic.
    *   A simple health check might just return `200 OK`.
    *   A more comprehensive one might check critical dependencies like database connectivity.

## 5. Logging in Production

*   **JSON Format**: Configure `LOG_FORMAT=json` for production. This structured format is easily parsable by log management systems (ELK Stack, Splunk, Datadog, Grafana Loki, etc.).
*   **Console Output**: In containerized environments, logging to `stdout` and `stderr` (which `LOG_OUTPUT=console` does) is standard. The container runtime/orchestrator then collects these logs.
*   **Log Level**: Set `LOG_LEVEL` appropriately (e.g., `info` or `warn` for production) to avoid excessive log volume, which can impact performance and cost.
*   **Log Aggregation**: Use a centralized log aggregation and analysis system to collect, search, and monitor logs from all your application instances.

## 6. Static Assets and Templates

If your Goe application serves frontend assets (CSS, JS, images) or HTML templates:

*   **Embedding**: For simpler deployments, you can embed these assets into your Go binary using Go's `embed` package (Go 1.16+). Your HTTP handlers would then serve them from the embedded filesystem.
*   **Separate Serving**: In larger applications or when using a dedicated frontend framework (React, Vue, Angular), it's common to:
    *   Serve static assets via a CDN.
    *   Have a reverse proxy (like Nginx) serve static assets directly and proxy API requests to your Goe application.
    *   If Goe serves a Single Page Application (SPA), ensure your routing is configured to serve the SPA's `index.html` for client-side routes. Fiber has middleware for this.

Choosing the right deployment strategy depends on your application's scale, your team's expertise, and your infrastructure. Start simple and evolve your deployment process as your needs grow.

Next, we will cover [Contributing to Goe](17-contributing.md).
```
