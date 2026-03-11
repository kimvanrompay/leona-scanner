package scanner

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"
)

// CycloneDX JSON structure
type CycloneDXBOM struct {
	BOMFormat   string `json:"bomFormat"`
	SpecVersion string `json:"specVersion"`
	Components  []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		CPE     string `json:"cpe"`
		PURL    string `json:"purl"`
		Licenses []struct {
			License struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"license"`
		} `json:"licenses"`
	} `json:"components"`
}

// CycloneDX XML structure
type CycloneDXXML struct {
	XMLName    xml.Name `xml:"bom"`
	Components struct {
		Component []struct {
			Name    string `xml:"name"`
			Version string `xml:"version"`
			CPE     string `xml:"cpe"`
			PURL    string `xml:"purl"`
			Licenses struct {
				License []struct {
					ID   string `xml:"id"`
					Name string `xml:"name"`
				} `xml:"license"`
			} `xml:"licenses"`
		} `xml:"component"`
	} `xml:"components"`
}

// SPDX JSON structure
type SPDXBOM struct {
	SPDXVersion string `json:"spdxVersion"`
	Packages    []struct {
		Name             string   `json:"name"`
		VersionInfo      string   `json:"versionInfo"`
		LicenseConcluded string   `json:"licenseConcluded"`
		LicenseDeclared  string   `json:"licenseDeclared"`
		ExternalRefs     []struct {
			ReferenceType string `json:"referenceType"`
			ReferenceLocator string `json:"referenceLocator"`
		} `json:"externalRefs"`
	} `json:"packages"`
}

// ParseSBOM detects format and parses SBOM data into Component slice
func ParseSBOM(data []byte) ([]Component, error) {
	// Try to detect format
	dataStr := string(data)
	
	// Check for CycloneDX JSON
	if strings.Contains(dataStr, "\"bomFormat\"") && strings.Contains(dataStr, "\"CycloneDX\"") {
		return parseCycloneDXJSON(data)
	}
	
	// Check for CycloneDX XML
	if strings.Contains(dataStr, "<bom") && strings.Contains(dataStr, "cyclonedx") {
		return parseCycloneDXXML(data)
	}
	
	// Check for SPDX JSON
	if strings.Contains(dataStr, "\"spdxVersion\"") || strings.Contains(dataStr, "\"SPDX") {
		return parseSPDXJSON(data)
	}
	
	return nil, fmt.Errorf("onbekend SBOM-formaat. Ondersteunde formaten: CycloneDX (JSON/XML), SPDX (JSON)")
}

func parseCycloneDXJSON(data []byte) ([]Component, error) {
	var bom CycloneDXBOM
	if err := json.Unmarshal(data, &bom); err != nil {
		return nil, fmt.Errorf("fout bij parsen CycloneDX JSON: %w", err)
	}
	
	var components []Component
	for _, c := range bom.Components {
		license := ""
		if len(c.Licenses) > 0 {
			if c.Licenses[0].License.ID != "" {
				license = c.Licenses[0].License.ID
			} else {
				license = c.Licenses[0].License.Name
			}
		}
		
		components = append(components, Component{
			Name:    c.Name,
			Version: c.Version,
			License: license,
			CPE:     c.CPE,
			PURL:    c.PURL,
		})
	}
	
	if len(components) == 0 {
		return nil, fmt.Errorf("geen componenten gevonden in SBOM")
	}
	
	return components, nil
}

func parseCycloneDXXML(data []byte) ([]Component, error) {
	var bom CycloneDXXML
	if err := xml.Unmarshal(data, &bom); err != nil {
		return nil, fmt.Errorf("fout bij parsen CycloneDX XML: %w", err)
	}
	
	var components []Component
	for _, c := range bom.Components.Component {
		license := ""
		if len(c.Licenses.License) > 0 {
			if c.Licenses.License[0].ID != "" {
				license = c.Licenses.License[0].ID
			} else {
				license = c.Licenses.License[0].Name
			}
		}
		
		components = append(components, Component{
			Name:    c.Name,
			Version: c.Version,
			License: license,
			CPE:     c.CPE,
			PURL:    c.PURL,
		})
	}
	
	if len(components) == 0 {
		return nil, fmt.Errorf("geen componenten gevonden in SBOM")
	}
	
	return components, nil
}

func parseSPDXJSON(data []byte) ([]Component, error) {
	var bom SPDXBOM
	if err := json.Unmarshal(data, &bom); err != nil {
		return nil, fmt.Errorf("fout bij parsen SPDX JSON: %w", err)
	}
	
	var components []Component
	for _, p := range bom.Packages {
		license := p.LicenseConcluded
		if license == "" || license == "NOASSERTION" {
			license = p.LicenseDeclared
		}
		
		// Extract CPE from external refs
		cpe := ""
		for _, ref := range p.ExternalRefs {
			if strings.Contains(strings.ToLower(ref.ReferenceType), "cpe") {
				cpe = ref.ReferenceLocator
				break
			}
		}
		
		components = append(components, Component{
			Name:    p.Name,
			Version: p.VersionInfo,
			License: license,
			CPE:     cpe,
			PURL:    "", // SPDX doesn't typically have PURL in same format
		})
	}
	
	if len(components) == 0 {
		return nil, fmt.Errorf("geen packages gevonden in SPDX")
	}
	
	return components, nil
}
