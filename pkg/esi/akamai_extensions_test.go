package esi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAkamaiExtensions(t *testing.T) {
	config := Config{Mode: "akamai", Debug: false}
	processor := NewProcessor(config)

	akamaiExt := NewAkamaiExtensions(processor)
	assert.NotNil(t, akamaiExt)
	assert.NotNil(t, akamaiExt.processor)
	assert.NotNil(t, akamaiExt.variables)
	assert.Equal(t, 0, len(akamaiExt.variables))
}

func TestAkamaiExtensions_ProcessAssign(t *testing.T) {
	config := Config{Mode: "akamai", Debug: false}
	processor := NewProcessor(config)

	tests := []struct {
		name             string
		input            string
		shouldNotContain []string
		shouldContain    []string
		vars             map[string]string
	}{
		{
			name:             "assign with value attribute",
			input:            `<html><body><esi:assign name="test" value="hello"></esi:assign><p>Content</p></body></html>`,
			shouldNotContain: []string{"esi:assign"},
			shouldContain:    []string{"<p>Content</p>"},
			vars:             map[string]string{"test": "hello"},
		},
		{
			name:             "assign with element content",
			input:            `<html><body><esi:assign name="test">world</esi:assign><p>Content</p></body></html>`,
			shouldNotContain: []string{"esi:assign"},
			shouldContain:    []string{"<p>Content</p>"},
			vars:             map[string]string{"test": "world"},
		},
		{
			name:             "assign with variable expansion",
			input:            `<html><body><esi:assign name="host" value="$(HTTP_HOST)"></esi:assign><p>Content</p></body></html>`,
			shouldNotContain: []string{"esi:assign"},
			shouldContain:    []string{"<p>Content</p>"},
			vars:             map[string]string{"host": "example.com"},
		},
		{
			name:             "assign without name attribute",
			input:            `<html><body><esi:assign value="test"></esi:assign><p>Content</p></body></html>`,
			shouldNotContain: []string{"esi:assign"},
			shouldContain:    []string{"<p>Content</p>"},
			vars:             map[string]string{},
		},
		{
			name:             "multiple assignments",
			input:            `<html><body><esi:assign name="var1" value="value1"></esi:assign><esi:assign name="var2" value="value2"></esi:assign><p>Content</p></body></html>`,
			shouldNotContain: []string{"esi:assign"},
			shouldContain:    []string{"<p>Content</p>"},
			vars:             map[string]string{"var1": "value1", "var2": "value2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor.akamaiExt.variables = make(map[string]string) // Reset variables
			context := ProcessContext{
				Headers: map[string]string{
					"Host": "example.com",
				},
				Cookies: make(map[string]string),
			}

			result, err := processor.Process(tt.input, context)
			require.NoError(t, err)

			for _, shouldNotContain := range tt.shouldNotContain {
				assert.NotContains(t, result, shouldNotContain)
			}
			for _, shouldContain := range tt.shouldContain {
				assert.Contains(t, result, shouldContain)
			}

			// Check assigned variables
			for key, expectedValue := range tt.vars {
				actualValue := processor.akamaiExt.variables[key]
				assert.Equal(t, expectedValue, actualValue, "Variable %s should have value %s", key, expectedValue)
			}
		})
	}
}

func TestAkamaiExtensions_ProcessEval(t *testing.T) {
	config := Config{Mode: "akamai", Debug: false}
	processor := NewProcessor(config)

	tests := []struct {
		name             string
		input            string
		shouldNotContain []string
		shouldContain    []string
		setup            func(*AkamaiExtensions)
	}{
		{
			name:             "eval simple expression",
			input:            `<html><body><esi:eval expr="'hello' == 'hello'"></esi:eval><p>Content</p></body></html>`,
			shouldNotContain: []string{"esi:eval"},
			shouldContain:    []string{"true", "<p>Content</p>"},
			setup:            func(ext *AkamaiExtensions) {},
		},
		{
			name:             "eval inequality expression",
			input:            `<html><body><esi:eval expr="'hello' != 'world'"></esi:eval><p>Content</p></body></html>`,
			shouldNotContain: []string{"esi:eval"},
			shouldContain:    []string{"true", "<p>Content</p>"},
			setup:            func(ext *AkamaiExtensions) {},
		},
		{
			name:             "eval with variable",
			input:            `<html><body><esi:eval expr="$(TEST_VAR) == 'test'"></esi:eval><p>Content</p></body></html>`,
			shouldNotContain: []string{"esi:eval"},
			shouldContain:    []string{"true", "<p>Content</p>"},
			setup: func(ext *AkamaiExtensions) {
				ext.variables["TEST_VAR"] = "test"
			},
		},
		{
			name:             "eval without expr attribute",
			input:            `<html><body><esi:eval></esi:eval><p>Content</p></body></html>`,
			shouldNotContain: []string{"esi:eval"},
			shouldContain:    []string{"<p>Content</p>"},
			setup:            func(ext *AkamaiExtensions) {},
		},
		{
			name:             "eval false expression",
			input:            `<html><body><esi:eval expr="'hello' == 'world'"></esi:eval><p>Content</p></body></html>`,
			shouldNotContain: []string{"esi:eval"},
			shouldContain:    []string{"false", "<p>Content</p>"},
			setup:            func(ext *AkamaiExtensions) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor.akamaiExt.variables = make(map[string]string)
			tt.setup(processor.akamaiExt)

			context := ProcessContext{
				Headers: make(map[string]string),
				Cookies: make(map[string]string),
			}

			result, err := processor.Process(tt.input, context)
			require.NoError(t, err)

			for _, shouldNotContain := range tt.shouldNotContain {
				assert.NotContains(t, result, shouldNotContain)
			}
			for _, shouldContain := range tt.shouldContain {
				assert.Contains(t, result, shouldContain)
			}
		})
	}
}

func TestAkamaiExtensions_ProcessFunction(t *testing.T) {
	config := Config{Mode: "akamai", Debug: false}
	processor := NewProcessor(config)

	tests := []struct {
		name             string
		input            string
		shouldNotContain []string
		shouldContain    []string
		timeTest         bool
	}{
		{
			name:             "base64_encode function",
			input:            `<html><body><esi:function name="base64_encode" input="hello"></esi:function><p>Content</p></body></html>`,
			shouldNotContain: []string{"esi:function"},
			shouldContain:    []string{"aGVsbG8=", "<p>Content</p>"},
		},
		{
			name:             "base64_decode function",
			input:            `<html><body><esi:function name="base64_decode" input="aGVsbG8="></esi:function><p>Content</p></body></html>`,
			shouldNotContain: []string{"esi:function"},
			shouldContain:    []string{"hello", "<p>Content</p>"},
		},
		{
			name:             "url_encode function",
			input:            `<html><body><esi:function name="url_encode" input="hello world"></esi:function><p>Content</p></body></html>`,
			shouldNotContain: []string{"esi:function"},
			shouldContain:    []string{"hello+world", "<p>Content</p>"},
		},
		{
			name:             "url_decode function",
			input:            `<html><body><esi:function name="url_decode" input="hello+world"></esi:function><p>Content</p></body></html>`,
			shouldNotContain: []string{"esi:function"},
			shouldContain:    []string{"hello world", "<p>Content</p>"},
		},
		{
			name:             "strlen function",
			input:            `<html><body><esi:function name="strlen" input="hello"></esi:function><p>Content</p></body></html>`,
			shouldNotContain: []string{"esi:function"},
			shouldContain:    []string{"5", "<p>Content</p>"},
		},
		{
			name:             "substr function",
			input:            `<html><body><esi:function name="substr" input="hello world" start="0" length="5"></esi:function><p>Content</p></body></html>`,
			shouldNotContain: []string{"esi:function"},
			shouldContain:    []string{"hello", "<p>Content</p>"},
		},
		{
			name:             "random function",
			input:            `<html><body><esi:function name="random" min="1" max="1"></esi:function><p>Content</p></body></html>`,
			shouldNotContain: []string{"esi:function"},
			shouldContain:    []string{"1", "<p>Content</p>"},
		},
		{
			name:             "time function",
			input:            `<html><body><esi:function name="time" format="2006"></esi:function><p>Content</p></body></html>`,
			shouldNotContain: []string{"esi:function"},
			shouldContain:    []string{"<p>Content</p>"},
			timeTest:         true,
		},
		{
			name:             "unknown function",
			input:            `<html><body><esi:function name="unknown"></esi:function><p>Content</p></body></html>`,
			shouldNotContain: []string{"esi:function"},
			shouldContain:    []string{"<p>Content</p>"},
		},
		{
			name:             "function without name attribute",
			input:            `<html><body><esi:function input="test"></esi:function><p>Content</p></body></html>`,
			shouldNotContain: []string{"esi:function"},
			shouldContain:    []string{"<p>Content</p>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := ProcessContext{
				Headers: make(map[string]string),
				Cookies: make(map[string]string),
			}

			result, err := processor.Process(tt.input, context)
			require.NoError(t, err)

			for _, shouldNotContain := range tt.shouldNotContain {
				assert.NotContains(t, result, shouldNotContain)
			}
			for _, shouldContain := range tt.shouldContain {
				assert.Contains(t, result, shouldContain)
			}

			if tt.timeTest {
				// Special check for time function - should contain 4 digits (year)
				assert.Regexp(t, `\d{4}`, result)
			}
		})
	}
}

func TestAkamaiExtensions_ProcessDictionary(t *testing.T) {
	config := Config{Mode: "akamai", Debug: false}
	processor := NewProcessor(config)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "dictionary with default value",
			input:    `<html><body><esi:dictionary src="/dict.txt" key="test" default="default_value"></esi:dictionary><p>Content</p></body></html>`,
			expected: `<html><head></head><body>default_value<p>Content</p></body></html>`,
		},
		{
			name:     "dictionary without default value",
			input:    `<html><body><esi:dictionary src="/dict.txt" key="test"></esi:dictionary><p>Content</p></body></html>`,
			expected: `<html><head></head><body><p>Content</p></body></html>`,
		},
		{
			name:     "dictionary missing src attribute",
			input:    `<html><body><esi:dictionary key="test"></esi:dictionary><p>Content</p></body></html>`,
			expected: `<html><head></head><body><p>Content</p></body></html>`,
		},
		{
			name:     "dictionary missing key attribute",
			input:    `<html><body><esi:dictionary src="/dict.txt"></esi:dictionary><p>Content</p></body></html>`,
			expected: `<html><head></head><body><p>Content</p></body></html>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := ProcessContext{
				Headers: make(map[string]string),
				Cookies: make(map[string]string),
			}

			result, err := processor.Process(tt.input, context)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAkamaiExtensions_ProcessDebug(t *testing.T) {
	config := Config{Mode: "akamai", Debug: true}
	processor := NewProcessor(config)

	tests := []struct {
		name     string
		input    string
		contains []string
		setup    func(*AkamaiExtensions, ProcessContext) ProcessContext
	}{
		{
			name:     "debug vars type",
			input:    `<html><body><esi:debug type="vars"></esi:debug><p>Content</p></body></html>`,
			contains: []string{"<!-- ESI DEBUG: Variables:", "testvar=testvalue", "-->"},
			setup: func(ext *AkamaiExtensions, ctx ProcessContext) ProcessContext {
				ext.variables["testvar"] = "testvalue"
				return ctx
			},
		},
		{
			name:     "debug headers type",
			input:    `<html><body><esi:debug type="headers"></esi:debug><p>Content</p></body></html>`,
			contains: []string{"<!-- ESI DEBUG: Headers:", "User-Agent=TestAgent", "-->"},
			setup: func(ext *AkamaiExtensions, ctx ProcessContext) ProcessContext {
				ctx.Headers["User-Agent"] = "TestAgent"
				return ctx
			},
		},
		{
			name:     "debug cookies type",
			input:    `<html><body><esi:debug type="cookies"></esi:debug><p>Content</p></body></html>`,
			contains: []string{"<!-- ESI DEBUG: Cookies:", "session=abc123", "-->"},
			setup: func(ext *AkamaiExtensions, ctx ProcessContext) ProcessContext {
				ctx.Cookies["session"] = "abc123"
				return ctx
			},
		},
		{
			name:     "debug time type",
			input:    `<html><body><esi:debug type="time"></esi:debug><p>Content</p></body></html>`,
			contains: []string{"<!-- ESI DEBUG:", "T", "-->"},
			setup:    func(ext *AkamaiExtensions, ctx ProcessContext) ProcessContext { return ctx },
		},
		{
			name:     "debug with custom content",
			input:    `<html><body><esi:debug>Custom debug message</esi:debug><p>Content</p></body></html>`,
			contains: []string{"<!-- ESI DEBUG: Custom debug message", "-->"},
			setup:    func(ext *AkamaiExtensions, ctx ProcessContext) ProcessContext { return ctx },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor.akamaiExt.variables = make(map[string]string)
			context := ProcessContext{
				Headers: make(map[string]string),
				Cookies: make(map[string]string),
			}
			context = tt.setup(processor.akamaiExt, context)

			result, err := processor.Process(tt.input, context)
			require.NoError(t, err)

			for _, contains := range tt.contains {
				assert.Contains(t, result, contains)
			}
		})
	}
}

func TestAkamaiExtensions_ProcessDebugDisabled(t *testing.T) {
	config := Config{Mode: "akamai", Debug: false}
	processor := NewProcessor(config)

	input := `<html><body><esi:debug type="vars"></esi:debug><p>Content</p></body></html>`
	expected := `<html><head></head><body><p>Content</p></body></html>`

	context := ProcessContext{
		Headers: make(map[string]string),
		Cookies: make(map[string]string),
	}

	result, err := processor.Process(input, context)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestAkamaiExtensions_ExpandVariables(t *testing.T) {
	config := Config{Mode: "akamai", Debug: false}
	processor := NewProcessor(config)
	akamaiExt := processor.akamaiExt

	// Set up some variables
	akamaiExt.variables["CUSTOM_VAR"] = "custom_value"

	context := ProcessContext{
		Headers: map[string]string{
			"Host":         "example.com",
			"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/91.0.4472.124",
			"Cookie":       "session=abc123; user=john",
			"Referer":      "https://google.com",
			"Query-String": "param1=value1&param2=value2",
			"Method":       "GET",
			"Request-URI":  "/path/to/resource",
		},
		Cookies: map[string]string{
			"session": "abc123",
			"user":    "john",
		},
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "custom variable",
			input:    "$(CUSTOM_VAR)",
			expected: "custom_value",
		},
		{
			name:     "HTTP_HOST",
			input:    "$(HTTP_HOST)",
			expected: "example.com",
		},
		{
			name:     "HTTP_USER_AGENT",
			input:    "$(HTTP_USER_AGENT)",
			expected: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/91.0.4472.124",
		},
		{
			name:     "HTTP_USER_AGENT with browser component",
			input:    "$(HTTP_USER_AGENT{browser})",
			expected: "CHROME",
		},
		{
			name:     "HTTP_USER_AGENT with os component",
			input:    "$(HTTP_USER_AGENT{os})",
			expected: "WIN",
		},
		{
			name:     "HTTP_COOKIE",
			input:    "$(HTTP_COOKIE)",
			expected: "session=abc123; user=john",
		},
		{
			name:     "HTTP_COOKIE with key",
			input:    "$(HTTP_COOKIE{session})",
			expected: "abc123",
		},
		{
			name:     "HTTP_REFERER",
			input:    "$(HTTP_REFERER)",
			expected: "https://google.com",
		},
		{
			name:     "QUERY_STRING",
			input:    "$(QUERY_STRING)",
			expected: "param1=value1&param2=value2",
		},
		{
			name:     "QUERY_STRING with parameter",
			input:    "$(QUERY_STRING{param1})",
			expected: "value1",
		},
		{
			name:     "REQUEST_METHOD",
			input:    "$(REQUEST_METHOD)",
			expected: "GET",
		},
		{
			name:     "REQUEST_URI",
			input:    "$(REQUEST_URI)",
			expected: "/path/to/resource",
		},
		{
			name:     "GEO_COUNTRY_CODE",
			input:    "$(GEO_COUNTRY_CODE)",
			expected: "US",
		},
		{
			name:     "GEO_COUNTRY_NAME",
			input:    "$(GEO_COUNTRY_NAME)",
			expected: "United States",
		},
		{
			name:     "GEO_REGION",
			input:    "$(GEO_REGION)",
			expected: "California",
		},
		{
			name:     "GEO_CITY",
			input:    "$(GEO_CITY)",
			expected: "San Francisco",
		},
		{
			name:     "CLIENT_IP with X-Forwarded-For",
			input:    "$(CLIENT_IP)",
			expected: "",
		},
		{
			name:     "unknown variable",
			input:    "$(UNKNOWN_VAR)",
			expected: "",
		},
		{
			name:     "multiple variables",
			input:    "Host: $(HTTP_HOST), Custom: $(CUSTOM_VAR)",
			expected: "Host: example.com, Custom: custom_value",
		},
		{
			name:     "no variables",
			input:    "Just plain text",
			expected: "Just plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := akamaiExt.expandVariables(tt.input, context)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAkamaiExtensions_GetUserAgentComponent(t *testing.T) {
	config := Config{Mode: "akamai", Debug: false}
	processor := NewProcessor(config)
	akamaiExt := processor.akamaiExt

	tests := []struct {
		name      string
		userAgent string
		component string
		expected  string
	}{
		{
			name:      "Chrome browser",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/91.0.4472.124",
			component: "browser",
			expected:  "CHROME",
		},
		{
			name:      "Firefox browser",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Firefox/89.0",
			component: "browser",
			expected:  "FIREFOX",
		},
		{
			name:      "Safari browser",
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Safari/537.36",
			component: "browser",
			expected:  "SAFARI",
		},
		{
			name:      "Edge browser",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Edge/91.0.864.59",
			component: "browser",
			expected:  "EDGE",
		},
		{
			name:      "Unknown browser",
			userAgent: "SomeOtherBrowser/1.0",
			component: "browser",
			expected:  "OTHER",
		},
		{
			name:      "Windows OS",
			userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/91.0.4472.124",
			component: "os",
			expected:  "WIN",
		},
		{
			name:      "Mac OS",
			userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Safari/537.36",
			component: "os",
			expected:  "MAC",
		},
		{
			name:      "Linux OS",
			userAgent: "Mozilla/5.0 (X11; Linux x86_64) Chrome/91.0.4472.124",
			component: "os",
			expected:  "UNIX",
		},
		{
			name:      "Unknown OS",
			userAgent: "SomeOS Browser/1.0",
			component: "os",
			expected:  "OTHER",
		},
		{
			name:      "Version component",
			userAgent: "Mozilla/5.0 Chrome/91.0.4472.124",
			component: "version",
			expected:  "1.0",
		},
		{
			name:      "Unknown component",
			userAgent: "Mozilla/5.0 Chrome/91.0.4472.124",
			component: "unknown",
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := akamaiExt.getUserAgentComponent(tt.userAgent, tt.component)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAkamaiExtensions_HasLanguage(t *testing.T) {
	config := Config{Mode: "akamai", Debug: false}
	processor := NewProcessor(config)
	akamaiExt := processor.akamaiExt

	tests := []struct {
		name       string
		acceptLang string
		lang       string
		expected   string
	}{
		{
			name:       "language present",
			acceptLang: "en-US,en;q=0.9,fr;q=0.8",
			lang:       "en",
			expected:   "true",
		},
		{
			name:       "language not present",
			acceptLang: "en-US,en;q=0.9,fr;q=0.8",
			lang:       "de",
			expected:   "false",
		},
		{
			name:       "empty accept language",
			acceptLang: "",
			lang:       "en",
			expected:   "false",
		},
		{
			name:       "specific locale present",
			acceptLang: "en-US,en;q=0.9,fr;q=0.8",
			lang:       "en-US",
			expected:   "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := akamaiExt.hasLanguage(tt.acceptLang, tt.lang)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAkamaiExtensions_GetQueryParam(t *testing.T) {
	config := Config{Mode: "akamai", Debug: false}
	processor := NewProcessor(config)
	akamaiExt := processor.akamaiExt

	tests := []struct {
		name        string
		queryString string
		key         string
		expected    string
	}{
		{
			name:        "existing parameter",
			queryString: "param1=value1&param2=value2",
			key:         "param1",
			expected:    "value1",
		},
		{
			name:        "non-existing parameter",
			queryString: "param1=value1&param2=value2",
			key:         "param3",
			expected:    "",
		},
		{
			name:        "empty query string",
			queryString: "",
			key:         "param1",
			expected:    "",
		},
		{
			name:        "URL encoded value",
			queryString: "param1=hello%20world",
			key:         "param1",
			expected:    "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := akamaiExt.getQueryParam(tt.queryString, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAkamaiExtensions_EvaluateExpression(t *testing.T) {
	config := Config{Mode: "akamai", Debug: false}
	processor := NewProcessor(config)
	akamaiExt := processor.akamaiExt

	// Set up a variable for testing
	akamaiExt.variables["TEST_VAR"] = "test_value"

	context := ProcessContext{
		Headers: map[string]string{
			"Host": "example.com",
		},
		Cookies: make(map[string]string),
	}

	tests := []struct {
		name     string
		expr     string
		expected string
	}{
		{
			name:     "simple equality true",
			expr:     "'hello' == 'hello'",
			expected: "true",
		},
		{
			name:     "simple equality false",
			expr:     "'hello' == 'world'",
			expected: "false",
		},
		{
			name:     "simple inequality true",
			expr:     "'hello' != 'world'",
			expected: "true",
		},
		{
			name:     "simple inequality false",
			expr:     "'hello' != 'hello'",
			expected: "false",
		},
		{
			name:     "variable expansion in expression",
			expr:     "$(TEST_VAR) == 'test_value'",
			expected: "true",
		},
		{
			name:     "HTTP variable in expression",
			expr:     "$(HTTP_HOST) == 'example.com'",
			expected: "true",
		},
		{
			name:     "no comparison operator",
			expr:     "$(HTTP_HOST)",
			expected: "example.com",
		},
		{
			name:     "plain text",
			expr:     "just text",
			expected: "just text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := akamaiExt.evaluateExpression(tt.expr, context)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAkamaiExtensions_Integration(t *testing.T) {
	config := Config{Mode: "akamai", Debug: false}
	processor := NewProcessor(config)

	input := `<html><body>
		<esi:assign name="site_name" value="MyWebsite"></esi:assign>
		<esi:assign name="user_browser" value="$(HTTP_USER_AGENT{browser})"></esi:assign>
		<h1>Welcome to $(site_name)</h1>
		<p>Your browser: <esi:eval expr="$(user_browser) == 'CHROME'"></esi:eval></p>
		<p>Encoded message: <esi:function name="base64_encode" input="Hello World"></esi:function></p>
		<esi:comment text="This comment should be removed"></esi:comment>
		<esi:remove>This content should be removed</esi:remove>
	</body></html>`

	context := ProcessContext{
		Headers: map[string]string{
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/91.0.4472.124",
		},
		Cookies: make(map[string]string),
	}

	result, err := processor.Process(input, context)
	require.NoError(t, err)

	// Check that variables were processed
	assert.Contains(t, result, "Welcome to MyWebsite")
	assert.Contains(t, result, "Your browser: true")
	assert.Contains(t, result, "Encoded message: SGVsbG8gV29ybGQ=")

	// Check that ESI elements were removed
	assert.NotContains(t, result, "esi:assign")
	assert.NotContains(t, result, "esi:eval")
	assert.NotContains(t, result, "esi:function")
	assert.NotContains(t, result, "esi:comment")
	assert.NotContains(t, result, "esi:remove")
	assert.NotContains(t, result, "This content should be removed")
}
