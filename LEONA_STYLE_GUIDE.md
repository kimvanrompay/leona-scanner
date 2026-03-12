# LEONA & CRAVIT - Official Style Guide
**Version:** 1.0  
**Last Updated:** March 12, 2026  
**Purpose:** Fixed design guidelines for consistent branding across all LEONA products

---

## 🎨 Color Palette

### Primary Colors
```css
/* Brand Blue - Primary brand color */
--brand-blue: #0033A0;           /* Deep royal blue for headers, buttons */
--brand-blue-light: #4169E1;     /* Lighter blue for gradients, accents */

/* Brand Orange - Primary accent & highlights */
--brand-orange: #FF6B35;         /* Davis Orange - ALL highlights, CTAs, urgency */
```

### Usage Rules
- **Orange (#FF6B35)** is the PRIMARY accent color for:
  - All section headers in dark backgrounds
  - All icon highlights (checkmarks, badges)
  - Urgent dates and deadlines
  - Primary CTAs and buttons with glow effects
  - Borders on featured cards
  - All pulsing/animated elements

- **Blue (#0033A0, #4169E1)** is for:
  - Navbar branding
  - Secondary buttons
  - Card backgrounds and gradients
  - Professional/corporate sections

### Text Colors - High Contrast Standards

#### On Dark Backgrounds (blue-900, blue-950, gray-900):
```css
--text-primary: text-white;           /* All headlines */
--text-secondary: text-gray-100;      /* Body text - NEVER use gray-300 or gray-400 */
--text-muted: text-gray-200;          /* Subtext - MINIMUM brightness */
```

#### On Light Backgrounds (white, gray-50):
```css
--text-primary: text-gray-900;        /* Headlines */
--text-body: text-gray-700;           /* Body text */
--text-muted: text-gray-600;          /* Subtext */
```

**RULE:** Never use `text-gray-400` or `text-gray-300` on dark backgrounds. Always use `text-gray-100` or brighter.

---

## 🔥 Gradient Patterns

### Hero Gradients
```html
<!-- Orange/Yellow gradient for emphasis -->
<span class="text-transparent bg-clip-text bg-gradient-to-r from-orange-400 via-yellow-300 to-white">
  Klaar of niet.
</span>

<!-- Blue gradient for professional sections -->
<div style="background: linear-gradient(135deg, #1e3a8a 0%, #1e40af 100%);">
```

### Background Gradients
```html
<!-- Dark sections with blue gradient blobs -->
<div class="bg-gray-900">
  <div class="bg-gradient-to-tr from-blue-400 to-blue-600 opacity-30 blur-3xl"></div>
</div>
```

---

## 💡 Component Patterns

### Featured Cards
```html
<!-- Orange-outlined featured card -->
<div class="rounded-2xl outline outline-2 -outline-offset-1" 
     style="background: linear-gradient(135deg, #1e3a8a 0%, #1e40af 100%); 
            outline-color: #FF6B35;">
  <h3 class="font-bold" style="color: #FF6B35;">Section Title</h3>
  <p class="text-gray-100">Body text</p>
</div>
```

### Icon Highlights
```html
<!-- ALL icons should use orange fill -->
<svg viewBox="0 0 20 20" fill="#FF6B35" class="h-6 w-5 flex-none">
  <path d="..."/>
</svg>
```

### Urgency Badges
```html
<!-- Orange pulsing badge -->
<div class="inline-flex items-center gap-3 rounded-full px-5 py-2.5 ring-2 ring-inset" 
     style="background-color: rgba(255, 107, 53, 0.15); border-color: #FF6B35;">
  <span class="relative flex h-3 w-3">
    <span class="animate-ping absolute inline-flex h-full w-full rounded-full opacity-75" 
          style="background-color: #FF6B35;"></span>
    <span class="relative inline-flex rounded-full h-3 w-3" 
          style="background-color: #FF6B35;"></span>
  </span>
  <span class="text-sm/5 font-semibold uppercase tracking-wider text-white">
    Urgent message
  </span>
</div>
```

### CTA Buttons

#### Primary CTA (Orange with glow)
```html
<a href="#" 
   class="w-full rounded-md px-3 py-2 text-center text-sm/6 font-bold text-white" 
   style="background-color: #FF6B35; box-shadow: 0 0 20px rgba(255, 107, 53, 0.5);"
   onmouseover="this.style.backgroundColor='#ff8555'" 
   onmouseout="this.style.backgroundColor='#FF6B35'">
  Primary Action
</a>
```

#### Secondary CTA (Blue)
```html
<a href="#" 
   class="rounded-lg bg-blue-500 px-5 py-2.5 text-sm font-semibold text-white hover:bg-blue-600">
  Secondary Action
</a>
```

### Advisory Notes
```html
<!-- Orange-bordered advisory box -->
<div class="rounded-2xl border-2 p-8" 
     style="background-color: rgba(255, 107, 53, 0.1); border-color: #FF6B35;">
  <p class="text-sm text-gray-100 leading-relaxed">
    <strong style="color: #FF6B35;">Technisch advies:</strong> 
    Advisory text content here.
  </p>
</div>
```

---

## 📐 Border & Outline Standards

### Card Outlines
```css
/* Standard card on dark background */
outline: outline-1 -outline-offset-1 outline-white/20;

/* Featured card with color */
outline: outline-2 -outline-offset-1;
outline-color: #FF6B35;

/* Border thickness */
border-2  /* For emphasis */
ring-2    /* For interactive elements */
```

### Dividers
```css
/* On dark backgrounds */
border-t border-white/10;  /* Minimum visibility */
divide-y divide-white/10;

/* On light backgrounds */
border-t border-gray-200;
divide-y divide-gray-200;
```

---

## 🎯 Timeline & Status Colors

### Date Colors (use semantic colors)
```html
<!-- Completed/Past -->
<svg fill="#10B981" class="h-6 w-5">...</svg>
<strong class="text-green-400">10 dec 2024:</strong>

<!-- Urgent/Current -->
<svg fill="#FF6B35" class="h-6 w-5">...</svg>
<strong style="color: #FF6B35;">11 sept 2026:</strong>

<!-- Future -->
<svg fill="#60A5FA" class="h-6 w-5">...</svg>
<strong class="text-blue-300">11 dec 2027:</strong>
```

---

## 🏷️ Framework Badge Colors

### Multi-Framework Cards
```html
<!-- CRA: Orange border + orange icon -->
<figure style="background-color: rgba(30, 58, 138, 0.4); border-color: #FF6B35;" 
        class="ring-2">
  <div style="background-color: #FF6B35;">Icon</div>
</figure>

<!-- CER: Green border + green icon with glow -->
<figure style="background-color: rgba(30, 58, 138, 0.4); border-color: #10B981;" 
        class="ring-2">
  <div class="bg-green-500" style="box-shadow: 0 0 15px rgba(16, 185, 129, 0.5);">Icon</div>
</figure>

<!-- NIS2: Purple border + purple icon with glow -->
<figure style="background-color: rgba(30, 58, 138, 0.4); border-color: #A855F7;" 
        class="ring-2">
  <div class="bg-purple-500" style="box-shadow: 0 0 15px rgba(168, 85, 247, 0.5);">Icon</div>
</figure>
```

---

## ✍️ Typography

### Font Stack
```css
font-family-display: 'Funnel Display', sans-serif;  /* Headlines */
font-family-sans: 'Switzer', system-ui, sans-serif;  /* Body text */
```

### Heading Hierarchy
```html
<!-- Hero headline -->
<h1 class="text-5xl sm:text-6xl font-semibold tracking-tight text-white">

<!-- Section headline -->
<h2 class="text-4xl sm:text-5xl lg:text-6xl font-bold tracking-tight">

<!-- Card/subsection headline -->
<h3 class="text-2xl font-semibold tracking-tight text-white">
```

### Body Text Sizing
```html
<!-- Hero subtext -->
<p class="text-lg sm:text-xl/8 font-medium text-gray-100">

<!-- Section description -->
<p class="text-lg text-gray-600 leading-relaxed">

<!-- Card body -->
<p class="text-sm leading-relaxed text-gray-100">
```

---

## 🎭 Background Patterns

### Dark Hero Sections
```html
<header class="bg-blue-950">
  <div class="bg-blue-900/25">
    <!-- Radial gradient blob -->
    <svg viewBox="0 0 1208 1024">
      <ellipse cx="604" cy="512" rx="604" ry="512" fill="url(#gradient)"/>
      <defs>
        <radialGradient id="gradient">
          <stop stop-color="#4169E1"/>
          <stop offset="1" stop-color="#60A5FA"/>
        </radialGradient>
      </defs>
    </svg>
  </div>
</header>
```

### Light Content Sections
```html
<section class="bg-white py-24 md:py-32">
  <!-- Content -->
</section>

<section class="bg-gray-50 py-24 md:py-32">
  <!-- Alternate section -->
</section>
```

---

## 🚀 Animation Standards

### Pulse Animation
```html
<!-- Orange pulsing dot -->
<span class="animate-ping absolute inline-flex h-full w-full rounded-full opacity-75" 
      style="background-color: #FF6B35;"></span>
```

### Hover Effects
```css
/* Button hover - lighten by 20% */
background-color: #FF6B35;
hover: #ff8555

/* Glow effect on hover */
box-shadow: 0 0 20px rgba(255, 107, 53, 0.5);
```

### Transition Speed
```css
transition-all  /* Default for most elements */
transition-colors  /* Text color changes only */
```

---

## ✅ Accessibility Rules

### Minimum Contrast Ratios
- **Text on dark backgrounds:** Use `text-gray-100` or brighter (never gray-300/400)
- **Text on light backgrounds:** Use `text-gray-700` or darker
- **Interactive elements:** Must have visible focus states

### Focus States
```css
focus-visible:outline 
focus-visible:outline-2 
focus-visible:outline-offset-2
```

---

## 🚫 Don'ts - Common Mistakes to Avoid

❌ **NEVER use gray-300 or gray-400 on dark backgrounds**  
✅ Use gray-100 or white

❌ **NEVER use multiple competing accent colors**  
✅ Use orange (#FF6B35) consistently for all highlights

❌ **NEVER use thin borders (white/5) on dark backgrounds**  
✅ Use white/10 minimum, white/20 for emphasis

❌ **NEVER use blue for urgency or warnings**  
✅ Use orange (#FF6B35) for all urgent/important elements

❌ **NEVER mix outline styles within the same section**  
✅ Be consistent: outline-1 for standard, outline-2 for featured

---

## 📱 Responsive Design

### Breakpoints (Tailwind defaults)
```
sm:  640px   /* Small devices */
md:  768px   /* Medium devices */
lg:  1024px  /* Large devices */
xl:  1280px  /* Extra large */
```

### Mobile-First Approach
```html
<!-- Start with mobile, add larger screens -->
<h1 class="text-4xl md:text-5xl lg:text-6xl">
<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
```

---

## 🎯 Implementation Checklist

When creating new pages or components:

- [ ] Use orange (#FF6B35) for ALL highlights and accents
- [ ] Use text-gray-100 or brighter on dark backgrounds
- [ ] Add glow effects to primary CTAs
- [ ] Use outline-2 for featured cards with orange border
- [ ] Apply consistent border thickness (white/10 minimum)
- [ ] Use semantic colors for timeline/status (green/orange/blue)
- [ ] Include hover states on interactive elements
- [ ] Test contrast ratios (text must be readable)
- [ ] Apply proper spacing (py-24 for sections)
- [ ] Use Funnel Display for headlines, Switzer for body

---

## 📄 File Reference

**Implementation Example:** `templates/index.html`  
**Color Variables:** Tailwind config in `<script>` tag  
**Typography:** Google Fonts (Funnel Display) + Fontshare (Switzer)

---

**Questions?** Contact the design team or reference the live implementation at `/templates/index.html`

---

*LEONA & CRAVIT Style Guide v1.0 - Consistency is key to professional branding.*
