# SOP: Converting HTML to Templ Components

Use this guide when you want to convert a raw HTML snippet (e.g., from Tailwind UI) into a reusable Go component using `templ`.

## 1. Create the Component File
Create a new file in `internal/components/` ending in `.templ`.  
*Example:* `internal/components/feature_workflow.templ`

**Template Structure:**
```templ
package components

// Optional: Define properties if you want to pass data
type FeatureWorkflowProps struct {
    Title string
}

// The component function
templ FeatureWorkflow() {
    <!-- PASTE YOUR HTML HERE -->
}
```

## 2. Generate Go Code
Run the `templ` generator to compile your `.templ` file into Go code. This creates a `_templ.go` file next to it.

**Command:**
```bash
go run github.com/a-h/templ/cmd/templ@latest generate
```
*Note: If you have `templ` installed globally, you can just run `templ generate`.*

## 3. Register in HTTP Handler
Open `internal/handler/http_handler_v2.go`.
Locate the handler function (e.g., `HandleIndex`) and add your component to the `data` map.

**Code:**
```go
// Inside HandleIndex...
data := map[string]interface{}{
    // ... existing items
    "MyNewSectionHTML": renderToString(components.FeatureWorkflow()),
}
```

## 4. Place in HTML Template
Open your page template (e.g., `templates/index.html`).
Insert the variable where you want the component to appear.

**Code:**
```html
<!-- ... previous content ... -->
{{.MyNewSectionHTML}}
<!-- ... next content ... -->
```

---

## 🛑 Important Rules & Gotchas

### 1. Curly Braces `{ }`
Templ uses `{ }` for dynamic Go expressions.
- **Problem:** Scripts or code blocks containing `{ }` will confuse the parser.
- **Solution:** Escape them using HTML entities.
  - Replace `{` with `&#123;`
  - Replace `}` with `&#125;`

### 2. Keywords at Start of Line
Templ treats `if`, `for`, `switch`, `case` as control flow keywords.
- **Problem:** A line of text starting with "if" (e.g. inside a `<p>` tag or code block) might be parsed as code.
- **Solution:** Escape the word as a string expression.
  - Change `if` to `{ "if" }`
  - Change `for` to `{ "for" }`

### 3. Do NOT Edit Generated Files
Never edit the files ending in `_templ.go`. They are auto-generated and will be overwritten. Always edit the `.templ` file instead.
