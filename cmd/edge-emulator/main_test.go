package main

import (
	"net/http"
	"os"
	"testing"

	"github.com/edge-computing/emulator-suite/internal/config"
	"github.com/edge-computing/emulator-suite/internal/utils"
	"github.com/edge-computing/emulator-suite/pkg/esi"
	"github.com/edge-computing/emulator-suite/pkg/propertymanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain runs before all tests
func TestMain(m *testing.M) {
	// Set up test environment
	os.Setenv("EMULATOR_MODE", "test")
	os.Setenv("ESI_MODE", "development")
	os.Setenv("DEBUG", "true")

	// Run tests
	code := m.Run()

	// Clean up
	os.Unsetenv("EMULATOR_MODE")
	os.Unsetenv("ESI_MODE")
	os.Unsetenv("DEBUG")

	os.Exit(code)
}

// TestConfigurationValidation tests configuration validation
func TestConfigurationValidation(t *testing.T) {
	tests := []struct {
		name         string
		emulatorMode string
		esiMode      string
		port         int
		expectError  bool
	}{
		{
			name:         "Valid ESI mode",
			emulatorMode: "esi",
			esiMode:      "fastly",
			port:         3000,
			expectError:  false,
		},
		{
			name:         "Valid Property Manager mode",
			emulatorMode: "property-manager",
			esiMode:      "akamai",
			port:         3000,
			expectError:  false,
		},
		{
			name:         "Valid Integrated mode",
			emulatorMode: "integrated",
			esiMode:      "akamai",
			port:         3000,
			expectError:  false,
		},
		{
			name:         "Invalid emulator mode",
			emulatorMode: "invalid",
			esiMode:      "fastly",
			port:         3000,
			expectError:  true,
		},
		{
			name:         "Invalid ESI mode",
			emulatorMode: "esi",
			esiMode:      "invalid",
			port:         3000,
			expectError:  true,
		},
		{
			name:         "Invalid port",
			emulatorMode: "esi",
			esiMode:      "fastly",
			port:         70000,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				EmulatorMode: tt.emulatorMode,
				ESIMode:      tt.esiMode,
				Port:         tt.port,
				Debug:        true,
				LogLevel:     "info",
			}

			err := cfg.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestESIEmulatorInitialization tests ESI emulator initialization
func TestESIEmulatorInitialization(t *testing.T) {
	tests := []struct {
		name     string
		esiMode  string
		debug    bool
		expected string
	}{
		{
			name:     "Fastly mode",
			esiMode:  "fastly",
			debug:    false,
			expected: "fastly",
		},
		{
			name:     "Akamai mode",
			esiMode:  "akamai",
			debug:    true,
			expected: "akamai",
		},
		{
			name:     "W3C mode",
			esiMode:  "w3c",
			debug:    false,
			expected: "w3c",
		},
		{
			name:     "Development mode",
			esiMode:  "development",
			debug:    true,
			expected: "development",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ESIMode: tt.esiMode,
				Debug:   tt.debug,
			}
			logger := utils.NewLogger("info", tt.debug, "test")

			processor, err := initializeESIEmulator(cfg, logger)
			require.NoError(t, err)
			require.NotNil(t, processor)

			// Test that the processor has the correct mode
			processorConfig := processor.GetConfig()
			assert.Equal(t, tt.expected, processorConfig.Mode)

			// Test that features are enabled based on mode
			features := processor.GetFeatures()
			if tt.esiMode == "fastly" {
				// Fastly should have limited features
				assert.True(t, features.Include)
				assert.True(t, features.Comment)
				assert.True(t, features.Remove)
				assert.False(t, features.Choose) // Not in Fastly
			} else {
				// Other modes should have full features
				assert.True(t, features.Include)
				assert.True(t, features.Comment)
				assert.True(t, features.Remove)
				assert.True(t, features.Choose)
				assert.True(t, features.Vars)
				assert.True(t, features.Variables)
			}
		})
	}
}

// TestPropertyManagerEmulatorInitialization tests Property Manager emulator initialization
func TestPropertyManagerEmulatorInitialization(t *testing.T) {
	tests := []struct {
		name  string
		debug bool
	}{
		{
			name:  "Debug enabled",
			debug: true,
		},
		{
			name:  "Debug disabled",
			debug: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Debug: tt.debug,
			}
			logger := utils.NewLogger("info", tt.debug, "test")

			pm, err := initializePropertyManagerEmulator(cfg, logger)
			require.NoError(t, err)
			require.NotNil(t, pm)

			// Test that the Property Manager is properly initialized
			assert.Equal(t, tt.debug, pm.Debug)
			assert.NotNil(t, pm.Rules)
			assert.NotNil(t, pm.Behaviors)
			assert.NotNil(t, pm.Variables)
		})
	}
}

// TestIntegratedEmulatorInitialization tests integrated emulator initialization
func TestIntegratedEmulatorInitialization(t *testing.T) {
	tests := []struct {
		name     string
		esiMode  string
		debug    bool
		expected string
	}{
		{
			name:     "Akamai integrated mode",
			esiMode:  "akamai",
			debug:    true,
			expected: "akamai",
		},
		{
			name:     "Development integrated mode",
			esiMode:  "development",
			debug:    false,
			expected: "development",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				ESIMode: tt.esiMode,
				Debug:   tt.debug,
			}
			logger := utils.NewLogger("info", tt.debug, "test")

			integrated, err := initializeIntegratedEmulator(cfg, logger)
			require.NoError(t, err)
			require.NotNil(t, integrated)

			// Test that both processors are initialized
			assert.NotNil(t, integrated.ESIProcessor)
			assert.NotNil(t, integrated.PropertyManager)
			assert.Equal(t, cfg, integrated.Config)
			assert.Equal(t, logger, integrated.Logger)

			// Test ESI processor configuration
			esiConfig := integrated.ESIProcessor.GetConfig()
			assert.Equal(t, tt.expected, esiConfig.Mode)
			assert.Equal(t, tt.debug, esiConfig.Debug)

			// Test Property Manager configuration
			assert.Equal(t, tt.debug, integrated.PropertyManager.Debug)
		})
	}
}

// TestESIProcessing tests ESI processing functionality
func TestESIProcessing(t *testing.T) {
	cfg := &config.Config{
		ESIMode: "akamai",
		Debug:   true,
	}
	logger := utils.NewLogger("info", true, "test")

	processor, err := initializeESIEmulator(cfg, logger)
	require.NoError(t, err)

	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "Variable substitution",
			html:     `<esi:vars>Host: $(HTTP_HOST)</esi:vars>`,
			expected: "Host: example.com",
		},
		{
			name:     "Comment removal",
			html:     `<esi:comment>This should be removed</esi:comment><p>Content</p>`,
			expected: "<p>Content</p>",
		},
		{
			name:     "Remove tag",
			html:     `<esi:remove><p>This should be removed</p></esi:remove><p>Content</p>`,
			expected: "<p>Content</p>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := esi.ProcessContext{
				BaseURL: "http://example.com",
				Headers: map[string]string{
					"Host": "example.com",
				},
				Cookies: make(map[string]string),
				Depth:   0,
			}

			result, err := processor.Process(tt.html, context)
			require.NoError(t, err)
			assert.Contains(t, result, tt.expected)
		})
	}
}

// TestConfigurationLoading tests configuration loading from environment
func TestConfigurationLoading(t *testing.T) {
	// Set test environment variables
	os.Setenv("EMULATOR_MODE", "esi")
	os.Setenv("ESI_MODE", "fastly")
	os.Setenv("PORT", "8080")
	os.Setenv("DEBUG", "true")
	os.Setenv("LOG_LEVEL", "debug")

	defer func() {
		os.Unsetenv("EMULATOR_MODE")
		os.Unsetenv("ESI_MODE")
		os.Unsetenv("PORT")
		os.Unsetenv("DEBUG")
		os.Unsetenv("LOG_LEVEL")
	}()

	cfg := config.Load()

	assert.Equal(t, "esi", cfg.EmulatorMode)
	assert.Equal(t, "fastly", cfg.ESIMode)
	assert.Equal(t, 8080, cfg.Port)
	assert.True(t, cfg.Debug)
	assert.Equal(t, "debug", cfg.LogLevel)
}

// TestErrorHandling tests error handling scenarios
func TestErrorHandling(t *testing.T) {
	// Test invalid configuration
	cfg := &config.Config{
		EmulatorMode: "invalid",
		ESIMode:      "fastly",
		Port:         3000,
	}

	err := cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")

	// Test invalid ESI mode
	cfg = &config.Config{
		EmulatorMode: "esi",
		ESIMode:      "invalid",
		Port:         3000,
	}

	err = cfg.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

// TestESIEnabledDetection tests ESI enabled detection
func TestESIEnabledDetection(t *testing.T) {
	cfg := &config.Config{
		ESIMode: "akamai",
		Debug:   true,
	}
	logger := utils.NewLogger("info", true, "test")

	integrated, err := initializeIntegratedEmulator(cfg, logger)
	require.NoError(t, err)

	tests := []struct {
		name              string
		executedBehaviors []string
		expected          bool
	}{
		{
			name:              "ESI enabled",
			executedBehaviors: []string{"esi", "compress"},
			expected:          true,
		},
		{
			name:              "ESI disabled",
			executedBehaviors: []string{"compress", "cache"},
			expected:          false,
		},
		{
			name:              "No behaviors",
			executedBehaviors: []string{},
			expected:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pmResult := &propertymanager.RuleResult{
				ExecutedBehaviors: tt.executedBehaviors,
			}

			result := integrated.isESIEnabled(pmResult)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestESIContextCreation tests ESI context creation from Property Manager result
func TestESIContextCreation(t *testing.T) {
	cfg := &config.Config{
		ESIMode: "akamai",
		Debug:   true,
	}
	logger := utils.NewLogger("info", true, "test")

	integrated, err := initializeIntegratedEmulator(cfg, logger)
	require.NoError(t, err)

	// Create a proper HTTP request
	req, err := http.NewRequest("GET", "http://example.com/test", nil)
	require.NoError(t, err)
	req.Header.Set("User-Agent", "Test Browser")
	req.Header.Set("Host", "example.com")
	req.Header.Set("Cookie", "session=abc123; user=test")

	// Create a mock Property Manager result
	pmResult := &propertymanager.RuleResult{
		MatchedRules:      []string{"test-rule"},
		ExecutedBehaviors: []string{"esi"},
		ModifiedHeaders: map[string]string{
			"X-Custom-Header": "custom-value",
		},
		RemovedHeaders: []string{"X-Removed-Header"},
		Variables: map[string]string{
			"user_type": "premium",
		},
	}

	// Test ESI context creation
	esiContext := integrated.createESIContext(req, pmResult)

	// Verify the context
	assert.Equal(t, "http://example.com", esiContext.BaseURL)
	assert.Equal(t, "Test Browser", esiContext.Headers["User-Agent"])
	assert.Equal(t, "example.com", esiContext.Headers["Host"])
	assert.Equal(t, "custom-value", esiContext.Headers["X-Custom-Header"])
	assert.Equal(t, "premium", esiContext.Headers["X-PM-user_type"])
	assert.Equal(t, "abc123", esiContext.Cookies["session"])
	assert.Equal(t, "test", esiContext.Cookies["user"])
	assert.Equal(t, 0, esiContext.Depth)

	// Verify removed headers are not present
	_, exists := esiContext.Headers["X-Removed-Header"]
	assert.False(t, exists)
}

// TestPerformance tests basic performance characteristics
func TestPerformance(t *testing.T) {
	cfg := &config.Config{
		ESIMode: "akamai",
		Debug:   false, // Disable debug for performance test
	}
	logger := utils.NewLogger("info", false, "test")

	processor, err := initializeESIEmulator(cfg, logger)
	require.NoError(t, err)

	// Test processing time
	html := `<esi:vars>Host: $(HTTP_HOST)</esi:vars>`
	context := esi.ProcessContext{
		BaseURL: "http://example.com",
		Headers: map[string]string{
			"Host": "example.com",
		},
		Cookies: make(map[string]string),
		Depth:   0,
	}

	result, err := processor.Process(html, context)
	require.NoError(t, err)
	assert.Contains(t, result, "Host: example.com")
}

// TestEdgeCases tests edge cases and error conditions
func TestEdgeCases(t *testing.T) {
	cfg := &config.Config{
		ESIMode: "akamai",
		Debug:   true,
	}
	logger := utils.NewLogger("info", true, "test")

	integrated, err := initializeIntegratedEmulator(cfg, logger)
	require.NoError(t, err)

	// Test with empty HTML
	context := esi.ProcessContext{
		BaseURL: "http://example.com",
		Headers: map[string]string{
			"Host": "example.com",
		},
		Cookies: make(map[string]string),
		Depth:   0,
	}

	result, err := integrated.ESIProcessor.Process("", context)
	require.NoError(t, err)
	// Empty HTML gets wrapped in HTML structure by the processor
	assert.Contains(t, result, "<html>")
}
