package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"esi-emulator/pkg/esi"

	"github.com/gin-gonic/gin"
)

// Config holds server configuration
type Config struct {
	Port  int    `json:"port"`
	Debug bool   `json:"debug"`
	Mode  string `json:"mode"`
}

// Server represents the ESI HTTP server
type Server struct {
	processor *esi.Processor
	config    Config
	router    *gin.Engine
	server    *http.Server
}

// ProcessRequest represents a request to process ESI content
type ProcessRequest struct {
	HTML    string              `json:"html" binding:"required"`
	Context *esi.ProcessContext `json:"context,omitempty"`
}

// ProcessResponse represents the response from processing ESI content
type ProcessResponse struct {
	Result string    `json:"result"`
	Stats  StatsInfo `json:"stats"`
}

// StatsInfo holds statistics information
type StatsInfo struct {
	ProcessingTime int64  `json:"processingTime"`
	Mode           string `json:"mode"`
	Requests       int64  `json:"requests"`
	CacheHits      int64  `json:"cacheHits"`
	CacheMiss      int64  `json:"cacheMiss"`
	Errors         int64  `json:"errors"`
	TotalTime      int64  `json:"totalTime"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Example represents an ESI example
type Example struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	HTML        string   `json:"html"`
	Modes       []string `json:"modes"`
}

// New creates a new ESI server
func New(processor *esi.Processor, config Config) *Server {
	// Set Gin mode
	if !config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	server := &Server{
		processor: processor,
		config:    config,
		router:    router,
	}

	server.setupRoutes()
	return server
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Root endpoint - status and configuration
	s.router.GET("/", s.handleRoot)

	// Process ESI content
	s.router.POST("/process", s.handleProcess)

	// Statistics endpoint
	s.router.GET("/stats", s.handleStats)

	// Cache management
	s.router.DELETE("/cache", s.handleClearCache)

	// Examples endpoints
	s.router.GET("/examples", s.handleListExamples)
	s.router.GET("/examples/:name", s.handleGetExample)

	// Test fragments endpoint
	s.router.GET("/fragments/:name", s.handleGetFragment)

	// Health check
	s.router.GET("/health", s.handleHealth)
}

// handleRoot returns server information and available endpoints
func (s *Server) handleRoot(c *gin.Context) {
	stats := s.processor.GetStats()
	features := s.processor.GetFeatures()

	c.JSON(http.StatusOK, gin.H{
		"name":     "ESI Emulator",
		"version":  "0.1.0",
		"mode":     s.config.Mode,
		"features": features,
		"stats": gin.H{
			"requests":  stats.Requests,
			"cacheHits": stats.CacheHits,
			"cacheMiss": stats.CacheMiss,
			"errors":    stats.Errors,
			"totalTime": stats.TotalTime,
		},
		"endpoints": gin.H{
			"/process":         "POST - Process ESI content",
			"/examples":        "GET - List available examples",
			"/examples/:name":  "GET - Get specific example",
			"/stats":           "GET - Get processing statistics",
			"/cache":           "DELETE - Clear cache",
			"/fragments/:name": "GET - Get test fragments",
			"/health":          "GET - Health check",
		},
	})
}

// handleProcess processes ESI content
func (s *Server) handleProcess(c *gin.Context) {
	var req ProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// Create default context if not provided
	if req.Context == nil {
		req.Context = &esi.ProcessContext{
			BaseURL: fmt.Sprintf("%s://%s", getScheme(c), c.Request.Host),
			Headers: make(map[string]string),
			Cookies: make(map[string]string),
			Depth:   0,
		}
	}

	// Add request headers to context
	if req.Context.Headers == nil {
		req.Context.Headers = make(map[string]string)
	}

	for key, values := range c.Request.Header {
		if len(values) > 0 {
			req.Context.Headers[key] = values[0]
		}
	}

	startTime := time.Now()
	result, err := s.processor.Process(req.HTML, *req.Context)
	processingTime := time.Since(startTime).Milliseconds()

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "ESI processing failed",
			Message: err.Error(),
		})
		return
	}

	stats := s.processor.GetStats()
	c.JSON(http.StatusOK, ProcessResponse{
		Result: result,
		Stats: StatsInfo{
			ProcessingTime: processingTime,
			Mode:           s.config.Mode,
			Requests:       stats.Requests,
			CacheHits:      stats.CacheHits,
			CacheMiss:      stats.CacheMiss,
			Errors:         stats.Errors,
			TotalTime:      stats.TotalTime,
		},
	})
}

// handleStats returns processing statistics
func (s *Server) handleStats(c *gin.Context) {
	stats := s.processor.GetStats()
	features := s.processor.GetFeatures()

	c.JSON(http.StatusOK, gin.H{
		"mode":     s.config.Mode,
		"features": features,
		"cache": gin.H{
			"size":    s.processor.GetCacheSize(),
			"enabled": s.processor.GetFeatures().Include, // Simplified check
		},
		"stats": gin.H{
			"requests":  stats.Requests,
			"cacheHits": stats.CacheHits,
			"cacheMiss": stats.CacheMiss,
			"errors":    stats.Errors,
			"totalTime": stats.TotalTime,
		},
	})
}

// handleClearCache clears the fragment cache
func (s *Server) handleClearCache(c *gin.Context) {
	s.processor.ClearCache()

	stats := s.processor.GetStats()
	c.JSON(http.StatusOK, gin.H{
		"message": "Cache cleared",
		"stats": gin.H{
			"requests":  stats.Requests,
			"cacheHits": stats.CacheHits,
			"cacheMiss": stats.CacheMiss,
			"errors":    stats.Errors,
			"totalTime": stats.TotalTime,
		},
	})
}

// handleListExamples returns available examples
func (s *Server) handleListExamples(c *gin.Context) {
	examples := []gin.H{
		{
			"name":        "basic-include",
			"description": "Basic ESI include example",
			"modes":       []string{"fastly", "akamai", "w3c"},
		},
		{
			"name":        "conditional",
			"description": "ESI conditional processing",
			"modes":       []string{"akamai", "w3c"},
		},
		{
			"name":        "error-handling",
			"description": "ESI error handling and fallbacks",
			"modes":       []string{"akamai", "w3c"},
		},
		{
			"name":        "variables",
			"description": "ESI variable substitution",
			"modes":       []string{"akamai", "w3c"},
		},
		{
			"name":        "ecommerce",
			"description": "E-commerce shopping cart example",
			"modes":       []string{"fastly", "akamai", "w3c"},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"examples": examples,
	})
}

// handleGetExample returns a specific example
func (s *Server) handleGetExample(c *gin.Context) {
	name := c.Param("name")
	examples := s.getExamples()

	example, exists := examples[name]
	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Example not found",
			Message: fmt.Sprintf("Available examples: %v", getMapKeys(examples)),
		})
		return
	}

	c.JSON(http.StatusOK, example)
}

// handleGetFragment returns test fragments
func (s *Server) handleGetFragment(c *gin.Context) {
	name := c.Param("name")
	fragments := s.getTestFragments()

	fragment, exists := fragments[name]
	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Fragment not found",
			Message: fmt.Sprintf("Available fragments: %v", getMapKeys(fragments)),
		})
		return
	}

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, fragment)
}

// handleHealth returns health status
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"uptime": time.Since(time.Now()).Seconds(), // Placeholder
		"mode":   s.config.Mode,
	})
}

// getExamples returns example ESI content for testing
func (s *Server) getExamples() map[string]Example {
	return map[string]Example{
		"basic-include": {
			Name:        "Basic Include",
			Description: "Simple ESI include example",
			HTML: `<!DOCTYPE html>
<html>
<head>
    <title>ESI Basic Include Example</title>
</head>
<body>
    <h1>Welcome to ESI Testing</h1>
    <esi:include src="/fragments/header" />
    <main>
        <p>This is the main content area.</p>
        <esi:include src="/fragments/content" />
    </main>
    <esi:include src="/fragments/footer" />
</body>
</html>`,
			Modes: []string{"fastly", "akamai", "w3c"},
		},
		"conditional": {
			Name:        "Conditional Processing",
			Description: "ESI choose/when/otherwise example",
			HTML: `<!DOCTYPE html>
<html>
<head>
    <title>ESI Conditional Example</title>
</head>
<body>
    <esi:choose>
        <esi:when test="$(HTTP_COOKIE{user_type})=='premium'">
            <esi:include src="/fragments/premium-header" />
        </esi:when>
        <esi:when test="$(HTTP_COOKIE{user_type})=='basic'">
            <esi:include src="/fragments/basic-header" />
        </esi:when>
        <esi:otherwise>
            <esi:include src="/fragments/guest-header" />
        </esi:otherwise>
    </esi:choose>
    
    <main>Content based on user type</main>
</body>
</html>`,
			Modes: []string{"akamai", "w3c"},
		},
		"error-handling": {
			Name:        "Error Handling",
			Description: "ESI try/attempt/except and onerror example",
			HTML: `<!DOCTYPE html>
<html>
<head>
    <title>ESI Error Handling Example</title>
</head>
<body>
    <h1>Error Handling Examples</h1>
    
    <!-- Example 1: onerror="continue" -->
    <div>
        <h2>With onerror continue:</h2>
        <esi:include src="/fragments/might-fail" onerror="continue" />
        <p>This will always show, even if include fails.</p>
    </div>
    
    <!-- Example 2: alt attribute -->
    <div>
        <h2>With fallback URL:</h2>
        <esi:include src="/fragments/might-fail" alt="/fragments/fallback" />
    </div>
    
    <!-- Example 3: try/attempt/except -->
    <div>
        <h2>With try/except block:</h2>
        <esi:try>
            <esi:attempt>
                <esi:include src="/fragments/might-fail" />
                <p>Primary content loaded successfully</p>
            </esi:attempt>
            <esi:except>
                <p>Fallback content - primary source failed</p>
            </esi:except>
        </esi:try>
    </div>
</body>
</html>`,
			Modes: []string{"akamai", "w3c"},
		},
		"akamai-extensions": {
			Name:        "Akamai Extensions",
			Description: "Akamai-specific ESI extensions showcase",
			HTML: `<!DOCTYPE html>
<html>
<head>
    <title>Akamai ESI Extensions</title>
</head>
<body>
    <h1>Akamai ESI Extensions Showcase</h1>
    
    <!-- Variable Assignment -->
    <h2>Variable Assignment</h2>
    <esi:assign name="user_level" value="premium" />
    <esi:assign name="site_name">Akamai Demo Site</esi:assign>
    
    <!-- Expression Evaluation -->
    <h2>Expression Evaluation</h2>
    <p>User level: <esi:eval expr="$(user_level)" /></p>
    <p>Is premium: <esi:eval expr="$(user_level) == 'premium'" /></p>
    
    <!-- Built-in Functions -->
    <h2>Built-in Functions</h2>
    <p>Current time: <esi:function name="time" format="2006-01-02 15:04:05" /></p>
    <p>Random number: <esi:function name="random" min="1" max="100" /></p>
    <p>URL encoded: <esi:function name="url_encode" input="Hello World!" /></p>
    <p>Base64 encoded: <esi:function name="base64_encode" input="test data" /></p>
    
    <!-- Geo Variables -->
    <h2>Geo-location Variables</h2>
    <p>Country: $(GEO_COUNTRY_NAME) ($(GEO_COUNTRY_CODE))</p>
    <p>Region: $(GEO_REGION)</p>
    <p>City: $(GEO_CITY)</p>
    <p>Client IP: $(CLIENT_IP)</p>
    
    <!-- Debug Information -->
    <h2>Debug Information</h2>
    <esi:debug type="vars" />
    <esi:debug type="time" />
    
    <!-- Variable Expansion -->
    <h2>Variable Expansion</h2>
    <p>Welcome to $(site_name)!</p>
    <p>Browser: $(HTTP_USER_AGENT{browser})</p>
    <p>OS: $(HTTP_USER_AGENT{os})</p>
</body>
</html>`,
			Modes: []string{"akamai"},
		},
		"variables": {
			Name:        "Variable Substitution",
			Description: "ESI variable substitution and expansion",
			HTML: `<!DOCTYPE html>
<html>
<head>
    <title>ESI Variables Example</title>
</head>
<body>
    <h1>ESI Variable Substitution</h1>
    
    <!-- Standard Variables -->
    <h2>Standard ESI Variables</h2>
    <p>Host: $(HTTP_HOST)</p>
    <p>User Agent: $(HTTP_USER_AGENT)</p>
    <p>Cookie: $(HTTP_COOKIE{user_id})</p>
    <p>Referer: $(HTTP_REFERER)</p>
    
    <!-- Akamai Extensions -->
    <h2>Akamai Extended Variables</h2>
    <p>Request Method: $(REQUEST_METHOD)</p>
    <p>Request URI: $(REQUEST_URI)</p>
    <p>Client IP: $(CLIENT_IP)</p>
    
    <!-- Custom Variables -->
    <h2>Custom Variables</h2>
    <esi:assign name="page_title" value="Dynamic Page" />
    <esi:assign name="current_user" value="$(HTTP_COOKIE{username})" />
    <p>Page: $(page_title)</p>
    <p>User: $(current_user)</p>
</body>
</html>`,
			Modes: []string{"akamai", "w3c"},
		},
		"ecommerce": {
			Name:        "E-commerce Example",
			Description: "Shopping cart with ESI includes",
			HTML: `<!DOCTYPE html>
<html>
<head>
    <title>Online Store</title>
</head>
<body>
    <header>
        <img src="/logo.png" alt="Store Logo" />
        <esi:include src="/fragments/shopping-cart" />
        <esi:include src="/fragments/user-menu" />
    </header>
    
    <main>
        <h1>Featured Products</h1>
        <esi:include src="/fragments/featured-products" />
        
        <h2>Recommendations</h2>
        <esi:include src="/fragments/recommendations" onerror="continue" />
    </main>
    
    <footer>
        <esi:include src="/fragments/footer" />
    </footer>
</body>
</html>`,
			Modes: []string{"fastly", "akamai", "w3c"},
		},
	}
}

// getTestFragments returns test fragments for includes
func (s *Server) getTestFragments() map[string]string {
	currentTime := time.Now().Format(time.RFC3339)

	return map[string]string{
		"header":            "<header><h2>Dynamic Header Content</h2><nav>Navigation here</nav></header>",
		"content":           fmt.Sprintf("<div><p>This is dynamically included content.</p><p>Generated at: %s</p></div>", currentTime),
		"footer":            "<footer><p>&copy; 2024 ESI Emulator. All rights reserved.</p></footer>",
		"shopping-cart":     "<div class=\"cart\">Cart: 3 items ($45.99) <a href=\"/cart\">View Cart</a></div>",
		"user-menu":         "<div class=\"user-menu\"><a href=\"/login\">Login</a> | <a href=\"/register\">Register</a></div>",
		"featured-products": "<div class=\"products\"><div class=\"product\">Product 1 - $19.99</div><div class=\"product\">Product 2 - $25.99</div></div>",
		"recommendations":   "<div class=\"recommendations\"><h3>You might also like:</h3><div class=\"product\">Recommended Product - $15.99</div></div>",
		"fallback":          "<div class=\"fallback\">This is fallback content when the primary source fails.</div>",
		"premium-header":    "<header class=\"premium\"><h2>Premium User Header</h2><div class=\"premium-badge\">PREMIUM</div></header>",
		"basic-header":      "<header class=\"basic\"><h2>Basic User Header</h2></header>",
		"guest-header":      "<header class=\"guest\"><h2>Welcome Guest</h2><a href=\"/login\">Login for more features</a></header>",
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr:    ":" + strconv.Itoa(s.config.Port),
		Handler: s.router,
	}

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	if s.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return s.server.Shutdown(ctx)
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// getScheme returns the request scheme
func getScheme(c *gin.Context) string {
	if c.Request.TLS != nil {
		return "https"
	}
	return "http"
}

// getMapKeys returns the keys of a map as a slice
func getMapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
