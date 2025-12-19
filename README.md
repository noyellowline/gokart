# GoKart Sidecar

A lightweight, production-ready HTTP proxy sidecar that handles cross-cutting concerns for your applications. Deploy it alongside your app to automatically add request tracing, logging, authentication, and security headers.

**Built with**: Go 1.21+ • Zero external dependencies for core functionality

## Features

- [x] **Configuration Engine** - YAML-based configuration with environment variable templating (`${VAR:DEFAULT}`), early validation and fail-fast approach.
- [x] **HTTP Proxy Core** - High-performance reverse proxy with connection pooling and timeout management.
- [x] **Request Tracing** - W3C-compliant traceparent generation/propagation for request correlation across services.
- [x] **Internal Headers** - Automatic injection of service identity headers (`X-Service-ID`, `X-Client-ID`) to backend.
- [ ] **Async Structured Logging** - Non-blocking logging pipeline with configurable outputs and log-level hot-reload.
- [ ] **Modular Feature System** - Pluggable architecture for features with self-contained config and lifecycle.
- [ ] **HTTP Logging Feature** - Request/response logging middleware with configurable format and path exclusions.
- [ ] **Security Headers Feature** - Advanced security headers management (CSP, HSTS, CORS) similar to Helmet.js.
- [ ] **Authentication Feature** - JWT validation with public paths whitelist capability.
- [ ] **Hot Reload Configuration** - Dynamic reload of config and feature state without downtime.
- [ ] **Health & Metrics Endpoints** - Built-in `/health` and `/metrics` for observability.

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
