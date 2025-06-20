package propertymanager

import (
	"fmt"
	"strings"
)

// executeSetResponseHeader sets a response header
func (pm *PropertyManager) executeSetResponseHeader(behavior *Behavior, context *HTTPContext, result *RuleResult) error {
	var headerName, headerValue string

	for _, option := range behavior.Option {
		switch option.Name {
		case "header_name":
			headerName = option.Value
		case "value":
			headerValue = option.Value
		}
	}

	if headerName != "" {
		result.ModifiedHeaders[headerName] = headerValue
		if pm.Debug {
			fmt.Printf("📝 Set response header: %s = %s\n", headerName, headerValue)
		}
	}

	return nil
}

// executeSetRequestHeader sets a request header
func (pm *PropertyManager) executeSetRequestHeader(behavior *Behavior, context *HTTPContext, result *RuleResult) error {
	var headerName, headerValue string

	for _, option := range behavior.Option {
		switch option.Name {
		case "header_name":
			headerName = option.Value
		case "value":
			headerValue = option.Value
		}
	}

	if headerName != "" && context.Request != nil {
		context.Request.Header.Set(headerName, headerValue)
		context.Headers[headerName] = headerValue
		if pm.Debug {
			fmt.Printf("📝 Set request header: %s = %s\n", headerName, headerValue)
		}
	}

	return nil
}

// executeSetVariable sets a variable
func (pm *PropertyManager) executeSetVariable(behavior *Behavior, context *HTTPContext, result *RuleResult) error {
	var varName, varValue string

	for _, option := range behavior.Option {
		switch option.Name {
		case "variable_name":
			varName = option.Value
		case "value":
			varValue = option.Value
		}
	}

	if varName != "" {
		context.Variables[varName] = varValue
		result.Variables[varName] = varValue
		if pm.Debug {
			fmt.Printf("📝 Set variable: %s = %s\n", varName, varValue)
		}
	}

	return nil
}

// executeRedirect performs a redirect
func (pm *PropertyManager) executeRedirect(behavior *Behavior, context *HTTPContext, result *RuleResult) error {
	var redirectURL, statusCode string

	for _, option := range behavior.Option {
		switch option.Name {
		case "destination":
			redirectURL = option.Value
		case "status_code":
			statusCode = option.Value
		}
	}

	if redirectURL != "" {
		if statusCode == "" {
			statusCode = "302"
		}

		result.ResponseContent = fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Redirecting...</title>
    <meta http-equiv="refresh" content="0;url=%s">
</head>
<body>
    <p>Redirecting to <a href="%s">%s</a>...</p>
</body>
</html>`, redirectURL, redirectURL, redirectURL)

		result.ModifiedHeaders["Location"] = redirectURL
		result.ModifiedHeaders["Status"] = statusCode

		if pm.Debug {
			fmt.Printf("🔄 Redirect: %s (Status: %s)\n", redirectURL, statusCode)
		}
	}

	return nil
}

// executeCacheKeyQueryParams modifies cache key based on query parameters
func (pm *PropertyManager) executeCacheKeyQueryParams(behavior *Behavior, context *HTTPContext, result *RuleResult) error {
	var behaviorType string

	for _, option := range behavior.Option {
		switch option.Name {
		case "behavior":
			behaviorType = option.Value
		}
	}

	if pm.Debug {
		fmt.Printf("🗄️  Cache key query params behavior: %s\n", behaviorType)
	}

	// This behavior would typically modify how the cache key is generated
	// For now, we'll just log it
	return nil
}

// executeOriginErrorPassThru handles origin error pass-through
func (pm *PropertyManager) executeOriginErrorPassThru(behavior *Behavior, context *HTTPContext, result *RuleResult) error {
	var enabled string

	for _, option := range behavior.Option {
		switch option.Name {
		case "enabled":
			enabled = option.Value
		}
	}

	if pm.Debug {
		fmt.Printf("⚠️  Origin error pass-thru: %s\n", enabled)
	}

	// This behavior would typically control how origin errors are handled
	// For now, we'll just log it
	return nil
}

// executeESI enables ESI processing
func (pm *PropertyManager) executeESI(behavior *Behavior, context *HTTPContext, result *RuleResult) error {
	var enabled string

	for _, option := range behavior.Option {
		switch option.Name {
		case "enabled":
			enabled = option.Value
		}
	}

	if pm.Debug {
		fmt.Printf("🔧 ESI processing: %s\n", enabled)
	}

	// This behavior would typically enable ESI processing
	// For now, we'll just log it
	return nil
}

// executeGzipResponse enables gzip compression
func (pm *PropertyManager) executeGzipResponse(behavior *Behavior, context *HTTPContext, result *RuleResult) error {
	var enabled string

	for _, option := range behavior.Option {
		switch option.Name {
		case "enabled":
			enabled = option.Value
		}
	}

	if enabled == "true" {
		result.ModifiedHeaders["Content-Encoding"] = "gzip"
		if pm.Debug {
			fmt.Printf("🗜️  Gzip compression enabled\n")
		}
	}

	return nil
}

// executeEdgeRedirector handles edge redirects
func (pm *PropertyManager) executeEdgeRedirector(behavior *Behavior, context *HTTPContext, result *RuleResult) error {
	var redirectType, destination string

	for _, option := range behavior.Option {
		switch option.Name {
		case "redirect_type":
			redirectType = option.Value
		case "destination":
			destination = option.Value
		}
	}

	if redirectType != "" && destination != "" {
		if pm.Debug {
			fmt.Printf("🔄 Edge redirect: %s -> %s\n", redirectType, destination)
		}

		// Handle different redirect types
		switch redirectType {
		case "permanent":
			result.ModifiedHeaders["Status"] = "301"
		case "temporary":
			result.ModifiedHeaders["Status"] = "302"
		case "found":
			result.ModifiedHeaders["Status"] = "302"
		case "see_other":
			result.ModifiedHeaders["Status"] = "303"
		}

		result.ModifiedHeaders["Location"] = destination
	}

	return nil
}

// executeOrigin sets the origin configuration
func (pm *PropertyManager) executeOrigin(behavior *Behavior, context *HTTPContext, result *RuleResult) error {
	var originType, hostname, port string

	for _, option := range behavior.Option {
		switch option.Name {
		case "origin_type":
			originType = option.Value
		case "hostname":
			hostname = option.Value
		case "port":
			port = option.Value
		}
	}

	if pm.Debug {
		fmt.Printf("🌐 Origin: %s (%s:%s)\n", originType, hostname, port)
	}

	// This behavior would typically configure the origin server
	// For now, we'll just log it
	return nil
}

// getBehaviorOption gets a behavior option value by name
func (pm *PropertyManager) getBehaviorOption(behavior *Behavior, optionName string) string {
	for _, option := range behavior.Option {
		if option.Name == optionName {
			return option.Value
		}
	}
	return ""
}

// expandVariables expands variables in a string
func (pm *PropertyManager) expandVariables(input string, context *HTTPContext) string {
	result := input

	// Simple variable expansion - in a real implementation, this would be more sophisticated
	for varName, varValue := range context.Variables {
		placeholder := fmt.Sprintf("$(%s)", varName)
		result = strings.ReplaceAll(result, placeholder, varValue)
	}

	// Expand common HTTP variables
	result = strings.ReplaceAll(result, "$(HTTP_HOST)", context.Host)
	result = strings.ReplaceAll(result, "$(HTTP_METHOD)", context.Method)
	result = strings.ReplaceAll(result, "$(HTTP_PATH)", context.Path)
	result = strings.ReplaceAll(result, "$(HTTP_QUERY)", context.Query)
	result = strings.ReplaceAll(result, "$(CLIENT_IP)", context.ClientIP)
	result = strings.ReplaceAll(result, "$(USER_AGENT)", context.UserAgent)

	return result
}
