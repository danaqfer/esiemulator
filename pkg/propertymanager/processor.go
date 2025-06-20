package propertymanager

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// processRules processes a list of rules recursively
func (pm *PropertyManager) processRules(rules []Rule, context *HTTPContext, result *RuleResult) error {
	for _, rule := range rules {
		if pm.evaluateRule(&rule, context) {
			if pm.Debug {
				fmt.Printf("🔍 Rule matched: %s\n", rule.Name)
			}

			result.MatchedRules = append(result.MatchedRules, rule.Name)

			// Execute behaviors for this rule
			if err := pm.executeBehaviors(rule.Behaviors, context, result); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Error executing behaviors for rule %s: %v", rule.Name, err))
			}

			// Process child rules
			if len(rule.Children) > 0 {
				if err := pm.processRules(rule.Children, context, result); err != nil {
					result.Errors = append(result.Errors, err.Error())
				}
			}
		}
	}
	return nil
}

// evaluateRule evaluates whether a rule should be executed based on its criteria
func (pm *PropertyManager) evaluateRule(rule *Rule, context *HTTPContext) bool {
	if len(rule.Criteria) == 0 {
		return true // No criteria means always match
	}

	// All criteria must match (AND logic)
	for _, criterion := range rule.Criteria {
		if !pm.evaluateCriterion(&criterion, context) {
			return false
		}
	}

	return true
}

// evaluateCriterion evaluates a single criterion
func (pm *PropertyManager) evaluateCriterion(criterion *Criterion, context *HTTPContext) bool {
	switch criterion.Name {
	case "path":
		return pm.evaluatePathCriterion(criterion, context)
	case "header":
		return pm.evaluateHeaderCriterion(criterion, context)
	case "method":
		return pm.evaluateMethodCriterion(criterion, context)
	case "host":
		return pm.evaluateHostCriterion(criterion, context)
	case "query":
		return pm.evaluateQueryCriterion(criterion, context)
	case "cookie":
		return pm.evaluateCookieCriterion(criterion, context)
	case "variable":
		return pm.evaluateVariableCriterion(criterion, context)
	case "client_ip":
		return pm.evaluateClientIPCriterion(criterion, context)
	case "user_agent":
		return pm.evaluateUserAgentCriterion(criterion, context)
	case "geo_country_code":
		return pm.evaluateGeoCountryCodeCriterion(criterion, context)
	case "geo_country_name":
		return pm.evaluateGeoCountryNameCriterion(criterion, context)
	case "geo_region":
		return pm.evaluateGeoRegionCriterion(criterion, context)
	case "geo_city":
		return pm.evaluateGeoCityCriterion(criterion, context)
	default:
		if pm.Debug {
			fmt.Printf("⚠️  Unknown criterion type: %s\n", criterion.Name)
		}
		return false
	}
}

// evaluatePathCriterion evaluates path-based criteria
func (pm *PropertyManager) evaluatePathCriterion(criterion *Criterion, context *HTTPContext) bool {
	path := context.Path
	value := criterion.Value

	switch criterion.Option {
	case "equals":
		return path == value
	case "not_equals":
		return path != value
	case "starts_with":
		return strings.HasPrefix(path, value)
	case "ends_with":
		return strings.HasSuffix(path, value)
	case "contains":
		return strings.Contains(path, value)
	case "regex":
		matched, _ := regexp.MatchString(value, path)
		return matched
	default:
		return path == value // Default to equals
	}
}

// evaluateHeaderCriterion evaluates header-based criteria
func (pm *PropertyManager) evaluateHeaderCriterion(criterion *Criterion, context *HTTPContext) bool {
	headerValue, exists := context.Headers[criterion.Option]
	if !exists {
		return false
	}

	value := criterion.Value
	if !criterion.Case {
		headerValue = strings.ToLower(headerValue)
		value = strings.ToLower(value)
	}

	switch criterion.Extract {
	case "equals":
		return headerValue == value
	case "not_equals":
		return headerValue != value
	case "starts_with":
		return strings.HasPrefix(headerValue, value)
	case "ends_with":
		return strings.HasSuffix(headerValue, value)
	case "contains":
		return strings.Contains(headerValue, value)
	case "regex":
		matched, _ := regexp.MatchString(value, headerValue)
		return matched
	default:
		return headerValue == value // Default to equals
	}
}

// evaluateMethodCriterion evaluates HTTP method criteria
func (pm *PropertyManager) evaluateMethodCriterion(criterion *Criterion, context *HTTPContext) bool {
	method := strings.ToUpper(context.Method)
	value := strings.ToUpper(criterion.Value)

	switch criterion.Option {
	case "equals":
		return method == value
	case "not_equals":
		return method != value
	default:
		return method == value
	}
}

// evaluateHostCriterion evaluates host-based criteria
func (pm *PropertyManager) evaluateHostCriterion(criterion *Criterion, context *HTTPContext) bool {
	host := context.Host
	value := criterion.Value

	if !criterion.Case {
		host = strings.ToLower(host)
		value = strings.ToLower(value)
	}

	switch criterion.Option {
	case "equals":
		return host == value
	case "not_equals":
		return host != value
	case "starts_with":
		return strings.HasPrefix(host, value)
	case "ends_with":
		return strings.HasSuffix(host, value)
	case "contains":
		return strings.Contains(host, value)
	default:
		return host == value
	}
}

// evaluateQueryCriterion evaluates query string criteria
func (pm *PropertyManager) evaluateQueryCriterion(criterion *Criterion, context *HTTPContext) bool {
	query := context.Query
	value := criterion.Value

	switch criterion.Option {
	case "equals":
		return query == value
	case "not_equals":
		return query != value
	case "contains":
		return strings.Contains(query, value)
	case "regex":
		matched, _ := regexp.MatchString(value, query)
		return matched
	default:
		return query == value
	}
}

// evaluateCookieCriterion evaluates cookie-based criteria
func (pm *PropertyManager) evaluateCookieCriterion(criterion *Criterion, context *HTTPContext) bool {
	cookieValue, exists := context.Cookies[criterion.Option]
	if !exists {
		return false
	}

	value := criterion.Value
	if !criterion.Case {
		cookieValue = strings.ToLower(cookieValue)
		value = strings.ToLower(value)
	}

	switch criterion.Extract {
	case "equals":
		return cookieValue == value
	case "not_equals":
		return cookieValue != value
	case "starts_with":
		return strings.HasPrefix(cookieValue, value)
	case "ends_with":
		return strings.HasSuffix(cookieValue, value)
	case "contains":
		return strings.Contains(cookieValue, value)
	default:
		return cookieValue == value
	}
}

// evaluateVariableCriterion evaluates variable-based criteria
func (pm *PropertyManager) evaluateVariableCriterion(criterion *Criterion, context *HTTPContext) bool {
	varValue, exists := context.Variables[criterion.Option]
	if !exists {
		return false
	}

	value := criterion.Value
	if !criterion.Case {
		varValue = strings.ToLower(varValue)
		value = strings.ToLower(value)
	}

	switch criterion.Extract {
	case "equals":
		return varValue == value
	case "not_equals":
		return varValue != value
	case "starts_with":
		return strings.HasPrefix(varValue, value)
	case "ends_with":
		return strings.HasSuffix(varValue, value)
	case "contains":
		return strings.Contains(varValue, value)
	default:
		return varValue == value
	}
}

// evaluateClientIPCriterion evaluates client IP criteria
func (pm *PropertyManager) evaluateClientIPCriterion(criterion *Criterion, context *HTTPContext) bool {
	clientIP := context.ClientIP
	value := criterion.Value

	switch criterion.Option {
	case "equals":
		return clientIP == value
	case "not_equals":
		return clientIP != value
	case "starts_with":
		return strings.HasPrefix(clientIP, value)
	case "ends_with":
		return strings.HasSuffix(clientIP, value)
	case "contains":
		return strings.Contains(clientIP, value)
	case "in":
		return pm.isIPInCIDR(clientIP, value)
	case "not_in":
		return !pm.isIPInCIDR(clientIP, value)
	case "regex":
		matched, _ := regexp.MatchString(value, clientIP)
		return matched
	default:
		return clientIP == value
	}
}

// isIPInCIDR checks if an IP is in a CIDR range
func (pm *PropertyManager) isIPInCIDR(ip, cidr string) bool {
	// Simple CIDR check implementation
	// For production, consider using a proper IP parsing library
	if strings.Contains(cidr, "/") {
		parts := strings.Split(cidr, "/")
		if len(parts) != 2 {
			return false
		}
		network := parts[0]
		// For simplicity, we'll do a basic prefix check
		// In a real implementation, you'd want proper IP/CIDR parsing
		return strings.HasPrefix(ip, network[:strings.LastIndex(network, ".")+1])
	}
	return ip == cidr
}

// evaluateGeoCountryCodeCriterion evaluates geo country code criteria
func (pm *PropertyManager) evaluateGeoCountryCodeCriterion(criterion *Criterion, context *HTTPContext) bool {
	// In a real implementation, this would use geo-location data
	// For now, we'll use a mock implementation
	geoCountryCode := context.Variables["GEO_COUNTRY_CODE"]
	if geoCountryCode == "" {
		geoCountryCode = "US" // Default for testing
	}

	value := criterion.Value
	if !criterion.Case {
		geoCountryCode = strings.ToLower(geoCountryCode)
		value = strings.ToLower(value)
	}

	switch criterion.Option {
	case "equals":
		return geoCountryCode == value
	case "not_equals":
		return geoCountryCode != value
	case "in":
		values := strings.Split(value, ",")
		for _, v := range values {
			if strings.TrimSpace(v) == geoCountryCode {
				return true
			}
		}
		return false
	case "not_in":
		values := strings.Split(value, ",")
		for _, v := range values {
			if strings.TrimSpace(v) == geoCountryCode {
				return false
			}
		}
		return true
	default:
		return geoCountryCode == value
	}
}

// evaluateGeoCountryNameCriterion evaluates geo country name criteria
func (pm *PropertyManager) evaluateGeoCountryNameCriterion(criterion *Criterion, context *HTTPContext) bool {
	geoCountryName := context.Variables["GEO_COUNTRY_NAME"]
	if geoCountryName == "" {
		geoCountryName = "United States" // Default for testing
	}

	value := criterion.Value
	if !criterion.Case {
		geoCountryName = strings.ToLower(geoCountryName)
		value = strings.ToLower(value)
	}

	switch criterion.Option {
	case "equals":
		return geoCountryName == value
	case "not_equals":
		return geoCountryName != value
	case "contains":
		return strings.Contains(geoCountryName, value)
	case "in":
		values := strings.Split(value, ",")
		for _, v := range values {
			if strings.Contains(geoCountryName, strings.TrimSpace(v)) {
				return true
			}
		}
		return false
	default:
		return geoCountryName == value
	}
}

// evaluateGeoRegionCriterion evaluates geo region criteria
func (pm *PropertyManager) evaluateGeoRegionCriterion(criterion *Criterion, context *HTTPContext) bool {
	geoRegion := context.Variables["GEO_REGION"]
	if geoRegion == "" {
		geoRegion = "California" // Default for testing
	}

	value := criterion.Value
	if !criterion.Case {
		geoRegion = strings.ToLower(geoRegion)
		value = strings.ToLower(value)
	}

	switch criterion.Option {
	case "equals":
		return geoRegion == value
	case "not_equals":
		return geoRegion != value
	case "contains":
		return strings.Contains(geoRegion, value)
	case "in":
		values := strings.Split(value, ",")
		for _, v := range values {
			if strings.Contains(geoRegion, strings.TrimSpace(v)) {
				return true
			}
		}
		return false
	default:
		return geoRegion == value
	}
}

// evaluateGeoCityCriterion evaluates geo city criteria
func (pm *PropertyManager) evaluateGeoCityCriterion(criterion *Criterion, context *HTTPContext) bool {
	geoCity := context.Variables["GEO_CITY"]
	if geoCity == "" {
		geoCity = "San Francisco" // Default for testing
	}

	value := criterion.Value
	if !criterion.Case {
		geoCity = strings.ToLower(geoCity)
		value = strings.ToLower(value)
	}

	switch criterion.Option {
	case "equals":
		return geoCity == value
	case "not_equals":
		return geoCity != value
	case "contains":
		return strings.Contains(geoCity, value)
	case "in":
		values := strings.Split(value, ",")
		for _, v := range values {
			if strings.Contains(geoCity, strings.TrimSpace(v)) {
				return true
			}
		}
		return false
	default:
		return geoCity == value
	}
}

// evaluateUserAgentCriterion evaluates user agent criteria
func (pm *PropertyManager) evaluateUserAgentCriterion(criterion *Criterion, context *HTTPContext) bool {
	userAgent := context.UserAgent
	value := criterion.Value

	if !criterion.Case {
		userAgent = strings.ToLower(userAgent)
		value = strings.ToLower(value)
	}

	switch criterion.Option {
	case "equals":
		return userAgent == value
	case "not_equals":
		return userAgent != value
	case "starts_with":
		return strings.HasPrefix(userAgent, value)
	case "ends_with":
		return strings.HasSuffix(userAgent, value)
	case "contains":
		return strings.Contains(userAgent, value)
	case "regex":
		matched, _ := regexp.MatchString(value, userAgent)
		return matched
	default:
		return userAgent == value
	}
}

// executeBehaviors executes a list of behaviors
func (pm *PropertyManager) executeBehaviors(behaviors []Behavior, context *HTTPContext, result *RuleResult) error {
	for _, behavior := range behaviors {
		if err := pm.executeBehavior(&behavior, context, result); err != nil {
			return err
		}
	}
	return nil
}

// executeBehavior executes a single behavior
func (pm *PropertyManager) executeBehavior(behavior *Behavior, context *HTTPContext, result *RuleResult) error {
	if pm.Debug {
		fmt.Printf("🔧 Executing behavior: %s\n", behavior.Name)
	}

	result.ExecutedBehaviors = append(result.ExecutedBehaviors, behavior.Name)

	switch behavior.Name {
	// Caching behaviors
	case "cache":
		return pm.executeCache(behavior, context, result)
	case "cache_bypass":
		return pm.executeCacheBypass(behavior, context, result)

	// Security behaviors
	case "access_control":
		return pm.executeAccessControl(behavior, context, result)
	case "rate_limit":
		return pm.executeRateLimit(behavior, context, result)

	// Performance behaviors
	case "compress":
		return pm.executeCompression(behavior, context, result)
	case "image_optimization":
		return pm.executeImageOptimization(behavior, context, result)

	// Content behaviors
	case "modify_headers":
		return pm.executeModifyHeaders(behavior, context, result)
	case "url_rewrite":
		return pm.executeURLRewrite(behavior, context, result)

	// Redirect behaviors
	case "redirect":
		return pm.executeRedirect(behavior, context, result)
	case "conditional_redirect":
		return pm.executeConditionalRedirect(behavior, context, result)

	// Legacy behaviors (for backward compatibility)
	case "set_response_header":
		return pm.executeSetResponseHeader(behavior, context, result)
	case "set_request_header":
		return pm.executeSetRequestHeader(behavior, context, result)
	case "set_variable":
		return pm.executeSetVariable(behavior, context, result)
	case "cache_key_query_params":
		return pm.executeCacheKeyQueryParams(behavior, context, result)
	case "origin_error_pass_thru":
		return pm.executeOriginErrorPassThru(behavior, context, result)
	case "esi":
		return pm.executeESI(behavior, context, result)
	case "gzip_response":
		return pm.executeGzipResponse(behavior, context, result)
	case "edge_redirector":
		return pm.executeEdgeRedirector(behavior, context, result)
	case "origin":
		return pm.executeOrigin(behavior, context, result)

	default:
		if pm.Debug {
			fmt.Printf("⚠️  Unknown behavior: %s\n", behavior.Name)
		}
		return nil
	}
}

// executeCache executes cache behavior
func (pm *PropertyManager) executeCache(behavior *Behavior, context *HTTPContext, result *RuleResult) error {
	if pm.Debug {
		fmt.Printf("🔧 Cache behavior: %+v\n", behavior.Options)
	}

	// Store cache settings in result for later use
	if result.CacheSettings == nil {
		result.CacheSettings = make(map[string]interface{})
	}

	for key, value := range behavior.Options {
		result.CacheSettings[key] = value
	}

	return nil
}

// executeCacheBypass executes cache bypass behavior
func (pm *PropertyManager) executeCacheBypass(behavior *Behavior, context *HTTPContext, result *RuleResult) error {
	if pm.Debug {
		fmt.Printf("🔧 Cache bypass behavior: %+v\n", behavior.Options)
	}

	// Mark as cache bypass
	if result.CacheSettings == nil {
		result.CacheSettings = make(map[string]interface{})
	}
	result.CacheSettings["bypass"] = true

	if reason, ok := behavior.Options["reason"].(string); ok {
		result.CacheSettings["bypass_reason"] = reason
	}

	return nil
}

// executeAccessControl executes access control behavior
func (pm *PropertyManager) executeAccessControl(behavior *Behavior, context *HTTPContext, result *RuleResult) error {
	if pm.Debug {
		fmt.Printf("🔧 Access control behavior: %+v\n", behavior.Options)
	}

	// Check allowed IPs
	if allowedIPs, ok := behavior.Options["allowed_ips"].(string); ok {
		ips := strings.Split(allowedIPs, ",")
		allowed := false
		for _, ip := range ips {
			if pm.isIPInCIDR(context.ClientIP, strings.TrimSpace(ip)) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("access denied: IP %s not in allowed list", context.ClientIP)
		}
	}

	// Check blocked IPs
	if blockedIPs, ok := behavior.Options["blocked_ips"].(string); ok {
		ips := strings.Split(blockedIPs, ",")
		for _, ip := range ips {
			if pm.isIPInCIDR(context.ClientIP, strings.TrimSpace(ip)) {
				return fmt.Errorf("access denied: IP %s is blocked", context.ClientIP)
			}
		}
	}

	// Check allowed countries
	if allowedCountries, ok := behavior.Options["allowed_countries"].(string); ok {
		countryCode := context.Variables["GEO_COUNTRY_CODE"]
		if countryCode == "" {
			countryCode = "US" // Default
		}
		countries := strings.Split(allowedCountries, ",")
		allowed := false
		for _, country := range countries {
			if strings.TrimSpace(country) == countryCode {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("access denied: country %s not allowed", countryCode)
		}
	}

	// Check blocked countries
	if blockedCountries, ok := behavior.Options["blocked_countries"].(string); ok {
		countryCode := context.Variables["GEO_COUNTRY_CODE"]
		if countryCode == "" {
			countryCode = "US" // Default
		}
		countries := strings.Split(blockedCountries, ",")
		for _, country := range countries {
			if strings.TrimSpace(country) == countryCode {
				return fmt.Errorf("access denied: country %s is blocked", countryCode)
			}
		}
	}

	return nil
}

// executeRateLimit executes rate limiting behavior
func (pm *PropertyManager) executeRateLimit(behavior *Behavior, context *HTTPContext, result *RuleResult) error {
	if pm.Debug {
		fmt.Printf("🔧 Rate limit behavior: %+v\n", behavior.Options)
	}

	// In a real implementation, this would check rate limits
	// For now, we'll just log the rate limit settings
	if rps, ok := behavior.Options["requests_per_second"].(float64); ok {
		if pm.Debug {
			fmt.Printf("Rate limit: %f requests per second\n", rps)
		}
	}

	if burst, ok := behavior.Options["burst_size"].(float64); ok {
		if pm.Debug {
			fmt.Printf("Burst size: %f\n", burst)
		}
	}

	return nil
}

// executeCompression executes compression behavior
func (pm *PropertyManager) executeCompression(behavior *Behavior, context *HTTPContext, result *RuleResult) error {
	if pm.Debug {
		fmt.Printf("🔧 Compression behavior: %+v\n", behavior.Options)
	}

	// Store compression settings in result
	if result.CompressionSettings == nil {
		result.CompressionSettings = make(map[string]interface{})
	}

	for key, value := range behavior.Options {
		result.CompressionSettings[key] = value
	}

	return nil
}

// executeImageOptimization executes image optimization behavior
func (pm *PropertyManager) executeImageOptimization(behavior *Behavior, context *HTTPContext, result *RuleResult) error {
	if pm.Debug {
		fmt.Printf("🔧 Image optimization behavior: %+v\n", behavior.Options)
	}

	// Store image optimization settings in result
	if result.ImageOptimizationSettings == nil {
		result.ImageOptimizationSettings = make(map[string]interface{})
	}

	for key, value := range behavior.Options {
		result.ImageOptimizationSettings[key] = value
	}

	return nil
}

// executeModifyHeaders executes header modification behavior
func (pm *PropertyManager) executeModifyHeaders(behavior *Behavior, context *HTTPContext, result *RuleResult) error {
	if pm.Debug {
		fmt.Printf("🔧 Modify headers behavior: %+v\n", behavior.Options)
	}

	// Add headers
	if addHeaders, ok := behavior.Options["add"].(string); ok {
		var headers map[string]string
		if err := json.Unmarshal([]byte(addHeaders), &headers); err == nil {
			for key, value := range headers {
				result.ModifiedHeaders[key] = value
			}
		}
	}

	// Remove headers
	if removeHeaders, ok := behavior.Options["remove"].(string); ok {
		var headers []string
		if err := json.Unmarshal([]byte(removeHeaders), &headers); err == nil {
			for _, header := range headers {
				result.RemovedHeaders = append(result.RemovedHeaders, header)
			}
		}
	}

	// Set headers
	if setHeaders, ok := behavior.Options["set"].(string); ok {
		var headers map[string]string
		if err := json.Unmarshal([]byte(setHeaders), &headers); err == nil {
			for key, value := range headers {
				result.ModifiedHeaders[key] = value
			}
		}
	}

	return nil
}

// executeURLRewrite executes URL rewriting behavior
func (pm *PropertyManager) executeURLRewrite(behavior *Behavior, context *HTTPContext, result *RuleResult) error {
	if pm.Debug {
		fmt.Printf("🔧 URL rewrite behavior: %+v\n", behavior.Options)
	}

	pattern, ok := behavior.Options["pattern"].(string)
	if !ok {
		return fmt.Errorf("URL rewrite: pattern is required")
	}

	replacement, ok := behavior.Options["replacement"].(string)
	if !ok {
		return fmt.Errorf("URL rewrite: replacement is required")
	}

	// Compile regex pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("URL rewrite: invalid regex pattern: %v", err)
	}

	// Apply rewrite
	newPath := re.ReplaceAllString(context.Path, replacement)
	if newPath != context.Path {
		result.RewrittenURL = newPath
		context.Path = newPath
	}

	// Check if redirect is needed
	if redirect, ok := behavior.Options["redirect"].(bool); ok && redirect {
		statusCode := 302 // Default
		if code, ok := behavior.Options["status_code"].(float64); ok {
			statusCode = int(code)
		}
		result.RedirectStatus = statusCode
		result.RedirectLocation = newPath
	}

	return nil
}

// executeConditionalRedirect executes conditional redirect behavior
func (pm *PropertyManager) executeConditionalRedirect(behavior *Behavior, context *HTTPContext, result *RuleResult) error {
	if pm.Debug {
		fmt.Printf("🔧 Conditional redirect behavior: %+v\n", behavior.Options)
	}

	conditions, ok := behavior.Options["conditions"].(string)
	if !ok {
		return fmt.Errorf("conditional redirect: conditions are required")
	}

	var redirectConditions []map[string]interface{}
	if err := json.Unmarshal([]byte(conditions), &redirectConditions); err != nil {
		return fmt.Errorf("conditional redirect: invalid conditions format: %v", err)
	}

	for _, condition := range redirectConditions {
		// Check if condition matches
		if pm.matchesRedirectCondition(condition, context) {
			if redirectTo, ok := condition["redirect_to"].(string); ok {
				result.RedirectLocation = redirectTo
				result.RedirectStatus = 302 // Default
				break
			}
		}
	}

	return nil
}

// matchesRedirectCondition checks if a redirect condition matches
func (pm *PropertyManager) matchesRedirectCondition(condition map[string]interface{}, context *HTTPContext) bool {
	header, ok := condition["header"].(string)
	if !ok {
		return false
	}

	headerValue, exists := context.Headers[header]
	if !exists {
		return false
	}

	// Check contains condition
	if contains, ok := condition["contains"].(string); ok {
		return strings.Contains(headerValue, contains)
	}

	// Check equals condition
	if equals, ok := condition["equals"].(string); ok {
		return headerValue == equals
	}

	// Check starts_with condition
	if startsWith, ok := condition["starts_with"].(string); ok {
		return strings.HasPrefix(headerValue, startsWith)
	}

	// Check ends_with condition
	if endsWith, ok := condition["ends_with"].(string); ok {
		return strings.HasSuffix(headerValue, endsWith)
	}

	return false
}
