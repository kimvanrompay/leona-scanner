# Quick Start: Rails-Style Components

## TL;DR

Build pages with **minimal code** using reusable components, just like Rails partials.

## 30-Second Example

**Page Template** (`pages/your-page.html`) - 20 lines:
```html
{{define "title"}}Your Page{{end}}

{{define "content"}}
    <div class="hero">Hero content</div>
    {{template "feature-grid" .Feature1}}
    {{template "cta-demo" .CTA}}
{{end}}
```

**Handler** (`http_handler_v2.go`) - 30 lines:
```go
func (h *HTTPHandlerV2) HandleYourPage(w http.ResponseWriter, r *http.Request) {
    tmpl, _ := template.ParseFiles(
        "templates/layouts/base.html",
        "templates/components/navbar.html",
        "templates/components/footer.html",
        "templates/components/cta-demo.html",
        "templates/components/feature-grid.html",
        "templates/pages/your-page.html",
    )
    
    data := NewSharedData("your-page")
    
    data["Feature1"] = NewFeatureSection(
        "Fast", "Lightning Speed", "So fast",
        []map[string]string{
            {"Title": "Real-time", "Description": "Instant"},
        },
        "",
    )
    
    data["CTA"] = NewCTAData("Title", "Description", "Button")
    
    tmpl.ExecuteTemplate(w, "base", data)
}
```

**Result:** Full page with navbar, footer, features, and CTA in ~50 total lines.

## Available Components

| Component | Usage | Data Helper |
|-----------|-------|-------------|
| Navbar | Auto-included | - |
| Footer | Auto-included | - |
| CTA Demo | `{{template "cta-demo" .CTA}}` | `NewCTAData()` |
| Feature Grid | `{{template "feature-grid" .Features}}` | `NewFeatureSection()` |

## Data Helpers

All in `internal/handler/template_data.go`:

- **NewSharedData(page)** - Site name, colors, contact, active page
- **NewCTAData(title, desc, button)** - CTA section with smart defaults
- **NewFeatureSection(...)** - Feature grid with bullet points

## File Locations

```
templates/
├── layouts/base.html           # Layout wrapper
├── components/                  # Reusable parts
│   ├── navbar.html
│   ├── footer.html
│   ├── cta-demo.html
│   └── feature-grid.html
└── pages/                       # Your pages (minimal)
    └── your-page.html

internal/handler/
└── template_data.go            # Data helpers
```

## Creating a New Page - 3 Steps

1. **Create template** in `templates/pages/new-page.html`
2. **Create handler** in `internal/handler/http_handler_v2.go`
3. **Register route** in `cmd/server/main.go`

That's it! 🎉

See `templates/README.md` for complete documentation.
