# Template System - Rails-Style Partials for Go

This project uses a component-based template system with Go's html/template package, inspired by Rails partials and renders.

## Directory Structure

```
templates/
├── layouts/              # Base layouts
│   └── base.html        # Main layout with <head>, <body>, navbar, footer
├── components/           # Reusable components (like Rails partials)
│   ├── navbar.html      # Shared navigation
│   ├── footer.html      # Shared footer
│   ├── cta-demo.html    # Call-to-action section
│   └── feature-grid.html # Feature section with grid
├── fragments/            # Small reusable pieces (AI token-efficient)
│   ├── icons.html       # Icon fragments (check, arrow, menu, etc)
│   └── ui.html          # UI fragments (buttons, badges, logo)
├── pages/                # Page-specific content (minimal code)
│   ├── products.html
│   └── products-simple.html # Example with component composition
└── *.html               # Legacy standalone templates (to be migrated)
```

## Three-Tier Architecture

1. **Fragments** (1-10 lines) - Tiny pieces: buttons, icons, badges
2. **Components** (50-200 lines) - Sections: navbar, footer, feature grids
3. **Pages** (30-50 lines) - Just data + component/fragment refs

## Helper Functions

`internal/handler/template_data.go` provides Rails-like helpers:

- `NewSharedData(activePage)` - Common data (site name, colors, contact)
- `NewCTAData(title, desc, button)` - CTA section data with defaults
- `NewFeatureSection(...)` - Feature grid data

## AI Token Efficiency 🤖

**Problem:** Long templates = expensive AI conversations

**Solution:** Fragment system reduces tokens by 83%!

**Example:**
```html
<!-- Before: 150 tokens -->
<svg viewBox="0 0 20 20" fill="currentColor" class="size-5">...</svg>

<!-- After: 8 tokens -->
{{template "icon-check" .}}
```

**See `AI-EFFICIENCY.md` for complete guide on working with AI efficiently.**

## How It Works

### Base Layout (`layouts/base.html`)
- Defines the overall HTML structure
- Includes common `<head>` elements (fonts, Tailwind, Alpine.js)
- Includes the navbar via `{{template "navbar" .}}`
- Provides blocks that pages can override:
  - `title` - Page title
  - `meta` - Additional meta tags
  - `extra_head` - Extra scripts/styles in head
  - `content` - Main page content
  - `styles` - Custom styles
  - `scripts` - Custom scripts at end of body

### Navbar Component (`components/navbar.html`)
- Shared navigation across all pages
- Uses `.ActivePage` data to highlight the current page
- Consistent mobile menu behavior

### Page Templates (`pages/*.html`)
- Define only the content specific to that page
- Override blocks from base layout as needed
- Example: `pages/products.html`

## Available Components

### 1. Navbar (`components/navbar.html`)
- Auto-included in base layout
- Highlights active page automatically
- Responsive mobile menu

### 2. Footer (`components/footer.html`)
- Auto-included in base layout
- Social links, site navigation
- Copyright info

### 3. CTA Demo (`components/cta-demo.html`)
```go
data["CTA"] = NewCTAData(
    "Your Title",
    "Your Description",
    "Button Text",
)
```
```html
{{template "cta-demo" .CTA}}
```

### 4. Feature Grid (`components/feature-grid.html`)
```go
data["Features"] = NewFeatureSection(
    "Eyebrow Text",
    "Main Title",
    "Description",
    []map[string]string{
        {"Title": "Feature 1", "Description": "Details"},
        {"Title": "Feature 2", "Description": "Details"},
    },
    "https://image-url.com/screenshot.png",
)
```
```html
{{template "feature-grid" .Features}}
```

## Creating a New Page

### 1. Create the Page Template

Create a file in `templates/pages/your-page.html`:

```html
{{define "title"}}Your Page Title{{end}}

{{define "content"}}
    <!-- Hero Section (custom) -->
    <div class="bg-gray-900 py-24">
        <h1>Your Hero</h1>
    </div>

    <!-- Feature Section (component) -->
    {{template "feature-grid" .Feature1}}
    
    <!-- CTA Section (component) -->
    {{template "cta-demo" .CTA}}
{{end}}
```

### 2. Create the Handler (Rails-Style)

In `internal/handler/http_handler_v2.go`:

```go
func (h *HTTPHandlerV2) HandleYourPage(w http.ResponseWriter, r *http.Request) {
    // Load all components you need
    tmpl, err := template.ParseFiles(
        "templates/layouts/base.html",
        "templates/components/navbar.html",
        "templates/components/footer.html",
        "templates/components/cta-demo.html",
        "templates/components/feature-grid.html",
        "templates/pages/your-page.html",
    )
    if err != nil {
        http.Error(w, "Template fout", http.StatusInternalServerError)
        log.Printf("Template parse error: %v", err)
        return
    }

    // Start with shared data (site name, colors, contact, etc.)
    data := NewSharedData("your-page")

    // Add component data
    data["Feature1"] = NewFeatureSection(
        "Fast",
        "Lightning Speed",
        "Our platform is incredibly fast",
        []map[string]string{
            {"Title": "Real-time", "Description": "Instant results"},
            {"Title": "Optimized", "Description": "Built for speed"},
        },
        "", // Uses default image
    )

    data["CTA"] = NewCTAData(
        "Ready to get started?",
        "Join thousands of happy customers",
        "Sign Up Now",
    )

    if err := tmpl.ExecuteTemplate(w, "base", data); err != nil {
        http.Error(w, "Template uitvoer fout", http.StatusInternalServerError)
        log.Printf("Template execute error: %v", err)
    }
}
```

### 3. Register the Route

In `cmd/server/main.go`:

```go
r.HandleFunc("/your-page", h.HandleYourPage).Methods("GET")
```

## Navbar Active Page

The navbar highlights the active page based on the `ActivePage` data:

```go
data := map[string]interface{}{
    "ActivePage": "producten",  // Highlights "Producten" in navbar
}
```

In the navbar template, this works via:
```html
<a href="/producten" class="{{if eq .ActivePage "producten"}}text-orange-600{{else}}text-gray-900{{end}}">
    Producten
</a>
```

## Migration Plan

Legacy standalone templates (in `templates/*.html`) can be migrated to this system by:

1. Moving page-specific content to `templates/pages/`
2. Removing duplicate navbar code
3. Using the base layout and shared navbar
4. Updating handlers to use `ParseFiles` with the layout system

## Benefits

### Like Rails Partials
- ✅ **DRY**: Components defined once, used everywhere
- ✅ **Composition**: Mix and match components like LEGO blocks
- ✅ **Shared Data**: Helpers provide consistent defaults
- ✅ **Minimal Code**: Pages are just data + component refs

### Example Comparison

**Before (200+ lines per page):**
```html
<!-- Full navbar HTML -->
<!-- Full hero HTML -->
<!-- Full feature section HTML -->
<!-- Full CTA HTML -->
<!-- Full footer HTML -->
```

**After (35 lines per page):**
```html
{{define "content"}}
    <div class="hero">...</div>
    {{template "feature-grid" .Feature1}}
    {{template "cta-demo" .CTA}}
{{end}}
```

### Real-World Example

See `templates/pages/products-simple.html` (34 lines) vs the old standalone approach.

The handler is also cleaner:
- Shared data via `NewSharedData()`
- Component data via helper functions
- No duplicate navbar/footer code

## Creating New Components

Add a new component in `templates/components/your-component.html`:

```html
{{define "your-component"}}
<div class="component">
    <h2>{{.Title}}</h2>
    <p>{{.Description}}</p>
    {{range .Items}}
        <div>{{.Name}}</div>
    {{end}}
</div>
{{end}}
```

Add a helper in `internal/handler/template_data.go`:

```go
func NewYourComponent(title, desc string, items []map[string]string) map[string]interface{} {
    return map[string]interface{}{
        "Title":       title,
        "Description": desc,
        "Items":       items,
    }
}
```

Use it:

```go
data["MyComponent"] = NewYourComponent(
    "Title",
    "Description",
    []map[string]string{{"Name": "Item 1"}},
)
```

```html
{{template "your-component" .MyComponent}}
```
