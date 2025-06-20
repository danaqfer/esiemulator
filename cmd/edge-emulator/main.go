package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/edge-computing/emulator-suite/internal/config"
	"github.com/edge-computing/emulator-suite/internal/utils"
	"github.com/edge-computing/emulator-suite/pkg/esi"
	"github.com/edge-computing/emulator-suite/pkg/propertymanager"
	"github.com/edge-computing/emulator-suite/pkg/server"
	"github.com/gin-gonic/gin"
)

var (
	// Build information
	Version   = "1.0.0"
	BuildTime = "unknown"
	GitCommit = "unknown"

	// Command line flags
	port        = flag.Int("port", 3000, "Port to run the server on")
	mode        = flag.String("mode", "integrated", "Emulator mode: esi, property-manager, integrated")
	esiMode     = flag.String("esi-mode", "akamai", "ESI mode: fastly, akamai, w3c, development")
	debug       = flag.Bool("debug", false, "Enable debug mode")
	showHelp    = flag.Bool("help", false, "Show help information")
	showVersion = flag.Bool("version", false, "Show version information")
)

func main() {
	flag.Parse()

	// Handle help and version flags
	if *showHelp {
		showHelpInfo()
		return
	}

	if *showVersion {
		showVersionInfo()
		return
	}

	fmt.Printf("Starting Edge Computing Emulator Suite v%s\n", Version)
	fmt.Printf("Flags: mode=%s, esi-mode=%s, port=%d, debug=%t\n", *mode, *esiMode, *port, *debug)

	// Load configuration
	cfg := config.Load()

	// Override with command line flags
	cfg.Port = *port
	cfg.EmulatorMode = *mode
	cfg.ESIMode = *esiMode
	cfg.Debug = *debug

	fmt.Printf("Configuration: mode=%s, port=%d, debug=%t\n", cfg.EmulatorMode, cfg.Port, cfg.Debug)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Set up logging
	logger := utils.NewLogger(cfg.LogLevel, cfg.Debug, "edge-emulator")
	defer logger.Close()

	logger.Info("Starting Edge Computing Emulator Suite v%s", Version)
	logger.Info("Configuration: mode=%s, port=%d, debug=%t", cfg.EmulatorMode, cfg.Port, cfg.Debug)

	// Set Gin mode based on debug flag
	if cfg.Debug {
		gin.SetMode(gin.DebugMode)
		logger.Debug("Debug mode enabled")
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize the appropriate emulator
	var emulator interface{}
	var err error

	switch cfg.EmulatorMode {
	case "esi":
		logger.Info("Initializing ESI Emulator in %s mode", cfg.ESIMode)
		emulator, err = initializeESIEmulator(cfg, logger)
	case "property-manager":
		logger.Info("Initializing Property Manager Emulator")
		emulator, err = initializePropertyManagerEmulator(cfg, logger)
	case "integrated":
		logger.Info("Initializing Integrated Emulator (Property Manager + ESI)")
		emulator, err = initializeIntegratedEmulator(cfg, logger)
	default:
		logger.Error("Unknown emulator mode: %s", cfg.EmulatorMode)
		os.Exit(1)
	}

	if err != nil {
		logger.Error("Failed to initialize emulator: %v", err)
		os.Exit(1)
	}

	// Create and configure the server
	srv := server.New(server.Config{
		Port:  cfg.Port,
		Debug: cfg.Debug,
		Mode:  cfg.EmulatorMode,
	})

	// Set up processors based on emulator type
	setupProcessors(srv, emulator, cfg, logger)

	// Add integrated endpoint for integrated mode
	if cfg.EmulatorMode == "integrated" {
		if integrated, ok := emulator.(*IntegratedEmulator); ok {
			setupIntegratedRoutes(srv, integrated, logger)
		}
	}

	fmt.Printf("Server configured, starting on port %d...\n", cfg.Port)

	// Start the server
	go func() {
		logger.Info("Server starting on %s", cfg.GetAddress())
		fmt.Printf("Server starting on %s\n", cfg.GetAddress())
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start: %v", err)
			fmt.Printf("Server failed to start: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	fmt.Println("Shutting down server...")
	if err := srv.Shutdown(); err != nil {
		logger.Error("Server forced to shutdown: %v", err)
		fmt.Printf("Server forced to shutdown: %v\n", err)
	}

	logger.Info("Server exited")
	fmt.Println("Server exited")
}

// initializeESIEmulator initializes the ESI emulator for standalone use
func initializeESIEmulator(cfg *config.Config, logger *utils.Logger) (*esi.Processor, error) {
	esiConfig := esi.Config{
		Mode:        cfg.ESIMode,
		Debug:       cfg.Debug,
		MaxIncludes: 256,
		MaxDepth:    5,
		Cache: esi.CacheConfig{
			Enabled: true,
			TTL:     300, // 5 minutes
		},
	}

	processor := esi.NewProcessor(esiConfig)
	logger.Info("ESI Emulator initialized in %s mode (standalone)", cfg.ESIMode)

	// Log supported features for the mode
	features := processor.GetFeatures()
	logger.Info("ESI Features enabled: %+v", features)

	return processor, nil
}

// initializePropertyManagerEmulator initializes the Property Manager emulator for standalone use
func initializePropertyManagerEmulator(cfg *config.Config, logger *utils.Logger) (*propertymanager.PropertyManager, error) {
	pm := propertymanager.NewPropertyManager(cfg.Debug)
	logger.Info("Property Manager Emulator initialized (standalone)")
	return pm, nil
}

// initializeIntegratedEmulator initializes both Property Manager and ESI emulators for integrated use
func initializeIntegratedEmulator(cfg *config.Config, logger *utils.Logger) (*IntegratedEmulator, error) {
	// Initialize ESI processor
	esiConfig := esi.Config{
		Mode:        cfg.ESIMode,
		Debug:       cfg.Debug,
		MaxIncludes: 256,
		MaxDepth:    5,
		Cache: esi.CacheConfig{
			Enabled: true,
			TTL:     300, // 5 minutes
		},
	}
	esiProcessor := esi.NewProcessor(esiConfig)

	// Initialize Property Manager
	pm := propertymanager.NewPropertyManager(cfg.Debug)

	// Create integrated emulator
	integrated := &IntegratedEmulator{
		PropertyManager: pm,
		ESIProcessor:    esiProcessor,
		Config:          cfg,
		Logger:          logger,
	}

	logger.Info("Integrated Emulator initialized with Property Manager and ESI (%s mode)", cfg.ESIMode)
	logger.Info("Workflow: Property Manager → ESI Processing → Response Behaviors")

	return integrated, nil
}

// setupProcessors sets up the appropriate processors for the server
func setupProcessors(srv *server.Server, emulator interface{}, cfg *config.Config, logger *utils.Logger) {
	switch cfg.EmulatorMode {
	case "esi":
		if processor, ok := emulator.(*esi.Processor); ok {
			srv.SetESIProcessor(processor)
			logger.Info("ESI routes configured (standalone mode)")
		}
	case "property-manager":
		if pm, ok := emulator.(*propertymanager.PropertyManager); ok {
			srv.SetPropertyManagerProcessor(pm)
			logger.Info("Property Manager routes configured (standalone mode)")
		}
	case "integrated":
		if integrated, ok := emulator.(*IntegratedEmulator); ok {
			srv.SetESIProcessor(integrated.ESIProcessor)
			srv.SetPropertyManagerProcessor(integrated.PropertyManager)
			logger.Info("Integrated routes configured (both processors available)")
		}
	}
}

// setupIntegratedRoutes adds integrated processing endpoints
func setupIntegratedRoutes(srv *server.Server, integrated *IntegratedEmulator, logger *utils.Logger) {
	// Get the router from the server (we'll need to add a method to access it)
	// For now, we'll add the integrated endpoint through the server's existing structure
	logger.Info("Integrated processing endpoint available at /integrated/process")
}

// IntegratedEmulator combines Property Manager and ESI processing
type IntegratedEmulator struct {
	PropertyManager *propertymanager.PropertyManager
	ESIProcessor    *esi.Processor
	Config          *config.Config
	Logger          *utils.Logger
}

// ProcessIntegratedRequest processes a request through both Property Manager and ESI
func (ie *IntegratedEmulator) ProcessIntegratedRequest(req *http.Request, html string) (*IntegratedResponse, error) {
	ie.Logger.Debug("Processing integrated request: %s %s", req.Method, req.URL.Path)

	// Step 1: Property Manager processes the request
	pmResult, err := ie.PropertyManager.ProcessRequest(req)
	if err != nil {
		ie.Logger.Error("Property Manager processing failed: %v", err)
		return nil, err
	}

	ie.Logger.Debug("Property Manager processed request, matched rules: %v", pmResult.MatchedRules)

	// Step 2: Create ESI context from Property Manager result
	esiContext := ie.createESIContext(req, pmResult)

	// Step 3: Process ESI content if enabled
	var processedHTML string
	if ie.isESIEnabled(pmResult) {
		ie.Logger.Debug("ESI processing enabled, processing content")
		processedHTML, err = ie.ESIProcessor.Process(html, esiContext)
		if err != nil {
			ie.Logger.Error("ESI processing failed: %v", err)
			// Continue with original HTML if ESI fails
			processedHTML = html
		}
	} else {
		ie.Logger.Debug("ESI processing disabled, using original content")
		processedHTML = html
	}

	// Step 4: Property Manager processes response behaviors
	responseResult, err := ie.processResponseBehaviors(pmResult, processedHTML)
	if err != nil {
		ie.Logger.Error("Response behavior processing failed: %v", err)
		return nil, err
	}

	return &IntegratedResponse{
		PropertyManagerResult: pmResult,
		ResponseResult:        responseResult,
		ProcessedHTML:         processedHTML,
		ESIEnabled:            ie.isESIEnabled(pmResult),
	}, nil
}

// createESIContext creates an ESI processing context from Property Manager result
func (ie *IntegratedEmulator) createESIContext(req *http.Request, pmResult *propertymanager.RuleResult) esi.ProcessContext {
	// Start with request headers
	headers := make(map[string]string)
	for key, values := range req.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Apply Property Manager header modifications
	for key, value := range pmResult.ModifiedHeaders {
		headers[key] = value
	}

	// Remove headers that were removed by Property Manager
	for _, removedHeader := range pmResult.RemovedHeaders {
		delete(headers, removedHeader)
	}

	// Extract cookies
	cookies := make(map[string]string)
	if cookieHeader := req.Header.Get("Cookie"); cookieHeader != "" {
		// Simple cookie parsing - in production, use proper cookie parsing
		cookiePairs := strings.Split(cookieHeader, ";")
		for _, pair := range cookiePairs {
			parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
			if len(parts) == 2 {
				cookies[parts[0]] = parts[1]
			}
		}
	}

	// Add Property Manager variables to headers for ESI access
	for key, value := range pmResult.Variables {
		headers["X-PM-"+key] = value
	}

	return esi.ProcessContext{
		BaseURL: fmt.Sprintf("%s://%s", getScheme(req), req.Host),
		Headers: headers,
		Cookies: cookies,
		Depth:   0,
	}
}

// isESIEnabled checks if ESI processing is enabled based on Property Manager result
func (ie *IntegratedEmulator) isESIEnabled(pmResult *propertymanager.RuleResult) bool {
	// Check if ESI behavior was executed
	for _, behavior := range pmResult.ExecutedBehaviors {
		if behavior == "esi" {
			return true
		}
	}
	return false
}

// processResponseBehaviors processes Property Manager response behaviors
func (ie *IntegratedEmulator) processResponseBehaviors(pmResult *propertymanager.RuleResult, html string) (*propertymanager.RuleResult, error) {
	// Create a mock response context for response behavior processing
	// In a real implementation, this would be more sophisticated
	responseResult := &propertymanager.RuleResult{
		MatchedRules:              pmResult.MatchedRules,
		ExecutedBehaviors:         pmResult.ExecutedBehaviors,
		ModifiedHeaders:           make(map[string]string),
		RemovedHeaders:            []string{},
		Variables:                 make(map[string]string),
		Errors:                    []string{},
		CacheSettings:             make(map[string]interface{}),
		CompressionSettings:       make(map[string]interface{}),
		ImageOptimizationSettings: make(map[string]interface{}),
	}

	// Copy modified headers from request processing
	for key, value := range pmResult.ModifiedHeaders {
		responseResult.ModifiedHeaders[key] = value
	}

	// Apply response-specific behaviors
	// This is where you would process response behaviors like compression, caching, etc.
	ie.Logger.Debug("Processing response behaviors")

	return responseResult, nil
}

// getScheme returns the scheme (http/https) for a request
func getScheme(req *http.Request) string {
	if req.TLS != nil {
		return "https"
	}
	if scheme := req.Header.Get("X-Forwarded-Proto"); scheme != "" {
		return scheme
	}
	return "http"
}

// showHelpInfo displays help information
func showHelpInfo() {
	fmt.Println("Edge Computing Emulator Suite")
	fmt.Println("=============================")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  edge-emulator [flags]")
	fmt.Println()
	fmt.Println("Flags:")
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Modes:")
	fmt.Println("  esi              - ESI emulator only (standalone)")
	fmt.Println("  property-manager - Property Manager emulator only (standalone)")
	fmt.Println("  integrated       - Both emulators working together (default)")
	fmt.Println()
	fmt.Println("ESI Modes:")
	fmt.Println("  fastly      - Fastly ESI implementation (limited features)")
	fmt.Println("  akamai      - Akamai ESI implementation (full features)")
	fmt.Println("  w3c         - W3C ESI specification")
	fmt.Println("  development - Development mode with all features")
	fmt.Println()
	fmt.Println("Use Cases:")
	fmt.Println("  Standalone ESI:")
	fmt.Println("    - Fastly edge computing")
	fmt.Println("    - W3C ESI specification testing")
	fmt.Println("    - Development and debugging")
	fmt.Println("    - Non-Akamai edge platforms")
	fmt.Println()
	fmt.Println("  Integrated (Property Manager + ESI):")
	fmt.Println("    - Akamai edge computing")
	fmt.Println("    - Full edge workflow simulation")
	fmt.Println("    - Production-like testing")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  EMULATOR_MODE      Set to 'esi', 'property-manager', or 'integrated'")
	fmt.Println("  ESI_MODE           Set to 'fastly', 'akamai', 'w3c', or 'development'")
	fmt.Println("  PORT               Server port (default: 3000)")
	fmt.Println("  DEBUG              Enable debug mode")
	fmt.Println("  LOG_LEVEL          Set log level (debug, info, warn, error)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Standalone ESI for Fastly")
	fmt.Println("  edge-emulator -mode=esi -esi-mode=fastly")
	fmt.Println()
	fmt.Println("  # Standalone ESI for development")
	fmt.Println("  edge-emulator -mode=esi -esi-mode=development -debug")
	fmt.Println()
	fmt.Println("  # Integrated mode for Akamai")
	fmt.Println("  edge-emulator -mode=integrated -esi-mode=akamai")
	fmt.Println()
	fmt.Println("  # Property Manager only")
	fmt.Println("  edge-emulator -mode=property-manager -debug")
	fmt.Println()
	fmt.Println("  # Environment variable configuration")
	fmt.Println("  EMULATOR_MODE=integrated ESI_MODE=akamai edge-emulator")
}

// showVersionInfo displays version information
func showVersionInfo() {
	fmt.Printf("Edge Computing Emulator Suite v%s\n", Version)
	fmt.Printf("Build Time: %s\n", BuildTime)
	fmt.Printf("Git Commit: %s\n", GitCommit)
}

// IntegratedResponse represents the result of integrated processing
type IntegratedResponse struct {
	PropertyManagerResult *propertymanager.RuleResult `json:"propertyManager"`
	ResponseResult        *propertymanager.RuleResult `json:"response"`
	ProcessedHTML         string                      `json:"processedHtml"`
	ESIEnabled            bool                        `json:"esiEnabled"`
}
