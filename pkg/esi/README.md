# ESI Emulator

A comprehensive Edge Side Include (ESI) emulator written in Go that supports both Fastly's limited ESI functionality and Akamai's extended ESI capabilities.

## Overview

Edge Side Includes (ESI) is a markup language that allows content assembly at the edge of the network, enabling dynamic content composition while maintaining high performance through selective caching. This emulator aims to provide a complete implementation that supports:

### Fastly ESI Support (Limited)
- `<esi:include>` - Basic content inclusion
- `<esi:comment>` - Comments 
- `<esi:remove>` - Conditional content removal

### Akamai ESI Support (Extended)
- Full ESI 1.0 specification compliance
- `<esi:include>` with `alt` and `onerror` attributes
- `<esi:inline>` - Embedded fragment storage
- `<esi:choose>`, `<esi:when>`, `<esi:otherwise>` - Conditional processing
- `<esi:try>`, `<esi:attempt>`, `<esi:except>` - Error handling
- `<esi:comment>` - Developer comments
- `<esi:remove>` - Content removal
- `<esi:vars>` - Variable substitution
- `<!--esi ...-->` - HTML comment ESI blocks
- ESI Variables and Expressions
- Variable substructure access (dictionaries and lists)

## Documentation

All ESI specifications and implementation guides are stored in the `resources/` directory:

- W3C ESI Language Specification 1.0
- Akamai ESI Developer's Guide
- Fastly ESI Implementation Details
- Edge Delivery Overview

## Architecture

The ESI emulator is built in Go and follows standard Go project structure:

- **Core ESI Processor** (`processor.go`): Handles parsing and processing of ESI tags with comprehensive tests
- **Akamai Extensions** (`akamai_extensions.go`): Extended functionality for Akamai-specific features
- **Comprehensive Tests**: Full test coverage with 140+ passing tests

Key features:
- Concurrent HTTP request handling with Gin web framework
- Thread-safe caching with configurable TTL
- Multiple ESI implementation modes (Fastly, Akamai, W3C)
- Built-in test fragments and examples
- Comprehensive error handling and logging

## Implementation Details

### Architecture

The ESI processor is built with a modular architecture:

- **Core Processor** (`processor.go`) - Main ESI processing engine
- **Akamai Extensions** (`akamai_extensions.go`) - Extended functionality
- **Cache System** - In-memory caching with TTL expiration
- **Statistics** - Request tracking and performance metrics

### Processing Pipeline

1. **Parse HTML** - Convert input to DOM using goquery
2. **Process Comment Blocks** - Handle `<!--esi ... -->` syntax  
3. **Process Extensions** - Apply Akamai-specific extensions (if enabled)
4. **Process Standard Elements** - Handle includes, conditionals, variables
5. **Generate Output** - Convert DOM back to HTML

### Performance Considerations

- **Concurrent Processing** - Thread-safe operations with mutex protection
- **Intelligent Caching** - Configurable TTL with cache hit/miss tracking
- **Resource Limits** - Configurable maximum includes and depth limits
- **Error Handling** - Graceful degradation with fallback support

## Akamai ESI Extensions

### Variable Assignment (`<esi:assign>`)

Assign values to custom variables for later use:

```xml
<!-- Direct value assignment -->
<esi:assign name="user_level" value="premium" />
<esi:assign name="api_key" value="$(HTTP_COOKIE{api_key})" />

<!-- Content-based assignment -->
<esi:assign name="welcome_msg">
    Welcome to our $(HTTP_HOST) website!
</esi:assign>

<!-- Use assigned variables -->
<p>User level: $(user_level)</p>
```

### Expression Evaluation (`<esi:eval>`)

Evaluate expressions and output the result:

```xml
<!-- Simple expression -->
<esi:eval expr="$(user_level)" />

<!-- Boolean comparison -->
<esi:eval expr="$(user_level) == 'premium'" />

<!-- Complex expression with variables -->
<esi:eval expr="$(HTTP_HOST) != ''" />
```

### Built-in Functions (`<esi:function>`)

Execute utility functions for data manipulation:

```xml
<!-- Encoding functions -->
<esi:function name="base64_encode" input="Hello World" />
<esi:function name="base64_decode" input="SGVsbG8gV29ybGQ=" />
<esi:function name="url_encode" input="hello world!" />
<esi:function name="url_decode" input="hello%20world%21" />

<!-- String functions -->
<esi:function name="strlen" input="$(HTTP_HOST)" />
<esi:function name="substr" input="$(HTTP_HOST)" start="0" length="3" />

<!-- Utility functions -->
<esi:function name="random" min="1" max="100" />
<esi:function name="time" format="2006-01-02 15:04:05" />
```

### Dictionary Lookup (`<esi:dictionary>`)

Perform key-value lookups (simplified implementation):

```xml
<esi:dictionary src="user_types" 
                key="$(HTTP_COOKIE{user_id})" 
                default="guest" />
```

### Debug Output (`<esi:debug>`)

Generate debugging information during development:

```xml
<!-- Debug specific information -->
<esi:debug type="vars" />    <!-- Show all variables -->
<esi:debug type="headers" /> <!-- Show HTTP headers -->
<esi:debug type="cookies" /> <!-- Show cookies -->
<esi:debug type="time" />    <!-- Show current timestamp -->

<!-- Custom debug content -->
<esi:debug>Custom debug: $(HTTP_HOST)</esi:debug>
```

### Extended Variables

Akamai mode supports additional ESI variables:

```xml
<!-- Geo-location variables -->
$(GEO_COUNTRY_CODE)   <!-- US -->
$(GEO_COUNTRY_NAME)   <!-- United States -->
$(GEO_REGION)         <!-- California -->
$(GEO_CITY)           <!-- San Francisco -->

<!-- Request information -->
$(REQUEST_METHOD)     <!-- GET, POST, etc. -->
$(REQUEST_URI)        <!-- /path/to/page -->
$(CLIENT_IP)          <!-- Client IP address -->

<!-- Enhanced user agent parsing -->
$(HTTP_USER_AGENT{browser})  <!-- CHROME, FIREFOX, etc. -->
$(HTTP_USER_AGENT{os})       <!-- WIN, MAC, UNIX -->
$(HTTP_USER_AGENT{version})  <!-- Browser version -->
```

### Enhanced Include Attributes

Akamai mode supports additional attributes on `<esi:include>`:

```xml
<esi:include src="/api/data" 
             timeout="5000"     <!-- Timeout in milliseconds -->
             cacheable="true"   <!-- Cache control -->
             method="GET"       <!-- HTTP method -->
             onerror="continue" />
```

## ESI Conditional Processing (`<esi:choose>`)

The emulator provides comprehensive conditional processing support through the `<esi:choose>`, `<esi:when>`, and `<esi:otherwise>` elements:

### Basic Conditional Logic

```xml
<esi:choose>
    <esi:when test="$(HTTP_HOST) == 'example.com'">
        <p>Welcome to Example.com!</p>
    </esi:when>
    <esi:otherwise>
        <p>Welcome to our site!</p>
    </esi:otherwise>
</esi:choose>
```

### Multiple Conditions

```xml
<esi:choose>
    <esi:when test="$(HTTP_HOST) == 'example.com'">
        <p>Example.com content</p>
    </esi:when>
    <esi:when test="$(HTTP_HOST) == 'test.com'">
        <p>Test.com content</p>
    </esi:when>
    <esi:when test="$(HTTP_COOKIE{logged_in})">
        <p>Welcome back, user!</p>
    </esi:when>
    <esi:otherwise>
        <p>Default content</p>
    </esi:otherwise>
</esi:choose>
```

### Supported Expression Operators

- **Equality**: `==` (e.g., `$(HTTP_HOST) == 'example.com'`)
- **Inequality**: `!=` (e.g., `$(HTTP_HOST) != 'example.com'`)
- **Boolean**: Direct variable evaluation (e.g., `$(HTTP_COOKIE{logged_in})`)

### Complex Conditions with Variables

```xml
<esi:choose>
    <esi:when test="$(HTTP_USER_AGENT{browser}) == 'Chrome'">
        <p>Chrome-specific content</p>
    </esi:when>
    <esi:when test="$(HTTP_ACCEPT_LANGUAGE{en})">
        <p>English content</p>
    </esi:when>
    <esi:when test="$(QUERY_STRING{debug}) == 'true'">
        <p>Debug information</p>
    </esi:when>
    <esi:otherwise>
        <p>Fallback content</p>
    </esi:otherwise>
</esi:choose>
```

## ESI Error Handling (`<esi:try>`)

The emulator provides robust error handling through the `<esi:try>`, `<esi:attempt>`, and `<esi:except>` elements:

### Basic Error Handling

```xml
<esi:try>
    <esi:attempt>
        <esi:include src="/fragments/header" />
    </esi:attempt>
    <esi:except>
        <p>Header could not be loaded</p>
    </esi:except>
</esi:try>
```

### Error Handling with Variables

```xml
<esi:try>
    <esi:attempt>
        <esi:vars>
            <p>Host: $(HTTP_HOST)</p>
            <p>User: $(HTTP_COOKIE{user_id})</p>
        </esi:vars>
    </esi:attempt>
    <esi:except>
        <p>Variable processing failed</p>
    </esi:except>
</esi:try>
```

### Nested Error Handling

```xml
<esi:try>
    <esi:attempt>
        <esi:try>
            <esi:attempt>
                <esi:include src="/fragments/critical" />
            </esi:attempt>
            <esi:except>
                <esi:include src="/fragments/fallback" />
            </esi:except>
        </esi:try>
    </esi:attempt>
    <esi:except>
        <p>All content failed to load</p>
    </esi:except>
</esi:try>
```

### Error Handling Best Practices

1. **Graceful Degradation**: Always provide meaningful fallback content
2. **Nested Protection**: Use nested try blocks for critical vs. non-critical content
3. **Variable Safety**: Wrap variable processing in try blocks when using complex expressions
4. **Include Protection**: Protect external includes that might fail

## ESI Comment Block Processing (`<!--esi ... -->`)

The emulator provides comprehensive support for ESI comment blocks, allowing ESI elements to be embedded within HTML comments for graceful degradation:

### Basic Comment Block Usage

```xml
<!--esi <esi:vars>Host: $(HTTP_HOST)</esi:vars> -->
<p>Content</p>
```

### Comment Blocks with Variables

```xml
<!--esi 
    <esi:vars>
        <p>User Agent: $(HTTP_USER_AGENT)</p>
        <p>Referer: $(HTTP_REFERER)</p>
    </esi:vars>
-->
<p>Main content</p>
```

### Comment Blocks with Conditional Logic

```xml
<!--esi 
    <esi:choose>
        <esi:when test="$(HTTP_HOST) == 'example.com'">
            <p>Welcome to Example.com!</p>
        </esi:when>
        <esi:otherwise>
            <p>Welcome to our site!</p>
        </esi:otherwise>
    </esi:choose>
-->
<p>Content</p>
```

### Comment Blocks with Error Handling

```xml
<!--esi 
    <esi:try>
        <esi:attempt>
            <esi:include src="/fragments/header" />
        </esi:attempt>
        <esi:except>
            <p>Header could not be loaded</p>
        </esi:except>
    </esi:try>
-->
<p>Content</p>
```

### Complex Comment Blocks with Multiple Elements

```xml
<!--esi 
    <esi:vars>Host: $(HTTP_HOST)</esi:vars>
    <esi:choose>
        <esi:when test="$(HTTP_COOKIE{logged_in})">
            <p>Welcome back, user!</p>
        </esi:when>
    </esi:choose>
    <esi:try>
        <esi:attempt>
            <esi:include src="/fragments/user-panel" />
        </esi:attempt>
        <esi:except>
            <p>User panel unavailable</p>
        </esi:except>
    </esi:try>
-->
<p>Main content</p>
```

### Comment Block Best Practices

1. **Graceful Degradation**: Comment blocks allow ESI content to be hidden from non-ESI processors
2. **Complex Nesting**: Support for nested ESI elements within comment blocks
3. **Whitespace Handling**: Robust handling of various whitespace patterns
4. **Error Recovery**: Empty or malformed comment blocks are safely removed
5. **Full ESI Support**: All ESI elements work within comment blocks including Akamai extensions

### Comment Block Processing Features

- **Enhanced Regex**: Robust pattern matching with `[\s\S]*?` for multiline content
- **Whitespace Flexibility**: Handles various whitespace patterns and indentation
- **Nested Processing**: Full ESI processing pipeline applied to comment content
- **Error Handling**: Graceful handling of processing errors within comments
- **Debug Support**: Comprehensive debug logging for comment block processing

## ESI Variable Substitution (`<esi:vars>`)

The emulator provides comprehensive ESI variable substitution support through the `<esi:vars>` element:

### Basic Variable Substitution

```xml
<esi:vars>
    <p>Host: $(HTTP_HOST)</p>
    <p>User Agent: $(HTTP_USER_AGENT)</p>
    <p>Referer: $(HTTP_REFERER)</p>
</esi:vars>
```

### Variable with Keys

Access specific values from complex variables:

```xml
<esi:vars>
    <!-- Cookie with specific key -->
    <p>Username: $(HTTP_COOKIE{username})</p>
    
    <!-- User Agent components -->
    <p>Browser: $(HTTP_USER_AGENT{browser})</p>
    <p>Operating System: $(HTTP_USER_AGENT{os})</p>
    
    <!-- Query string parameters -->
    <p>Product ID: $(QUERY_STRING{id})</p>
    
    <!-- Accept Language check -->
    <p>Supports English: $(HTTP_ACCEPT_LANGUAGE{en})</p>
</esi:vars>
```

### Default Values

Provide fallback values when variables are missing:

```xml
<esi:vars>
    <!-- Simple default -->
    <p>User: $(HTTP_COOKIE{username}|guest)</p>
    
    <!-- Quoted default values -->
    <p>Name: $(HTTP_COOKIE{name}|'Anonymous User')</p>
    
    <!-- Multiple variables with defaults -->
    <p>Welcome $(HTTP_COOKIE{name}|'Guest') from $(HTTP_HOST|'unknown')</p>
</esi:vars>
```

### Multiple Variable Blocks

Use multiple `<esi:vars>` blocks in the same document:

```xml
<html>
<body>
    <esi:vars>
        <header>Welcome to $(HTTP_HOST)</header>
    </esi:vars>
    
    <main>
        <esi:vars>
            <p>Request Method: $(REQUEST_METHOD)</p>
            <p>Request URI: $(REQUEST_URI)</p>
        </esi:vars>
    </main>
</body>
</html>
```

### Integration with Akamai Extensions

Combine with `<esi:assign>` for custom variables:

```xml
<esi:assign name="site_name" value="My Awesome Site" />
<esi:assign name="user_level" value="$(HTTP_COOKIE{level}|basic)" />

<esi:vars>
    <h1>Welcome to $(site_name)</h1>
    <p>Your level: $(user_level)</p>
    <p>Current time: $(TIME)</p>
</esi:vars>
```

### Supported Variables

| Variable | Description | Key Support | Default Support |
|----------|-------------|-------------|-----------------|
| `HTTP_HOST` | Request host header | ❌ | ✅ |
| `HTTP_USER_AGENT` | User agent string | ✅ (browser, os, version) | ✅ |
| `HTTP_COOKIE` | Cookie header | ✅ (cookie name) | ✅ |
| `HTTP_REFERER` | Referer header | ❌ | ✅ |
| `HTTP_ACCEPT_LANGUAGE` | Accept-Language header | ✅ (language code) | ✅ |
| `QUERY_STRING` | Query string | ✅ (parameter name) | ✅ |
| `REQUEST_METHOD` | HTTP method | ❌ | ✅ |
| `REQUEST_URI` | Request URI | ❌ | ✅ |
| `GEO_COUNTRY_CODE` | Country code (Akamai) | ❌ | ✅ |
| `GEO_COUNTRY_NAME` | Country name (Akamai) | ❌ | ✅ |
| `GEO_REGION` | Region (Akamai) | ❌ | ✅ |
| `GEO_CITY` | City (Akamai) | ❌ | ✅ |
| `CLIENT_IP` | Client IP address (Akamai) | ❌ | ✅ |

### Variable Patterns

The emulator supports these variable patterns:

- **Simple**: `$(VARIABLE_NAME)`
- **With Key**: `$(VARIABLE_NAME{key})`
- **With Default**: `$(VARIABLE_NAME|default_value)`
- **With Key and Default**: `$(VARIABLE_NAME{key}|default_value)`
- **Quoted Defaults**: `$(VARIABLE_NAME|'quoted default')`

## Usage

### Programmatic Usage

```go
package main

import (
    "fmt"
    "esi-emulator/pkg/esi"
)

func main() {
    // Create ESI processor configuration
    config := esi.Config{
        Mode:        "akamai",
        Debug:       true,
        MaxIncludes: 256,
        MaxDepth:    5,
        Cache: esi.CacheConfig{
            Enabled: true,
            TTL:     300, // 5 minutes
        },
    }

    // Create processor
    processor := esi.NewProcessor(config)

    // Create processing context
    context := esi.ProcessContext{
        BaseURL: "https://example.com",
        Headers: map[string]string{
            "User-Agent": "Mozilla/5.0...",
            "Host":       "example.com",
        },
        Cookies: map[string]string{
            "session": "abc123",
        },
        Depth: 0,
    }

    // Process ESI content
    html := `<esi:vars>Host: $(HTTP_HOST)</esi:vars>`
    result, err := processor.Process(html, context)
    if err != nil {
        panic(err)
    }

    fmt.Println(result)
}
```

### HTTP API Usage

The ESI emulator can be used as part of the main server application:

```bash
# Start the server with ESI support
go run main.go -mode akamai

# Process ESI content via HTTP API
curl -X POST http://localhost:3000/process \
  -H "Content-Type: application/json" \
  -d '{"html": "<esi:include src=\"/fragments/header\" />Hello World!"}'
```

## Current Status

This is a **comprehensive ESI implementation** with full test coverage and production-ready features:

### ✅ Fully Implemented and Tested
- **Core ESI Processing**: Complete Go implementation with goroutine-safe operations
- **Basic ESI Elements**: `<esi:include>`, `<esi:comment>`, `<esi:remove>`
- **Multiple Implementation Modes**: Fastly, Akamai, W3C, Development
- **HTTP Fragment Fetching**: Concurrent requests with configurable timeouts
- **Advanced Caching**: Thread-safe caching with TTL expiration and statistics
- **Error Handling**: Comprehensive error handling with `alt` URLs and `onerror` attributes
- **Built-in Examples**: Comprehensive test cases and example fragments
- **Performance Monitoring**: Request statistics and timing metrics
- **Akamai Extensions**: Full implementation of `<esi:assign>`, `<esi:eval>`, `<esi:function>`, `<esi:dictionary>`, `<esi:debug>`
- **Variable System**: Complete ESI variable expansion with HTTP headers, geo data, and custom variables
- **Expression Engine**: Boolean expression evaluation for conditionals
- **ESI Variable Substitution**: Full `<esi:vars>` implementation with default values, keys, and complex variable patterns
- **Conditional Processing**: Complete `<esi:choose>/<esi:when>/<esi:otherwise>` implementation with expression evaluation
- **Error Handling Blocks**: Full `<esi:try>/<esi:attempt>/<esi:except>` implementation for graceful error handling
- **Comment Block Processing**: Complete `<!--esi ... -->` implementation with full ESI element support
- **Comprehensive Test Suite**: 140+ passing tests covering all functionality

### 📋 Future Enhancements
- **Streaming ESI Processing**: Stream-based processing for large documents
- **Advanced Caching Strategies**: Redis/Memcached backends, cache invalidation
- **ESI Validation Tools**: Syntax validation and debugging utilities
- **Performance Profiling**: Detailed performance analysis and bottleneck detection
- **Web-based Testing Interface**: Browser-based ESI testing and visualization
- **Integration Examples**: Docker, Kubernetes, reverse proxy configurations
- **Load Testing**: Concurrent request testing and stress testing capabilities

## Features

### Core ESI Elements (All Modes)
- **`<esi:include>`** - Include external content with caching support
- **`<esi:comment>`** - Developer comments (removed from output) 
- **`<esi:remove>`** - Content removal when ESI is not processed

### Extended ESI Elements (Akamai/W3C/Development Modes)
- **`<esi:choose>/<esi:when>/<esi:otherwise>`** - Conditional processing
- **`<esi:try>/<esi:attempt>/<esi:except>`** - Error handling blocks
- **`<esi:vars>`** - Variable substitution in content
- **Variables** - ESI variable expansion `$(HTTP_COOKIE{name})`
- **Expressions** - Boolean expressions in conditionals
- **Comment blocks** - `<!--esi ... -->` processing

### Akamai-Specific Extensions
- **`<esi:assign>`** - Variable assignment and custom variables
- **`<esi:eval>`** - Expression evaluation and output
- **`<esi:function>`** - Built-in functions (base64, url_encode, time, etc.)
- **`<esi:dictionary>`** - Key-value dictionary lookups
- **`<esi:debug>`** - Development debugging output
- **Extended Variables** - Geo-location and client information
- **Enhanced Include** - Timeout, caching, and method attributes

## References

- [W3C ESI Language Specification 1.0](https://www.w3.org/TR/esi-lang/)
- [Fastly ESI Documentation](https://www.fastly.com/documentation/reference/vcl/statements/esi/)
- [Akamai ESI Developer's Guide](https://www.akamai.com/site/zh/documents/technical-publication/akamai-esi-developers-guide-technical-publication.pdf) 