package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestProcessContainerConfig_BasicFunctionality(t *testing.T) {
	config := ContainerConfig{
		Pixels: []Pixel{
			{
				ID:   "test1",
				URL:  "https://example.com/pixel.gif?evid=~~evid~~&time=~~r~~",
				TYPE: "dir",
			},
			{
				ID:   "test2",
				URL:  "https://example.com/iframe.html",
				TYPE: "frm",
			},
			{
				ID:   "test3",
				URL:  "https://example.com/script.js",
				TYPE: "script",
			},
		},
	}

	esiConfig := ESIConfig{
		BrowserVars: false,
		MaxWait:     0,
	}

	esiContent, browserConfig, err := ProcessContainerConfig(config, esiConfig)
	if err != nil {
		t.Fatalf("ProcessContainerConfig failed: %v", err)
	}

	// Check that ESI content contains the dir pixel
	if !strings.Contains(esiContent, "https://example.com/pixel.gif?evid=$(PMUSER_EVID)&time=$(TIME)") {
		t.Errorf("ESI content should contain processed dir pixel URL")
	}

	// Check that browser config contains frm and script pixels
	if len(browserConfig.Pixels) != 2 {
		t.Errorf("Expected 2 browser pixels, got %d", len(browserConfig.Pixels))
	}

	// Verify browser pixels are frm and script types
	browserTypes := make(map[string]bool)
	for _, pixel := range browserConfig.Pixels {
		browserTypes[pixel.TYPE] = true
	}
	if !browserTypes["frm"] || !browserTypes["script"] {
		t.Errorf("Browser config should contain frm and script pixels")
	}
}

func TestProcessMacros_BasicMacros(t *testing.T) {
	esiConfig := ESIConfig{BrowserVars: false}

	testCases := []struct {
		input    string
		expected string
	}{
		{"~~r~~", "$(TIME)"},
		{"~~evid~~", "$(PMUSER_EVID)"},
		{"~~cs~~", "$(HTTP_COOKIE{consent})"},
		{"~~cc~~", "$(GEO_COUNTRY)"},
		{"~~uu~~", "$(PMUSER_UU)"},
		{"~~suu~~", "$(PMUSER_SUU)"},
	}

	for _, tc := range testCases {
		result, err := processMacros(tc.input, esiConfig)
		if err != nil {
			t.Errorf("processMacros failed for %s: %v", tc.input, err)
		}
		if result != tc.expected {
			t.Errorf("processMacros(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestProcessMacros_CookieMacros(t *testing.T) {
	esiConfig := ESIConfig{BrowserVars: false}

	testCases := []struct {
		input    string
		expected string
	}{
		{"~~c~userid~~", "$(HTTP_COOKIE{userid})"},
		{"~~c~sessionid~hpr~path~~", "$(PMUSER_COOKIE_HASH_PR_{sessionid}_{path})"},
		{"~~c~sessionid~hpo~domain~~", "$(PMUSER_COOKIE_HASH_PO_{sessionid}_{domain})"},
	}

	for _, tc := range testCases {
		result, err := processMacros(tc.input, esiConfig)
		if err != nil {
			t.Errorf("processMacros failed for %s: %v", tc.input, err)
		}
		if result != tc.expected {
			t.Errorf("processMacros(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestProcessMacros_DecodeMacros(t *testing.T) {
	esiConfig := ESIConfig{BrowserVars: false}

	testCases := []struct {
		input    string
		expected string
	}{
		{"~~dl:qs~~", "$(PMUSER_DECODED_QUERY_STRING)"},
		{"~~dl:qs~utm_source~~", "$(PMUSER_DECODED_QS_{utm_source})"},
		{"~~dl:qs~campaign~~", "$(PMUSER_DECODED_QS_{campaign})"},
	}

	for _, tc := range testCases {
		result, err := processMacros(tc.input, esiConfig)
		if err != nil {
			t.Errorf("processMacros failed for %s: %v", tc.input, err)
		}
		if result != tc.expected {
			t.Errorf("processMacros(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestProcessMacros_UserVariables(t *testing.T) {
	esiConfig := ESIConfig{BrowserVars: false}

	testCases := []struct {
		input    string
		expected string
	}{
		{"~~u1~~", "$(PMUSER_V1)"},
		{"~~u2~~", "$(PMUSER_V2)"},
		{"~~u10~~", "$(PMUSER_V10)"},
		{"~~customvar~~", "$(PMUSER_CUSTOMVAR)"},
	}

	for _, tc := range testCases {
		result, err := processMacros(tc.input, esiConfig)
		if err != nil {
			t.Errorf("processMacros failed for %s: %v", tc.input, err)
		}
		if result != tc.expected {
			t.Errorf("processMacros(%s) = %s, expected %s", tc.input, result, tc.expected)
		}
	}
}

func TestProcessMacros_ComplexURL(t *testing.T) {
	esiConfig := ESIConfig{BrowserVars: false}

	input := "https://example.com/pixel.gif?evid=~~evid~~&time=~~r~~&country=~~cc~~&user=~~uu~~&fingerprint=~~suu~~&cookie=~~c~userid~~&hash=~~c~session~hpr~path~~&decoded=~~dl:qs~utm_source~~"
	expected := "https://example.com/pixel.gif?evid=$(PMUSER_EVID)&time=$(TIME)&country=$(GEO_COUNTRY)&user=$(PMUSER_UU)&fingerprint=$(PMUSER_SUU)&cookie=$(HTTP_COOKIE{userid})&hash=$(PMUSER_COOKIE_HASH_PR_{session}_{path})&decoded=$(PMUSER_DECODED_QS_{utm_source})"

	result, err := processMacros(input, esiConfig)
	if err != nil {
		t.Fatalf("processMacros failed: %v", err)
	}

	if result != expected {
		t.Errorf("processMacros failed for complex URL")
		t.Errorf("Expected: %s", expected)
		t.Errorf("Got:      %s", result)
	}
}

func TestGenerateFingerprintID(t *testing.T) {
	ip := "192.168.1.1"
	accept := "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"

	// Test comprehensive fingerprinting
	fingerprint := GenerateFingerprintID(ip, accept, userAgent)

	// Check that fingerprint is a valid MD5 hash (32 hex characters)
	if len(fingerprint) != 32 {
		t.Errorf("Fingerprint should be 32 characters long, got %d", len(fingerprint))
	}

	// Check that it's a valid hex string
	for _, char := range fingerprint {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			t.Errorf("Fingerprint should contain only hex characters, got %c", char)
		}
	}

	// Test that same inputs produce same fingerprint
	fingerprint2 := GenerateFingerprintID(ip, accept, userAgent)
	if fingerprint != fingerprint2 {
		t.Errorf("Same inputs should produce same fingerprint")
	}

	// Test that different inputs produce different fingerprints
	fingerprint3 := GenerateFingerprintID("192.168.1.2", accept, userAgent)
	if fingerprint == fingerprint3 {
		t.Errorf("Different inputs should produce different fingerprints")
	}
}

func TestGenerateFingerprintID_EmptyValues(t *testing.T) {
	// Test with empty values
	fingerprint := GenerateFingerprintID("", "", "")

	// Should still generate a valid hash even with empty values
	if len(fingerprint) != 32 {
		t.Errorf("Fingerprint with empty values should be 32 characters long, got %d", len(fingerprint))
	}

	// Test with partial values
	fingerprint2 := GenerateFingerprintID("192.168.1.1", "", "")
	if len(fingerprint2) != 32 {
		t.Errorf("Fingerprint with partial values should be 32 characters long, got %d", len(fingerprint2))
	}
}

func TestGenerateCookieHash(t *testing.T) {
	cookieValue := "test_cookie_value"
	salt := "test_salt"

	// Test hpr (path + cookie)
	hashHpr := GenerateCookieHash(cookieValue, salt, "hpr")
	expectedHpr := GenerateFingerprintID(salt+cookieValue, "", "")
	if hashHpr != expectedHpr {
		t.Errorf("HPR hash mismatch: got %s, expected %s", hashHpr, expectedHpr)
	}

	// Test hpo (cookie + path)
	hashHpo := GenerateCookieHash(cookieValue, salt, "hpo")
	expectedHpo := GenerateFingerprintID(cookieValue+salt, "", "")
	if hashHpo != expectedHpo {
		t.Errorf("HPO hash mismatch: got %s, expected %s", hashHpo, expectedHpo)
	}

	// Test default (cookie only)
	hashDefault := GenerateCookieHash(cookieValue, salt, "unknown")
	expectedDefault := GenerateFingerprintID(cookieValue, "", "")
	if hashDefault != expectedDefault {
		t.Errorf("Default hash mismatch: got %s, expected %s", hashDefault, expectedDefault)
	}
}

func TestURLDecode(t *testing.T) {
	testCases := []struct {
		encoded  string
		expected string
	}{
		{"hello%20world", "hello world"},
		{"test%2Bvalue", "test+value"},
		{"%3Fquery%3Dvalue", "?query=value"},
		{"no%20encoding", "no encoding"},
		{"", ""},
	}

	for _, tc := range testCases {
		result, err := URLDecode(tc.encoded)
		if err != nil {
			t.Errorf("URLDecode failed for %s: %v", tc.encoded, err)
		}
		if result != tc.expected {
			t.Errorf("URLDecode(%s) = %s, expected %s", tc.encoded, result, tc.expected)
		}
	}
}

func TestGenerateESIFunctions(t *testing.T) {
	functions := GenerateESIFunctions()

	// Check that essential functions are present
	expectedFunctions := []string{
		"generate_suu",
		"cookie_hash",
		"url_decode",
		"query_param_decode",
		"default_value",
	}

	for _, funcName := range expectedFunctions {
		if !strings.Contains(functions, funcName) {
			t.Errorf("ESI functions should contain %s function", funcName)
		}
	}

	// Check that it's valid ESI syntax
	if !strings.Contains(functions, "<esi:function") {
		t.Errorf("ESI functions should contain <esi:function tags")
	}
}

func TestProcessContainerConfig_DefaultValues(t *testing.T) {
	config := ContainerConfig{
		Pixels: []Pixel{
			{
				ID:   "test_defaults",
				URL:  "https://example.com/pixel.gif",
				TYPE: "", // Empty type should default to "dir"
			},
		},
	}

	esiConfig := ESIConfig{BrowserVars: false}
	esiContent, browserConfig, err := ProcessContainerConfig(config, esiConfig)
	if err != nil {
		t.Fatalf("ProcessContainerConfig failed: %v", err)
	}

	// Check that empty TYPE defaults to "dir" and gets converted to ESI
	if !strings.Contains(esiContent, "https://example.com/pixel.gif") {
		t.Errorf("Empty TYPE should default to 'dir' and be converted to ESI")
	}

	if len(browserConfig.Pixels) != 0 {
		t.Errorf("Empty TYPE should not be added to browser config")
	}
}

func TestProcessContainerConfig_JSONRoundTrip(t *testing.T) {
	originalConfig := ContainerConfig{
		Pixels: []Pixel{
			{
				ID:   "test1",
				URL:  "https://example.com/pixel.gif",
				TYPE: "frm",
				REQ:  true,
				PCT:  80,
				CAP:  2,
				RC:   "analytics",
			},
			{
				ID:     "test2",
				URL:    "https://example.com/script.js",
				TYPE:   "script",
				SCRIPT: "console.log('test');",
			},
		},
	}

	esiConfig := ESIConfig{BrowserVars: false}
	_, browserConfig, err := ProcessContainerConfig(originalConfig, esiConfig)
	if err != nil {
		t.Fatalf("ProcessContainerConfig failed: %v", err)
	}

	// Test JSON marshaling/unmarshaling of browser config
	jsonData, err := json.Marshal(browserConfig)
	if err != nil {
		t.Fatalf("JSON marshaling failed: %v", err)
	}

	var unmarshaledConfig ContainerConfig
	err = json.Unmarshal(jsonData, &unmarshaledConfig)
	if err != nil {
		t.Fatalf("JSON unmarshaling failed: %v", err)
	}

	// Verify the round trip preserved the data
	if len(unmarshaledConfig.Pixels) != len(browserConfig.Pixels) {
		t.Errorf("JSON round trip changed pixel count: got %d, expected %d",
			len(unmarshaledConfig.Pixels), len(browserConfig.Pixels))
	}

	for i, pixel := range unmarshaledConfig.Pixels {
		if pixel.ID != browserConfig.Pixels[i].ID {
			t.Errorf("JSON round trip changed pixel ID: got %s, expected %s",
				pixel.ID, browserConfig.Pixels[i].ID)
		}
		if pixel.TYPE != browserConfig.Pixels[i].TYPE {
			t.Errorf("JSON round trip changed pixel TYPE: got %s, expected %s",
				pixel.TYPE, browserConfig.Pixels[i].TYPE)
		}
	}
}

func TestProcessMacros_ErrorHandling(t *testing.T) {
	esiConfig := ESIConfig{BrowserVars: false}

	// Test empty macro
	result, err := processMacros("~~", esiConfig)
	if err != nil {
		t.Errorf("Empty macro should not cause error")
	}
	if result != "~~" {
		t.Errorf("Empty macro should return unchanged")
	}

	// Test malformed macro (should return original)
	result, err = processMacros("~~invalid~macro~with~too~many~parts~~", esiConfig)
	if err != nil {
		t.Errorf("Malformed macro should not cause error")
	}
	if !strings.Contains(result, "$(PMUSER_INVALID)") {
		t.Errorf("Malformed macro should be processed as simple variable")
	}
}
