package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// CVEService handles CVE lookups from NVD API 2.0
type CVEService struct {
	client      *http.Client
	rateLimiter *RateLimiter
	cache       *CVECache
	apiKey      string // Optional - increases rate limit to 50 requests/30s
}

// CVEResult represents a single CVE finding
type CVEResult struct {
	CVEID       string  `json:"cve_id"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"` // CRITICAL, HIGH, MEDIUM, LOW
	BaseScore   float64 `json:"base_score"`
	Vector      string  `json:"vector"`
	Published   string  `json:"published"`
	Modified    string  `json:"modified"`
}

// NVDResponse represents NVD API 2.0 response structure
type NVDResponse struct {
	ResultsPerPage int `json:"resultsPerPage"`
	TotalResults   int `json:"totalResults"`
	Vulnerabilities []struct {
		CVE struct {
			ID          string `json:"id"`
			Descriptions []struct {
				Lang  string `json:"lang"`
				Value string `json:"value"`
			} `json:"descriptions"`
			Published string `json:"published"`
			Modified  string `json:"lastModified"`
			Metrics   struct {
				CVSSMetricV31 []struct {
					CVSSData struct {
						BaseScore      float64 `json:"baseScore"`
						BaseSeverity   string  `json:"baseSeverity"`
						VectorString   string  `json:"vectorString"`
					} `json:"cvssData"`
				} `json:"cvssMetricV31"`
				CVSSMetricV2 []struct {
					CVSSData struct {
						BaseScore    float64 `json:"baseScore"`
						VectorString string  `json:"vectorString"`
					} `json:"cvssData"`
				} `json:"cvssMetricV2"`
			} `json:"metrics"`
		} `json:"cve"`
	} `json:"vulnerabilities"`
}

// RateLimiter implements token bucket algorithm for NVD API rate limiting
type RateLimiter struct {
	tokens     int
	maxTokens  int
	refillRate time.Duration
	mu         sync.Mutex
	lastRefill time.Time
}

// CVECache stores CVE lookups to reduce API calls
type CVECache struct {
	data map[string][]CVEResult
	mu   sync.RWMutex
	ttl  time.Duration
}

// NewCVEService creates a new CVE lookup service
func NewCVEService(apiKey string) *CVEService {
	maxTokens := 5 // Free tier: 5 requests per 30 seconds
	if apiKey != "" {
		maxTokens = 50 // With API key: 50 requests per 30 seconds
	}

	return &CVEService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		rateLimiter: &RateLimiter{
			tokens:     maxTokens,
			maxTokens:  maxTokens,
			refillRate: 30 * time.Second,
			lastRefill: time.Now(),
		},
		cache: &CVECache{
			data: make(map[string][]CVEResult),
			ttl:  24 * time.Hour, // Cache for 24 hours
		},
		apiKey: apiKey,
	}
}

// LookupByCPE queries NVD for CVEs matching a CPE identifier
func (s *CVEService) LookupByCPE(cpe string) ([]CVEResult, error) {
	// Check cache first
	if cached := s.cache.Get(cpe); cached != nil {
		return cached, nil
	}

	// Wait for rate limit token
	if err := s.rateLimiter.Wait(); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Build NVD API URL
	baseURL := "https://services.nvd.nist.gov/rest/json/cves/2.0"
	params := url.Values{}
	params.Add("cpeName", cpe)

	req, err := http.NewRequest("GET", baseURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key if available
	if s.apiKey != "" {
		req.Header.Set("apiKey", s.apiKey)
	}

	// Execute request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("NVD API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("NVD API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var nvdResp NVDResponse
	if err := json.NewDecoder(resp.Body).Decode(&nvdResp); err != nil {
		return nil, fmt.Errorf("failed to decode NVD response: %w", err)
	}

	// Convert to CVEResult
	results := make([]CVEResult, 0)
	for _, vuln := range nvdResp.Vulnerabilities {
		result := CVEResult{
			CVEID:     vuln.CVE.ID,
			Published: vuln.CVE.Published,
			Modified:  vuln.CVE.Modified,
		}

		// Get English description
		for _, desc := range vuln.CVE.Descriptions {
			if desc.Lang == "en" {
				result.Description = desc.Value
				break
			}
		}

		// Extract CVSS v3.1 metrics (preferred)
		if len(vuln.CVE.Metrics.CVSSMetricV31) > 0 {
			cvss := vuln.CVE.Metrics.CVSSMetricV31[0].CVSSData
			result.BaseScore = cvss.BaseScore
			result.Severity = cvss.BaseSeverity
			result.Vector = cvss.VectorString
		} else if len(vuln.CVE.Metrics.CVSSMetricV2) > 0 {
			// Fallback to CVSS v2
			cvss := vuln.CVE.Metrics.CVSSMetricV2[0].CVSSData
			result.BaseScore = cvss.BaseScore
			result.Vector = cvss.VectorString
			result.Severity = scoreToCVSSv3Severity(cvss.BaseScore)
		}

		results = append(results, result)
	}

	// Cache results
	s.cache.Set(cpe, results)

	return results, nil
}

// LookupByComponent looks up CVEs for a component with name and version
func (s *CVEService) LookupByComponent(name, version string) ([]CVEResult, error) {
	// Try to construct CPE from component name
	// This is a heuristic - ideally the SBOM already has CPE
	cpe := fmt.Sprintf("cpe:2.3:a:*:%s:%s:*:*:*:*:*:*:*", name, version)
	return s.LookupByCPE(cpe)
}

// scoreToCVSSv3Severity converts CVSS v2 score to v3 severity rating
func scoreToCVSSv3Severity(score float64) string {
	switch {
	case score >= 9.0:
		return "CRITICAL"
	case score >= 7.0:
		return "HIGH"
	case score >= 4.0:
		return "MEDIUM"
	case score > 0.0:
		return "LOW"
	default:
		return "NONE"
	}
}

// Wait blocks until a rate limit token is available
func (rl *RateLimiter) Wait() error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens if needed
	now := time.Now()
	if now.Sub(rl.lastRefill) >= rl.refillRate {
		rl.tokens = rl.maxTokens
		rl.lastRefill = now
	}

	// Check if token available
	if rl.tokens <= 0 {
		waitTime := rl.refillRate - now.Sub(rl.lastRefill)
		return fmt.Errorf("rate limit exceeded, retry in %v", waitTime)
	}

	// Consume token
	rl.tokens--
	return nil
}

// Get retrieves cached CVE results
func (c *CVECache) Get(key string) []CVEResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data[key]
}

// Set stores CVE results in cache
func (c *CVECache) Set(key string, results []CVEResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = results
}

// Clear removes all cached entries
func (c *CVECache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string][]CVEResult)
}
