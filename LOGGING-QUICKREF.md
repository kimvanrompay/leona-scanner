# Logging Quick Reference

**TL;DR:** Automatic HTTP logging with two modes.

## Commands

```bash
# Default mode (compact)
make run

# Verbose mode (detailed)
make run-verbose

# Or set environment variable
export LOG_VERBOSE=true
make run
```

## What You See

### Compact Mode (Default)
```
→ 📖 GET /
← ✅ 200 GET / [12ms]
```

### Verbose Mode (LOG_VERBOSE=true)
```
┌────────────────────────────────────────────
│ → INCOMING REQUEST 📖
├────────────────────────────────────────────
│ Method:     GET
│ Path:       /
│ Client IP:  127.0.0.1
│ User-Agent: Mozilla/5.0...
└────────────────────────────────────────────

┌────────────────────────────────────────────
│ ← RESPONSE ✅
├────────────────────────────────────────────
│ Status:     200 OK
│ Size:       45 KB
│ Duration:   12ms
│ Path:       GET /
└────────────────────────────────────────────
```

## Method Emojis

| GET | POST | PUT | PATCH | DELETE |
|-----|------|-----|-------|--------|
| 📖 | 📝 | ✏️ | 🔧 | 🗑️ |

## Status Emojis

| 2xx | 3xx | 4xx | 5xx |
|-----|-----|-----|-----|
| ✅ | ↪️ | ⚠️ | ❌ |

## Features

✅ Response timing (ms/s)  
✅ Response size (KB/MB/GB)  
✅ Client IP (proxy-aware)  
✅ Status codes  
✅ User-Agent  
✅ Query params  

❌ Never logs: passwords, tokens, request bodies, cookies

## Parse Logs

```bash
# Find errors
grep "❌" app.log

# Find 404s
grep "⚠️ 404" app.log

# Find slow requests (>1s)
grep -E "\[[0-9]+\.[0-9]+s\]" app.log

# Count POST requests
grep "📝 POST" app.log | wc -l
```

## Production Tips

- **Dev:** Use verbose (`LOG_VERBOSE=true`)
- **Staging/Prod:** Use compact (default)
- **Performance:** ~50µs overhead (negligible)

See `LOGGING.md` for full documentation.
