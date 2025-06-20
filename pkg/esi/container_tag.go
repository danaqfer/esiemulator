package esi

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// PartnerBeacon represents a single partner beacon configuration
type PartnerBeacon struct {
	ID          string            `json:"id"`          // Unique identifier for the beacon
	Name        string            `json:"name"`        // Human-readable name
	URL         string            `json:"url"`         // The beacon URL to fire
	Method      string            `json:"method"`      // HTTP method (GET, POST, etc.)
	Headers     map[string]string `json:"headers"`     // Additional headers to send
	Parameters  map[string]string `json:"parameters"`  // Query parameters or POST data
	Timeout     int               `json:"timeout"`     // Timeout in milliseconds
	Enabled     bool              `json:"enabled"`     // Whether this beacon is enabled
	Conditions  map[string]string `json:"conditions"`  // Conditional firing rules
	Frequency   string            `json:"frequency"`   // Firing frequency (always, once, etc.)
	Priority    int               `json:"priority"`    // Priority for firing order
	Category    string            `json:"category"`    // Beacon category (analytics, advertising, etc.)
	Description string            `json:"description"` // Description of the beacon
}

// ContainerTagConfig represents the overall container tag configuration
type ContainerTagConfig struct {
	ClientID    string            `json:"clientId"`    // Client identifier
	PropertyID  string            `json:"propertyId"`  // Property identifier
	Environment string            `json:"environment"` // Environment (dev, staging, prod)
	Version     string            `json:"version"`     // Configuration version
	CreatedAt   time.Time         `json:"createdAt"`   // Configuration creation time
	Beacons     []PartnerBeacon   `json:"beacons"`     // List of partner beacons
	Settings    ContainerSettings `json:"settings"`    // Container-wide settings
	Macros      map[string]string `json:"macros"`      // Macro definitions for substitution
}

// ContainerSettings holds container-wide configuration
type ContainerSettings struct {
	MaxConcurrentBeacons int    `json:"maxConcurrentBeacons"` // Maximum concurrent beacon fires
	DefaultTimeout       int    `json:"defaultTimeout"`       // Default timeout in milliseconds
	FireAndForget        bool   `json:"fireAndForget"`        // Whether to use fire-and-forget mode
	MaxWait              int    `json:"maxWait"`              // MAXWAIT value for ESI includes
	EnableLogging        bool   `json:"enableLogging"`        // Whether to enable logging
	EnableErrorHandling  bool   `json:"enableErrorHandling"`  // Whether to handle errors
	DefaultMethod        string `json:"defaultMethod"`        // Default HTTP method
}

// ContainerTagProcessor handles the conversion of JSON beacon configurations to ESI
type ContainerTagProcessor struct {
	config ContainerTagConfig
	esi    *Processor
}

// NewContainerTagProcessor creates a new container tag processor
func NewContainerTagProcessor(config ContainerTagConfig, esiProcessor *Processor) *ContainerTagProcessor {
	return &ContainerTagProcessor{
		config: config,
		esi:    esiProcessor,
	}
}

// LoadConfigFromJSON loads a container tag configuration from JSON
func LoadConfigFromJSON(jsonData []byte) (*ContainerTagConfig, error) {
	var config ContainerTagConfig
	if err := json.Unmarshal(jsonData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse container tag config: %w", err)
	}

	// Set defaults if not provided
	if config.Settings.DefaultTimeout == 0 {
		config.Settings.DefaultTimeout = 5000 // 5 seconds
	}
	if config.Settings.MaxWait == 0 {
		config.Settings.MaxWait = 0 // Fire and forget
	}
	if config.Settings.DefaultMethod == "" {
		config.Settings.DefaultMethod = "GET"
	}
	if config.Settings.MaxConcurrentBeacons == 0 {
		config.Settings.MaxConcurrentBeacons = 10
	}

	return &config, nil
}

// GenerateESIFromConfig generates ESI content from the container tag configuration
func (ctp *ContainerTagProcessor) GenerateESIFromConfig() (string, error) {
	var esiContent strings.Builder

	// Add ESI comment header
	esiContent.WriteString("<!--esi Container Tag Generated ESI -->\n")
	esiContent.WriteString("<!--esi Client: " + ctp.config.ClientID + " -->\n")
	esiContent.WriteString("<!--esi Property: " + ctp.config.PropertyID + " -->\n")
	esiContent.WriteString("<!--esi Environment: " + ctp.config.Environment + " -->\n")
	esiContent.WriteString("<!--esi Generated: " + time.Now().Format(time.RFC3339) + " -->\n\n")

	// Process each beacon
	for _, beacon := range ctp.config.Beacons {
		if !beacon.Enabled {
			continue
		}

		// Check conditions if any
		if !ctp.evaluateConditions(beacon.Conditions) {
			continue
		}

		// Generate ESI include for this beacon
		esiInclude := ctp.generateESIInclude(beacon)
		esiContent.WriteString(esiInclude)
		esiContent.WriteString("\n")
	}

	return esiContent.String(), nil
}

// generateESIInclude generates an ESI include element for a single beacon
func (ctp *ContainerTagProcessor) generateESIInclude(beacon PartnerBeacon) string {
	var include strings.Builder

	// Start ESI include
	include.WriteString("<esi:include src=\"")

	// Build the URL with parameters
	url := ctp.buildBeaconURL(beacon)
	include.WriteString(url)

	// Add MAXWAIT=0 for fire-and-forget
	include.WriteString("\" maxwait=\"0\"")

	// Add timeout if specified
	if beacon.Timeout > 0 {
		include.WriteString(fmt.Sprintf(" timeout=\"%d\"", beacon.Timeout))
	} else if ctp.config.Settings.DefaultTimeout > 0 {
		include.WriteString(fmt.Sprintf(" timeout=\"%d\"", ctp.config.Settings.DefaultTimeout))
	}

	// Add method if not GET
	if beacon.Method != "" && beacon.Method != "GET" {
		include.WriteString(fmt.Sprintf(" method=\"%s\"", beacon.Method))
	} else if ctp.config.Settings.DefaultMethod != "GET" {
		include.WriteString(fmt.Sprintf(" method=\"%s\"", ctp.config.Settings.DefaultMethod))
	}

	// Add onerror="continue" for fire-and-forget
	include.WriteString(" onerror=\"continue\"")

	// Add alt attribute for fallback
	include.WriteString(" alt=\"\"")

	// Close the include
	include.WriteString(" />")

	// Add comment for debugging
	if ctp.config.Settings.EnableLogging {
		include.WriteString(fmt.Sprintf(" <!-- Beacon: %s (%s) -->", beacon.Name, beacon.ID))
	}

	return include.String()
}

// buildBeaconURL constructs the full URL for a beacon with parameters
func (ctp *ContainerTagProcessor) buildBeaconURL(beacon PartnerBeacon) string {
	url := beacon.URL

	// Add query parameters
	if len(beacon.Parameters) > 0 {
		params := make([]string, 0, len(beacon.Parameters))
		for key, value := range beacon.Parameters {
			// Apply macro substitution
			substitutedValue := ctp.substituteMacros(value)
			params = append(params, fmt.Sprintf("%s=%s", key, substitutedValue))
		}

		separator := "?"
		if strings.Contains(url, "?") {
			separator = "&"
		}
		url += separator + strings.Join(params, "&")
	}

	return url
}

// substituteMacros replaces macro placeholders with their values
func (ctp *ContainerTagProcessor) substituteMacros(input string) string {
	result := input

	for macro, value := range ctp.config.Macros {
		placeholder := fmt.Sprintf("${%s}", macro)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	// Add common ESI variable substitutions
	commonMacros := map[string]string{
		"CLIENT_ID":   ctp.config.ClientID,
		"PROPERTY_ID": ctp.config.PropertyID,
		"ENVIRONMENT": ctp.config.Environment,
		"TIMESTAMP":   fmt.Sprintf("%d", time.Now().Unix()),
		"RANDOM":      fmt.Sprintf("%d", time.Now().UnixNano()),
	}

	for macro, value := range commonMacros {
		placeholder := fmt.Sprintf("${%s}", macro)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// evaluateConditions checks if beacon conditions are met
func (ctp *ContainerTagProcessor) evaluateConditions(conditions map[string]string) bool {
	if len(conditions) == 0 {
		return true // No conditions means always fire
	}

	// For now, implement basic condition evaluation
	// This could be expanded to support more complex logic
	for condition, value := range conditions {
		switch condition {
		case "country":
			// Example: only fire for specific countries
			// This would need to be implemented with actual geo-detection
			if value != "" {
				// For now, assume condition is met
				// In a real implementation, you'd check against actual geo data
			}
		case "consent":
			// Example: only fire if user has given consent
			if value == "required" {
				// Check consent status
				// For now, assume consent is given
			}
		case "frequency":
			// Example: control firing frequency
			switch value {
			case "once":
				// Check if already fired in this session
				// For now, assume it's the first time
			case "always":
				// Always fire
			}
		}
	}

	return true // For now, assume all conditions are met
}

// GenerateCompleteESIHTML generates a complete HTML document with ESI includes
func (ctp *ContainerTagProcessor) GenerateCompleteESIHTML() (string, error) {
	esiContent, err := ctp.GenerateESIFromConfig()
	if err != nil {
		return "", err
	}

	var html strings.Builder

	html.WriteString("<!DOCTYPE html>\n")
	html.WriteString("<html>\n")
	html.WriteString("<head>\n")
	html.WriteString("    <title>Container Tag ESI</title>\n")
	html.WriteString("    <meta charset=\"utf-8\">\n")
	html.WriteString("</head>\n")
	html.WriteString("<body>\n")
	html.WriteString("    <!-- Container Tag Generated Content -->\n")
	html.WriteString("    " + esiContent + "\n")
	html.WriteString("    <!-- End Container Tag Content -->\n")
	html.WriteString("</body>\n")
	html.WriteString("</html>")

	return html.String(), nil
}

// ProcessContainerTagESI processes ESI content that contains container tag includes
func (ctp *ContainerTagProcessor) ProcessContainerTagESI(esiContent string, context ProcessContext) (string, error) {
	// Use the existing ESI processor to handle the includes
	return ctp.esi.Process(esiContent, context)
}

// GetBeaconStats returns statistics about the beacons in the configuration
func (ctp *ContainerTagProcessor) GetBeaconStats() map[string]interface{} {
	totalBeacons := len(ctp.config.Beacons)
	enabledBeacons := 0
	categories := make(map[string]int)

	for _, beacon := range ctp.config.Beacons {
		if beacon.Enabled {
			enabledBeacons++
		}
		if beacon.Category != "" {
			categories[beacon.Category]++
		}
	}

	return map[string]interface{}{
		"totalBeacons":    totalBeacons,
		"enabledBeacons":  enabledBeacons,
		"disabledBeacons": totalBeacons - enabledBeacons,
		"categories":      categories,
		"clientId":        ctp.config.ClientID,
		"propertyId":      ctp.config.PropertyID,
		"environment":     ctp.config.Environment,
		"version":         ctp.config.Version,
	}
}
