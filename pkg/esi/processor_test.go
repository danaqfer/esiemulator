package esi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProcessor(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   Features
	}{
		{
			name: "fastly mode",
			config: Config{
				Mode:  "fastly",
				Debug: false,
			},
			want: Features{
				Include: true,
				Comment: true,
				Remove:  true,
			},
		},
		{
			name: "akamai mode",
			config: Config{
				Mode:  "akamai",
				Debug: false,
			},
			want: Features{
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
			},
		},
		{
			name: "development mode",
			config: Config{
				Mode:  "development",
				Debug: true,
			},
			want: Features{
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewProcessor(tt.config)
			assert.NotNil(t, processor)
			assert.Equal(t, tt.config, processor.GetConfig())
			assert.Equal(t, tt.want, processor.GetFeatures())
			assert.NotNil(t, processor.akamaiExt)
		})
	}
}

func TestProcessor_ProcessComments(t *testing.T) {
	processor := NewProcessor(Config{Mode: "akamai", Debug: false})

	tests := []struct {
		name             string
		input            string
		shouldNotContain []string
		shouldContain    []string
	}{
		{
			name:             "remove esi:comment elements",
			input:            `<html><body><esi:comment text="This is a comment"></esi:comment><p>Content</p></body></html>`,
			shouldNotContain: []string{"esi:comment"},
			shouldContain:    []string{"<p>Content</p>"},
		},
		{
			name:             "remove comment elements",
			input:            `<html><body><comment text="This is a comment"></comment><p>Content</p></body></html>`,
			shouldNotContain: []string{"<comment"},
			shouldContain:    []string{"<p>Content</p>"},
		},
		{
			name:             "multiple comments",
			input:            `<html><body><esi:comment text="Comment 1"></esi:comment><p>Content</p><comment text="Comment 2"></comment></body></html>`,
			shouldNotContain: []string{"esi:comment", "comment text="},
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
		})
	}
}

func TestProcessor_ProcessRemove(t *testing.T) {
	processor := NewProcessor(Config{Mode: "akamai", Debug: false})

	tests := []struct {
		name             string
		input            string
		shouldNotContain []string
		shouldContain    []string
	}{
		{
			name:             "remove esi:remove elements",
			input:            `<html><body><esi:remove><p>Remove this</p></esi:remove><p>Keep this</p></body></html>`,
			shouldNotContain: []string{"esi:remove", "Remove this"},
			shouldContain:    []string{"<p>Keep this</p>"},
		},
		{
			name:             "remove elements",
			input:            `<html><body><remove><p>Remove this</p></remove><p>Keep this</p></body></html>`,
			shouldNotContain: []string{"<remove>", "Remove this"},
			shouldContain:    []string{"<p>Keep this</p>"},
		},
		{
			name:             "multiple remove blocks",
			input:            `<html><body><esi:remove>Block 1</esi:remove><p>Content</p><remove>Block 2</remove></body></html>`,
			shouldNotContain: []string{"esi:remove", "Block 1", "Block 2"},
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
		})
	}
}

func TestProcessor_ProcessCommentBlocks(t *testing.T) {
	processor := NewProcessor(Config{Mode: "akamai", Debug: false})

	tests := []struct {
		name             string
		input            string
		shouldNotContain []string
		shouldContain    []string
	}{
		{
			name:             "convert comment blocks to esi elements",
			input:            `<html><body><!--esi <remove>test</remove> --><p>Content</p></body></html>`,
			shouldNotContain: []string{"<!--esi", "test"},
			shouldContain:    []string{"<p>Content</p>"},
		},
		{
			name:             "multiple comment blocks",
			input:            `<html><body><!--esi <remove>test1</remove> --><!--esi <comment text="test2"></comment> --><p>Content</p></body></html>`,
			shouldNotContain: []string{"<!--esi", "test1", "test2"},
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
		})
	}
}

func TestProcessor_ProcessIncludes(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/fragment.html":
			w.WriteHeader(200)
			w.Write([]byte("<p>Fragment content</p>"))
		case "/error":
			w.WriteHeader(500)
			w.Write([]byte("Internal Server Error"))
		case "/alt-fragment.html":
			w.WriteHeader(200)
			w.Write([]byte("<p>Alternative content</p>"))
		default:
			w.WriteHeader(404)
			w.Write([]byte("Not Found"))
		}
	}))
	defer server.Close()

	processor := NewProcessor(Config{
		Mode:        "akamai",
		Debug:       false,
		MaxIncludes: 10,
		BaseURL:     server.URL,
	})

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "successful include",
			input:    `<html><body><esi:include src="/fragment.html"></esi:include><p>Main content</p></body></html>`,
			expected: `<html><head></head><body><p>Fragment content</p><p>Main content</p></body></html>`,
		},
		{
			name:     "include with alt fallback",
			input:    `<html><body><esi:include src="/error" alt="/alt-fragment.html"></esi:include><p>Main content</p></body></html>`,
			expected: `<html><head></head><body><p>Alternative content</p><p>Main content</p></body></html>`,
		},
		{
			name:     "include with onerror continue",
			input:    `<html><body><esi:include src="/error" onerror="continue"></esi:include><p>Main content</p></body></html>`,
			expected: `<html><head></head><body><p>Main content</p></body></html>`,
		},
		{
			name:     "missing src attribute",
			input:    `<html><body><esi:include></esi:include><p>Main content</p></body></html>`,
			expected: `<html><head></head><body><p>Main content</p></body></html>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := ProcessContext{
				Headers: map[string]string{
					"User-Agent": "ESI-Test/1.0",
				},
				Cookies: make(map[string]string),
			}

			result, err := processor.Process(tt.input, context)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessor_Cache(t *testing.T) {
	// Create a test server with a counter
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(200)
		w.Write([]byte("<p>Fragment content</p>"))
	}))
	defer server.Close()

	processor := NewProcessor(Config{
		Mode:        "akamai",
		Debug:       false,
		MaxIncludes: 10,
		BaseURL:     server.URL,
		Cache: CacheConfig{
			Enabled: true,
			TTL:     60, // 60 seconds
		},
	})

	context := ProcessContext{
		BaseURL: server.URL,
		Headers: make(map[string]string),
		Cookies: make(map[string]string),
	}

	input := `<html><body><esi:include src="/fragment.html"></esi:include></body></html>`
	expected := `<html><head></head><body><p>Fragment content</p></body></html>`

	// First call should hit the server
	result1, err := processor.Process(input, context)
	require.NoError(t, err)
	assert.Equal(t, expected, result1)
	assert.Equal(t, 1, callCount)

	// Second call should use cache
	result2, err := processor.Process(input, context)
	require.NoError(t, err)
	assert.Equal(t, expected, result2)
	assert.Equal(t, 1, callCount) // Should still be 1

	// Check cache statistics
	stats := processor.GetStats()
	assert.Equal(t, int64(2), stats.Requests)
	assert.Equal(t, int64(1), stats.CacheHits)
	assert.Equal(t, int64(1), stats.CacheMiss)

	// Clear cache and verify
	processor.ClearCache()
	assert.Equal(t, 0, processor.GetCacheSize())
}

func TestProcessor_MaxIncludes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("<p>Fragment</p>"))
	}))
	defer server.Close()

	processor := NewProcessor(Config{
		Mode:        "akamai",
		Debug:       false,
		MaxIncludes: 2,
		BaseURL:     server.URL,
	})

	input := `<html><body>
		<esi:include src="/fragment1.html"></esi:include>
		<esi:include src="/fragment2.html"></esi:include>
		<esi:include src="/fragment3.html"></esi:include>
	</body></html>`

	context := ProcessContext{
		BaseURL: server.URL,
		Headers: make(map[string]string),
		Cookies: make(map[string]string),
	}

	result, err := processor.Process(input, context)
	require.NoError(t, err)

	// Should only process the first 2 includes
	fragmentCount := strings.Count(result, "<p>Fragment</p>")
	assert.Equal(t, 2, fragmentCount)
}

func TestProcessor_MaxDepth(t *testing.T) {
	processor := NewProcessor(Config{
		Mode:     "akamai",
		Debug:    false,
		MaxDepth: 2,
	})

	context := ProcessContext{
		Headers: make(map[string]string),
		Cookies: make(map[string]string),
		Depth:   3, // Exceeds max depth
	}

	input := `<html><body><p>Test</p></body></html>`

	_, err := processor.Process(input, context)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum include depth exceeded")
}

func TestProcessor_ResolveURL(t *testing.T) {
	processor := NewProcessor(Config{
		Mode:    "akamai",
		BaseURL: "",
	})

	tests := []struct {
		name     string
		url      string
		baseURL  string
		expected string
		hasError bool
	}{
		{
			name:     "absolute URL",
			url:      "https://other.com/path",
			baseURL:  "",
			expected: "https://other.com/path",
			hasError: false,
		},
		{
			name:     "relative URL with base",
			url:      "/path/to/resource",
			baseURL:  "https://example.com",
			expected: "https://example.com/path/to/resource",
			hasError: false,
		},
		{
			name:     "relative URL with config base",
			url:      "/path/to/resource",
			baseURL:  "",
			expected: "/path/to/resource",
			hasError: false,
		},
		{
			name:     "empty URL",
			url:      "",
			baseURL:  "https://example.com",
			expected: "",
			hasError: true,
		},
		{
			name:     "relative URL without base",
			url:      "/path/to/resource",
			baseURL:  "",
			expected: "/path/to/resource",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.resolveURL(tt.url, tt.baseURL)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestProcessor_Statistics(t *testing.T) {
	processor := NewProcessor(Config{Mode: "akamai", Debug: false})

	// Initial stats should be zero
	stats := processor.GetStats()
	assert.Equal(t, int64(0), stats.Requests)
	assert.Equal(t, int64(0), stats.CacheHits)
	assert.Equal(t, int64(0), stats.CacheMiss)
	assert.Equal(t, int64(0), stats.Errors)

	// Process some content
	context := ProcessContext{
		Headers: make(map[string]string),
		Cookies: make(map[string]string),
	}

	input := `<html><body><p>Test</p></body></html>`
	_, err := processor.Process(input, context)
	require.NoError(t, err)

	// Check that requests counter increased
	stats = processor.GetStats()
	assert.Equal(t, int64(1), stats.Requests)
	assert.True(t, stats.TotalTime >= 0)
}

func TestProcessor_InvalidHTML(t *testing.T) {
	processor := NewProcessor(Config{Mode: "akamai", Debug: false})

	context := ProcessContext{
		Headers: make(map[string]string),
		Cookies: make(map[string]string),
	}

	// This HTML should still be processable by goquery
	input := `<html><body><p>Test</p><unclosed-tag><esi:comment text="test"/></body></html>`

	result, err := processor.Process(input, context)
	require.NoError(t, err)
	assert.Contains(t, result, "<p>Test</p>")
	assert.NotContains(t, result, "esi:comment")
}

func TestProcessor_TruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "string shorter than max",
			input:    "short",
			maxLen:   10,
			expected: "short",
		},
		{
			name:     "string longer than max",
			input:    "this is a very long string",
			maxLen:   10,
			expected: "this is a ...",
		},
		{
			name:     "string exactly max length",
			input:    "exactly10c",
			maxLen:   10,
			expected: "exactly10c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessor_ProcessVars(t *testing.T) {
	tests := []struct {
		name             string
		mode             string
		html             string
		context          ProcessContext
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name: "basic variable substitution",
			mode: "w3c",
			html: `<html><body><esi:vars><p>Host: $(HTTP_HOST)</p></esi:vars></body></html>`,
			context: ProcessContext{
				Headers: map[string]string{
					"Host": "www.example.com",
				},
			},
			shouldContain:    []string{"<p>Host: www.example.com</p>"},
			shouldNotContain: []string{"<esi:vars>", "</esi:vars>", "$(HTTP_HOST)"},
		},
		{
			name: "multiple variables in esi:vars",
			mode: "w3c",
			html: `<html><body><esi:vars><p>Host: $(HTTP_HOST)</p><p>UA: $(HTTP_USER_AGENT)</p></esi:vars></body></html>`,
			context: ProcessContext{
				Headers: map[string]string{
					"Host":       "www.example.com",
					"User-Agent": "Mozilla/5.0",
				},
			},
			shouldContain:    []string{"Host: www.example.com", "UA: Mozilla/5.0"},
			shouldNotContain: []string{"<esi:vars>", "$(HTTP_HOST)", "$(HTTP_USER_AGENT)"},
		},
		{
			name: "cookie variables with keys",
			mode: "w3c",
			html: `<html><body><esi:vars><p>User: $(HTTP_COOKIE{username})</p></esi:vars></body></html>`,
			context: ProcessContext{
				Cookies: map[string]string{
					"username": "john_doe",
				},
			},
			shouldContain:    []string{"User: john_doe"},
			shouldNotContain: []string{"$(HTTP_COOKIE{username})"},
		},
		{
			name: "variable with default value (missing variable)",
			mode: "w3c",
			html: `<html><body><esi:vars><p>User: $(HTTP_COOKIE{missing}|guest)</p></esi:vars></body></html>`,
			context: ProcessContext{
				Cookies: map[string]string{},
			},
			shouldContain:    []string{"User: guest"},
			shouldNotContain: []string{"$(HTTP_COOKIE{missing}|guest)"},
		},
		{
			name: "variable with default value (existing variable)",
			mode: "w3c",
			html: `<html><body><esi:vars><p>User: $(HTTP_COOKIE{username}|guest)</p></esi:vars></body></html>`,
			context: ProcessContext{
				Cookies: map[string]string{
					"username": "jane_doe",
				},
			},
			shouldContain:    []string{"User: jane_doe"},
			shouldNotContain: []string{"guest", "$(HTTP_COOKIE{username}|guest)"},
		},
		{
			name: "variable with quoted default value",
			mode: "w3c",
			html: `<html><body><esi:vars><p>Name: $(HTTP_COOKIE{name}|'Anonymous User')</p></esi:vars></body></html>`,
			context: ProcessContext{
				Cookies: map[string]string{},
			},
			shouldContain:    []string{"Name: &#39;Anonymous User&#39;"},
			shouldNotContain: []string{"$(HTTP_COOKIE{name}|'Anonymous User')"},
		},
		{
			name: "user agent components",
			mode: "w3c",
			html: `<html><body><esi:vars><p>Browser: $(HTTP_USER_AGENT{browser})</p><p>OS: $(HTTP_USER_AGENT{os})</p></esi:vars></body></html>`,
			context: ProcessContext{
				Headers: map[string]string{
					"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
				},
			},
			shouldContain:    []string{"Browser: CHROME", "OS: WIN"},
			shouldNotContain: []string{"$(HTTP_USER_AGENT{browser})", "$(HTTP_USER_AGENT{os})"},
		},
		{
			name: "accept language list check",
			mode: "w3c",
			html: `<html><body><esi:vars><p>Has EN: $(HTTP_ACCEPT_LANGUAGE{en})</p></esi:vars></body></html>`,
			context: ProcessContext{
				Headers: map[string]string{
					"Accept-Language": "en-US,en;q=0.9,es;q=0.8",
				},
			},
			shouldContain:    []string{"Has EN: true"},
			shouldNotContain: []string{"$(HTTP_ACCEPT_LANGUAGE{en})"},
		},
		{
			name: "query string parameter",
			mode: "w3c",
			html: `<html><body><esi:vars><p>ID: $(QUERY_STRING{id})</p></esi:vars></body></html>`,
			context: ProcessContext{
				Headers: map[string]string{
					"Query-String": "id=123&category=tech",
				},
			},
			shouldContain:    []string{"ID: 123"},
			shouldNotContain: []string{"$(QUERY_STRING{id})"},
		},
		{
			name: "multiple esi:vars blocks",
			mode: "w3c",
			html: `<html><body><esi:vars><p>Host: $(HTTP_HOST)</p></esi:vars><esi:vars><p>Method: $(REQUEST_METHOD)</p></esi:vars></body></html>`,
			context: ProcessContext{
				Headers: map[string]string{
					"Host":   "example.com",
					"Method": "POST",
				},
			},
			shouldContain:    []string{"Host: example.com", "Method: POST"},
			shouldNotContain: []string{"<esi:vars>", "$(HTTP_HOST)", "$(REQUEST_METHOD)"},
		},
		{
			name:             "akamai mode with custom variables",
			mode:             "akamai",
			html:             `<html><body><esi:assign name="site_name" value="My Site"></esi:assign><esi:vars><p>Welcome to $(site_name)!</p></esi:vars></body></html>`,
			context:          ProcessContext{},
			shouldContain:    []string{"Welcome to My Site!"},
			shouldNotContain: []string{"<esi:vars>", "$(site_name)", "<esi:assign>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewProcessor(Config{
				Mode:        tt.mode,
				Debug:       false,
				MaxIncludes: 10,
				Cache: CacheConfig{
					Enabled: false,
				},
			})

			result, err := processor.Process(tt.html, tt.context)
			assert.NoError(t, err)

			for _, shouldContain := range tt.shouldContain {
				assert.Contains(t, result, shouldContain, "Result should contain: %s", shouldContain)
			}

			for _, shouldNotContain := range tt.shouldNotContain {
				assert.NotContains(t, result, shouldNotContain, "Result should not contain: %s", shouldNotContain)
			}
		})
	}
}

func TestProcessor_ExpandESIVariables(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		context  ProcessContext
		expected string
	}{
		{
			name:  "simple variable",
			input: "Host: $(HTTP_HOST)",
			context: ProcessContext{
				Headers: map[string]string{
					"Host": "example.com",
				},
			},
			expected: "Host: example.com",
		},
		{
			name:  "variable with key",
			input: "User: $(HTTP_COOKIE{username})",
			context: ProcessContext{
				Cookies: map[string]string{
					"username": "john",
				},
			},
			expected: "User: john",
		},
		{
			name:  "variable with default (missing)",
			input: "User: $(HTTP_COOKIE{missing}|guest)",
			context: ProcessContext{
				Cookies: map[string]string{},
			},
			expected: "User: guest",
		},
		{
			name:  "variable with default (exists)",
			input: "User: $(HTTP_COOKIE{username}|guest)",
			context: ProcessContext{
				Cookies: map[string]string{
					"username": "jane",
				},
			},
			expected: "User: jane",
		},
		{
			name:  "multiple variables",
			input: "$(HTTP_HOST) - $(REQUEST_METHOD)",
			context: ProcessContext{
				Headers: map[string]string{
					"Host":   "example.com",
					"Method": "GET",
				},
			},
			expected: "example.com - GET",
		},
		{
			name:     "unknown variable",
			input:    "Unknown: $(UNKNOWN_VAR)",
			context:  ProcessContext{},
			expected: "Unknown: ",
		},
		{
			name:     "unknown variable with default",
			input:    "Unknown: $(UNKNOWN_VAR|default)",
			context:  ProcessContext{},
			expected: "Unknown: default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := NewProcessor(Config{
				Mode:  "w3c",
				Debug: false,
			})

			result := processor.ExpandESIVariables(tt.input, tt.context)
			assert.Equal(t, tt.expected, result)
		})
	}
}
