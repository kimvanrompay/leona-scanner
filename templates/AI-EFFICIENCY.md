# AI-Efficient Template Development

## Problem: Long Template Files = Expensive AI Tokens

When working with AI on templates, you want to **minimize context** while maximizing clarity.

## Solution: Fragment System

Break down templates into tiny, reusable pieces in `templates/fragments/`.

### Token Savings Example

**Before (150 tokens):**
```html
<svg viewBox="0 0 20 20" fill="currentColor" class="size-5" style="color: #FF6B35;">
    <path fill-rule="evenodd" d="M16.704 4.153a.75.75 0 01.143 1.052l-8 10.5a.75.75 0 01-1.127.075l-4.5-4.5a.75.75 0 011.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 011.05-.143z" clip-rule="evenodd" />
</svg>
```

**After (8 tokens):**
```html
{{template "icon-check" (dict "Style" "color: #FF6B35;")}}
```

**Savings: 94% fewer tokens!**

## Available Fragments

### Icons (`fragments/icons.html`)

```html
{{template "icon-check" .}}
{{template "icon-arrow-right" .}}
{{template "icon-menu" .}}
{{template "icon-close" .}}
{{template "icon-linkedin" .}}
{{template "icon-github" .}}
```

Pass custom parameters:
```html
{{template "icon-check" (dict "Class" "size-6" "Style" "color: #FF6B35;")}}
```

**Note:** Go templates use `{{if .Var}}{{.Var}}{{else}}default{{end}}` syntax for defaults, not `| default`. All fragments handle this correctly.

### UI Elements (`fragments/ui.html`)

```html
<!-- Primary button -->
{{template "button-primary" (dict "Text" "Get Started" "Link" "/signup")}}

<!-- Secondary button -->
{{template "button-secondary" (dict "Text" "Learn more" "Link" "/docs")}}

<!-- Badges -->
{{template "badge-orange" (dict "Text" "New")}}
{{template "badge-blue" (dict "Text" "Beta")}}

<!-- Logo -->
{{template "logo-sphere" (dict "Size" "h-10 w-10" "TextSize" "text-lg")}}

<!-- Section heading -->
{{template "section-heading" (dict 
    "Eyebrow" "Features"
    "Title" "Everything you need"
    "Description" "Build faster"
    "Align" "text-center"
)}}

<!-- Grid background pattern -->
{{template "grid-pattern" (dict "ID" "unique-pattern-id")}}
```

## Working with AI

### ❌ Bad: Sending Full Templates

```
AI, update the button in this file:
[pastes 300 line template]
```
**Cost:** ~500 tokens input

### ✅ Good: Reference Fragments

```
AI, update button-primary fragment to have rounded-full instead of rounded-md
```
**Cost:** ~30 tokens input

### ✅ Even Better: Describe in Terms of Fragments

```
AI, create a hero section using:
- grid-pattern background
- section-heading with "Fast", "Lightning Speed", "So fast"
- button-primary "Get Started"
```
**Cost:** ~40 tokens, AI generates ~80 tokens

## Creating New Fragments

### When to Create a Fragment

Create a fragment when:
- ✅ Used 3+ times across templates
- ✅ Complex SVG/HTML (>50 tokens)
- ✅ Likely to change together (buttons, badges, icons)
- ✅ AI frequently needs to reference it

Don't create a fragment when:
- ❌ Used only once
- ❌ Very simple (1-2 lines)
- ❌ Page-specific content

### Template

```html
{{define "fragment-name"}}
<div class="{{.Class | default "default-class"}}">
    {{.Content}}
</div>
{{end}}
```

## Usage in Handlers

Load fragments in ParseFiles:

```go
tmpl, err := template.ParseFiles(
    "templates/layouts/base.html",
    "templates/components/navbar.html",
    "templates/components/footer.html",
    "templates/fragments/icons.html",    // ← Add fragments
    "templates/fragments/ui.html",       // ← Add fragments
    "templates/pages/your-page.html",
)
```

## AI Conversation Tips

### Instead of:
> "Change the LinkedIn icon SVG to have a different color"

### Say:
> "Update icon-linkedin fragment to accept a Color parameter, defaulting to currentColor"

### Instead of:
> "Add a CTA button that says 'Get Started' with orange background"

### Say:
> "Use button-primary fragment with Text='Get Started'"

## Token Budget Comparison

### Traditional Approach
- Page file: 200 lines = ~1000 tokens
- Component file: 100 lines = ~500 tokens
- **Total per edit**: ~1500 tokens

### Fragment Approach
- Page file: 40 lines = ~200 tokens
- Reference to fragments: ~50 tokens
- **Total per edit**: ~250 tokens

**Result: 83% token reduction!**

## Benefits

1. **Cheaper AI conversations** - Less context to send
2. **Faster iterations** - AI works with smaller files
3. **Clearer prompts** - "Update button-primary" vs pasting HTML
4. **Consistent changes** - Update fragment once, applies everywhere
5. **Better versioning** - Small, focused commits

## Quick Reference

### Load Order
```go
1. layouts/base.html           // Wrapper
2. components/*.html           // Big sections
3. fragments/*.html            // Small pieces
4. pages/*.html                // Your content
```

### In Templates
```html
<!-- Component (50-200 lines) -->
{{template "feature-grid" .Features}}

<!-- Fragment (1-10 lines) -->
{{template "button-primary" (dict "Text" "Click me")}}
```

### With AI
```
"Use icon-check fragment with orange color"
vs
"Add this SVG: <svg>...</svg>" [pastes 150 tokens]
```

---

**Remember:** The smaller and more modular your templates, the cheaper and faster your AI conversations become! 🚀
