# Akamai Property Manager Emulator

A comprehensive Akamai Property Manager emulator written in Go that simulates Akamai's edge computing platform for testing and development purposes.

## Modes of Operation

- **Standalone Property Manager Mode**: Run only the Property Manager emulator for traffic management and content delivery.
- **Integrated Mode (with ESI)**: Run Property Manager together with the ESI emulator, simulating the full Akamai edge workflow (Property Manager processes the request, invokes ESI, then applies response behaviors).

### How to Run

After building the project (see root README), use the following commands from the project root:

```sh
# Standalone Property Manager Mode
bin/edge-emulator -mode=property-manager -debug -port=3002

# Integrated Mode (Property Manager + ESI)
bin/edge-emulator -mode=integrated -esi-mode=akamai -debug -port=3003
```

- All binaries are in the `bin/` directory.
- All test output is in `cmd/edge-emulator/test_output/` (untracked).

## Overview

Akamai Property Manager is a powerful edge computing platform that allows you to configure complex traffic management, content delivery, and security policies at the edge of the network. This emulator provides a complete implementation that supports:

### Core Property Manager Features
- **Rule-based Traffic Management** - Hierarchical rule processing with criteria evaluation
- **Behavior Execution** - Comprehensive behavior library for content manipulation
- **HTTP Context Processing** - Full HTTP request/response context simulation
- **Performance Optimization** - Efficient rule matching and behavior execution
- **Debug and Monitoring** - Comprehensive logging and statistics

### Supported Criteria Types
- **Path-based** - URL path matching with various operators
- **Header-based** - HTTP header evaluation and manipulation
- **Method-based** - HTTP method filtering
- **Host-based** - Host header processing
- **Query-based** - Query string parameter evaluation
- **Cookie-based** - Cookie value processing
- **Variable-based** - Custom variable evaluation
- **Client IP-based** - IP address filtering and geo-location
- **User Agent-based** - Browser and device detection

### Supported Behaviors
- **Caching Behaviors** - Cache control and optimization
- **Security Behaviors** - Access control and protection
- **Performance Behaviors** - Compression and optimization
- **Content Behaviors** - Content manipulation and transformation
- **Redirect Behaviors** - URL redirection and rewriting

## Architecture

The Property Manager emulator is built in Go and follows standard Go project structure:

- **Core Processor** (`processor.go`): Main rule processing engine with criteria evaluation
- **Behavior System** (`behaviors.go`): Comprehensive behavior library and execution
- **Type Definitions** (`types.go`): Complete type system for rules, criteria, and behaviors
- **Comprehensive Tests** (`propertymanager_test.go`): Full test coverage with 700+ test cases

Key features:
- Hierarchical rule processing with parent-child relationships
- Thread-safe concurrent processing
- Configurable performance limits and timeouts
- Comprehensive error handling and logging
- Built-in test scenarios and examples

## Implementation Details

### Architecture

The Property Manager processor is built with a modular architecture:

- **Core Processor** (`processor.go`) - Main rule processing engine
- **Behavior System** (`behaviors.go`) - Behavior library and execution engine
- **Type System** (`types.go`) - Complete type definitions and interfaces
- **Statistics** - Request tracking and performance metrics

### Processing Pipeline

1. **Context Creation** - Build HTTP context from request data
2. **Rule Evaluation** - Process hierarchical rules with criteria matching
3. **Behavior Execution** - Execute matched behaviors in order
4. **Result Generation** - Generate final response with applied behaviors
5. **Statistics Update** - Track performance and usage metrics

### Performance Considerations

- **Concurrent Processing** - Thread-safe operations with mutex protection
- **Efficient Matching** - Optimized criteria evaluation algorithms
- **Resource Limits** - Configurable maximum rules and depth limits
- **Error Handling** - Graceful degradation with fallback support

## Rule-Based Traffic Management

### Basic Rule Structure

```go
rule := Rule{
    Name: "Mobile Users",
    Criteria: []Criterion{
        {
            Name:   "user_agent",
            Option: "contains",
            Value:  "Mobile",
        },
    },
    Behaviors: []Behavior{
        {
            Name: "compress",
            Options: map[string]interface{}{
                "gzip": true,
            },
        },
    },
    Children: []Rule{
        // Child rules for more specific conditions
    },
}
```

### Hierarchical Rule Processing

The emulator supports complex hierarchical rule structures:

```go
// Parent rule - matches all mobile users
parentRule := Rule{
    Name: "Mobile Traffic",
    Criteria: []Criterion{
        {Name: "user_agent", Option: "contains", Value: "Mobile"},
    },
    Children: []Rule{
        // Child rule - specific mobile browsers
        {
            Name: "Chrome Mobile",
            Criteria: []Criterion{
                {Name: "user_agent", Option: "contains", Value: "Chrome"},
            },
            Behaviors: []Behavior{
                {Name: "compress", Options: map[string]interface{}{"gzip": true}},
                {Name: "cache", Options: map[string]interface{}{"ttl": 3600}},
            },
        },
        // Another child rule - other mobile browsers
        {
            Name: "Other Mobile",
            Behaviors: []Behavior{
                {Name: "compress", Options: map[string]interface{}{"gzip": true}},
            },
        },
    },
}
```

### Criteria Evaluation

The emulator supports comprehensive criteria evaluation:

```go
// Path-based criteria
pathCriterion := Criterion{
    Name:   "path",
    Option: "starts_with",
    Value:  "/api/",
}

// Header-based criteria
headerCriterion := Criterion{
    Name:   "header",
    Option: "Authorization",
    Extract: "starts_with",
    Value:  "Bearer ",
    Case:   false,
}

// Cookie-based criteria
cookieCriterion := Criterion{
    Name:   "cookie",
    Option: "session_id",
    Extract: "not_equals",
    Value:  "",
}

// Query string criteria
queryCriterion := Criterion{
    Name:   "query",
    Option: "contains",
    Value:  "debug=true",
}
```

## Behavior System

### Caching Behaviors

```go
// Cache control behavior
cacheBehavior := Behavior{
    Name: "cache",
    Options: map[string]interface{}{
        "ttl":          3600,
        "cacheable":    true,
        "max_age":      1800,
        "stale_while_revalidate": 300,
    },
}

// Cache bypass behavior
cacheBypassBehavior := Behavior{
    Name: "cache_bypass",
    Options: map[string]interface{}{
        "reason": "dynamic_content",
    },
}
```

### Security Behaviors

```go
// Access control behavior
accessControlBehavior := Behavior{
    Name: "access_control",
    Options: map[string]interface{}{
        "allowed_ips": []string{"192.168.1.0/24", "10.0.0.0/8"},
        "blocked_ips": []string{"203.0.113.0/24"},
    },
}

// Rate limiting behavior
rateLimitBehavior := Behavior{
    Name: "rate_limit",
    Options: map[string]interface{}{
        "requests_per_second": 100,
        "burst_size":          50,
    },
}
```

### Performance Behaviors

```go
// Compression behavior
compressionBehavior := Behavior{
    Name: "compress",
    Options: map[string]interface{}{
        "gzip":     true,
        "brotli":   true,
        "min_size": 1024,
    },
}

// Image optimization behavior
imageOptimizationBehavior := Behavior{
    Name: "image_optimization",
    Options: map[string]interface{}{
        "webp":     true,
        "quality":  85,
        "strip_metadata": true,
    },
}
```

### Content Behaviors

```go
// Header modification behavior
headerModificationBehavior := Behavior{
    Name: "modify_headers",
    Options: map[string]interface{}{
        "add": map[string]string{
            "X-Custom-Header": "value",
            "X-Processed-By":  "akamai-emulator",
        },
        "remove": []string{"X-Debug-Header"},
    },
}

// URL rewriting behavior
urlRewriteBehavior := Behavior{
    Name: "url_rewrite",
    Options: map[string]interface{}{
        "pattern":     "/old/(.*)",
        "replacement": "/new/$1",
        "redirect":    false,
    },
}
```

### Redirect Behaviors

```go
// HTTP redirect behavior
redirectBehavior := Behavior{
    Name: "redirect",
    Options: map[string]interface{}{
        "status_code": 301,
        "location":    "https://new-domain.com$1",
        "preserve_query": true,
    },
}

// Conditional redirect behavior
conditionalRedirectBehavior := Behavior{
    Name: "conditional_redirect",
    Options: map[string]interface{}{
        "conditions": []map[string]interface{}{
            {
                "header": "User-Agent",
                "contains": "Mobile",
                "redirect_to": "/mobile$1",
            },
            {
                "header": "Accept-Language",
                "contains": "es",
                "redirect_to": "/es$1",
            },
        },
    },
}
```

## HTTP Context Processing

### Context Creation

```go
// Create HTTP context from request
context := &HTTPContext{
    Method:  "GET",
    Path:    "/api/users/123",
    Host:    "example.com",
    Query:   "debug=true&format=json",
    Headers: map[string]string{
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
        "Accept":     "application/json",
        "Authorization": "Bearer abc123",
    },
    Cookies: map[string]string{
        "session_id": "xyz789",
        "user_id":    "12345",
    },
    ClientIP: "192.168.1.100",
    Variables: map[string]string{
        "custom_var": "custom_value",
    },
}
```

### Context Evaluation

The emulator provides comprehensive context evaluation capabilities:

```go
// Evaluate path criteria
pathMatch := pm.evaluatePathCriterion(&Criterion{
    Name:   "path",
    Option: "regex",
    Value:  "/api/users/(\\d+)",
}, context)

// Evaluate header criteria
headerMatch := pm.evaluateHeaderCriterion(&Criterion{
    Name:   "header",
    Option: "Authorization",
    Extract: "starts_with",
    Value:  "Bearer ",
    Case:   false,
}, context)

// Evaluate cookie criteria
cookieMatch := pm.evaluateCookieCriterion(&Criterion{
    Name:   "cookie",
    Option: "session_id",
    Extract: "not_equals",
    Value:  "",
}, context)
```

## Usage

### Programmatic Usage

```go
package main

import (
    "fmt"
    "esi-emulator/pkg/propertymanager"
)

func main() {
    // Create Property Manager configuration
    config := propertymanager.Config{
        Debug:       true,
        MaxRules:    1000,
        MaxDepth:    10,
        Timeout:     5000, // milliseconds
    }

    // Create Property Manager instance
    pm := propertymanager.New(config)

    // Define rules
    rules := []propertymanager.Rule{
        {
            Name: "API Rate Limiting",
            Criteria: []propertymanager.Criterion{
                {
                    Name:   "path",
                    Option: "starts_with",
                    Value:  "/api/",
                },
            },
            Behaviors: []propertymanager.Behavior{
                {
                    Name: "rate_limit",
                    Options: map[string]interface{}{
                        "requests_per_second": 100,
                    },
                },
            },
        },
        {
            Name: "Mobile Optimization",
            Criteria: []propertymanager.Criterion{
                {
                    Name:   "user_agent",
                    Option: "contains",
                    Value:  "Mobile",
                },
            },
            Behaviors: []propertymanager.Behavior{
                {
                    Name: "compress",
                    Options: map[string]interface{}{
                        "gzip": true,
                    },
                },
            },
        },
    }

    // Set rules
    pm.SetRules(rules)

    // Create HTTP context
    context := &propertymanager.HTTPContext{
        Method:  "GET",
        Path:    "/api/users/123",
        Host:    "example.com",
        Headers: map[string]string{
            "User-Agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)",
        },
    }

    // Process request
    result, err := pm.ProcessRequest(context)
    if err != nil {
        panic(err)
    }

    // Access results
    fmt.Printf("Matched rules: %v\n", result.MatchedRules)
    fmt.Printf("Applied behaviors: %v\n", result.AppliedBehaviors)
    fmt.Printf("Processing time: %dms\n", result.ProcessingTime)
}
```

### HTTP API Usage

The Property Manager emulator can be used as part of the main server application:

```bash
# Start the server with Property Manager support
go run main.go -mode property-manager

# Process Property Manager rules via HTTP API
curl -X POST http://localhost:3000/property-manager/process \
  -H "Content-Type: application/json" \
  -d '{
    "rules": [
      {
        "name": "Mobile Users",
        "criteria": [
          {
            "name": "user_agent",
            "option": "contains",
            "value": "Mobile"
          }
        ],
        "behaviors": [
          {
            "name": "compress",
            "options": {"gzip": true}
          }
        ]
      }
    ],
    "context": {
      "method": "GET",
      "path": "/api/data",
      "headers": {
        "User-Agent": "Mozilla/5.0 (iPhone; CPU iPhone OS 14_0 like Mac OS X)"
      }
    }
  }'
```

## Current Status

This is a **comprehensive Property Manager implementation** with full test coverage and production-ready features:

### ✅ Fully Implemented and Tested
- **Core Rule Processing**: Complete Go implementation with hierarchical rule evaluation
- **Criteria Evaluation**: All major criteria types with comprehensive operators
- **Behavior System**: Complete behavior library with execution engine
- **HTTP Context Processing**: Full request/response context simulation
- **Performance Optimization**: Efficient rule matching and behavior execution
- **Error Handling**: Comprehensive error handling with graceful degradation
- **Statistics and Monitoring**: Request tracking and performance metrics
- **Comprehensive Test Suite**: 700+ passing tests covering all functionality
- **Thread Safety**: Concurrent processing with proper synchronization
- **Configuration Management**: Flexible configuration system
- **Debug Support**: Comprehensive logging and debugging capabilities

### 📋 Future Enhancements
- **Advanced Caching**: Redis/Memcached backends for distributed caching
- **Load Balancing**: Advanced load balancing algorithms and health checks
- **Security Features**: WAF, DDoS protection, and advanced security behaviors
- **Performance Profiling**: Detailed performance analysis and bottleneck detection
- **Web-based Configuration Interface**: Browser-based rule configuration and testing
- **Integration Examples**: Docker, Kubernetes, reverse proxy configurations
- **Load Testing**: Concurrent request testing and stress testing capabilities
- **Rule Import/Export**: JSON/YAML configuration file support
- **Real-time Monitoring**: Live statistics and performance dashboards

## Features

### Core Rule Processing
- **Hierarchical Rules**: Parent-child rule relationships with inheritance
- **Criteria Evaluation**: Comprehensive criteria matching with multiple operators
- **Behavior Execution**: Complete behavior library with execution engine
- **Performance Optimization**: Efficient rule matching algorithms
- **Error Handling**: Graceful degradation with fallback support

### Supported Criteria Types
- **Path-based**: URL path matching with regex, prefix, suffix, and contains
- **Header-based**: HTTP header evaluation with case-sensitive/insensitive options
- **Method-based**: HTTP method filtering (GET, POST, PUT, DELETE, etc.)
- **Host-based**: Host header processing with various matching options
- **Query-based**: Query string parameter evaluation and extraction
- **Cookie-based**: Cookie value processing with secure and http-only options
- **Variable-based**: Custom variable evaluation and manipulation
- **Client IP-based**: IP address filtering with CIDR notation support
- **User Agent-based**: Browser and device detection with parsing

### Supported Behaviors
- **Caching Behaviors**: Cache control, TTL management, cache bypass
- **Security Behaviors**: Access control, rate limiting, IP blocking
- **Performance Behaviors**: Compression, image optimization, minification
- **Content Behaviors**: Header modification, content transformation
- **Redirect Behaviors**: HTTP redirects, conditional redirects, URL rewriting

### Advanced Features
- **Concurrent Processing**: Thread-safe operations with proper synchronization
- **Statistics and Monitoring**: Request tracking and performance metrics
- **Debug Support**: Comprehensive logging and debugging capabilities
- **Configuration Management**: Flexible configuration system
- **Error Recovery**: Graceful error handling and recovery mechanisms

## References

- [Akamai Property Manager Documentation](https://techdocs.akamai.com/property-manager/docs)
- [Akamai Edge Computing Platform](https://www.akamai.com/products/edge-computing)
- [Akamai Traffic Management](https://www.akamai.com/products/traffic-management) 