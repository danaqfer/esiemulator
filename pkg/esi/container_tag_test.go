package esi

import (
	"strings"
	"testing"
)

func TestLoadConfigFromJSON(t *testing.T) {
	jsonData := `{
		"clientId": "test-client",
		"propertyId": "test-property",
		"environment": "development",
		"version": "1.0.0",
		"beacons": [
			{
				"id": "beacon1",
				"name": "Test Beacon 1",
				"url": "https://example.com/beacon1",
				"method": "GET",
				"enabled": true,
				"category": "analytics"
			},
			{
				"id": "beacon2",
				"name": "Test Beacon 2",
				"url": "https://example.com/beacon2",
				"method": "POST",
				"enabled": false,
				"category": "advertising"
			}
		],
		"settings": {
			"maxConcurrentBeacons": 5,
			"defaultTimeout": 3000,
			"fireAndForget": true,
			"maxWait": 0,
			"enableLogging": true
		},
		"macros": {
			"USER_ID": "12345",
			"SITE_ID": "example.com"
		}
	}`

	config, err := LoadConfigFromJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test basic fields
	if config.ClientID != "test-client" {
		t.Errorf("Expected ClientID 'test-client', got '%s'", config.ClientID)
	}
	if config.PropertyID != "test-property" {
		t.Errorf("Expected PropertyID 'test-property', got '%s'", config.PropertyID)
	}
	if config.Environment != "development" {
		t.Errorf("Expected Environment 'development', got '%s'", config.Environment)
	}

	// Test beacons
	if len(config.Beacons) != 2 {
		t.Errorf("Expected 2 beacons, got %d", len(config.Beacons))
	}

	// Test settings
	if config.Settings.MaxConcurrentBeacons != 5 {
		t.Errorf("Expected MaxConcurrentBeacons 5, got %d", config.Settings.MaxConcurrentBeacons)
	}
	if config.Settings.DefaultTimeout != 3000 {
		t.Errorf("Expected DefaultTimeout 3000, got %d", config.Settings.DefaultTimeout)
	}
	if !config.Settings.FireAndForget {
		t.Error("Expected FireAndForget to be true")
	}
	if config.Settings.MaxWait != 0 {
		t.Errorf("Expected MaxWait 0, got %d", config.Settings.MaxWait)
	}

	// Test macros
	if config.Macros["USER_ID"] != "12345" {
		t.Errorf("Expected USER_ID macro '12345', got '%s'", config.Macros["USER_ID"])
	}
}

func TestLoadConfigFromJSONWithDefaults(t *testing.T) {
	jsonData := `{
		"clientId": "test-client",
		"propertyId": "test-property",
		"environment": "development",
		"beacons": []
	}`

	config, err := LoadConfigFromJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test default values
	if config.Settings.DefaultTimeout != 5000 {
		t.Errorf("Expected default timeout 5000, got %d", config.Settings.DefaultTimeout)
	}
	if config.Settings.MaxWait != 0 {
		t.Errorf("Expected default maxWait 0, got %d", config.Settings.MaxWait)
	}
	if config.Settings.DefaultMethod != "GET" {
		t.Errorf("Expected default method 'GET', got '%s'", config.Settings.DefaultMethod)
	}
	if config.Settings.MaxConcurrentBeacons != 10 {
		t.Errorf("Expected default maxConcurrentBeacons 10, got %d", config.Settings.MaxConcurrentBeacons)
	}
}

func TestGenerateESIFromConfig(t *testing.T) {
	config := ContainerTagConfig{
		ClientID:    "test-client",
		PropertyID:  "test-property",
		Environment: "development",
		Beacons: []PartnerBeacon{
			{
				ID:     "beacon1",
				Name:   "Test Beacon 1",
				URL:    "https://example.com/beacon1",
				Method: "GET",
				Parameters: map[string]string{
					"param1": "value1",
					"param2": "value2",
				},
				Enabled:  true,
				Category: "analytics",
			},
			{
				ID:       "beacon2",
				Name:     "Test Beacon 2",
				URL:      "https://example.com/beacon2",
				Method:   "POST",
				Enabled:  false,
				Category: "advertising",
			},
			{
				ID:       "beacon3",
				Name:     "Test Beacon 3",
				URL:      "https://example.com/beacon3",
				Method:   "GET",
				Timeout:  2000,
				Enabled:  true,
				Category: "analytics",
			},
		},
		Settings: ContainerSettings{
			DefaultTimeout:       5000,
			FireAndForget:        true,
			MaxWait:              0,
			EnableLogging:        true,
			EnableErrorHandling:  true,
			DefaultMethod:        "GET",
			MaxConcurrentBeacons: 10,
		},
		Macros: map[string]string{
			"USER_ID": "12345",
			"SITE_ID": "example.com",
		},
	}

	esiProcessor := NewProcessor(Config{Mode: "akamai", Debug: false})
	ctp := NewContainerTagProcessor(config, esiProcessor)

	esiContent, err := ctp.GenerateESIFromConfig()
	if err != nil {
		t.Fatalf("Failed to generate ESI: %v", err)
	}

	// Check that ESI content contains expected elements
	if !strings.Contains(esiContent, "<!--esi Container Tag Generated ESI -->") {
		t.Error("ESI content should contain header comment")
	}
	if !strings.Contains(esiContent, "<!--esi Client: test-client -->") {
		t.Error("ESI content should contain client ID")
	}
	if !strings.Contains(esiContent, "<!--esi Property: test-property -->") {
		t.Error("ESI content should contain property ID")
	}

	// Check for beacon1 include (enabled)
	if !strings.Contains(esiContent, `<esi:include src="https://example.com/beacon1?param1=value1&param2=value2"`) {
		t.Error("ESI content should contain beacon1 include")
	}
	if !strings.Contains(esiContent, `maxwait="0"`) {
		t.Error("ESI content should contain maxwait=\"0\"")
	}
	if !strings.Contains(esiContent, `onerror="continue"`) {
		t.Error("ESI content should contain onerror=\"continue\"")
	}

	// Check for beacon2 include (disabled - should not be present)
	if strings.Contains(esiContent, "beacon2") {
		t.Error("ESI content should not contain disabled beacon2")
	}

	// Check for beacon3 include (enabled with custom timeout)
	if !strings.Contains(esiContent, `<esi:include src="https://example.com/beacon3"`) {
		t.Error("ESI content should contain beacon3 include")
	}
	if !strings.Contains(esiContent, `timeout="2000"`) {
		t.Error("ESI content should contain custom timeout for beacon3")
	}

	// Check for logging comments
	if !strings.Contains(esiContent, `<!-- Beacon: Test Beacon 1 (beacon1) -->`) {
		t.Error("ESI content should contain logging comments when enabled")
	}
}

func TestGenerateESIInclude(t *testing.T) {
	config := ContainerTagConfig{
		ClientID:    "test-client",
		PropertyID:  "test-property",
		Environment: "development",
		Settings: ContainerSettings{
			DefaultTimeout: 5000,
			FireAndForget:  true,
			MaxWait:        0,
			EnableLogging:  true,
		},
		Macros: map[string]string{
			"USER_ID": "12345",
			"SITE_ID": "example.com",
		},
	}

	esiProcessor := NewProcessor(Config{Mode: "akamai", Debug: false})
	ctp := NewContainerTagProcessor(config, esiProcessor)

	beacon := PartnerBeacon{
		ID:     "test-beacon",
		Name:   "Test Beacon",
		URL:    "https://example.com/beacon",
		Method: "GET",
		Parameters: map[string]string{
			"user_id":   "${USER_ID}",
			"site_id":   "${SITE_ID}",
			"client":    "${CLIENT_ID}",
			"timestamp": "${TIMESTAMP}",
		},
		Enabled: true,
	}

	include := ctp.generateESIInclude(beacon)

	// Check basic structure
	if !strings.Contains(include, `<esi:include src="https://example.com/beacon`) {
		t.Error("Include should contain the beacon URL")
	}
	if !strings.Contains(include, `maxwait="0"`) {
		t.Error("Include should contain maxwait=\"0\"")
	}
	if !strings.Contains(include, `onerror="continue"`) {
		t.Error("Include should contain onerror=\"continue\"")
	}
	if !strings.Contains(include, `alt=""`) {
		t.Error("Include should contain alt attribute")
	}

	// Check parameter substitution
	if !strings.Contains(include, `user_id=12345`) {
		t.Error("Include should substitute USER_ID macro")
	}
	if !strings.Contains(include, `site_id=example.com`) {
		t.Error("Include should substitute SITE_ID macro")
	}
	if !strings.Contains(include, `client=test-client`) {
		t.Error("Include should substitute CLIENT_ID macro")
	}

	// Check logging comment
	if !strings.Contains(include, `<!-- Beacon: Test Beacon (test-beacon) -->`) {
		t.Error("Include should contain logging comment")
	}
}

func TestBuildBeaconURL(t *testing.T) {
	config := ContainerTagConfig{
		ClientID:    "test-client",
		PropertyID:  "test-property",
		Environment: "development",
		Macros: map[string]string{
			"USER_ID": "12345",
			"SITE_ID": "example.com",
		},
	}

	esiProcessor := NewProcessor(Config{Mode: "akamai", Debug: false})
	ctp := NewContainerTagProcessor(config, esiProcessor)

	// Test URL with no parameters
	beacon1 := PartnerBeacon{
		URL: "https://example.com/beacon",
	}
	url1 := ctp.buildBeaconURL(beacon1)
	if url1 != "https://example.com/beacon" {
		t.Errorf("Expected URL 'https://example.com/beacon', got '%s'", url1)
	}

	// Test URL with parameters
	beacon2 := PartnerBeacon{
		URL: "https://example.com/beacon",
		Parameters: map[string]string{
			"param1": "value1",
			"param2": "value2",
		},
	}
	url2 := ctp.buildBeaconURL(beacon2)
	expected2 := "https://example.com/beacon?param1=value1&param2=value2"
	if url2 != expected2 {
		t.Errorf("Expected URL '%s', got '%s'", expected2, url2)
	}

	// Test URL with existing query parameters
	beacon3 := PartnerBeacon{
		URL: "https://example.com/beacon?existing=param",
		Parameters: map[string]string{
			"param1": "value1",
		},
	}
	url3 := ctp.buildBeaconURL(beacon3)
	expected3 := "https://example.com/beacon?existing=param&param1=value1"
	if url3 != expected3 {
		t.Errorf("Expected URL '%s', got '%s'", expected3, url3)
	}

	// Test URL with macro substitution
	beacon4 := PartnerBeacon{
		URL: "https://example.com/beacon",
		Parameters: map[string]string{
			"user_id": "${USER_ID}",
			"site_id": "${SITE_ID}",
			"client":  "${CLIENT_ID}",
		},
	}
	url4 := ctp.buildBeaconURL(beacon4)
	if !strings.Contains(url4, "user_id=12345") {
		t.Error("URL should substitute USER_ID macro")
	}
	if !strings.Contains(url4, "site_id=example.com") {
		t.Error("URL should substitute SITE_ID macro")
	}
	if !strings.Contains(url4, "client=test-client") {
		t.Error("URL should substitute CLIENT_ID macro")
	}
}

func TestSubstituteMacros(t *testing.T) {
	config := ContainerTagConfig{
		ClientID:    "test-client",
		PropertyID:  "test-property",
		Environment: "development",
		Macros: map[string]string{
			"USER_ID": "12345",
			"SITE_ID": "example.com",
		},
	}

	esiProcessor := NewProcessor(Config{Mode: "akamai", Debug: false})
	ctp := NewContainerTagProcessor(config, esiProcessor)

	// Test basic macro substitution
	result := ctp.substituteMacros("user_id=${USER_ID}&site_id=${SITE_ID}")
	expected := "user_id=12345&site_id=example.com"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test common macros
	result2 := ctp.substituteMacros("client=${CLIENT_ID}&property=${PROPERTY_ID}&env=${ENVIRONMENT}")
	if !strings.Contains(result2, "client=test-client") {
		t.Error("Should substitute CLIENT_ID")
	}
	if !strings.Contains(result2, "property=test-property") {
		t.Error("Should substitute PROPERTY_ID")
	}
	if !strings.Contains(result2, "env=development") {
		t.Error("Should substitute ENVIRONMENT")
	}

	// Test timestamp and random macros
	result3 := ctp.substituteMacros("timestamp=${TIMESTAMP}&random=${RANDOM}")
	if !strings.Contains(result3, "timestamp=") {
		t.Error("Should substitute TIMESTAMP")
	}
	if !strings.Contains(result3, "random=") {
		t.Error("Should substitute RANDOM")
	}

	// Test no macros
	result4 := ctp.substituteMacros("no_macros_here")
	if result4 != "no_macros_here" {
		t.Errorf("Expected 'no_macros_here', got '%s'", result4)
	}
}

func TestEvaluateConditions(t *testing.T) {
	config := ContainerTagConfig{
		ClientID:    "test-client",
		PropertyID:  "test-property",
		Environment: "development",
	}

	esiProcessor := NewProcessor(Config{Mode: "akamai", Debug: false})
	ctp := NewContainerTagProcessor(config, esiProcessor)

	// Test no conditions
	result := ctp.evaluateConditions(nil)
	if !result {
		t.Error("No conditions should evaluate to true")
	}

	result2 := ctp.evaluateConditions(map[string]string{})
	if !result2 {
		t.Error("Empty conditions should evaluate to true")
	}

	// Test with conditions (currently always returns true)
	conditions := map[string]string{
		"country":   "US",
		"consent":   "required",
		"frequency": "once",
	}
	result3 := ctp.evaluateConditions(conditions)
	if !result3 {
		t.Error("Conditions should evaluate to true (for now)")
	}
}

func TestGenerateCompleteESIHTML(t *testing.T) {
	config := ContainerTagConfig{
		ClientID:    "test-client",
		PropertyID:  "test-property",
		Environment: "development",
		Beacons: []PartnerBeacon{
			{
				ID:      "beacon1",
				Name:    "Test Beacon 1",
				URL:     "https://example.com/beacon1",
				Method:  "GET",
				Enabled: true,
			},
		},
		Settings: ContainerSettings{
			DefaultTimeout: 5000,
			FireAndForget:  true,
			MaxWait:        0,
			EnableLogging:  true,
		},
	}

	esiProcessor := NewProcessor(Config{Mode: "akamai", Debug: false})
	ctp := NewContainerTagProcessor(config, esiProcessor)

	html, err := ctp.GenerateCompleteESIHTML()
	if err != nil {
		t.Fatalf("Failed to generate HTML: %v", err)
	}

	// Check HTML structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("HTML should contain DOCTYPE")
	}
	if !strings.Contains(html, "<html>") {
		t.Error("HTML should contain html tag")
	}
	if !strings.Contains(html, "<head>") {
		t.Error("HTML should contain head tag")
	}
	if !strings.Contains(html, "<body>") {
		t.Error("HTML should contain body tag")
	}
	if !strings.Contains(html, "Container Tag Generated Content") {
		t.Error("HTML should contain container tag content")
	}

	// Check ESI content
	if !strings.Contains(html, `<esi:include src="https://example.com/beacon1"`) {
		t.Error("HTML should contain ESI include")
	}
	if !strings.Contains(html, `maxwait="0"`) {
		t.Error("HTML should contain maxwait=\"0\"")
	}
}

func TestGetBeaconStats(t *testing.T) {
	config := ContainerTagConfig{
		ClientID:    "test-client",
		PropertyID:  "test-property",
		Environment: "development",
		Version:     "1.0.0",
		Beacons: []PartnerBeacon{
			{
				ID:       "beacon1",
				Name:     "Test Beacon 1",
				URL:      "https://example.com/beacon1",
				Enabled:  true,
				Category: "analytics",
			},
			{
				ID:       "beacon2",
				Name:     "Test Beacon 2",
				URL:      "https://example.com/beacon2",
				Enabled:  false,
				Category: "advertising",
			},
			{
				ID:       "beacon3",
				Name:     "Test Beacon 3",
				URL:      "https://example.com/beacon3",
				Enabled:  true,
				Category: "analytics",
			},
		},
	}

	esiProcessor := NewProcessor(Config{Mode: "akamai", Debug: false})
	ctp := NewContainerTagProcessor(config, esiProcessor)

	stats := ctp.GetBeaconStats()

	// Check basic stats
	if stats["totalBeacons"] != 3 {
		t.Errorf("Expected totalBeacons 3, got %v", stats["totalBeacons"])
	}
	if stats["enabledBeacons"] != 2 {
		t.Errorf("Expected enabledBeacons 2, got %v", stats["enabledBeacons"])
	}
	if stats["disabledBeacons"] != 1 {
		t.Errorf("Expected disabledBeacons 1, got %v", stats["disabledBeacons"])
	}

	// Check categories
	categories := stats["categories"].(map[string]int)
	if categories["analytics"] != 2 {
		t.Errorf("Expected 2 analytics beacons, got %d", categories["analytics"])
	}
	if categories["advertising"] != 1 {
		t.Errorf("Expected 1 advertising beacon, got %d", categories["advertising"])
	}

	// Check other fields
	if stats["clientId"] != "test-client" {
		t.Errorf("Expected clientId 'test-client', got '%v'", stats["clientId"])
	}
	if stats["propertyId"] != "test-property" {
		t.Errorf("Expected propertyId 'test-property', got '%v'", stats["propertyId"])
	}
	if stats["environment"] != "development" {
		t.Errorf("Expected environment 'development', got '%v'", stats["environment"])
	}
	if stats["version"] != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%v'", stats["version"])
	}
}

func TestProcessContainerTagESI(t *testing.T) {
	config := ContainerTagConfig{
		ClientID:    "test-client",
		PropertyID:  "test-property",
		Environment: "development",
		Beacons: []PartnerBeacon{
			{
				ID:      "beacon1",
				Name:    "Test Beacon 1",
				URL:     "https://example.com/beacon1",
				Method:  "GET",
				Enabled: true,
			},
		},
		Settings: ContainerSettings{
			DefaultTimeout: 5000,
			FireAndForget:  true,
			MaxWait:        0,
			EnableLogging:  true,
		},
	}

	esiProcessor := NewProcessor(Config{Mode: "akamai", Debug: false})
	ctp := NewContainerTagProcessor(config, esiProcessor)

	esiContent := `<esi:include src="https://example.com/beacon1" maxwait="0" onerror="continue" alt="" />`

	context := ProcessContext{
		BaseURL: "https://example.com",
		Headers: map[string]string{
			"User-Agent": "Test Browser",
		},
		Cookies: map[string]string{
			"session": "test-session",
		},
		Depth: 0,
	}

	result, err := ctp.ProcessContainerTagESI(esiContent, context)
	if err != nil {
		t.Fatalf("Failed to process ESI: %v", err)
	}

	// The result should be empty or contain minimal content since we're using fire-and-forget
	// and the beacon URLs are external
	if result == "" {
		t.Log("ESI processing returned empty result (expected for fire-and-forget)")
	}
}
