package handler

// SharedData contains common data used across templates
type SharedData struct {
	ActivePage string
	SiteName   string
	SiteURL    string
	Colors     Colors
	Contact    Contact
}

// Colors contains brand colors
type Colors struct {
	Primary   string // #FF6B35 (Orange)
	Secondary string // #1428A0 (Royal Blue)
	Accent    string // #4169E1 (Blue Light)
}

// Contact contains contact information
type Contact struct {
	Email    string
	Phone    string
	LinkedIn string
	GitHub   string
}

// Feature represents a feature item
type Feature struct {
	Title       string
	Description string
	Icon        string
}

// CTAData represents call-to-action data
type CTAData struct {
	Title       string
	Description string
	ButtonText  string
	ButtonLink  string
}

// FeatureSection represents a feature grid section
type FeatureSection struct {
	Eyebrow     string
	Title       string
	Description string
	Features    []Feature
	ImageUrl    string
}

// NewSharedData creates shared data with default values
func NewSharedData(activePage string) map[string]interface{} {
	return map[string]interface{}{
		"ActivePage": activePage,
		"SiteName":   "LEONA & CRAVIT",
		"SiteURL":    "https://leona-cravit.be",
		"Colors": map[string]string{
			"Primary":   "#FF6B35",
			"Secondary": "#1428A0",
			"Accent":    "#4169E1",
		},
		"Contact": map[string]string{
			"Email":    "info@craleona.be",
			"Phone":    "+32 xxx xxx xxx",
			"LinkedIn": "https://linkedin.com/company/leona-cravit",
			"GitHub":   "https://github.com/leona-cravit",
		},
	}
}

// NewCTAData creates CTA data with defaults
func NewCTAData(title, description, buttonText string) map[string]interface{} {
	if title == "" {
		title = "Start vandaag met CRA compliance"
	}
	if description == "" {
		description = "Vraag een demo aan en ontdek hoe LEONA uw embedded Linux systemen CRA-compliant maakt."
	}
	if buttonText == "" {
		buttonText = "Demo Aanvragen"
	}

	return map[string]interface{}{
		"Title":       title,
		"Description": description,
		"ButtonText":  buttonText,
	}
}

// NewFeatureSection creates feature section data
func NewFeatureSection(eyebrow, title, description string, features []map[string]string, imageUrl string) map[string]interface{} {
	if imageUrl == "" {
		imageUrl = "https://tailwindcss.com/plus-assets/img/component-images/dark-project-app-screenshot.png"
	}

	return map[string]interface{}{
		"Eyebrow":     eyebrow,
		"Title":       title,
		"Description": description,
		"Features":    features,
		"ImageUrl":    imageUrl,
	}
}
