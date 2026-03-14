package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	DefaultLanguage    string   `json:"default_language"`
	SupportedLanguages []string `json:"supported_languages"`
	FallbackLanguage   string   `json:"fallback_language"`
}

type I18n struct {
	config       Config
	translations map[string]map[string]interface{} // lang -> file -> data
	basePath     string
}

// New creates a new i18n manager
func New(basePath string) (*I18n, error) {
	i := &I18n{
		translations: make(map[string]map[string]interface{}),
		basePath:     basePath,
	}

	// Load config
	configPath := filepath.Join(basePath, "config.json")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	if err := json.Unmarshal(configData, &i.config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Load all translations
	for _, lang := range i.config.SupportedLanguages {
		if err := i.loadLanguage(lang); err != nil {
			return nil, fmt.Errorf("failed to load language %s: %w", lang, err)
		}
	}

	return i, nil
}

func (i *I18n) loadLanguage(lang string) error {
	langDir := filepath.Join(i.basePath, lang)
	files, err := os.ReadDir(langDir)
	if err != nil {
		return err
	}

	i.translations[lang] = make(map[string]interface{})

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(langDir, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		var content interface{}
		if err := json.Unmarshal(data, &content); err != nil {
			return err
		}

		// Store without .json extension
		key := strings.TrimSuffix(file.Name(), ".json")
		i.translations[lang][key] = content
	}

	return nil
}

// Get retrieves translation data for a specific namespace (file)
func (i *I18n) Get(lang, namespace string) interface{} {
	if langData, ok := i.translations[lang]; ok {
		if data, ok := langData[namespace]; ok {
			return data
		}
	}

	// Fallback to default language
	if lang != i.config.FallbackLanguage {
		return i.Get(i.config.FallbackLanguage, namespace)
	}

	return nil
}

// GetAll returns all translations for a language as a flat map
// This makes it easy to pass to templates
func (i *I18n) GetAll(lang string) map[string]interface{} {
	if data, ok := i.translations[lang]; ok {
		return data
	}
	return i.translations[i.config.DefaultLanguage]
}

// DefaultLang returns the default language
func (i *I18n) DefaultLang() string {
	return i.config.DefaultLanguage
}

// SupportedLangs returns list of supported languages
func (i *I18n) SupportedLangs() []string {
	return i.config.SupportedLanguages
}

// GetLanguageFromRequest determines language from request (header, query param, cookie)
func (i *I18n) GetLanguageFromRequest(acceptLang string, queryLang string) string {
	// 1. Check query parameter first
	if queryLang != "" {
		for _, supported := range i.config.SupportedLanguages {
			if supported == queryLang {
				return queryLang
			}
		}
	}

	// 2. Parse Accept-Language header
	if acceptLang != "" {
		langs := strings.Split(acceptLang, ",")
		for _, lang := range langs {
			// Extract language code (e.g., "nl-BE" -> "nl")
			langCode := strings.Split(strings.TrimSpace(lang), ";")[0]
			langCode = strings.Split(langCode, "-")[0]

			for _, supported := range i.config.SupportedLanguages {
				if supported == langCode {
					return langCode
				}
			}
		}
	}

	// 3. Fall back to default
	return i.config.DefaultLanguage
}
