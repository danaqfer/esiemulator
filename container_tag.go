package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// ContainerConfig represents the JSON configuration for partner beacons
type ContainerConfig struct {
	Pixels []Pixel `json:"pixels"`
}

// Pixel represents a single partner beacon configuration
type Pixel struct {
	ID             string                 `json:"ID"`
	URL            string                 `json:"URL"`
	TYPE           string                 `json:"TYPE"`
	REQ            bool                   `json:"REQ,omitempty"`
	PCT            int                    `json:"PCT,omitempty"`
	CAP            int                    `json:"CAP,omitempty"`
	RC             string                 `json:"RC,omitempty"`
	CONTINENT_FREQ map[string]int         `json:"CONTINENT_FREQ,omitempty"`
	FIRE_EXPR      string                 `json:"FIRE_EXPR,omitempty"`
	SCRIPT         string                 `json:"SCRIPT,omitempty"`
	Extra          map[string]interface{} `json:"-"`
}

// ESIConfig represents the configuration for ESI generation
type ESIConfig struct {
	BrowserVars bool
	MaxWait     int
}

// ProcessContainerConfig processes the JSON configuration and generates ESI includes
func ProcessContainerConfig(config ContainerConfig, esiConfig ESIConfig) (string, ContainerConfig, error) {
	var esiIncludes []string
	var browserPixels []Pixel

	// Process each pixel
	for _, pixel := range config.Pixels {
		// Set defaults if not provided
		if pixel.TYPE == "" {
			pixel.TYPE = "dir"
		}
		if pixel.REQ == false {
			pixel.REQ = true
		}
		if pixel.PCT == 0 {
			pixel.PCT = 100
		}
		if pixel.CAP == 0 {
			pixel.CAP = 1
		}
		if pixel.RC == "" {
			pixel.RC = "default"
		}

		// Filter pixels: keep frm and script types for browser execution
		if pixel.TYPE == "frm" || pixel.TYPE == "script" {
			browserPixels = append(browserPixels, pixel)
			continue
		}

		// Process dir type pixels for ESI conversion
		if pixel.TYPE == "dir" {
			esiInclude, err := generateESIInclude(pixel, esiConfig)
			if err != nil {
				return "", ContainerConfig{}, fmt.Errorf("error generating ESI for pixel %s: %w", pixel.ID, err)
			}
			esiIncludes = append(esiIncludes, esiInclude)
		}
	}

	// Generate the ESI content
	esiContent := generateESIContent(esiIncludes, esiConfig)

	// Create new config with only browser-executed pixels
	browserConfig := ContainerConfig{
		Pixels: browserPixels,
	}

	return esiContent, browserConfig, nil
}

// generateESIInclude generates an ESI include for a single pixel
func generateESIInclude(pixel Pixel, config ESIConfig) (string, error) {
	// Process URL with macro substitution
	processedURL, err := processMacros(pixel.URL, config)
	if err != nil {
		return "", fmt.Errorf("error processing macros in URL: %w", err)
	}

	// Generate ESI include with MAXWAIT=0 for fire-and-forget
	esiInclude := fmt.Sprintf(`<esi:include src="%s" maxwait="%d" />`, processedURL, config.MaxWait)

	return esiInclude, nil
}

// generateESIContent generates the complete ESI content
func generateESIContent(includes []string, config ESIConfig) string {
	var content strings.Builder

	content.WriteString("<!-- ESI Container Generated Content -->\n")
	content.WriteString("<!-- Fire-and-forget pixels with MAXWAIT=0 -->\n\n")

	for _, include := range includes {
		content.WriteString(include)
		content.WriteString("\n")
	}

	return content.String()
}

// processMacros processes macro substitution in URLs
func processMacros(urlStr string, config ESIConfig) (string, error) {
	// Use a non-greedy match to capture everything between ~~ and ~~
	macroRegex := regexp.MustCompile(`~~(.*?)~~`)

	processedURL := macroRegex.ReplaceAllStringFunc(urlStr, func(match string) string {
		// Extract the macro content (remove ~~)
		macroContent := match[2 : len(match)-2]

		// Process the macro
		replacement, err := processMacro(macroContent, config)
		if err != nil {
			// If there's an error, return the original macro
			return match
		}

		return replacement
	})

	return processedURL, nil
}

// processMacro processes a single macro and returns its replacement
func processMacro(macro string, config ESIConfig) (string, error) {
	// Support macros like dl:qs and dl:qs~utm_source
	if strings.Contains(macro, ":") {
		parts := strings.SplitN(macro, ":", 2)
		macroType := parts[0]
		rest := parts[1]
		// Rebuild the parts slice as [macroType, rest split by ~...]
		restParts := strings.Split(rest, "~")
		allParts := append([]string{macroType}, restParts...)
		return processMacroParts(allParts, config)
	}
	// Otherwise, split by ~ as before
	parts := strings.Split(macro, "~")
	return processMacroParts(parts, config)
}

// Helper to process macro parts
func processMacroParts(parts []string, config ESIConfig) (string, error) {
	if len(parts) == 0 {
		return "", fmt.Errorf("empty macro")
	}
	macroType := parts[0]
	switch macroType {
	case "r":
		return "$(TIME)", nil
	case "evid":
		return "$(PMUSER_EVID)", nil
	case "cs":
		return "$(HTTP_COOKIE{consent})", nil
	case "cc":
		return "$(GEO_COUNTRY)", nil
	case "uu":
		return "$(PMUSER_UU)", nil
	case "suu":
		return "$(PMUSER_SUU)", nil
	case "c":
		return processCookieMacro(parts, config)
	case "dl":
		return processDecodeMacro(parts, config)
	default:
		if strings.HasPrefix(macroType, "u") {
			userVar := strings.Replace(macroType, "u", "v", 1)
			return "$(PMUSER_" + strings.ToUpper(userVar) + ")", nil
		}
		return "$(PMUSER_" + strings.ToUpper(macroType) + ")", nil
	}
}

// processCookieMacro processes cookie-related macros
func processCookieMacro(parts []string, config ESIConfig) (string, error) {
	if len(parts) < 2 {
		return "", fmt.Errorf("cookie macro requires at least cookie name")
	}

	cookieName := parts[1]

	// Check for hash directives
	if len(parts) >= 4 {
		hashType := parts[2]
		salt := parts[3]

		switch hashType {
		case "hpr":
			// Hash: path + cookie value
			return fmt.Sprintf("$(PMUSER_COOKIE_HASH_PR_{%s}_{%s})", cookieName, salt), nil
		case "hpo":
			// Hash: cookie value + path
			return fmt.Sprintf("$(PMUSER_COOKIE_HASH_PO_{%s}_{%s})", cookieName, salt), nil
		}
	}

	// Simple cookie value
	return "$(HTTP_COOKIE{" + cookieName + "})", nil
}

// processDecodeMacro processes decode-related macros
func processDecodeMacro(parts []string, config ESIConfig) (string, error) {
	if len(parts) < 2 {
		return "", fmt.Errorf("decode macro requires type")
	}

	decodeType := parts[1]

	switch decodeType {
	case "qs":
		if len(parts) >= 3 {
			// Specific query parameter
			paramName := parts[2]
			return fmt.Sprintf("$(PMUSER_DECODED_QS_{%s})", paramName), nil
		}
		// Full query string
		return "$(PMUSER_DECODED_QUERY_STRING)", nil

	default:
		return "", fmt.Errorf("unknown decode type: %s", decodeType)
	}
}

// GenerateFingerprintID generates a fingerprint ID based on IP, Accept headers, and User-Agent
func GenerateFingerprintID(ipAddress, acceptHeaders, userAgent string) string {
	// Combine the three values
	combined := ipAddress + acceptHeaders + userAgent

	// Generate MD5 hash
	hash := md5.Sum([]byte(combined))

	// Return hex string
	return hex.EncodeToString(hash[:])
}

// URLDecode decodes a URL-encoded string
func URLDecode(encoded string) (string, error) {
	return url.QueryUnescape(encoded)
}

// GenerateCookieHash generates an MD5 hash for cookie values
func GenerateCookieHash(cookieValue, salt string, hashType string) string {
	var combined string

	switch hashType {
	case "hpr":
		// path + cookie value
		combined = salt + cookieValue
	case "hpo":
		// cookie value + path
		combined = cookieValue + salt
	default:
		// default to cookie value only
		combined = cookieValue
	}

	hash := md5.Sum([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// GenerateESIFunctions generates ESI functions for advanced macro processing
func GenerateESIFunctions() string {
	return `<!-- ESI Functions for Advanced Macro Processing -->

<esi:function name="generate_suu">
    <esi:assign name="ip" value="$(CLIENT_IP)" />
    <esi:assign name="accept" value="$(HTTP_ACCEPT)" />
    <esi:assign name="ua" value="$(HTTP_USER_AGENT)" />
    <esi:assign name="combined" value="$(ip)$(accept)$(ua)" />
    <esi:return value="$digest_md5_hex($(combined))" />
</esi:function>

<esi:function name="cookie_hash">
    <esi:assign name="cookie_name" value="$(ARGS{0})" />
    <esi:assign name="hash_type" value="$(ARGS{1})" />
    <esi:assign name="salt" value="$(ARGS{2})" />
    <esi:assign name="cookie_value" value="$(HTTP_COOKIE{$(cookie_name)})" />
    
    <esi:choose>
        <esi:when test="$(hash_type)=='hpr'">
            <esi:return value="$digest_md5_hex($(salt)$(cookie_value))" />
        </esi:when>
        <esi:when test="$(hash_type)=='hpo'">
            <esi:return value="$digest_md5_hex($(cookie_value)$(salt))" />
        </esi:when>
        <esi:otherwise>
            <esi:return value="$(cookie_value)" />
        </esi:otherwise>
    </esi:choose>
</esi:function>

<esi:function name="url_decode">
    <esi:assign name="encoded" value="$(ARGS{0})" />
    <esi:return value="$url_decode($(encoded))" />
</esi:function>

<esi:function name="query_param_decode">
    <esi:assign name="param_name" value="$(ARGS{0})" />
    <esi:assign name="param_value" value="$(QUERY_STRING{$(param_name)})" />
    <esi:return value="$url_decode($(param_value))" />
</esi:function>

<esi:function name="default_value">
    <esi:assign name="value" value="$(ARGS{0})" />
    <esi:assign name="default" value="$(ARGS{1})" />
    <esi:choose>
        <esi:when test="$is_empty($(value))">
            <esi:return value="$(default)" />
        </esi:when>
        <esi:otherwise>
            <esi:return value="$(value)" />
        </esi:otherwise>
    </esi:choose>
</esi:function>`
}
