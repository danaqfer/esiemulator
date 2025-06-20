# Akamai Property Manager Specification

## Overview

Akamai Property Manager is a powerful edge computing platform that allows you to configure complex traffic management, content delivery, and security policies at the edge of the network. This document provides the complete specification for behaviors, rules, and criteria based on the latest Akamai Property Manager documentation.

## Table of Contents

1. [Rules and Criteria](#rules-and-criteria)
2. [Behaviors](#behaviors)
3. [Variables](#variables)
4. [Property Configuration](#property-configuration)
5. [Processing Pipeline](#processing-pipeline)
6. [API Reference](#api-reference)

## Rules and Criteria

### Rule Structure

```xml
<rule name="rule-name" comment="optional comment">
    <criteria>
        <criteria name="criteria-type" option="option-value" value="match-value" />
    </criteria>
    <behaviors>
        <behavior name="behavior-name">
            <option name="option-name" value="option-value" />
        </behavior>
    </behaviors>
    <children>
        <!-- Child rules -->
    </children>
</rule>
```

### Supported Criteria Types

#### 1. Path-based Criteria

**Name:** `path`

**Options:**
- `equals` - Exact path match
- `not_equals` - Path does not equal
- `starts_with` - Path starts with value
- `ends_with` - Path ends with value
- `contains` - Path contains value
- `regex` - Regular expression match

**Example:**
```xml
<criteria name="path" option="starts_with" value="/api/" />
<criteria name="path" option="regex" value="/users/(\d+)" />
```

#### 2. Header-based Criteria

**Name:** `header`

**Options:**
- `option` - Header name to match
- `extract` - Extraction method (equals, not_equals, starts_with, ends_with, contains, regex)
- `value` - Value to match
- `case` - Case sensitivity (true/false)

**Example:**
```xml
<criteria name="header" option="User-Agent" extract="contains" value="Mobile" case="false" />
<criteria name="header" option="Authorization" extract="starts_with" value="Bearer " />
```

#### 3. Method-based Criteria

**Name:** `method`

**Options:**
- `equals` - HTTP method equals
- `not_equals` - HTTP method does not equal

**Example:**
```xml
<criteria name="method" option="equals" value="POST" />
<criteria name="method" option="not_equals" value="GET" />
```

#### 4. Host-based Criteria

**Name:** `host`

**Options:**
- `equals` - Host equals
- `not_equals` - Host does not equal
- `starts_with` - Host starts with
- `ends_with` - Host ends with
- `contains` - Host contains
- `case` - Case sensitivity (true/false)

**Example:**
```xml
<criteria name="host" option="equals" value="api.example.com" />
<criteria name="host" option="ends_with" value=".example.com" case="false" />
```

#### 5. Query-based Criteria

**Name:** `query`

**Options:**
- `equals` - Query string equals
- `not_equals` - Query string does not equal
- `contains` - Query string contains
- `regex` - Regular expression match

**Example:**
```xml
<criteria name="query" option="contains" value="debug=true" />
<criteria name="query" option="regex" value="id=(\d+)" />
```

#### 6. Cookie-based Criteria

**Name:** `cookie`

**Options:**
- `option` - Cookie name
- `extract` - Extraction method (equals, not_equals, starts_with, ends_with, contains, regex)
- `value` - Value to match

**Example:**
```xml
<criteria name="cookie" option="session_id" extract="not_equals" value="" />
<criteria name="cookie" option="user_type" extract="equals" value="premium" />
```

#### 7. Variable-based Criteria

**Name:** `variable`

**Options:**
- `option` - Variable name
- `extract` - Extraction method
- `value` - Value to match

**Example:**
```xml
<criteria name="variable" option="GEO_COUNTRY_CODE" extract="equals" value="US" />
<criteria name="variable" option="CLIENT_IP" extract="starts_with" value="192.168." />
```

#### 8. Client IP-based Criteria

**Name:** `client_ip`

**Options:**
- `equals` - IP equals
- `not_equals` - IP does not equal
- `starts_with` - IP starts with
- `ends_with` - IP ends with
- `contains` - IP contains
- `in` - IP in CIDR range
- `not_in` - IP not in CIDR range

**Example:**
```xml
<criteria name="client_ip" option="in" value="192.168.1.0/24" />
<criteria name="client_ip" option="not_in" value="10.0.0.0/8" />
```

#### 9. User Agent-based Criteria

**Name:** `user_agent`

**Options:**
- `equals` - User agent equals
- `not_equals` - User agent does not equal
- `starts_with` - User agent starts with
- `ends_with` - User agent ends with
- `contains` - User agent contains
- `regex` - Regular expression match
- `case` - Case sensitivity (true/false)

**Example:**
```xml
<criteria name="user_agent" option="contains" value="Mobile" case="false" />
<criteria name="user_agent" option="regex" value="Chrome/(\d+)" />
```

## Behaviors

### Caching Behaviors

#### 1. Cache Control

**Name:** `cache`

**Options:**
- `ttl` - Time to live in seconds
- `cacheable` - Whether content is cacheable (true/false)
- `max_age` - Max age in seconds
- `stale_while_revalidate` - Stale while revalidate time in seconds
- `stale_if_error` - Stale if error time in seconds

**Example:**
```xml
<behavior name="cache">
    <option name="ttl" value="3600" />
    <option name="cacheable" value="true" />
    <option name="max_age" value="1800" />
</behavior>
```

#### 2. Cache Bypass

**Name:** `cache_bypass`

**Options:**
- `reason` - Reason for bypass

**Example:**
```xml
<behavior name="cache_bypass">
    <option name="reason" value="dynamic_content" />
</behavior>
```

### Security Behaviors

#### 1. Access Control

**Name:** `access_control`

**Options:**
- `allowed_ips` - Comma-separated list of allowed IP ranges
- `blocked_ips` - Comma-separated list of blocked IP ranges
- `allowed_countries` - Comma-separated list of allowed country codes
- `blocked_countries` - Comma-separated list of blocked country codes

**Example:**
```xml
<behavior name="access_control">
    <option name="allowed_ips" value="192.168.1.0/24,10.0.0.0/8" />
    <option name="blocked_countries" value="XX,YY" />
</behavior>
```

#### 2. Rate Limiting

**Name:** `rate_limit`

**Options:**
- `requests_per_second` - Maximum requests per second
- `burst_size` - Burst size allowance
- `window_size` - Time window in seconds

**Example:**
```xml
<behavior name="rate_limit">
    <option name="requests_per_second" value="100" />
    <option name="burst_size" value="50" />
</behavior>
```

### Performance Behaviors

#### 1. Compression

**Name:** `compress`

**Options:**
- `gzip` - Enable gzip compression (true/false)
- `brotli` - Enable brotli compression (true/false)
- `min_size` - Minimum size for compression in bytes

**Example:**
```xml
<behavior name="compress">
    <option name="gzip" value="true" />
    <option name="brotli" value="true" />
    <option name="min_size" value="1024" />
</behavior>
```

#### 2. Image Optimization

**Name:** `image_optimization`

**Options:**
- `webp` - Enable WebP conversion (true/false)
- `quality` - Image quality (1-100)
- `strip_metadata` - Strip metadata (true/false)
- `resize` - Resize dimensions (widthxheight)

**Example:**
```xml
<behavior name="image_optimization">
    <option name="webp" value="true" />
    <option name="quality" value="85" />
    <option name="strip_metadata" value="true" />
</behavior>
```

### Content Behaviors

#### 1. Header Modification

**Name:** `modify_headers`

**Options:**
- `add` - Headers to add (JSON object)
- `remove` - Headers to remove (array)
- `set` - Headers to set (JSON object)

**Example:**
```xml
<behavior name="modify_headers">
    <option name="add" value='{"X-Custom-Header": "value"}' />
    <option name="remove" value='["X-Debug-Header"]' />
</behavior>
```

#### 2. URL Rewriting

**Name:** `url_rewrite`

**Options:**
- `pattern` - Regular expression pattern
- `replacement` - Replacement string
- `redirect` - Whether to redirect (true/false)
- `status_code` - Redirect status code

**Example:**
```xml
<behavior name="url_rewrite">
    <option name="pattern" value="/old/(.*)" />
    <option name="replacement" value="/new/$1" />
    <option name="redirect" value="false" />
</behavior>
```

### Redirect Behaviors

#### 1. HTTP Redirect

**Name:** `redirect`

**Options:**
- `status_code` - HTTP status code (301, 302, 307, 308)
- `location` - Redirect location
- `preserve_query` - Preserve query string (true/false)

**Example:**
```xml
<behavior name="redirect">
    <option name="status_code" value="301" />
    <option name="location" value="https://new-domain.com$1" />
    <option name="preserve_query" value="true" />
</behavior>
```

#### 2. Conditional Redirect

**Name:** `conditional_redirect`

**Options:**
- `conditions` - Array of redirect conditions

**Example:**
```xml
<behavior name="conditional_redirect">
    <option name="conditions" value='[{"header": "User-Agent", "contains": "Mobile", "redirect_to": "/mobile$1"}]' />
</behavior>
```

## Variables

### Built-in Variables

- `GEO_COUNTRY_CODE` - Country code
- `GEO_COUNTRY_NAME` - Country name
- `GEO_REGION` - Region
- `GEO_CITY` - City
- `CLIENT_IP` - Client IP address
- `REQUEST_METHOD` - HTTP method
- `REQUEST_URI` - Request URI
- `HTTP_HOST` - Host header
- `HTTP_USER_AGENT` - User agent
- `HTTP_REFERER` - Referer header
- `HTTP_COOKIE` - Cookie header
- `QUERY_STRING` - Query string

### Custom Variables

```xml
<variables>
    <variable name="custom_var" value="custom_value" type="string" />
    <variable name="api_key" value="$(HTTP_COOKIE{api_key})" type="string" />
</variables>
```

## Property Configuration

### Property Structure

```xml
<property name="property-name" version="1">
    <rules>
        <!-- Rule definitions -->
    </rules>
    <behaviors>
        <!-- Behavior definitions -->
    </behaviors>
    <variables>
        <!-- Variable definitions -->
    </variables>
    <comments>Optional comments</comments>
</property>
```

## Processing Pipeline

1. **Request Reception** - HTTP request received
2. **Context Creation** - Build HTTP context from request
3. **Rule Evaluation** - Process hierarchical rules with criteria matching
4. **Behavior Execution** - Execute matched behaviors in order
5. **Response Generation** - Generate final response with applied behaviors
6. **Statistics Update** - Track performance and usage metrics

## API Reference

### Process Request

**Endpoint:** `POST /property-manager/process`

**Request Body:**
```json
{
    "rules": [
        {
            "name": "rule-name",
            "criteria": [
                {
                    "name": "path",
                    "option": "starts_with",
                    "value": "/api/"
                }
            ],
            "behaviors": [
                {
                    "name": "compress",
                    "options": {
                        "gzip": true
                    }
                }
            ]
        }
    ],
    "context": {
        "method": "GET",
        "path": "/api/users/123",
        "host": "example.com",
        "headers": {
            "User-Agent": "Mozilla/5.0..."
        },
        "cookies": {
            "session_id": "abc123"
        },
        "client_ip": "192.168.1.100"
    }
}
```

**Response:**
```json
{
    "result": {
        "matched_rules": ["rule-name"],
        "executed_behaviors": ["compress"],
        "modified_headers": {},
        "variables": {},
        "errors": []
    },
    "stats": {
        "processing_time": 5,
        "mode": "property-manager",
        "requests": 1,
        "cache_hits": 0,
        "cache_miss": 0,
        "errors": 0,
        "total_time": 5
    }
}
```

## Compliance Notes

This specification is based on the latest Akamai Property Manager documentation and should be used as the authoritative reference for implementing Property Manager functionality. All implementations should ensure compliance with these specifications to maintain compatibility with Akamai's edge computing platform.

## References

- [Akamai Property Manager Documentation](https://techdocs.akamai.com/property-manager/docs)
- [Akamai Property Manager API Reference](https://techdocs.akamai.com/property-manager/reference)
- [Akamai Edge Computing Platform](https://www.akamai.com/products/edge-computing) 