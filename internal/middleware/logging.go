package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// ResponseWriter wrapper to capture status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

// LoggingMiddleware logs all HTTP requests with verbose output
func LoggingMiddleware(next http.Handler) http.Handler {
	verbose := os.Getenv("LOG_VERBOSE") == "true"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code and size
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     200, // default if WriteHeader not called
			bytesWritten:   0,
		}

		// Log incoming request
		if verbose {
			logVerboseRequest(r)
		} else {
			logBasicRequest(r)
		}

		// Call next handler
		next.ServeHTTP(wrapped, r)

		// Log response
		duration := time.Since(start)
		logResponse(r, wrapped, duration, verbose)
	})
}

// logBasicRequest logs minimal request info (default)
func logBasicRequest(r *http.Request) {
	method := r.Method
	path := r.URL.Path

	// Add emoji based on method
	emoji := getMethodEmoji(method)

	log.Printf("→ %s %s %s", emoji, method, path)
}

// logVerboseRequest logs detailed request info (when LOG_VERBOSE=true)
func logVerboseRequest(r *http.Request) {
	method := r.Method
	path := r.URL.Path
	query := r.URL.RawQuery
	ip := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	emoji := getMethodEmoji(method)

	log.Println("┌─────────────────────────────────────────────────────────────────────")
	log.Printf("│ → INCOMING REQUEST %s", emoji)
	log.Println("├─────────────────────────────────────────────────────────────────────")
	log.Printf("│ Method:     %s", method)
	log.Printf("│ Path:       %s", path)

	if query != "" {
		log.Printf("│ Query:      %s", query)
	}

	log.Printf("│ Client IP:  %s", ip)

	if userAgent != "" {
		// Shorten user agent for readability
		ua := userAgent
		if len(ua) > 60 {
			ua = ua[:57] + "..."
		}
		log.Printf("│ User-Agent: %s", ua)
	}

	// Log important headers
	if contentType := r.Header.Get("Content-Type"); contentType != "" {
		log.Printf("│ Content-Type: %s", contentType)
	}

	if referer := r.Header.Get("Referer"); referer != "" {
		log.Printf("│ Referer:    %s", referer)
	}

	log.Println("└─────────────────────────────────────────────────────────────────────")
}

// logResponse logs response with timing and status
func logResponse(r *http.Request, rw *responseWriter, duration time.Duration, verbose bool) {
	method := r.Method
	path := r.URL.Path
	status := rw.statusCode
	size := rw.bytesWritten

	// Status emoji
	statusEmoji := getStatusEmoji(status)

	// Duration color coding (for verbose mode)
	durationStr := formatDuration(duration)

	if verbose {
		log.Println("┌─────────────────────────────────────────────────────────────────────")
		log.Printf("│ ← RESPONSE %s", statusEmoji)
		log.Println("├─────────────────────────────────────────────────────────────────────")
		log.Printf("│ Status:     %d %s", status, http.StatusText(status))
		log.Printf("│ Size:       %s", formatBytes(size))
		log.Printf("│ Duration:   %s", durationStr)
		log.Printf("│ Path:       %s %s", method, path)
		log.Println("└─────────────────────────────────────────────────────────────────────")
	} else {
		// Compact single-line format
		log.Printf("← %s %d %s %s [%s]", statusEmoji, status, method, path, durationStr)
	}
}

// getMethodEmoji returns emoji for HTTP method
func getMethodEmoji(method string) string {
	switch method {
	case "GET":
		return "📖"
	case "POST":
		return "📝"
	case "PUT":
		return "✏️"
	case "PATCH":
		return "🔧"
	case "DELETE":
		return "🗑️"
	case "OPTIONS":
		return "🔍"
	default:
		return "📡"
	}
}

// getStatusEmoji returns emoji for status code
func getStatusEmoji(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "✅" // Success
	case status >= 300 && status < 400:
		return "↪️" // Redirect
	case status >= 400 && status < 500:
		return "⚠️" // Client error
	case status >= 500:
		return "❌" // Server error
	default:
		return "❓"
	}
}

// formatDuration formats duration with appropriate unit
func formatDuration(d time.Duration) string {
	switch {
	case d < time.Microsecond:
		return d.String()
	case d < time.Millisecond:
		return d.Round(time.Microsecond).String()
	case d < time.Second:
		return d.Round(time.Millisecond).String()
	default:
		return d.Round(10 * time.Millisecond).String()
	}
}

// formatBytes formats bytes into human-readable size
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return "< 1 KB"
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return formatFloat(float64(bytes)/float64(div)) + " " + []string{"KB", "MB", "GB"}[exp]
}

// formatFloat formats float with 1 decimal place
func formatFloat(f float64) string {
	if f < 10 {
		return trimTrailingZeros(formatFloatPrecision(f, 1))
	}
	return trimTrailingZeros(formatFloatPrecision(f, 0))
}

// formatFloatPrecision formats float with specified precision
func formatFloatPrecision(f float64, precision int) string {
	switch precision {
	case 0:
		return formatInt(int(f + 0.5))
	case 1:
		return formatInt(int(f)) + "." + formatInt(int((f-float64(int(f)))*10+0.5))
	default:
		return formatInt(int(f))
	}
}

// formatInt converts int to string
func formatInt(i int) string {
	if i < 10 {
		return string(rune('0' + i))
	}
	return formatInt(i/10) + string(rune('0'+i%10))
}

// trimTrailingZeros removes trailing .0
func trimTrailingZeros(s string) string {
	if strings.HasSuffix(s, ".0") {
		return s[:len(s)-2]
	}
	return s
}

// getClientIP extracts real client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return xff
	}

	// Check X-Real-IP header (used by nginx)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		return ip[:idx]
	}
	return ip
}
