# Shared Data - NewSharedData()

This guide explains the `NewSharedData()` function and when you need to modify it.

## What is NewSharedData?

`NewSharedData()` is a helper function in `internal/handler/template_data.go` that creates a **data package** sent to every page template. It's like a "toolbox" that all your pages can access.

## What's Inside?

```go
func NewSharedData(activePage string) map[string]interface{} {
    return map[string]interface{}{
        // Navigation & Layout
        "NavbarDarkHTML":      "...",  // Dark navbar component
        "FooterFourColumnHTML": "...",  // Footer component
        
        // Reusable Components
        "TrustedTeamsHTML":     "...",
        "KnowledgeHTML":        "...",
        "FeatureWorkflowHTML":  "...",
        "PricingThreeTiersHTML": "...",
        // ... more components
        
        // Page Info
        "ActivePage": "index",
        "SiteName":   "LEONA & CRAVIT",
        "SiteURL":    "https://leona-cravit.be",
        
        // Branding
        "Colors": {
            "Primary":   "#FF6B35",   // Orange
            "Secondary": "#1428A0",   // Royal Blue
            "Accent":    "#4169E1",   // Blue Light
        },
        
        // Contact Info
        "Contact": {
            "Email":    "info@craleona.be",
            "Phone":    "+32 xxx xxx xxx",
            "LinkedIn": "...",
            "GitHub":   "...",
        }
    }
}
```

## How It Works

### 1. HandlePage uses it

```go
func (h *HTTPHandlerV2) HandlePage(pageName string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Load shared data
        data := NewSharedData(pageName)
        
        // Render with base layout
        tmpl.ExecuteTemplate(w, "base", data)
    }
}
```

### 2. Templates access it

```html
{{define "content"}}
<!-- Access site data -->
<h1>Welcome to {{.SiteName}}</h1>
<p style="color: {{.Colors.Primary}}">Orange text</p>

<!-- Use components -->
{{.FeatureWorkflowHTML}}
{{.PricingThreeTiersHTML}}

<!-- Access contact -->
<a href="mailto:{{.Contact.Email}}">Email us</a>
{{end}}
```

### 3. Templ components are rendered

```go
// This function converts .templ components to HTML strings
func renderToString(c templ.Component) template.HTML {
    var buf bytes.Buffer
    c.Render(context.Background(), &buf)
    return template.HTML(buf.String())
}

// Used like:
"NavbarDarkHTML": renderToString(components.NavbarDark()),
```

## When to Modify NewSharedData

### ✅ You DON'T need to modify it when:

- **Creating new pages**
  ```bash
  # Just create the file and add route - that's it!
  templates/pages/about.html
  r.HandleFunc("/about", h.HandlePage("about"))
  ```

- **Using existing components**
  ```html
  {{.FeatureWorkflowHTML}}  <!-- Already available! -->
  ```

- **Accessing existing data**
  ```html
  {{.SiteName}}  {{.Colors.Primary}}  <!-- Already there! -->
  ```

### ⚠️ You DO need to modify it when:

#### 1. Adding a NEW templ component globally

When you create a new component in `internal/components/my_component.templ`:

```go
// 1. Create component
templ MyNewComponent() {
    <div>My new component</div>
}

// 2. Generate Go code
// Run: go run github.com/a-h/templ/cmd/templ@latest generate

// 3. Add to NewSharedData in template_data.go:
"MyNewComponentHTML": renderToString(components.MyNewComponent()),
```

#### 2. Adding new site-wide data

```go
// Add new color
"Colors": map[string]string{
    "Primary":   "#FF6B35",
    "Secondary": "#1428A0",
    "Accent":    "#4169E1",
    "Tertiary":  "#10B981",  // NEW!
},

// Add new contact method
"Contact": map[string]string{
    "Email":    "info@craleona.be",
    "Phone":    "+32 xxx xxx xxx",
    "WhatsApp": "+32 xxx xxx xxx",  // NEW!
},
```

#### 3. Updating existing values

```go
// Change site name, URL, contact info, etc.
"SiteName": "LEONA & CRAVIT CRA Scanner",  // Updated
"Contact": map[string]string{
    "Email": "new-email@craleona.be",  // Updated
},
```

## Complete Workflow Examples

### Example 1: Creating a New Page (No NewSharedData Changes)

```bash
# 1. Create page template
# templates/pages/contact.html
```

```html
{{define "title"}}Contact Us | LEONA & CRAVIT{{end}}

{{define "content"}}
<div class="bg-white py-24">
    <h1>Contact {{.SiteName}}</h1>
    <a href="mailto:{{.Contact.Email}}">{{.Contact.Email}}</a>
</div>
{{end}}
```

```go
// 2. Add route in cmd/server/main.go
r.HandleFunc("/contact", h.HandlePage("contact")).Methods("GET")

// 3. Done! No changes to NewSharedData needed
```

### Example 2: Adding a New Global Component

```bash
# 1. Create templ component
# internal/components/hero_banner.templ
```

```templ
package components

templ HeroBanner() {
    <div class="bg-blue-950 py-20">
        <h1 class="text-white text-5xl">Welcome!</h1>
    </div>
}
```

```bash
# 2. Generate Go code
go run github.com/a-h/templ/cmd/templ@latest generate
```

```go
// 3. Update NewSharedData in internal/handler/template_data.go
func NewSharedData(activePage string) map[string]interface{} {
    return map[string]interface{}{
        // ... existing components
        "HeroBannerHTML": renderToString(components.HeroBanner()),  // ADD THIS
        // ... rest of data
    }
}
```

```html
<!-- 4. Use in any page -->
{{define "content"}}
{{.HeroBannerHTML}}
<p>Rest of content...</p>
{{end}}
```

## Available Data Reference

Here's everything available in `{{.___}}` on every page:

### Components (HTML)
```
{{.NavbarDarkHTML}}
{{.FooterFourColumnHTML}}
{{.TrustedTeamsHTML}}
{{.KnowledgeHTML}}
{{.FeatureWorkflowHTML}}
{{.FeatureScreenshotHTML}}
{{.FeatureBorderedScreenshotHTML}}
{{.FeatureThreeColHTML}}
{{.FeatureTestimonialHTML}}
{{.PricingThreeTiersHTML}}
{{.PricingComparisonHTML}}
{{.Page404BackgroundHTML}}
```

### Site Info
```
{{.ActivePage}}  → Current page name
{{.SiteName}}    → "LEONA & CRAVIT"
{{.SiteURL}}     → "https://leona-cravit.be"
```

### Colors
```
{{.Colors.Primary}}    → "#FF6B35" (Orange)
{{.Colors.Secondary}}  → "#1428A0" (Royal Blue)
{{.Colors.Accent}}     → "#4169E1" (Blue Light)
```

### Contact
```
{{.Contact.Email}}     → "info@craleona.be"
{{.Contact.Phone}}     → "+32 xxx xxx xxx"
{{.Contact.LinkedIn}}  → LinkedIn URL
{{.Contact.GitHub}}    → GitHub URL
```

## File Locations

- **Shared Data Function**: `internal/handler/template_data.go`
- **Templ Components**: `internal/components/*.templ`
- **Page Templates**: `templates/pages/*.html`
- **Base Layout**: `templates/layouts/base.html`

## Summary

**Think of NewSharedData as:**
- 📦 A toolbox that every page gets automatically
- 🔧 Set it up once, use everywhere
- 🍳 Like a kitchen pantry - stock once, use by everyone
- ⚙️ Only update when adding new tools/ingredients

**Normal workflow:** Create pages → No touching NewSharedData  
**Special cases:** New component/data → Update NewSharedData once → Available everywhere

---

**See also:**
- `docs/adding-pages.md` - How to create new pages
- `templ_sop.md` - How to create templ components

Last Updated: March 2026
