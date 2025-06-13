package esi

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// ProcessorInterface defines the interface needed by Akamai extensions
type ProcessorInterface interface {
	GetConfig() Config
	GetESIVariable(varName, key string, context ProcessContext) string
}

// AkamaiExtensions contains Akamai-specific ESI extensions
type AkamaiExtensions struct {
	processor ProcessorInterface
	variables map[string]string // Storage for assigned variables
}

// NewAkamaiExtensions creates a new Akamai extensions handler
func NewAkamaiExtensions(processor ProcessorInterface) *AkamaiExtensions {
	return &AkamaiExtensions{
		processor: processor,
		variables: make(map[string]string),
	}
}

// ProcessAkamaiExtensions processes Akamai-specific ESI elements
func (a *AkamaiExtensions) ProcessAkamaiExtensions(doc *goquery.Document, context ProcessContext) error {
	if a.processor.GetConfig().Debug {
		fmt.Println("🔍 Processing Akamai ESI extensions...")
	}

	// Process esi:assign elements
	if err := a.processAssign(doc, context); err != nil {
		return err
	}

	// Process esi:eval elements
	if err := a.processEval(doc, context); err != nil {
		return err
	}

	// Process esi:function elements
	if err := a.processFunction(doc, context); err != nil {
		return err
	}

	// Process esi:dictionary elements
	if err := a.processDictionary(doc, context); err != nil {
		return err
	}

	// Process esi:debug elements
	if err := a.processDebug(doc, context); err != nil {
		return err
	}

	// Process extended esi:include features
	if err := a.processExtendedInclude(doc, context); err != nil {
		return err
	}

	return nil
}

// processAssign handles esi:assign elements for variable assignment
func (a *AkamaiExtensions) processAssign(doc *goquery.Document, context ProcessContext) error {
	doc.Find("esi\\:assign, assign").Each(func(i int, s *goquery.Selection) {
		name, nameExists := s.Attr("name")
		value, valueExists := s.Attr("value")

		if !nameExists || name == "" {
			if a.processor.GetConfig().Debug {
				fmt.Println("⚠️  esi:assign missing name attribute")
			}
			s.Remove()
			return
		}

		if valueExists {
			// Direct value assignment
			expandedValue := a.expandVariables(value, context)
			a.variables[name] = expandedValue
		} else {
			// Use element content as value
			content := s.Text()
			expandedValue := a.expandVariables(content, context)
			a.variables[name] = expandedValue
		}

		if a.processor.GetConfig().Debug {
			fmt.Printf("📝 Assigned variable %s = %s\n", name, a.variables[name])
		}

		s.Remove()
	})

	return nil
}

// processEval handles esi:eval elements for expression evaluation
func (a *AkamaiExtensions) processEval(doc *goquery.Document, context ProcessContext) error {
	doc.Find("esi\\:eval, eval").Each(func(i int, s *goquery.Selection) {
		expr, exists := s.Attr("expr")
		if !exists || expr == "" {
			if a.processor.GetConfig().Debug {
				fmt.Println("⚠️  esi:eval missing expr attribute")
			}
			s.Remove()
			return
		}

		result := a.evaluateExpression(expr, context)
		s.ReplaceWithHtml(result)

		if a.processor.GetConfig().Debug {
			fmt.Printf("🧮 Evaluated expression: %s = %s\n", expr, result)
		}
	})

	return nil
}

// processFunction handles esi:function elements for built-in functions
func (a *AkamaiExtensions) processFunction(doc *goquery.Document, context ProcessContext) error {
	doc.Find("esi\\:function, function").Each(func(i int, s *goquery.Selection) {
		name, nameExists := s.Attr("name")
		if !nameExists || name == "" {
			if a.processor.GetConfig().Debug {
				fmt.Println("⚠️  esi:function missing name attribute")
			}
			s.Remove()
			return
		}

		result := a.executeFunction(name, s, context)
		s.ReplaceWithHtml(result)

		if a.processor.GetConfig().Debug {
			fmt.Printf("⚙️  Executed function: %s = %s\n", name, result)
		}
	})

	return nil
}

// processDictionary handles esi:dictionary elements for key-value lookups
func (a *AkamaiExtensions) processDictionary(doc *goquery.Document, context ProcessContext) error {
	doc.Find("esi\\:dictionary, dictionary").Each(func(i int, s *goquery.Selection) {
		src, srcExists := s.Attr("src")
		key, keyExists := s.Attr("key")
		defaultVal, _ := s.Attr("default")

		if !srcExists || !keyExists {
			if a.processor.GetConfig().Debug {
				fmt.Println("⚠️  esi:dictionary missing src or key attribute")
			}
			s.Remove()
			return
		}

		result := a.dictionaryLookup(src, key, defaultVal, context)
		s.ReplaceWithHtml(result)

		if a.processor.GetConfig().Debug {
			fmt.Printf("📚 Dictionary lookup: %s[%s] = %s\n", src, key, result)
		}
	})

	return nil
}

// processDebug handles esi:debug elements for development debugging
func (a *AkamaiExtensions) processDebug(doc *goquery.Document, context ProcessContext) error {
	doc.Find("esi\\:debug, debug").Each(func(i int, s *goquery.Selection) {
		if !a.processor.GetConfig().Debug {
			s.Remove()
			return
		}

		debugType, _ := s.Attr("type")
		content := s.Text()

		var debugOutput string
		switch debugType {
		case "vars":
			debugOutput = a.generateVariableDebugOutput(context)
		case "headers":
			debugOutput = a.generateHeaderDebugOutput(context)
		case "cookies":
			debugOutput = a.generateCookieDebugOutput(context)
		case "time":
			debugOutput = time.Now().Format(time.RFC3339)
		default:
			debugOutput = a.expandVariables(content, context)
		}

		debugHtml := fmt.Sprintf("<!-- ESI DEBUG: %s -->", debugOutput)
		s.ReplaceWithHtml(debugHtml)
	})

	return nil
}

// processExtendedInclude handles extended esi:include features specific to Akamai
func (a *AkamaiExtensions) processExtendedInclude(doc *goquery.Document, _ ProcessContext) error {
	doc.Find("esi\\:include, include").Each(func(i int, s *goquery.Selection) {
		// Handle timeout attribute (Akamai extension)
		if timeout, exists := s.Attr("timeout"); exists {
			if a.processor.GetConfig().Debug {
				fmt.Printf("⏱️  Include timeout: %s\n", timeout)
			}
			// TODO: Implement custom timeout handling
		}

		// Handle cacheable attribute (Akamai extension)
		if cacheable, exists := s.Attr("cacheable"); exists {
			if a.processor.GetConfig().Debug {
				fmt.Printf("💾 Include cacheable: %s\n", cacheable)
			}
			// TODO: Implement cacheable directive
		}

		// Handle method attribute (Akamai extension)
		if method, exists := s.Attr("method"); exists && method != "GET" {
			if a.processor.GetConfig().Debug {
				fmt.Printf("🌐 Include method: %s\n", method)
			}
			// TODO: Implement POST/PUT support
		}
	})

	return nil
}

// expandVariables expands ESI variables in a string
func (a *AkamaiExtensions) expandVariables(input string, context ProcessContext) string {
	// Regex to match $(VARIABLE), $(VARIABLE{key}), and $(VARIABLE|default) patterns
	varRegex := regexp.MustCompile(`\$\(([A-Za-z_]+)(?:\{([^}]+)\})?(?:\|([^)]+))?\)`)

	return varRegex.ReplaceAllStringFunc(input, func(match string) string {
		matches := varRegex.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}

		varName := matches[1]
		var key string
		var defaultValue string

		if len(matches) > 2 && matches[2] != "" {
			key = matches[2]
		}
		if len(matches) > 3 && matches[3] != "" {
			defaultValue = strings.Trim(matches[3], "'\"")
		}

		// Check for assigned variables first
		if val, exists := a.variables[varName]; exists {
			return val
		}

		// Check for Akamai-specific variables
		value := a.getESIVariable(varName, key, context)
		if value != "" {
			return value
		}

		// Delegate standard ESI variables to processor
		if processor, ok := a.processor.(*Processor); ok {
			value = processor.GetESIVariable(varName, key, context)
		}

		// Return default value if variable is empty and default is specified
		if value == "" && defaultValue != "" {
			return defaultValue
		}

		return value
	})
}

// getESIVariable returns the value of an ESI variable
func (a *AkamaiExtensions) getESIVariable(varName, _ string, context ProcessContext) string {
	// Check for assigned variables first
	if val, exists := a.variables[varName]; exists {
		return val
	}

	// Handle Akamai-specific variables only
	switch varName {
	case "GEO_COUNTRY_CODE":
		return a.getGeoVariable("country_code", context)
	case "GEO_COUNTRY_NAME":
		return a.getGeoVariable("country_name", context)
	case "GEO_REGION":
		return a.getGeoVariable("region", context)
	case "GEO_CITY":
		return a.getGeoVariable("city", context)
	case "CLIENT_IP":
		if ip, exists := context.Headers["X-Forwarded-For"]; exists {
			return strings.Split(ip, ",")[0]
		}
		if ip, exists := context.Headers["X-Real-IP"]; exists {
			return ip
		}
		return ""
	default:
		// Unknown variable - don't delegate to processor to avoid infinite recursion
		if a.processor.GetConfig().Debug {
			fmt.Printf("⚠️  Unknown Akamai ESI variable: %s\n", varName)
		}
		return ""
	}
}

// evaluateExpression evaluates a simple ESI expression
func (a *AkamaiExtensions) evaluateExpression(expr string, context ProcessContext) string {
	// Expand variables first
	expanded := a.expandVariables(expr, context)

	// Simple expression evaluation
	// This is a basic implementation - a full parser would be more robust
	if strings.Contains(expanded, "==") {
		parts := strings.Split(expanded, "==")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			// Remove surrounding quotes if present
			left = strings.Trim(left, "'\"")
			right = strings.Trim(right, "'\"")

			if left == right {
				return "true"
			}
			return "false"
		}
	}

	if strings.Contains(expanded, "!=") {
		parts := strings.Split(expanded, "!=")
		if len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])

			// Remove surrounding quotes if present
			left = strings.Trim(left, "'\"")
			right = strings.Trim(right, "'\"")

			if left != right {
				return "true"
			}
			return "false"
		}
	}

	return expanded
}

// executeFunction executes built-in ESI functions
func (a *AkamaiExtensions) executeFunction(name string, s *goquery.Selection, context ProcessContext) string {
	switch name {
	case "base64_encode":
		input, _ := s.Attr("input")
		expanded := a.expandVariables(input, context)
		return base64.StdEncoding.EncodeToString([]byte(expanded))

	case "base64_decode":
		input, _ := s.Attr("input")
		expanded := a.expandVariables(input, context)
		if decoded, err := base64.StdEncoding.DecodeString(expanded); err == nil {
			return string(decoded)
		}
		return ""

	case "url_encode":
		input, _ := s.Attr("input")
		expanded := a.expandVariables(input, context)
		return url.QueryEscape(expanded)

	case "url_decode":
		input, _ := s.Attr("input")
		expanded := a.expandVariables(input, context)
		if decoded, err := url.QueryUnescape(expanded); err == nil {
			return decoded
		}
		return expanded

	case "strlen":
		input, _ := s.Attr("input")
		expanded := a.expandVariables(input, context)
		return strconv.Itoa(len(expanded))

	case "substr":
		input, _ := s.Attr("input")
		start, _ := s.Attr("start")
		length, _ := s.Attr("length")

		expanded := a.expandVariables(input, context)
		startInt, _ := strconv.Atoi(start)
		lengthInt, _ := strconv.Atoi(length)

		if startInt < 0 || startInt >= len(expanded) {
			return ""
		}

		end := startInt + lengthInt
		if end > len(expanded) {
			end = len(expanded)
		}

		return expanded[startInt:end]

	case "random":
		min, _ := s.Attr("min")
		max, _ := s.Attr("max")
		minInt, _ := strconv.Atoi(min)
		maxInt, _ := strconv.Atoi(max)

		if maxInt <= minInt {
			return strconv.Itoa(minInt)
		}

		// Simple pseudo-random (not cryptographically secure)
		result := minInt + int(time.Now().UnixNano())%(maxInt-minInt+1)
		return strconv.Itoa(result)

	case "time":
		format, exists := s.Attr("format")
		if !exists {
			format = "2006-01-02 15:04:05"
		}
		return time.Now().Format(format)

	default:
		if a.processor.GetConfig().Debug {
			fmt.Printf("⚠️  Unknown ESI function: %s\n", name)
		}
		return ""
	}
}

// dictionaryLookup performs a dictionary lookup (simplified implementation)
func (a *AkamaiExtensions) dictionaryLookup(src, key, defaultVal string, _ ProcessContext) string {
	// This is a simplified implementation
	// In a real implementation, this would fetch and parse the dictionary
	if a.processor.GetConfig().Debug {
		fmt.Printf("📚 Dictionary lookup: src=%s, key=%s\n", src, key)
	}

	// Return default value for now
	if defaultVal != "" {
		return defaultVal
	}

	return ""
}

// Helper functions
func (a *AkamaiExtensions) getUserAgentComponent(userAgent, component string) string {
	if userAgent == "" {
		return ""
	}

	switch component {
	case "browser":
		if strings.Contains(userAgent, "Chrome") {
			return "CHROME"
		} else if strings.Contains(userAgent, "Firefox") {
			return "FIREFOX"
		} else if strings.Contains(userAgent, "Safari") {
			return "SAFARI"
		} else if strings.Contains(userAgent, "Edge") {
			return "EDGE"
		}
		return "OTHER"

	case "os":
		if strings.Contains(userAgent, "Windows") {
			return "WIN"
		} else if strings.Contains(userAgent, "Mac") {
			return "MAC"
		} else if strings.Contains(userAgent, "Linux") {
			return "UNIX"
		}
		return "OTHER"

	case "version":
		// Simplified version extraction
		return "1.0"

	default:
		return ""
	}
}

func (a *AkamaiExtensions) hasLanguage(acceptLang, lang string) string {
	if acceptLang == "" {
		return "false"
	}

	if strings.Contains(acceptLang, lang) {
		return "true"
	}
	return "false"
}

func (a *AkamaiExtensions) getQueryParam(queryString, key string) string {
	if queryString == "" {
		return ""
	}

	values, err := url.ParseQuery(queryString)
	if err != nil {
		return ""
	}

	return values.Get(key)
}

func (a *AkamaiExtensions) getGeoVariable(component string, _ ProcessContext) string {
	// Simplified geo implementation - would integrate with real GeoIP service
	switch component {
	case "country_code":
		return "US"
	case "country_name":
		return "United States"
	case "region":
		return "California"
	case "city":
		return "San Francisco"
	default:
		return ""
	}
}

func (a *AkamaiExtensions) generateVariableDebugOutput(_ ProcessContext) string {
	var output strings.Builder
	output.WriteString("Variables: ")

	for name, value := range a.variables {
		output.WriteString(fmt.Sprintf("%s=%s ", name, value))
	}

	return output.String()
}

func (a *AkamaiExtensions) generateHeaderDebugOutput(context ProcessContext) string {
	var output strings.Builder
	output.WriteString("Headers: ")

	for name, value := range context.Headers {
		output.WriteString(fmt.Sprintf("%s=%s ", name, value))
	}

	return output.String()
}

func (a *AkamaiExtensions) generateCookieDebugOutput(context ProcessContext) string {
	var output strings.Builder
	output.WriteString("Cookies: ")

	for name, value := range context.Cookies {
		output.WriteString(fmt.Sprintf("%s=%s ", name, value))
	}

	return output.String()
}
