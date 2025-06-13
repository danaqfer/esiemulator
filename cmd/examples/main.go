package main

import (
	"fmt"
	"log"
	"time"

	"esi-simulator/pkg/esi"
)

func main() {
	fmt.Println("🚀 ESI Simulator - Go Examples")

	// Run different examples
	if err := basicExample(); err != nil {
		log.Printf("Basic example failed: %v", err)
	}

	if err := advancedExample(); err != nil {
		log.Printf("Advanced example failed: %v", err)
	}

	if err := performanceExample(); err != nil {
		log.Printf("Performance example failed: %v", err)
	}

	// Run Akamai-specific examples
	if err := akamaiExtensionsExample(); err != nil {
		log.Printf("Akamai extensions example failed: %v", err)
	}

	if err := compareModes(); err != nil {
		log.Printf("Mode comparison example failed: %v", err)
	}

	if err := performanceComparison(); err != nil {
		log.Printf("Performance comparison example failed: %v", err)
	}

	fmt.Println("\n🎉 All examples completed!")
	fmt.Println("💡 To test interactively, run: go run main.go")
}

func basicExample() error {
	fmt.Println("📊 Basic ESI Processing Example")
	fmt.Println(repeat("═", 50))

	// Example ESI content
	esiContent := `<!DOCTYPE html>
<html xmlns:esi="http://www.edge-delivery.org/esi/1.0">
<head>
    <title>ESI Example</title>
</head>
<body>
    <h1>Welcome to ESI Testing</h1>
    
    <!-- Basic include -->
    <esi:include src="https://httpbin.org/html" />
    
    <!-- Comment (will be removed) -->
    <esi:comment text="This is an ESI comment" />
    
    <!-- Content that shows when ESI is not processed -->
    <esi:remove>
        <p>This content is only shown when ESI is NOT processed</p>
    </esi:remove>
    
    <p>This is regular HTML content.</p>
</body>
</html>`

	// Test different modes
	modes := []string{"fastly", "akamai"}

	for _, mode := range modes {
		fmt.Printf("\n🎯 Testing %s mode:\n", mode)

		// Create processor for this mode
		processor := esi.NewProcessor(esi.Config{
			Mode:        mode,
			Debug:       true,
			MaxIncludes: 256,
			MaxDepth:    5,
			Cache: esi.CacheConfig{
				Enabled: true,
				TTL:     60,
			},
		})

		// Create processing context
		context := esi.ProcessContext{
			BaseURL: "https://example.com",
			Headers: map[string]string{
				"User-Agent": "ESI-Simulator-Go/1.0",
				"Cookie":     "user_type=premium; session_id=abc123",
			},
			Cookies: map[string]string{
				"user_type":  "premium",
				"session_id": "abc123",
			},
			Depth: 0,
		}

		startTime := time.Now()
		result, err := processor.Process(esiContent, context)
		processingTime := time.Since(startTime)

		if err != nil {
			fmt.Printf("❌ Error in %s mode: %v\n", mode, err)
			continue
		}

		fmt.Printf("✅ Processing completed in %v\n", processingTime)

		stats := processor.GetStats()
		fmt.Printf("📈 Stats: Requests=%d, CacheHits=%d, CacheMiss=%d, Errors=%d\n",
			stats.Requests, stats.CacheHits, stats.CacheMiss, stats.Errors)

		features := processor.GetFeatures()
		fmt.Printf("🎯 Supported features: Include=%t, Choose=%t, Try=%t, Vars=%t\n",
			features.Include, features.Choose, features.Try, features.Vars)

		// Show a preview of the result
		preview := result
		if len(result) > 200 {
			preview = result[:200] + "..."
		}
		fmt.Printf("📄 Result preview:\n%s\n", preview)
	}

	return nil
}

func advancedExample() error {
	fmt.Println("\n\n🔧 Advanced ESI Example (Akamai mode with conditionals)")
	fmt.Println(repeat("═", 60))

	advancedEsiContent := `<!DOCTYPE html>
<html xmlns:esi="http://www.edge-delivery.org/esi/1.0">
<head>
    <title>Advanced ESI Example</title>
</head>
<body>
    <h1>Advanced ESI Features</h1>
    
    <!-- Conditional content based on cookie -->
    <esi:choose>
        <esi:when test="$(HTTP_COOKIE{user_type})=='premium'">
            <div class="premium-content">
                <h2>Premium User Content</h2>
                <p>Welcome, premium user!</p>
            </div>
        </esi:when>
        <esi:when test="$(HTTP_COOKIE{user_type})=='basic'">
            <div class="basic-content">
                <h2>Basic User Content</h2>
                <p>Welcome, basic user!</p>
            </div>
        </esi:when>
        <esi:otherwise>
            <div class="guest-content">
                <h2>Guest Content</h2>
                <p>Please log in for personalized content.</p>
            </div>
        </esi:otherwise>
    </esi:choose>
    
    <!-- Error handling example -->
    <esi:try>
        <esi:attempt>
            <esi:include src="https://httpbin.org/status/404" />
        </esi:attempt>
        <esi:except>
            <div class="fallback">
                <p>Sorry, the content could not be loaded. Here's some fallback content.</p>
            </div>
        </esi:except>
    </esi:try>
    
    <!-- Variable substitution -->
    <esi:vars>
        <p>Your User-Agent: $(HTTP_USER_AGENT)</p>
        <p>Request Host: $(HTTP_HOST)</p>
    </esi:vars>
</body>
</html>`

	processor := esi.NewProcessor(esi.Config{
		Mode:  "akamai",
		Debug: true,
		Cache: esi.CacheConfig{
			Enabled: true,
			TTL:     300,
		},
	})

	context := esi.ProcessContext{
		BaseURL: "https://example.com",
		Headers: map[string]string{
			"User-Agent": "Mozilla/5.0 (ESI-Simulator-Go)",
			"Cookie":     "user_type=premium; session_id=xyz789",
			"Host":       "www.example.com",
		},
		Cookies: map[string]string{
			"user_type":  "premium",
			"session_id": "xyz789",
		},
		Depth: 0,
	}

	_, err := processor.Process(advancedEsiContent, context)
	if err != nil {
		return fmt.Errorf("advanced example error: %w", err)
	}

	fmt.Println("✅ Advanced processing completed")
	stats := processor.GetStats()
	fmt.Printf("📈 Stats: Requests=%d, CacheHits=%d, CacheMiss=%d, Errors=%d, TotalTime=%d\n",
		stats.Requests, stats.CacheHits, stats.CacheMiss, stats.Errors, stats.TotalTime)

	fmt.Println("\n📝 Note: Advanced features like conditionals and variables are")
	fmt.Println("   currently placeholder implementations. They will be fully")
	fmt.Println("   implemented in the next development phase.")

	return nil
}

func performanceExample() error {
	fmt.Println("\n\n⚡ Performance Testing Example")
	fmt.Println(repeat("═", 40))

	processor := esi.NewProcessor(esi.Config{
		Mode:  "fastly",
		Debug: false, // Disable debug for performance testing
		Cache: esi.CacheConfig{
			Enabled: true,
			TTL:     30,
		},
	})

	simpleEsi := `<html>
<body>
    <h1>Performance Test</h1>
    <esi:include src="https://httpbin.org/delay/1" />
    <esi:include src="https://httpbin.org/json" />
    <p>End of content</p>
</body>
</html>`

	context := esi.ProcessContext{
		BaseURL: "https://example.com",
		Headers: map[string]string{
			"User-Agent": "ESI-Simulator-Performance-Test",
		},
		Depth: 0,
	}

	fmt.Println("🏃 Running performance test (first run - no cache):")
	startTime := time.Now()
	_, err := processor.Process(simpleEsi, context)
	firstRun := time.Since(startTime)

	if err != nil {
		fmt.Printf("⚠️  First run failed: %v\n", err)
	}

	fmt.Println("🏃 Running performance test (second run - with cache):")
	startTime = time.Now()
	_, err = processor.Process(simpleEsi, context)
	secondRun := time.Since(startTime)

	if err != nil {
		fmt.Printf("⚠️  Second run failed: %v\n", err)
	}

	improvement := float64(firstRun-secondRun) / float64(firstRun) * 100

	fmt.Printf("📊 Performance Results:\n")
	fmt.Printf("  - First run (no cache): %v\n", firstRun)
	fmt.Printf("  - Second run (cached):  %v\n", secondRun)
	fmt.Printf("  - Cache improvement:    %.1f%%\n", improvement)

	stats := processor.GetStats()
	fmt.Printf("  - Final stats: Requests=%d, CacheHits=%d, CacheMiss=%d, Errors=%d, TotalTime=%d\n",
		stats.Requests, stats.CacheHits, stats.CacheMiss, stats.Errors, stats.TotalTime)

	return nil
}

// Helper function since strings.Repeat doesn't exist in basic Go
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// Akamai Extensions Example
func akamaiExtensionsExample() error {
	fmt.Println("\n\n🔧 Akamai ESI Extensions Example")
	fmt.Println(repeat("═", 50))

	// Create processor in Akamai mode
	processor := esi.NewProcessor(esi.Config{
		Mode:        "akamai",
		Debug:       true,
		MaxIncludes: 256,
		MaxDepth:    5,
		Cache: esi.CacheConfig{
			Enabled: true,
			TTL:     300,
		},
	})

	// Advanced ESI content with Akamai extensions
	akamaiESI := `<!DOCTYPE html>
<html xmlns:esi="http://www.edge-delivery.org/esi/1.0">
<head>
    <title>Akamai ESI Extensions Demo</title>
</head>
<body>
    <h1>Akamai ESI Extensions Showcase</h1>
    
    <!-- Variable Assignment -->
    <h2>1. Variable Assignment (esi:assign)</h2>
    <esi:assign name="user_level" value="premium" />
    <esi:assign name="site_name">My Awesome Site</esi:assign>
    <esi:assign name="greeting" value="Hello $(HTTP_COOKIE{username})" />
    
    <!-- Expression Evaluation -->
    <h2>2. Expression Evaluation (esi:eval)</h2>
    <p>User level: <esi:eval expr="$(user_level)" /></p>
    <p>Is premium user: <esi:eval expr="$(user_level) == 'premium'" /></p>
    <p>Site: <esi:eval expr="$(site_name)" /></p>
    
    <!-- Built-in Functions -->
    <h2>3. Built-in Functions (esi:function)</h2>
    <p>URL Encoded Site: <esi:function name="url_encode" input="$(site_name)" /></p>
    <p>Base64 Encoded: <esi:function name="base64_encode" input="Hello World" /></p>
    <p>String Length: <esi:function name="strlen" input="$(site_name)" /></p>
    <p>Substring: <esi:function name="substr" input="$(site_name)" start="0" length="2" /></p>
    <p>Random Number: <esi:function name="random" min="1" max="100" /></p>
    <p>Current Time: <esi:function name="time" format="2006-01-02 15:04:05" /></p>
    
    <!-- Dictionary Lookup -->
    <h2>4. Dictionary Lookup (esi:dictionary)</h2>
    <p>User Type: <esi:dictionary src="user_types" key="$(HTTP_COOKIE{user_id})" default="guest" /></p>
    
    <!-- Geo Variables -->
    <h2>5. Geo-location Variables</h2>
    <p>Country: $(GEO_COUNTRY_NAME) ($(GEO_COUNTRY_CODE))</p>
    <p>Region: $(GEO_REGION)</p>
    <p>City: $(GEO_CITY)</p>
    <p>Client IP: $(CLIENT_IP)</p>
    
    <!-- Extended Include Features -->
    <h2>6. Extended Include Features</h2>
    <esi:include src="/fragments/user-profile" 
                 timeout="5000" 
                 cacheable="true" 
                 method="GET" 
                 onerror="continue" />
    
    <!-- Debug Information -->
    <h2>7. Debug Information (esi:debug)</h2>
    <esi:debug type="vars" />
    <esi:debug type="headers" />
    <esi:debug type="cookies" />
    <esi:debug type="time" />
    
    <!-- Variable Expansion in Content -->
    <h2>8. Variable Expansion</h2>
    <p>Welcome to $(site_name)!</p>
    <p>User Agent Browser: $(HTTP_USER_AGENT{browser})</p>
    <p>Operating System: $(HTTP_USER_AGENT{os})</p>
    <p>Request Method: $(REQUEST_METHOD)</p>
    
    <!-- Complex Expressions -->
    <h2>9. Complex Expressions</h2>
    <esi:assign name="is_mobile" value="$(HTTP_USER_AGENT{os}) == 'mobile'" />
    <esi:eval expr="$(is_mobile) == 'true'" />
</body>
</html>`

	// Create context with sample data
	context := esi.ProcessContext{
		BaseURL: "https://example.com",
		Headers: map[string]string{
			"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
			"Cookie":          "username=john_doe; user_id=12345; user_type=premium",
			"Host":            "www.example.com",
			"Referer":         "https://google.com/search?q=test",
			"Accept-Language": "en-US,en;q=0.9,es;q=0.8",
			"X-Forwarded-For": "203.0.113.42",
			"Method":          "GET",
			"Request-URI":     "/test-page",
		},
		Cookies: map[string]string{
			"username":  "john_doe",
			"user_id":   "12345",
			"user_type": "premium",
		},
		Depth: 0,
	}

	fmt.Println("🔄 Processing Akamai ESI extensions...")
	startTime := time.Now()
	result, err := processor.Process(akamaiESI, context)
	processingTime := time.Since(startTime)

	if err != nil {
		return fmt.Errorf("Akamai extensions example failed: %w", err)
	}

	fmt.Printf("✅ Processing completed in %v\n", processingTime)

	// Show features
	features := processor.GetFeatures()
	fmt.Printf("🎯 Akamai Features Available:\n")
	fmt.Printf("  - Variable Assignment: %t\n", features.Assign)
	fmt.Printf("  - Expression Evaluation: %t\n", features.Eval)
	fmt.Printf("  - Built-in Functions: %t\n", features.Function)
	fmt.Printf("  - Dictionary Lookup: %t\n", features.Dictionary)
	fmt.Printf("  - Debug Output: %t\n", features.Debug)
	fmt.Printf("  - Geo Variables: %t\n", features.GeoVariables)
	fmt.Printf("  - Extended Variables: %t\n", features.ExtendedVars)

	// Show stats
	stats := processor.GetStats()
	fmt.Printf("📈 Processing Stats: Requests=%d, CacheHits=%d, CacheMiss=%d, Errors=%d, TotalTime=%d\n",
		stats.Requests, stats.CacheHits, stats.CacheMiss, stats.Errors, stats.TotalTime)

	// Show a preview of the result
	preview := result
	if len(result) > 500 {
		preview = result[:500] + "..."
	}
	fmt.Printf("📄 Result preview:\n%s\n", preview)

	return nil
}

func compareModes() error {
	fmt.Println("\n\n⚡ Mode Comparison Example")
	fmt.Println(repeat("═", 40))

	// Simple ESI content to test across modes
	testESI := `<html>
<body>
    <h1>ESI Mode Comparison</h1>
    
    <!-- Basic features (supported by all) -->
    <esi:include src="https://httpbin.org/json" />
    <esi:comment text="This is a comment" />
    <esi:remove>This shows only when ESI is not processed</esi:remove>
    
    <!-- Akamai extensions (only in Akamai mode) -->
    <esi:assign name="test_var" value="Akamai Extension" />
    <esi:eval expr="$(test_var)" />
    <esi:function name="time" />
    <p>Geo: $(GEO_COUNTRY_CODE)</p>
</body>
</html>`

	modes := []string{"fastly", "akamai"}
	context := esi.ProcessContext{
		BaseURL: "https://example.com",
		Headers: map[string]string{
			"User-Agent": "ESI-Test/1.0",
		},
		Depth: 0,
	}

	for _, mode := range modes {
		fmt.Printf("\n🎯 Testing %s mode:\n", mode)

		processor := esi.NewProcessor(esi.Config{
			Mode:        mode,
			Debug:       false, // Reduce noise for comparison
			MaxIncludes: 10,
			MaxDepth:    3,
			Cache: esi.CacheConfig{
				Enabled: true,
				TTL:     60,
			},
		})

		startTime := time.Now()
		_, err := processor.Process(testESI, context)
		processingTime := time.Since(startTime)

		if err != nil {
			fmt.Printf("❌ Error in %s mode: %v\n", mode, err)
			continue
		}

		features := processor.GetFeatures()
		fmt.Printf("✅ Processing completed in %v\n", processingTime)
		fmt.Printf("📊 Features: Include=%t, Assign=%t, Function=%t, GeoVars=%t\n",
			features.Include, features.Assign, features.Function, features.GeoVariables)

		// Count how many Akamai extensions were processed
		akamaiFeatures := 0
		if features.Assign {
			akamaiFeatures++
		}
		if features.Function {
			akamaiFeatures++
		}
		if features.GeoVariables {
			akamaiFeatures++
		}

		fmt.Printf("🎯 Akamai extensions available: %d\n", akamaiFeatures)
	}

	return nil
}

func performanceComparison() error {
	fmt.Println("\n\n⚡ Performance Impact of Akamai Extensions")
	fmt.Println(repeat("═", 50))

	// Create test content with varying complexity
	testCases := map[string]string{
		"Basic ESI":      `<html><body><esi:include src="https://httpbin.org/json" /><esi:comment text="test" /></body></html>`,
		"With Variables": `<html><body><esi:assign name="test" value="hello" /><esi:eval expr="$(test)" /><p>$(HTTP_HOST)</p></body></html>`,
		"With Functions": `<html><body><esi:function name="time" /><esi:function name="random" min="1" max="100" /><esi:function name="base64_encode" input="test" /></body></html>`,
		"Complex": `<html><body>
			<esi:assign name="user" value="$(HTTP_COOKIE{user})" />
			<esi:eval expr="$(user) != ''" />
			<esi:function name="url_encode" input="$(user)" />
			<p>Geo: $(GEO_COUNTRY_CODE) - $(GEO_CITY)</p>
			<esi:debug type="vars" />
		</body></html>`,
	}

	context := esi.ProcessContext{
		BaseURL: "https://example.com",
		Headers: map[string]string{
			"User-Agent": "ESI-Performance-Test/1.0",
			"Cookie":     "user=test_user",
			"Host":       "www.example.com",
		},
		Cookies: map[string]string{
			"user": "test_user",
		},
		Depth: 0,
	}

	processor := esi.NewProcessor(esi.Config{
		Mode:        "akamai",
		Debug:       false,
		MaxIncludes: 50,
		MaxDepth:    3,
		Cache: esi.CacheConfig{
			Enabled: true,
			TTL:     300,
		},
	})

	fmt.Printf("%-15s | %-12s | %-10s\n", "Test Case", "Time", "Features")
	fmt.Printf("%s\n", repeat("-", 40))

	for name, content := range testCases {
		startTime := time.Now()
		_, err := processor.Process(content, context)
		duration := time.Since(startTime)

		status := "✅"
		if err != nil {
			status = "❌"
		}

		fmt.Printf("%-15s | %-12v | %s\n", name, duration, status)
	}

	return nil
}
