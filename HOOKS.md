# Pre-Commit Hooks

Automatic code quality checks that run before each `git commit`.

## TL;DR

```bash
# Install full hook (recommended)
make hook-install

# Or install light version (faster)
make hook-install-light

# Test it without committing
make hook-test

# Remove if needed
make hook-uninstall
```

## Two Versions Available

### Full Hook (Recommended)

**Checks:** Format + Lint + Tests  
**Install:** `make hook-install`  
**Time:** ~5-10 seconds

✅ **Pros:**
- Catches all issues before commit
- Ensures tests pass
- Most thorough

❌ **Cons:**
- Slower for quick commits
- Runs full test suite

### Light Hook (Fast)

**Checks:** Format + Lint only  
**Install:** `make hook-install-light`  
**Time:** ~2-3 seconds

✅ **Pros:**
- Much faster
- Good for quick fixes
- Still catches code quality issues

❌ **Cons:**
- Skips tests
- Might break CI if tests fail

## What Happens When You Commit

### Example: Successful Commit

```bash
$ git commit -m "Add new feature"

🔍 Running pre-commit checks...
📝 Files to check:
  - internal/handler/new_handler.go

🎨 Checking formatting...
✅ Formatting OK

🔍 Running linter...
✅ Linter OK

🧪 Running tests...
✅ Tests OK

✨ All checks passed! Committing...
[master 1a2b3c4] Add new feature
 1 file changed, 50 insertions(+)
```

### Example: Failed Commit (Unformatted Code)

```bash
$ git commit -m "Quick fix"

🔍 Running pre-commit checks...
📝 Files to check:
  - internal/handler/fix.go

🎨 Checking formatting...
❌ These files are not formatted:
  - internal/handler/fix.go

💡 Run: make format
```

**Fix it:**
```bash
$ make format
✨ Formatting code...

$ git add .
$ git commit -m "Quick fix"
✨ All checks passed! Committing...
```

### Example: Failed Commit (Linter Issues)

```bash
$ git commit -m "Refactor handler"

🔍 Running pre-commit checks...
📝 Files to check:
  - internal/handler/http_handler_v2.go

🎨 Checking formatting...
✅ Formatting OK

🔍 Running linter...
internal/handler/http_handler_v2.go:145:1: Function 'HandleProducts' is too long (92 > 80) (funlen)

❌ Linter found issues
💡 Run: make lint-fix (auto-fix) or make lint (see all issues)
```

**Fix it:**
```bash
# Try auto-fix first
$ make lint-fix

# If that doesn't work, refactor the function
# See GO-QUALITY.md for refactoring tips
```

## Bypassing the Hook (Emergency Only)

Sometimes you need to commit without checks (e.g., work-in-progress):

```bash
# Use --no-verify flag
git commit -m "WIP: incomplete feature" --no-verify
```

⚠️ **Warning:** Only use this for WIP commits. Fix issues before merging!

## Testing the Hook

Test the hook without actually committing:

```bash
# Stage some Go files
git add internal/handler/new_handler.go

# Test the hook
make hook-test
```

This runs all checks as if you were committing.

## Hook Management Commands

| Command | What it does |
|---------|-------------|
| `make hook-install` | Install full hook (format + lint + tests) |
| `make hook-install-light` | Install light hook (format + lint only) |
| `make hook-uninstall` | Remove the hook completely |
| `make hook-test` | Test hook without committing |

## Switching Between Versions

You can switch anytime:

```bash
# Currently using light, want full
make hook-install

# Currently using full, want light
make hook-install-light

# Don't want any hook
make hook-uninstall
```

## Team Setup

### For Your Team

If you want all team members to use the same hook:

**Option 1: Manual (one-time per developer)**
```bash
# Add to onboarding docs
make hook-install
```

**Option 2: Automatic (using git config)**
```bash
# Add to your README.md setup section
git config core.hooksPath .githooks
```

Then commit hooks to `.githooks/` directory (not `.git/hooks/`).

## CI/CD vs Pre-Commit

**Pre-commit hooks:** Catch issues BEFORE they reach the repo  
**CI/CD:** Final safety net for things that slip through

Both are important:

```
Developer → Pre-commit → git push → CI/CD → Deploy
              ↑                        ↑
           Fast checks            Full test suite
           (5 seconds)             (30+ seconds)
```

## Troubleshooting

### Hook not running?

```bash
# Check if hook is executable
ls -la .git/hooks/pre-commit

# Should show: -rwxr-xr-x
# If not, run:
make hook-install
```

### Hook running but failing on fresh clone?

```bash
# Install tools first
make install-tools

# Then install hook
make hook-install
```

### Hook too slow?

```bash
# Switch to light version
make hook-install-light

# Or bypass occasionally
git commit --no-verify
```

### Want to see what the hook does?

```bash
# Read the script
cat .git/hooks/pre-commit

# Or the light version
cat scripts/pre-commit-light
```

## Advanced: Custom Checks

Want to add custom checks? Edit `.git/hooks/pre-commit`:

```bash
# Example: Check for TODO comments
echo "🔍 Checking for TODOs..."
if grep -r "TODO" $STAGED_GO_FILES; then
    echo "⚠️  Found TODO comments"
    # Don't fail, just warn
fi
```

## Philosophy

### Why Pre-Commit Hooks?

**Rob Pike (Go creator):**
> "A little help at commit time is worth a lot of debugging later."

**Benefits:**
- ✅ Catch issues early (cheapest to fix)
- ✅ Keep git history clean
- ✅ Faster code reviews
- ✅ CI/CD runs faster (fewer failures)
- ✅ Better code quality habits

**Trade-off:**
- ⏱️ Adds 2-10 seconds per commit
- 🧠 Have to think about code quality
- 🔄 Occasional friction for WIP commits

**Verdict:** Worth it for professional projects.

## Resources

- [Git Hooks Documentation](https://git-scm.com/book/en/v2/Customizing-Git-Git-Hooks)
- [golangci-lint](https://golangci-lint.run/)
- [Effective Go](https://go.dev/doc/effective_go)

---

**Remember:** The hook is there to help you, not annoy you. If it's consistently failing, that's a signal to improve code quality, not to bypass it! 🚀
