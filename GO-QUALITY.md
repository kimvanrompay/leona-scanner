# Go Code Quality Setup

Professional Go linting and code quality configuration for the LEONA scanner project.

## TL;DR

```bash
# Install tools
make install-tools

# Check code quality
make lint

# Auto-fix issues
make lint-fix

# Full check (format + lint + test)
make check
```

## Metrics & Limits

Following Go community best practices:

| Metric | Limit | What it measures |
|--------|-------|------------------|
| **Function Length** | 80 lines | Total lines in a function |
| **Statements** | 50 statements | Actual code statements (not comments/whitespace) |
| **Cognitive Complexity** | 20 | How hard code is to understand ⭐ |
| **Cyclomatic Complexity** | 15 | Number of execution paths |
| **Line Length** | 120 chars | Max characters per line |

⭐ **Cognitive Complexity** is the most important - it measures mental effort needed to understand code.

## The "Pro" Way: Breaking Up Long Functions

### ❌ Before: Everything in one handler (100+ lines)

```go
func HandleScan(w http.ResponseWriter, r *http.Request) {
    // 20 lines of request parsing
    // 30 lines of validation
    // 25 lines of database operations
    // 15 lines of business logic
    // 20 lines of response formatting
    // Total: 110 lines - TOO LONG!
}
```

### ✅ After: Service layer pattern (20 lines)

```go
func HandleScan(w http.ResponseWriter, r *http.Request) {
    req, err := parseRequest(r)
    if err != nil {
        respondError(w, err, http.StatusBadRequest)
        return
    }

    result, err := scannerService.ProcessScan(req)
    if err != nil {
        respondError(w, err, http.StatusInternalServerError)
        return
    }

    respondJSON(w, result, http.StatusOK)
}

// Helper functions (each <20 lines)
func parseRequest(r *http.Request) (*ScanRequest, error) { ... }
func respondError(w http.ResponseWriter, err error, code int) { ... }
func respondJSON(w http.ResponseWriter, data interface{}, code int) { ... }
```

## Configuration Details

### `.golangci.yml`

Located at project root. Configures all linters with professional limits.

**Enabled linters:**
- **Complexity:** `funlen`, `gocognit`, `cyclop`, `gocyclo`, `nestif`
- **Quality:** `revive`, `govet`, `errcheck`, `staticcheck`, `unused`
- **Style:** `lll`, `gofmt`, `goimports`, `misspell`, `whitespace`
- **Security:** `gosec`, `bodyclose`, `unconvert`

### Makefile Commands

```bash
make help          # Show all commands
make install-tools # Install golangci-lint
make lint          # Check code quality
make lint-fix      # Auto-fix issues
make format        # Format all code
make test          # Run tests
make test-coverage # Tests with coverage report
make run           # Start server
make build         # Build binary
make check         # Full check (format + lint + test)
```

## Common Issues & Fixes

### 1. Function Too Long

**Issue:**
```
http_handler_v2.go:45: Function 'HandleScan' is too long (92 > 80)
```

**Fix:** Extract helpers or move logic to service layer

```go
// Move business logic to service
func (s *ScannerService) ProcessScan(req *ScanRequest) (*ScanResult, error) {
    // Complex logic here
}

// Keep handler thin
func HandleScan(w http.ResponseWriter, r *http.Request) {
    // Just coordination
}
```

### 2. Cognitive Complexity Too High

**Issue:**
```
validator.go:23: cognitive complexity 25 of func `validateRequest` is high (> 20)
```

**Fix:** Extract nested logic into separate functions

```go
// Before: Nested ifs (complexity: 25)
func validateRequest(req *Request) error {
    if req.Field1 != "" {
        if req.Field2 != "" {
            if req.Field3 != "" {
                // More nesting...
            }
        }
    }
}

// After: Early returns (complexity: 8)
func validateRequest(req *Request) error {
    if req.Field1 == "" {
        return errors.New("field1 required")
    }
    if req.Field2 == "" {
        return errors.New("field2 required")
    }
    if req.Field3 == "" {
        return errors.New("field3 required")
    }
    return nil
}
```

### 3. Line Too Long

**Issue:**
```
main.go:45: line is 135 characters (> 120)
```

**Fix:** Break into multiple lines

```go
// Before: 135 chars
tmpl, err := template.ParseFiles("templates/layouts/base.html", "templates/components/navbar.html", "templates/components/footer.html")

// After: Multi-line
tmpl, err := template.ParseFiles(
    "templates/layouts/base.html",
    "templates/components/navbar.html",
    "templates/components/footer.html",
)
```

## Pre-Commit Hook (Optional)

Auto-run linter before each commit:

```bash
# Create .git/hooks/pre-commit
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/sh
make lint
EOF

chmod +x .git/hooks/pre-commit
```

## IDE Integration

### VS Code

Install the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.Go) and add to `.vscode/settings.json`:

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.formatTool": "goimports",
  "editor.formatOnSave": true
}
```

### GoLand / IntelliJ

1. Settings → Tools → Go Linter
2. Select "golangci-lint"
3. Enable "Run on save"

## CI/CD Integration

Add to GitHub Actions / GitLab CI:

```yaml
- name: Lint
  run: make lint

- name: Test
  run: make test
```

## Why These Limits?

**Rob Pike (Go creator):**
> "Clear is better than clever."

**Go Proverbs:**
- "A little copying is better than a little dependency."
- "The bigger the interface, the weaker the abstraction."
- "Errors are values."

### Function Length: 80 lines

Go's verbose error handling (`if err != nil`) makes functions naturally longer than other languages. 80 lines is the sweet spot between "too strict" and "spaghetti code."

### Cognitive Complexity: 20

Research shows humans can hold ~7 items in working memory. Cognitive complexity >20 means the function is doing too much.

### Line Length: 120 chars

Modern monitors are wider. 120 is standard for Go projects (vs 80 in older languages).

## Resources

- [golangci-lint docs](https://golangci-lint.run/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Cognitive Complexity paper](https://www.sonarsource.com/docs/CognitiveComplexity.pdf)

---

**Remember:** These are guidelines, not hard rules. The goal is **readable, maintainable code**, not passing arbitrary metrics. Use judgment! 🧠
