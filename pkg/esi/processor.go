package esi

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Config holds the ESI processor configuration
type Config struct {
	Mode        string      `json:"mode"`        // fastly, akamai, w3c, development
	Debug       bool        `json:"debug"`       // Enable debug logging
	MaxIncludes int         `json:"maxIncludes"` // Maximum number of includes per request
	MaxDepth    int         `json:"maxDepth"`    // Maximum include depth
	BaseURL     string      `json:"baseUrl"`     // Base URL for relative includes
	Cache       CacheConfig `json:"cache"`       // Cache configuration
}

// CacheConfig holds cache-related configuration
type CacheConfig struct {
	Enabled bool `json:"enabled"` // Whether caching is enabled
	TTL     int  `json:"ttl"`     // Time to live in seconds
}

// Features represents the supported ESI features for each mode
type Features struct {
	Include       bool `json:"include"`       // <esi:include>
	Comment       bool `json:"comment"`       // <esi:comment>
	Remove        bool `json:"remove"`        // <esi:remove>
	Inline        bool `json:"inline"`        // <esi:inline>
	Choose        bool `json:"choose"`        // <esi:choose>/<esi:when>/<esi:otherwise>
	Try           bool `json:"try"`           // <esi:try>/<esi:attempt>/<esi:except>
	Vars          bool `json:"vars"`          // <esi:vars>
	Variables     bool `json:"variables"`     // ESI variables $(...)
	Expressions   bool `json:"expressions"`   // ESI expressions
	CommentBlocks bool `json:"commentBlocks"` // <!--esi ...-->
	// Akamai-specific extensions
	Assign       bool `json:"assign"`       // <esi:assign> - Variable assignment
	Eval         bool `json:"eval"`         // <esi:eval> - Expression evaluation
	Function     bool `json:"function"`     // <esi:function> - Built-in functions
	Dictionary   bool `json:"dictionary"`   // <esi:dictionary> - Key-value lookups
	Debug        bool `json:"debug"`        // <esi:debug> - Debug output
	GeoVariables bool `json:"geoVariables"` // Geo-location variables
	ExtendedVars bool `json:"extendedVars"` // Extended variable set
}

// Stats holds processing statistics
type Stats struct {
	Requests  int64 `json:"requests"`
	CacheHits int64 `json:"cacheHits"`
	CacheMiss int64 `json:"cacheMiss"`
	Errors    int64 `json:"errors"`
	TotalTime int64 `json:"totalTime"` // Total processing time in milliseconds
	mutex     sync.RWMutex
}

// CacheEntry represents a cached fragment
type CacheEntry struct {
	Content   string    `json:"content"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// ProcessContext holds context for ESI processing
type ProcessContext struct {
	BaseURL string            `json:"baseUrl"`
	Headers map[string]string `json:"headers"`
	Cookies map[string]string `json:"cookies"`
	Depth   int               `json:"depth"`
}

// Processor is the main ESI processing engine
type Processor struct {
	config    Config
	features  Features
	stats     Stats
	cache     map[string]CacheEntry
	mutex     sync.RWMutex
	client    *http.Client
	akamaiExt *AkamaiExtensions // Akamai extensions handler
}

// NewProcessor creates a new ESI processor with the given configuration
func NewProcessor(config Config) *Processor {
	processor := &Processor{
		config: config,
		cache:  make(map[string]CacheEntry),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	processor.features = processor.getSupportedFeatures()
	processor.akamaiExt = NewAkamaiExtensions(processor) // Initialize Akamai extensions
	return processor
}

// getSupportedFeatures returns the features supported by the current mode
func (p *Processor) getSupportedFeatures() Features {
	base := Features{
		Include: true,
		Comment: true,
		Remove:  true,
	}

	switch p.config.Mode {
	case "fastly":
		return base
	case "akamai", "w3c", "development":
		return Features{
			Include:       true,
			Comment:       true,
			Remove:        true,
			Inline:        true,
			Choose:        true,
			Try:           true,
			Vars:          true,
			Variables:     true,
			Expressions:   true,
			CommentBlocks: true,
			Assign:        true,
			Eval:          true,
			Function:      true,
			Dictionary:    true,
			Debug:         true,
			GeoVariables:  true,
			ExtendedVars:  true,
		}
	default:
		return base
	}
}

// Process processes ESI content and returns the processed HTML
func (p *Processor) Process(html string, context ProcessContext) (string, error) {
	startTime := time.Now()

	p.stats.mutex.Lock()
	p.stats.Requests++
	p.stats.mutex.Unlock()

	if p.config.Debug {
		fmt.Printf("🔄 Processing ESI content (mode: %s): %s...\n",
			p.config.Mode, truncateString(html, 100))
	}

	// Check depth limit
	if context.Depth > p.config.MaxDepth {
		return html, fmt.Errorf("maximum include depth exceeded: %d", p.config.MaxDepth)
	}

	// Process ESI comment blocks first (<!--esi ...-->)
	if p.features.CommentBlocks {
		html = p.processCommentBlocks(html, context)
	}

	// Parse HTML with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		p.incrementErrors()
		return html, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Process ESI elements
	if err := p.processESIElements(doc, context); err != nil {
		p.incrementErrors()
		return html, err
	}

	// Get the processed HTML
	result, err := doc.Html()
	if err != nil {
		p.incrementErrors()
		return html, fmt.Errorf("failed to generate HTML: %w", err)
	}

	// Final variable expansion for Akamai mode
	if (p.config.Mode == "akamai" || p.config.Mode == "development") && p.akamaiExt != nil {
		result = p.akamaiExt.expandVariables(result, context)
	}

	// Update statistics
	processingTime := time.Since(startTime).Milliseconds()
	p.stats.mutex.Lock()
	p.stats.TotalTime += processingTime
	p.stats.mutex.Unlock()

	if p.config.Debug {
		fmt.Printf("✅ Processing completed in %dms\n", processingTime)
	}

	return result, nil
}

// processCommentBlocks processes <!--esi ...--> blocks
func (p *Processor) processCommentBlocks(html string, _ ProcessContext) string {
	// Regex to match <!--esi ... --> blocks
	re := regexp.MustCompile(`<!--esi\s+(.*?)\s+-->`)

	return re.ReplaceAllStringFunc(html, func(match string) string {
		// Extract the ESI content from the comment
		re := regexp.MustCompile(`<!--esi\s+(.*?)\s+-->`)
		matches := re.FindStringSubmatch(match)
		if len(matches) > 1 {
			return matches[1] // Return the ESI content without comment wrapper
		}
		return match
	})
}

// processESIElements processes all ESI elements in the document
func (p *Processor) processESIElements(doc *goquery.Document, context ProcessContext) error {
	// Process Akamai-specific extensions first if in Akamai mode
	if p.config.Mode == "akamai" || p.config.Mode == "development" {
		if err := p.akamaiExt.ProcessAkamaiExtensions(doc, context); err != nil {
			return err
		}
	}

	// Process different ESI elements based on supported features
	if p.features.Include {
		if err := p.processIncludes(doc, context); err != nil {
			return err
		}
	}

	if p.features.Choose {
		if err := p.processChoose(doc, context); err != nil {
			return err
		}
	}

	if p.features.Try {
		if err := p.processTry(doc, context); err != nil {
			return err
		}
	}

	if p.features.Vars {
		if err := p.processVars(doc, context); err != nil {
			return err
		}
	}

	if p.features.Comment {
		p.processComments(doc)
	}

	if p.features.Remove {
		p.processRemove(doc)
	}

	return nil
}

// processIncludes handles esi:include elements
func (p *Processor) processIncludes(doc *goquery.Document, context ProcessContext) error {
	var includeCount int

	doc.Find("esi\\:include, include").Each(func(i int, s *goquery.Selection) {
		includeCount++
		if includeCount > p.config.MaxIncludes {
			if p.config.Debug {
				fmt.Printf("⚠️  Maximum includes exceeded: %d\n", p.config.MaxIncludes)
			}
			return
		}

		src, exists := s.Attr("src")
		if !exists || src == "" {
			if p.config.Debug {
				fmt.Println("⚠️  esi:include missing src attribute")
			}
			s.Remove()
			return
		}

		alt, _ := s.Attr("alt")
		onerror, _ := s.Attr("onerror")

		// Try to fetch the content
		content, err := p.fetchInclude(src, context)
		if err != nil {
			if p.config.Debug {
				fmt.Printf("⚠️  Include failed for %s: %v\n", src, err)
			}

			// Try alt URL if available
			if alt != "" && p.features.Include {
				if altContent, altErr := p.fetchInclude(alt, context); altErr == nil {
					s.ReplaceWithHtml(altContent)
					return
				} else if p.config.Debug {
					fmt.Printf("⚠️  Alt include failed for %s: %v\n", alt, altErr)
				}
			}

			// Handle onerror="continue"
			if onerror == "continue" {
				s.Remove()
			} else {
				if p.config.Debug {
					s.ReplaceWithHtml(fmt.Sprintf("<!-- ESI include error: %v -->", err))
				} else {
					s.Remove()
				}
			}
			return
		}

		// Replace with fetched content
		s.ReplaceWithHtml(content)
	})

	return nil
}

// fetchInclude fetches content for an ESI include
func (p *Processor) fetchInclude(src string, context ProcessContext) (string, error) {
	// Resolve relative URLs
	resolvedURL, err := p.resolveURL(src, context.BaseURL)
	if err != nil {
		return "", fmt.Errorf("failed to resolve URL %s: %w", src, err)
	}

	// Check cache first
	if p.config.Cache.Enabled {
		p.mutex.RLock()
		if entry, exists := p.cache[resolvedURL]; exists && time.Now().Before(entry.ExpiresAt) {
			p.mutex.RUnlock()
			p.incrementCacheHits()
			return entry.Content, nil
		}
		p.mutex.RUnlock()
	}

	p.incrementCacheMiss()

	// Create HTTP request
	req, err := http.NewRequest("GET", resolvedURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers from context
	for key, value := range context.Headers {
		req.Header.Set(key, value)
	}

	// Perform request
	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch %s: %w", resolvedURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	content := string(body)

	// Cache the result
	if p.config.Cache.Enabled {
		p.mutex.Lock()
		p.cache[resolvedURL] = CacheEntry{
			Content:   content,
			ExpiresAt: time.Now().Add(time.Duration(p.config.Cache.TTL) * time.Second),
		}
		p.mutex.Unlock()
	}

	return content, nil
}

// processChoose handles esi:choose/when/otherwise elements for conditional processing
func (p *Processor) processChoose(doc *goquery.Document, context ProcessContext) error {
	if p.config.Debug {
		fmt.Println("🔍 Processing esi:choose elements")
	}

	doc.Find("esi\\:choose, choose").Each(func(i int, chooseSelection *goquery.Selection) {
		// Find all esi:when elements within this choose block
		whenElements := chooseSelection.Find("esi\\:when, when")
		otherwiseElement := chooseSelection.Find("esi\\:otherwise, otherwise").First()

		var selectedContent string
		var foundMatch bool

		// Evaluate each when condition
		whenElements.Each(func(j int, whenSelection *goquery.Selection) {
			if foundMatch {
				return // Skip if we already found a match
			}

			// Get the test attribute
			test, exists := whenSelection.Attr("test")
			if !exists || test == "" {
				if p.config.Debug {
					fmt.Println("⚠️  esi:when missing test attribute")
				}
				return
			}

			// Evaluate the test expression
			result := p.evaluateExpression(test, context)
			if result == "true" {
				// Get the content of this when block
				content, err := whenSelection.Html()
				if err != nil {
					if p.config.Debug {
						fmt.Printf("⚠️  Failed to get esi:when content: %v\n", err)
					}
					return
				}

				selectedContent = content
				foundMatch = true

				if p.config.Debug {
					fmt.Printf("✅ esi:when condition '%s' matched\n", test)
				}
			}
		})

		// If no when condition matched, use otherwise
		if !foundMatch && otherwiseElement.Length() > 0 {
			content, err := otherwiseElement.Html()
			if err != nil {
				if p.config.Debug {
					fmt.Printf("⚠️  Failed to get esi:otherwise content: %v\n", err)
				}
			} else {
				selectedContent = content
				if p.config.Debug {
					fmt.Println("✅ Using esi:otherwise content")
				}
			}
		}

		// Replace the entire choose block with the selected content
		if selectedContent != "" {
			chooseSelection.ReplaceWithHtml(selectedContent)
		} else {
			// Remove the choose block if no content was selected
			chooseSelection.Remove()
		}

		if p.config.Debug {
			fmt.Printf("📝 Processed esi:choose block: %s\n", truncateString(selectedContent, 50))
		}
	})

	return nil
}

// processTry handles esi:try/attempt/except elements for error handling
func (p *Processor) processTry(doc *goquery.Document, context ProcessContext) error {
	if p.config.Debug {
		fmt.Println("🔍 Processing esi:try elements")
	}

	doc.Find("esi\\:try, try").Each(func(i int, trySelection *goquery.Selection) {
		// Find attempt and except elements
		attemptElement := trySelection.Find("esi\\:attempt, attempt").First()
		exceptElement := trySelection.Find("esi\\:except, except").First()

		var finalContent string
		var processingError error

		// Try to process the attempt block
		if attemptElement.Length() > 0 {
			content, err := attemptElement.Html()
			if err != nil {
				if p.config.Debug {
					fmt.Printf("⚠️  Failed to get esi:attempt content: %v\n", err)
				}
				processingError = err
			} else {
				// Create a temporary processor to process the attempt content
				// This allows us to catch errors from includes, vars, etc.
				tempProcessor := NewProcessor(p.config)

				// Process the attempt content
				processedContent, err := tempProcessor.Process(content, context)
				if err != nil {
					if p.config.Debug {
						fmt.Printf("⚠️  Error processing esi:attempt content: %v\n", err)
					}
					processingError = err
				} else {
					// Check if the processed content contains error indicators
					if strings.Contains(processedContent, "ESI include error") ||
						strings.Contains(processedContent, "failed to fetch") ||
						strings.Contains(processedContent, "HTTP 4") ||
						strings.Contains(processedContent, "HTTP 5") {
						processingError = fmt.Errorf("include processing failed")
					} else {
						finalContent = processedContent
						if p.config.Debug {
							fmt.Println("✅ esi:attempt content processed successfully")
						}
					}
				}
			}
		}

		// If there was an error and we have an except block, use it
		if processingError != nil && exceptElement.Length() > 0 {
			content, err := exceptElement.Html()
			if err != nil {
				if p.config.Debug {
					fmt.Printf("⚠️  Failed to get esi:except content: %v\n", err)
				}
			} else {
				// Process the except content
				processedContent, err := p.Process(content, context)
				if err != nil {
					if p.config.Debug {
						fmt.Printf("⚠️  Error processing esi:except content: %v\n", err)
					}
				} else {
					finalContent = processedContent
					if p.config.Debug {
						fmt.Println("✅ Using esi:except content due to error")
					}
				}
			}
		}

		// Replace the entire try block with the final content
		if finalContent != "" {
			trySelection.ReplaceWithHtml(finalContent)
		} else {
			// Remove the try block if no content was processed
			trySelection.Remove()
		}

		if p.config.Debug {
			fmt.Printf("📝 Processed esi:try block: %s\n", truncateString(finalContent, 50))
		}
	})

	return nil
}

// processVars handles esi:vars elements for variable substitution
func (p *Processor) processVars(doc *goquery.Document, context ProcessContext) error {
	if p.config.Debug {
		fmt.Println("🔍 Processing esi:vars elements")
	}

	doc.Find("esi\\:vars, vars").Each(func(i int, s *goquery.Selection) {
		// Get the content inside the esi:vars element
		content, err := s.Html()
		if err != nil {
			if p.config.Debug {
				fmt.Printf("⚠️  Failed to get esi:vars content: %v\n", err)
			}
			s.Remove()
			return
		}

		// Expand variables in the content
		expandedContent := p.ExpandESIVariables(content, context)

		// Replace the esi:vars element with the expanded content
		s.ReplaceWithHtml(expandedContent)

		if p.config.Debug {
			fmt.Printf("📝 Processed esi:vars: %s -> %s\n",
				truncateString(content, 50), truncateString(expandedContent, 50))
		}
	})

	return nil
}

// ExpandESIVariables expands ESI variables in content with support for default values
func (p *Processor) ExpandESIVariables(input string, context ProcessContext) string {
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

		// Get variable value
		value := p.GetESIVariable(varName, key, context)

		// Return default value if variable is empty and default is specified
		if value == "" && defaultValue != "" {
			return defaultValue
		}

		return value
	})
}

// GetESIVariable returns the value of a standard ESI variable
func (p *Processor) GetESIVariable(varName, key string, context ProcessContext) string {
	switch varName {
	case "HTTP_HOST":
		if host, exists := context.Headers["Host"]; exists {
			return host
		}
		return ""

	case "HTTP_USER_AGENT":
		if key != "" {
			return p.getUserAgentComponent(context.Headers["User-Agent"], key)
		}
		if ua, exists := context.Headers["User-Agent"]; exists {
			return ua
		}
		return ""

	case "HTTP_COOKIE":
		if key != "" {
			return p.getCookieValue(context.Cookies, key)
		}
		if cookie, exists := context.Headers["Cookie"]; exists {
			return cookie
		}
		return ""

	case "HTTP_REFERER":
		if referer, exists := context.Headers["Referer"]; exists {
			return referer
		}
		return ""

	case "HTTP_ACCEPT_LANGUAGE":
		if key != "" {
			return p.hasLanguage(context.Headers["Accept-Language"], key)
		}
		if lang, exists := context.Headers["Accept-Language"]; exists {
			return lang
		}
		return ""

	case "QUERY_STRING":
		if key != "" {
			return p.getQueryParam(context.Headers["Query-String"], key)
		}
		if qs, exists := context.Headers["Query-String"]; exists {
			return qs
		}
		return ""

	case "REQUEST_METHOD":
		if method, exists := context.Headers["Method"]; exists {
			return method
		}
		return "GET"

	case "REQUEST_URI":
		if uri, exists := context.Headers["Request-URI"]; exists {
			return uri
		}
		return ""

	default:
		// Delegate to Akamai extensions for non-standard variables in Akamai/development mode
		if (p.config.Mode == "akamai" || p.config.Mode == "development") && p.akamaiExt != nil {
			return p.akamaiExt.getESIVariable(varName, key, context)
		}
		if p.config.Debug {
			fmt.Printf("⚠️  Unknown ESI variable: %s\n", varName)
		}
		return ""
	}
}

// processComments removes esi:comment elements
func (p *Processor) processComments(doc *goquery.Document) {
	doc.Find("esi\\:comment, comment").Remove()
}

// processRemove removes esi:remove elements
func (p *Processor) processRemove(doc *goquery.Document) {
	doc.Find("esi\\:remove, remove").Remove()
}

// resolveURL resolves a relative URL against a base URL
func (p *Processor) resolveURL(urlStr, baseURL string) (string, error) {
	if urlStr == "" {
		return "", fmt.Errorf("empty URL")
	}

	// If already absolute, return as-is
	if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") {
		return urlStr, nil
	}

	// Use base URL from context or config
	base := baseURL
	if base == "" {
		base = p.config.BaseURL
	}
	if base == "" {
		return urlStr, nil // Return relative URL as-is
	}

	baseURL_parsed, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	resolved, err := baseURL_parsed.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("failed to resolve URL: %w", err)
	}

	return resolved.String(), nil
}

// GetStats returns current processing statistics
func (p *Processor) GetStats() Stats {
	p.stats.mutex.RLock()
	defer p.stats.mutex.RUnlock()

	// Return a copy without the mutex to avoid copy lock error
	return Stats{
		Requests:  p.stats.Requests,
		CacheHits: p.stats.CacheHits,
		CacheMiss: p.stats.CacheMiss,
		Errors:    p.stats.Errors,
		TotalTime: p.stats.TotalTime,
		// Note: mutex is not copied
	}
}

// GetFeatures returns supported features for the current mode
func (p *Processor) GetFeatures() Features {
	return p.features
}

// ClearCache clears the fragment cache
func (p *Processor) ClearCache() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.cache = make(map[string]CacheEntry)
}

// GetCacheSize returns the current number of cached items
func (p *Processor) GetCacheSize() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return len(p.cache)
}

// GetConfig returns the processor configuration (implements ProcessorInterface)
func (p *Processor) GetConfig() Config {
	return p.config
}

// Helper methods for statistics
func (p *Processor) incrementCacheHits() {
	p.stats.mutex.Lock()
	defer p.stats.mutex.Unlock()
	p.stats.CacheHits++
}

func (p *Processor) incrementCacheMiss() {
	p.stats.mutex.Lock()
	defer p.stats.mutex.Unlock()
	p.stats.CacheMiss++
}

func (p *Processor) incrementErrors() {
	p.stats.mutex.Lock()
	defer p.stats.mutex.Unlock()
	p.stats.Errors++
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Helper functions for ESI variable processing

// getUserAgentComponent extracts specific components from User-Agent header
func (p *Processor) getUserAgentComponent(userAgent, component string) string {
	if userAgent == "" {
		return ""
	}

	switch component {
	case "browser":
		if strings.Contains(userAgent, "Chrome") {
			return "CHROME"
		} else if strings.Contains(userAgent, "Firefox") {
			return "FIREFOX"
		} else if strings.Contains(userAgent, "Safari") && !strings.Contains(userAgent, "Chrome") {
			return "SAFARI"
		} else if strings.Contains(userAgent, "Edge") {
			return "EDGE"
		} else if strings.Contains(userAgent, "MSIE") || strings.Contains(userAgent, "Trident") {
			return "MSIE"
		} else if strings.Contains(userAgent, "Mozilla") {
			return "MOZILLA"
		}
		return "OTHER"

	case "os":
		if strings.Contains(userAgent, "Windows") {
			return "WIN"
		} else if strings.Contains(userAgent, "Mac") {
			return "MAC"
		} else if strings.Contains(userAgent, "Linux") || strings.Contains(userAgent, "Unix") {
			return "UNIX"
		}
		return "OTHER"

	case "version":
		// Basic version extraction - could be enhanced
		if strings.Contains(userAgent, "Chrome/") {
			if parts := strings.Split(userAgent, "Chrome/"); len(parts) > 1 {
				if version := strings.Split(parts[1], " ")[0]; version != "" {
					return strings.Split(version, ".")[0]
				}
			}
		} else if strings.Contains(userAgent, "Firefox/") {
			if parts := strings.Split(userAgent, "Firefox/"); len(parts) > 1 {
				if version := strings.Split(parts[1], " ")[0]; version != "" {
					return strings.Split(version, ".")[0]
				}
			}
		} else if strings.Contains(userAgent, "Safari/") && !strings.Contains(userAgent, "Chrome") {
			if parts := strings.Split(userAgent, "Version/"); len(parts) > 1 {
				if version := strings.Split(parts[1], " ")[0]; version != "" {
					return strings.Split(version, ".")[0]
				}
			}
		}
		return "1.0" // Default fallback

	default:
		return ""
	}
}

// getCookieValue extracts a specific cookie value
func (p *Processor) getCookieValue(cookies map[string]string, key string) string {
	if val, exists := cookies[key]; exists {
		return val
	}
	return ""
}

// hasLanguage checks if a language is present in Accept-Language header (returns boolean as string)
func (p *Processor) hasLanguage(acceptLang, lang string) string {
	if acceptLang == "" {
		return "false"
	}

	// Parse Accept-Language header properly
	langs := strings.Split(acceptLang, ",")
	for _, l := range langs {
		// Clean up whitespace and quality values
		cleanLang := strings.TrimSpace(strings.Split(l, ";")[0])
		if strings.EqualFold(cleanLang, lang) || strings.HasPrefix(strings.ToLower(cleanLang), strings.ToLower(lang)) {
			return "true"
		}
	}
	return "false"
}

// getQueryParam extracts a query parameter value
func (p *Processor) getQueryParam(queryString, key string) string {
	if queryString == "" {
		return ""
	}

	values, err := url.ParseQuery(queryString)
	if err != nil {
		return ""
	}

	return values.Get(key)
}

// evaluateExpression evaluates a simple ESI expression
func (p *Processor) evaluateExpression(expr string, context ProcessContext) string {
	// Expand variables first
	expanded := p.ExpandESIVariables(expr, context)

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

	// Check for simple boolean values
	if expanded == "true" || expanded == "1" {
		return "true"
	}
	if expanded == "false" || expanded == "0" || expanded == "" {
		return "false"
	}

	// If it's not empty, consider it true
	if expanded != "" {
		return "true"
	}

	return "false"
}
