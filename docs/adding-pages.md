# Adding New Pages - Quick Reference

This guide explains how to add new pages to the LEONA scanner application using our convention-based routing system.

## Architecture Overview

```
Project Structure:
├── internal/components/         ← Templ components (.templ files)
│   ├── navbar_dark.templ       ← Dark navbar (rendered to HTML)
│   ├── footer_four_column.templ
│   └── ...
├── templates/
│   ├── layouts/
│   │   └── base.html           ← Base layout (navbar + footer + page content)
│   ├── components/
│   │   ├── navbar.html         ← Thin wrapper: outputs {{.NavbarDarkHTML}}
│   │   └── footer.html         ← Thin wrapper: outputs footer component
│   └── pages/
│       ├── services.html       ← Page-specific content
│       ├── about.html
│       └── ...
```

## How It Works

1. **Templ components** (`internal/components/*.templ`) are compiled to Go code
2. **Shared data** (`NewSharedData()`) renders templ components to HTML strings
3. **Base layout** (`templates/layouts/base.html`) includes navbar + footer + page content
4. **Page templates** (`templates/pages/*.html`) define only their unique content

## Adding a New Page (3 Steps)

### Step 1: Create Page Template

Create `templates/pages/your-page.html`:

```html
{{define "title"}}Your Page Title | LEONA & CRAVIT{{end}}

{{define "meta"}}
<meta name="description" content="Your page description for SEO">
{{end}}

{{define "content"}}
<!-- Your page content here -->
<div class="bg-white py-24 sm:py-32">
    <div class="mx-auto max-w-7xl px-6 lg:px-8">
        <h1 class="text-4xl font-bold tracking-tight text-gray-900">
            Your Page Heading
        </h1>
        <p class="mt-6 text-lg text-gray-600">
            Your content...
        </p>
    </div>
</div>
{{end}}
```

### Step 2: Register Route

In `cmd/server/main.go`, add ONE line in the routes section:

```go
// Page routes using base layout
r.HandleFunc("/your-page", h.HandlePage("your-page")).Methods("GET")
```

### Step 3: Done! 🎉

Your page is now live at `http://localhost:8080/your-page` with:
- ✅ Dark navbar (from `navbar_dark.templ`)
- ✅ Footer (from templ components)
- ✅ Base layout styling
- ✅ Shared data (colors, contact info, etc.)

## Template Blocks Available

Your page template can define these blocks:

- `{{define "title"}}...{{end}}` - Page title (appears in browser tab)
- `{{define "meta"}}...{{end}}` - Meta tags for SEO
- `{{define "content"}}...{{end}}` - Main page content (required)
- `{{define "extra_head"}}...{{end}}` - Additional head content
- `{{define "styles"}}...{{end}}` - Page-specific CSS
- `{{define "scripts"}}...{{end}}` - Page-specific JavaScript

## Accessing Shared Data

In your page templates, you can access:

```html
{{.SiteName}}     → "LEONA & CRAVIT"
{{.SiteURL}}      → "https://leona-cravit.be"
{{.ActivePage}}   → The page name you passed to HandlePage()
{{.Colors.Primary}}   → "#FF6B35" (Orange)
{{.Colors.Secondary}} → "#1428A0" (Royal Blue)
{{.Contact.Email}}    → "info@craleona.be"
```

## Creating Reusable Components (Advanced)

If you need a new reusable component:

### 1. Create Templ Component

Create `internal/components/my_component.templ`:

```templ
package components

templ MyComponent() {
    <div class="my-component">
        <h2>My Reusable Component</h2>
    </div>
}
```

### 2. Generate Go Code

```bash
go run github.com/a-h/templ/cmd/templ@latest generate
```

### 3. Add to Shared Data

In `internal/handler/template_data.go`, update `NewSharedData()`:

```go
func NewSharedData(activePage string) map[string]interface{} {
    return map[string]interface{}{
        "NavbarDarkHTML":   renderToString(components.NavbarDark()),
        "MyComponentHTML":  renderToString(components.MyComponent()),  // Add this
        // ... existing data
    }
}
```

### 4. Use in Pages

In any page template:

```html
{{define "content"}}
<div>
    {{.MyComponentHTML}}
</div>
{{end}}
```

## Page Template Starter

Copy this starter template for new pages:

```html
{{define "title"}}Page Title | LEONA & CRAVIT{{end}}

{{define "meta"}}
<meta name="description" content="Page description">
{{end}}

{{define "content"}}
<!-- Hero Section -->
<div class="bg-gray-900 py-24 sm:py-32">
    <div class="mx-auto max-w-7xl px-6 lg:px-8">
        <h1 class="text-4xl font-bold tracking-tight text-white sm:text-6xl">
            Your Hero Title
        </h1>
        <p class="mt-6 text-lg text-gray-200">
            Your hero description
        </p>
    </div>
</div>

<!-- Content Section -->
<div class="bg-white py-24 sm:py-32">
    <div class="mx-auto max-w-7xl px-6 lg:px-8">
        <div class="mx-auto max-w-2xl">
            <h2 class="text-3xl font-bold tracking-tight text-gray-900">
                Section Title
            </h2>
            <p class="mt-6 text-lg text-gray-600">
                Your content...
            </p>
        </div>
    </div>
</div>
{{end}}
```

## Examples

### Simple About Page

```go
// In cmd/server/main.go
r.HandleFunc("/about", h.HandlePage("about")).Methods("GET")
```

```html
<!-- In templates/pages/about.html -->
{{define "title"}}About Us | LEONA & CRAVIT{{end}}

{{define "content"}}
<div class="bg-white py-24">
    <div class="mx-auto max-w-7xl px-6">
        <h1 class="text-4xl font-bold">About LEONA & CRAVIT</h1>
        <p class="mt-6 text-lg">We help companies achieve CRA compliance...</p>
    </div>
</div>
{{end}}
```

### Contact Page with Form

```go
r.HandleFunc("/contact", h.HandlePage("contact")).Methods("GET")
```

```html
{{define "title"}}Contact Us | LEONA & CRAVIT{{end}}

{{define "content"}}
<div class="bg-white py-24">
    <div class="mx-auto max-w-2xl px-6">
        <h1 class="text-3xl font-bold">Contact Us</h1>
        <form action="/api/contact" method="POST" class="mt-8 space-y-6">
            <input type="email" name="email" placeholder="Your email" 
                   class="w-full px-4 py-2 border rounded-lg" />
            <textarea name="message" rows="4" placeholder="Your message"
                      class="w-full px-4 py-2 border rounded-lg"></textarea>
            <button type="submit" class="px-6 py-3 bg-orange-500 text-white rounded-lg">
                Send Message
            </button>
        </form>
    </div>
</div>
{{end}}
```

## Tips

1. **Keep it simple**: Only define `content` block for basic pages
2. **Use components**: For reusable UI, create templ components
3. **Follow SOP**: See `templ_sop.md` for converting HTML to templ
4. **Test locally**: Run `go run cmd/server/main.go` to test changes
5. **Consistent naming**: Use kebab-case for page names (e.g., `my-page`)

## Related Files

- `templ_sop.md` - Guide for creating templ components
- `templates/layouts/base.html` - Base layout structure
- `internal/handler/template_data.go` - Shared data configuration
- `internal/components/` - All templ components

---

Last Updated: March 2026
