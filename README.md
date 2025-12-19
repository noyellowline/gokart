# GoKart Sidecar

A lightweight, production-ready HTTP proxy sidecar that handles cross-cutting concerns for your applications. Deploy it alongside your app to automatically add request tracing, logging, authentication, and security headers.

**Built with**: Go 1.21+ • Zero external dependencies for core functionality

## Features

- [x] **Configuration Engine** - YAML-based configuration with environment variable templating (`${VAR:DEFAULT}`), early validation and fail-fast approach.
- [x] **HTTP Proxy Core** - High-performance reverse proxy with connection pooling and timeout management.
- [x] **Request Tracing** - W3C-compliant traceparent generation/propagation for request correlation across services.
- [x] **Internal Headers** - Automatic injection of service identity headers (`X-Service-ID`, `X-Client-ID`) to backend.
- [x] **Health Endpoint** - Production-ready `/health` endpoint for Kubernetes liveness/readiness probes.
- [x] **HTTP Logging** - Structured HTTP access logs with trace correlation, smart log levels, and sensitive data redaction.
- [ ] **Security Headers** - Advanced security headers management (CSP, HSTS, CORS) similar to Helmet.js.
- [ ] **Authentication** - JWT validation with public paths whitelist and client identity extraction.
- [ ] **Hot Reload** - Runtime configuration reload without downtime via signal handlers.
- [ ] **Metrics Endpoint** - Prometheus-compatible `/metrics` for request rates, latencies, and error rates.

## Configuration

The proxy uses a secure, validation-first YAML configuration system.

- **Templates & Environment**: Uses `${VAR}` (required) and `${VAR:DEFAULT}` syntax to inject environment variables.
- **Validation**: All configuration is validated on startup with `go-playground/validator`.
- **Resolution**: **Environment Variable** → **YAML Value** → **Code Default**.

**Example (`configs/sidecar.yaml`)**:

```yaml
core:
  server:
    addr: "${SERVER_ADDR:}:${PORT:8080}"
    read_timeout: "${READ_TIMEOUT:10s}"
    write_timeout: "${WRITE_TIMEOUT:10s}"
  proxy:
    target: "${TARGET_URL:http://localhost:3000}"
    max_idle_conns: ${MAX_IDLE_CONNS:100}
features:
  http_logging:
    enabled: true
    config:
      format: "json"
      exclude_paths: ["/health", "/metrics"]
```

**Environment Variables**:

- `LOG_LEVEL` - Sidecar internal log level: `debug`, `info`, `warn`, `error` (default: `info`)
- `SERVICE_ID` - Service identifier for backend correlation (default: `unknown-service`)
- `CONFIG_PATH` - Path to config file (default: `configs/sidecar.yaml`)

**Template Syntax**:

- `${VAR}` - Required variable (fails if not set)
- `${VAR:default}` - Optional with default value
- `${VAR:}` - Optional, empty string if not set
- Split host:port → `${HOST:}:${PORT:8080}` allows optional host with mandatory port default

## Request Tracing

The sidecar automatically manages distributed tracing using W3C-compliant `traceparent` headers.

**Behavior**:

- If incoming request has `traceparent` → **propagates** it to backend
- If missing → **generates** new W3C-compliant traceparent
- Format: `00-{trace_id}-{span_id}-01` (32+16 hex chars)

**Backend receives**:

```http
traceparent: 00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01
X-Service-ID: my-service
```

This enables request correlation across services and centralized log aggregation.

## HTTP Logging

The sidecar provides production-grade structured HTTP logging with automatic request/response tracking, sensitive data redaction, and intelligent log level assignment based on response status codes.

### Key Features

**Structured Logging**: All HTTP logs follow a consistent JSON schema with standardized fields for easy parsing and aggregation in centralized logging systems (ELK, Datadog, CloudWatch, etc.).

**Automatic Correlation**: Every HTTP log includes the `trace_id` from the W3C traceparent header, enabling request correlation across distributed services without any application changes.

**Smart Log Levels**: Response status codes automatically determine log levels:

- `2xx/3xx` responses → `INFO` level
- `4xx` responses → `WARN` level
- `5xx` responses → `ERROR` level

**Sensitive Data Protection**: Configurable keyword-based redaction automatically masks sensitive data in headers, request bodies, and response bodies. All fields matching configured keywords are replaced with `[REDACTED]`.

**Debug Mode**: When `LOG_LEVEL=debug`, the sidecar captures request/response headers and bodies (with size limits and content-type filtering) for deep debugging without code changes.

**Zero Application Changes**: Applications behind the sidecar don't need logging middleware - the sidecar handles all HTTP access logging uniformly across all services.

### Log Schema

Each HTTP request produces a structured log entry with the following fields:

```json
{
  "timestamp": "2025-12-19T10:30:45.123Z",
  "level": "INFO",
  "type": "http.access",
  "message": "http request",
  "trace_id": "0af7651916cd43dd8448eb211c80319c",
  "service_id": "my-service",
  "http": {
    "method": "POST",
    "path": "/api/users",
    "status": 201,
    "duration_ms": 45,
    "request_size": 256,
    "response_size": 512,
    "remote_addr": "192.168.1.100:54321",
    "user_agent": "Mozilla/5.0..."
  }
}
```

**Error responses** (5xx) use type `http.error` for easier filtering:

```json
{
  "level": "ERROR",
  "type": "http.error",
  "message": "http request",
  "http": {
    "status": 502,
    "error": "backend unavailable"
  }
}
```

### Configuration

Enable HTTP logging in `configs/sidecar.yaml`:

```yaml
features:
  http_logging:
    enabled: true
    config:
      # Paths to exclude from logging (health checks, metrics)
      exclude_paths: ["/health", "/metrics", "/favicon.ico"]

      # Sensitive data redaction
      # Any header or body field matching these keywords will be replaced with [REDACTED]
      redact_keywords:
        - password
        - token
        - secret
        - authorization
        - cookie
        - api_key
        - bearer
        - x-api-key
        - x-auth-token
        - credit_card
        - ssn

      # Debug mode configuration (only active when LOG_LEVEL=debug)
      debug:
        log_request_headers: true
        log_response_headers: false
        log_request_body: true
        log_response_body: false
        max_body_size: 4096 # Maximum body size to capture (bytes)
```

### Usage Examples

**Basic Access Logging** (`LOG_LEVEL=info`):

```bash
$ SERVICE_ID=my-service LOG_LEVEL=info go run ./cmd/proxy
```

Request: `GET /api/users`

Log output:

```json
{
  "timestamp": "2025-12-19T10:30:45Z",
  "level": "INFO",
  "type": "http.access",
  "message": "http request",
  "trace_id": "0af7651916cd43dd8448eb211c80319c",
  "service_id": "my-service",
  "http": {
    "method": "GET",
    "path": "/api/users",
    "status": 200,
    "duration_ms": 23,
    "request_size": 0,
    "response_size": 1543
  }
}
```

**Debug Mode with Request Details** (`LOG_LEVEL=debug`):

```bash
$ SERVICE_ID=my-service LOG_LEVEL=debug go run ./cmd/proxy
```

Request: `POST /api/login` with body `{"username":"john","password":"secret123"}`

Log output (with redaction):

```json
{
  "level": "INFO",
  "type": "http.access",
  "message": "http request",
  "trace_id": "abc123...",
  "http": {
    "method": "POST",
    "path": "/api/login",
    "status": 200,
    "duration_ms": 150
  },
  "request_headers": {
    "Content-Type": "application/json",
    "Authorization": "[REDACTED]",
    "User-Agent": "curl/7.68.0"
  },
  "request_body": "{\"username\":\"john\",\"password\":\"[REDACTED]\"}"
}
```

**Error Logging**:

When backend returns 5xx or proxy errors occur:

```json
{
  "level": "ERROR",
  "type": "http.error",
  "message": "http request",
  "trace_id": "def456...",
  "http": {
    "method": "GET",
    "path": "/api/orders",
    "status": 502,
    "duration_ms": 5000,
    "error": "backend unavailable"
  }
}
```

### Path Exclusions

Exclude noisy endpoints from logs to reduce volume:

```yaml
exclude_paths:
  - /health # Kubernetes liveness/readiness probes
  - /metrics # Prometheus scraping endpoint
  - /favicon.ico # Browser requests
  - /_next/static # Next.js static assets (pattern matching future)
```

### Redaction in Action

**Before Redaction**:

```json
{
  "Authorization": "Bearer eyJhbGc...",
  "X-API-Key": "sk_live_abc123",
  "Content-Type": "application/json"
}
```

**After Redaction** (keywords: `authorization`, `api_key`):

```json
{
  "Authorization": "[REDACTED]",
  "X-API-Key": "[REDACTED]",
  "Content-Type": "application/json"
}
```

Body redaction works recursively on JSON objects:

```json
{
  "user": {
    "email": "john@example.com",
    "password": "[REDACTED]",
    "api_token": "[REDACTED]"
  }
}
```

### Context Enrichment

Features can enrich HTTP logs by adding values to the request context. For example, the authentication feature (future) adds `client_id`:

```json
{
  "level": "INFO",
  "type": "http.access",
  "trace_id": "abc123...",
  "client_id": "user_9876", // Added by auth feature
  "http": {
    "method": "GET",
    "path": "/api/profile"
  }
}
```

This allows features to contribute contextual information without tight coupling.

### Performance Considerations

- **Minimal Overhead**: Logging middleware adds ~50-100µs per request in production mode
- **Zero Allocation** for excluded paths (fast return)
- **Body Capture**: Only enabled in debug mode, with configurable size limits (default 4KB)
- **Async Logging**: slog handlers can be configured for non-blocking writes
- **Content-Type Filtering**: Bodies are only captured for `application/json` and `text/*` content types

### Best Practices

1. **Production**: Use `LOG_LEVEL=info` for access logs only (no body/header logging)
2. **Staging**: Use `LOG_LEVEL=warn` to log only errors (4xx/5xx)
3. **Debug**: Temporarily set `LOG_LEVEL=debug` to investigate specific issues with full request/response details
4. **Redaction**: Always include sensitive keywords in `redact_keywords` for your domain (e.g., API keys, auth tokens)
5. **Exclusions**: Add noisy endpoints to `exclude_paths` to reduce log volume and storage costs
6. **Centralization**: Send logs to centralized systems (ELK, Datadog) using `trace_id` for correlation
