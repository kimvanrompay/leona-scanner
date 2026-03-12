# Server Logging

Comprehensive HTTP request/response logging with two modes: **compact** (default) and **verbose** (detailed).

## Quick Start

```bash
# Default mode (compact, one-line logs)
make run

# Verbose mode (detailed request/response logs)
LOG_VERBOSE=true make run

# Or add to .env file
echo "LOG_VERBOSE=true" >> .env
```

## Two Logging Modes

### Compact Mode (Default)

**Best for:** Production, normal development

**Example output:**
```
→ 📖 GET /
← ✅ 200 GET / [12ms]

→ 📖 GET /static/css/styles.css
← ✅ 200 GET /static/css/styles.css [3ms]

→ 📝 POST /api/scan
← ✅ 200 POST /api/scan [245ms]

→ 📖 GET /api/pdf/download/abc123
← ✅ 200 GET /api/pdf/download/abc123 [1.8s]
```

**Format:**
```
→ [emoji] [METHOD] [PATH]                    # Request
← [status_emoji] [CODE] [METHOD] [PATH] [DURATION]  # Response
```

### Verbose Mode

**Best for:** Debugging, troubleshooting, development

**Example output:**
```
┌─────────────────────────────────────────────────────────────────────
│ → INCOMING REQUEST 📝
├─────────────────────────────────────────────────────────────────────
│ Method:     POST
│ Path:       /api/scan
│ Query:      tier=2&email=user@example.com
│ Client IP:  192.168.1.100
│ User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)...
│ Content-Type: application/json
│ Referer:    http://localhost:8080/
└─────────────────────────────────────────────────────────────────────

[Your application logs here]

┌─────────────────────────────────────────────────────────────────────
│ ← RESPONSE ✅
├─────────────────────────────────────────────────────────────────────
│ Status:     200 OK
│ Size:       15 KB
│ Duration:   245ms
│ Path:       POST /api/scan
└─────────────────────────────────────────────────────────────────────
```

## What Gets Logged

### Compact Mode
- ✅ HTTP method (with emoji)
- ✅ Request path
- ✅ Response status code
- ✅ Response time
- ✅ Status emoji (success/error indicator)

### Verbose Mode (all of above PLUS)
- ✅ Client IP address (with proxy support)
- ✅ User-Agent string
- ✅ Query parameters
- ✅ Content-Type header
- ✅ Referer header
- ✅ Response size (in KB/MB)
- ✅ Detailed timing

## Method Emojis

| Method | Emoji | Meaning |
|--------|-------|---------|
| GET | 📖 | Reading data |
| POST | 📝 | Creating/submitting data |
| PUT | ✏️ | Updating data |
| PATCH | 🔧 | Partial update |
| DELETE | 🗑️ | Deleting data |
| OPTIONS | 🔍 | CORS preflight |

## Status Emojis

| Status Range | Emoji | Meaning |
|--------------|-------|---------|
| 200-299 | ✅ | Success |
| 300-399 | ↪️ | Redirect |
| 400-499 | ⚠️ | Client error |
| 500+ | ❌ | Server error |

## Real-World Examples

### Example 1: Homepage Visit

**Compact:**
```
→ 📖 GET /
← ✅ 200 GET / [12ms]
→ 📖 GET /static/css/styles.css
← ✅ 200 GET /static/css/styles.css [3ms]
→ 📖 GET /static/js/main.js
← ✅ 200 GET /static/js/main.js [2ms]
```

**Verbose:**
```
┌─────────────────────────────────────────────────────────────────────
│ → INCOMING REQUEST 📖
├─────────────────────────────────────────────────────────────────────
│ Method:     GET
│ Path:       /
│ Client IP:  127.0.0.1
│ User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)...
└─────────────────────────────────────────────────────────────────────

┌─────────────────────────────────────────────────────────────────────
│ ← RESPONSE ✅
├─────────────────────────────────────────────────────────────────────
│ Status:     200 OK
│ Size:       45 KB
│ Duration:   12ms
│ Path:       GET /
└─────────────────────────────────────────────────────────────────────
```

### Example 2: SBOM Scan (Success)

**Compact:**
```
→ 📝 POST /api/scan
← ✅ 200 POST /api/scan [2.3s]
```

**Verbose:**
```
┌─────────────────────────────────────────────────────────────────────
│ → INCOMING REQUEST 📝
├─────────────────────────────────────────────────────────────────────
│ Method:     POST
│ Path:       /api/scan
│ Client IP:  192.168.1.50
│ User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64)...
│ Content-Type: multipart/form-data; boundary=----WebKitFormBoundary
│ Referer:    http://localhost:8080/demo
└─────────────────────────────────────────────────────────────────────

🔍 [Handler logs] Parsing SBOM file...
🔍 [Handler logs] Found 247 components
🔍 [Handler logs] Checking CVEs...
🔍 [Handler logs] Found 12 vulnerabilities

┌─────────────────────────────────────────────────────────────────────
│ ← RESPONSE ✅
├─────────────────────────────────────────────────────────────────────
│ Status:     200 OK
│ Size:       89 KB
│ Duration:   2.3s
│ Path:       POST /api/scan
└─────────────────────────────────────────────────────────────────────
```

### Example 3: Client Error (404)

**Compact:**
```
→ 📖 GET /nonexistent-page
← ⚠️ 404 GET /nonexistent-page [1ms]
```

**Verbose:**
```
┌─────────────────────────────────────────────────────────────────────
│ → INCOMING REQUEST 📖
├─────────────────────────────────────────────────────────────────────
│ Method:     GET
│ Path:       /nonexistent-page
│ Client IP:  127.0.0.1
└─────────────────────────────────────────────────────────────────────

┌─────────────────────────────────────────────────────────────────────
│ ← RESPONSE ⚠️
├─────────────────────────────────────────────────────────────────────
│ Status:     404 Not Found
│ Size:       < 1 KB
│ Duration:   1ms
│ Path:       GET /nonexistent-page
└─────────────────────────────────────────────────────────────────────
```

### Example 4: Server Error (500)

**Compact:**
```
→ 📝 POST /api/checkout/tier1
← ❌ 500 POST /api/checkout/tier1 [45ms]
```

**Verbose:**
```
┌─────────────────────────────────────────────────────────────────────
│ → INCOMING REQUEST 📝
├─────────────────────────────────────────────────────────────────────
│ Method:     POST
│ Path:       /api/checkout/tier1
│ Client IP:  192.168.1.75
│ Content-Type: application/json
└─────────────────────────────────────────────────────────────────────

❌ [Handler logs] Payment provider error: API key invalid

┌─────────────────────────────────────────────────────────────────────
│ ← RESPONSE ❌
├─────────────────────────────────────────────────────────────────────
│ Status:     500 Internal Server Error
│ Size:       < 1 KB
│ Duration:   45ms
│ Path:       POST /api/checkout/tier1
└─────────────────────────────────────────────────────────────────────
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `LOG_VERBOSE` | `false` | Enable verbose logging |

### Ways to Enable Verbose Logging

**Option 1: Command line (temporary)**
```bash
LOG_VERBOSE=true go run cmd/server/main.go
```

**Option 2: .env file (persistent)**
```bash
echo "LOG_VERBOSE=true" >> .env
make run
```

**Option 3: Export (session-wide)**
```bash
export LOG_VERBOSE=true
make run
```

**Option 4: Makefile target (create new target)**
```makefile
run-verbose:
	LOG_VERBOSE=true go run cmd/server/main.go
```

## Production Recommendations

### Development Environment
```bash
# Use verbose for debugging
LOG_VERBOSE=true
```

### Staging Environment
```bash
# Use compact for cleaner logs
# LOG_VERBOSE=false (or unset)
```

### Production Environment
```bash
# Use compact + structured logging to file
# LOG_VERBOSE=false
# LOG_FILE=/var/log/leona-scanner/app.log
```

## Additional Logging Features

### 1. Client IP Detection

Works behind proxies and load balancers:
- Checks `X-Forwarded-For` header (Cloudflare, AWS ALB)
- Checks `X-Real-IP` header (nginx)
- Falls back to `RemoteAddr`

Example:
```
Client IP:  203.0.113.45  (from X-Forwarded-For)
```

### 2. Response Size Formatting

Human-readable sizes:
```
< 1 KB      (for tiny responses)
5 KB        (kilobytes)
1.2 MB      (megabytes)
3 GB        (gigabytes)
```

### 3. Duration Formatting

Auto-scales based on time:
```
500µs       (microseconds for fast requests)
12ms        (milliseconds for normal requests)
1.8s        (seconds for slow requests)
```

### 4. User-Agent Truncation

Long user-agent strings are shortened in verbose mode:
```
Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Apple...
```

## What's NOT Logged (For Security)

For security and privacy, the following are **NEVER** logged:
- ❌ Request body contents (may contain passwords, API keys)
- ❌ Authorization headers
- ❌ Cookie values
- ❌ POST/PUT data payloads
- ❌ File upload contents

If you need to log request bodies for debugging, add custom logging in your handlers, not middleware.

## Combining with Application Logs

The logging middleware works alongside your application logs:

```go
// In your handler
func (h *HTTPHandler) HandleScan(w http.ResponseWriter, r *http.Request) {
    // Middleware logs: → 📝 POST /api/scan
    
    log.Println("🔍 Parsing SBOM file...")
    // ... your code ...
    
    log.Println("✅ Scan complete")
    
    // Middleware logs: ← ✅ 200 POST /api/scan [2.3s]
}
```

## Performance Impact

| Mode | Overhead | Impact |
|------|----------|--------|
| Compact | ~50µs per request | Negligible |
| Verbose | ~200µs per request | Minimal |

Both modes are production-safe and have minimal performance impact.

## Integration with Log Management

### Send logs to file
```bash
# Redirect stdout to file
./leona-scanner 2>&1 | tee app.log

# Or with systemd
StandardOutput=append:/var/log/leona-scanner/app.log
StandardError=append:/var/log/leona-scanner/error.log
```

### Parse logs with tools

**Compact format** is easy to parse:
```bash
# Extract all POST requests
grep "📝 POST" app.log

# Find slow requests (>1s)
grep -E "← .* \[[0-9]+\.[0-9]+s\]" app.log

# Count errors
grep "❌" app.log | wc -l

# Find 404s
grep "⚠️ 404" app.log
```

**Verbose format** is easy to read but harder to parse programmatically.

## Troubleshooting

### Logs not appearing?

```bash
# Check if middleware is registered
grep "LoggingMiddleware" cmd/server/main.go
```

### Want to disable logging temporarily?

```bash
# Redirect to /dev/null (Unix)
./leona-scanner > /dev/null 2>&1

# Or modify code temporarily
# Comment out: r.Use(middleware.LoggingMiddleware)
```

### Want custom log format?

Edit `internal/middleware/logging.go` and modify:
- `logBasicRequest()` - for compact format
- `logVerboseRequest()` - for verbose format
- `logResponse()` - for response format

## Best Practices

### ✅ DO
- Use compact mode in production
- Use verbose mode when debugging specific issues
- Monitor response times to catch slow endpoints
- Watch for 4xx/5xx errors
- Set up log rotation for production

### ❌ DON'T
- Don't log sensitive data (passwords, tokens, API keys)
- Don't parse verbose logs programmatically
- Don't leave verbose mode on in production (too much data)
- Don't rely solely on logs for monitoring (use metrics too)

## Future Enhancements

Possible additions (not implemented yet):
- Structured JSON logging (for log aggregation tools)
- Request ID tracking (for distributed tracing)
- Slow query highlighting (auto-warn on >1s requests)
- Custom log levels (DEBUG, INFO, WARN, ERROR)
- Log sampling (log 1 in N requests under high load)

## Examples in Production

### Cloudflare + nginx + Go

```
Client → Cloudflare → nginx → Go App
         └─ Sets CF-Connecting-IP
                    └─ Sets X-Real-IP
                               └─ Logs client IP correctly
```

### AWS ALB + Go

```
Client → ALB → Go App
         └─ Sets X-Forwarded-For
                └─ Logs client IP correctly
```

Both scenarios work automatically with the logging middleware.

## Summary

**Compact mode (default):**
- ✅ Clean, one-line logs
- ✅ Production-ready
- ✅ Easy to parse with grep/awk

**Verbose mode (LOG_VERBOSE=true):**
- ✅ Detailed request/response info
- ✅ Perfect for debugging
- ✅ Shows client IP, headers, timing

Choose the mode that fits your needs! 🚀
