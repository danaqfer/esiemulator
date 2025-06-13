# Fastly ESI Implementation

## Overview

Fastly's Edge Side Include (ESI) support provides a subset of the full ESI 1.0 specification, optimized for high-performance edge processing. ESI processing occurs after the `vcl_deliver` subroutine has finished and is only effective when executed in `vcl_fetch`.

## Supported ESI Directives

Fastly only supports the following ESI directives:

- **`<esi:include>`** - Basic content inclusion
- **`<esi:comment>`** - Developer comments (removed from output)
- **`<esi:remove>`** - Content removal when ESI is not processed

## ESI Include Syntax

```xml
<esi:include src="URL" />
```

The `src` URL can be either absolute or relative:

- **Absolute URLs:** Must be prefixed with `http://` or `https://`
- **Relative URLs:** Resolved relative to the current request

Example:
```xml
<esi:include src="https://example.com/index.html" />
<esi:include src="/fragments/header.html" />
```

## Request Processing

ESI requests are made to your Fastly service as new client requests with:

- Copy of HTTP header information from the original client request
- `Host` header set to the hostname from the ESI src URL (if specified)
- `req.topurl` variable set to the original request's `req.url` value

## VCL Integration

To enable ESI processing, use the `esi;` statement in VCL:

```vcl
sub vcl_fetch {
  esi;
}
```

## Best Practices

### Content Type Filtering
Apply ESI only to response types likely to contain ESI directives:

```vcl
sub vcl_fetch {
  if (beresp.http.content-type ~ "text/html") {
    esi;
  }
}
```

### Compression Considerations

ESI requires uncompressed content to detect `<esi:...>` tags:

- **Origin compression:** Remove `Accept-Encoding` header before backend requests:
  ```vcl
  sub vcl_miss {
    unset bereq.http.accept-encoding;
  }
  ```
- **Static edge compression:** Will fail with ESI (compression occurs before processing)
- **Dynamic compression:** Compatible (compression occurs after ESI processing)

## Limitations and Constraints

### Processing Model
- **Sequential processing:** ESI tags processed one at a time, not concurrently
- **No shared state:** Child requests cannot pass data back to parent
- **CDATA disabled:** Set `esi.allow_inside_cdata` to `true` if needed

### Limits
- **Maximum include depth:** 5 levels
- **Maximum total includes:** 256 per request

### Error Handling
Limited compared to full ESI specification:
- No `alt` attribute support
- No `onerror` attribute support
- No `<esi:try>` / `<esi:attempt>` / `<esi:except>` blocks

### Variables and Conditionals
Not supported:
- ESI variables (`$(HTTP_COOKIE{...})`)
- Conditional processing (`<esi:choose>`, `<esi:when>`, `<esi:otherwise>`)
- Variable substitution (`<esi:vars>`)

## Example Use Cases

### Basic Fragment Inclusion

**Template:**
```html
<html>
<head>
  <title>My Site</title>
</head>
<body>
  <esi:include src="/header.html" />
  <main>Static content here</main>
  <esi:include src="/footer.html" />
</body>
</html>
```

### E-commerce Shopping Cart

**Before ESI (entire page uncacheable):**
```html
<div id="header">
  <img src="/logo.jpg" />
  Shopping Cart: 3 items ($45.99)
</div>
```

**After ESI (base page cacheable):**
```html
<div id="header">
  <img src="/logo.jpg" />
  <esi:include src="/shopping_cart" />
</div>
```

### JSON with Dynamic Content

**Template with mixed static/dynamic JSON:**
```json
{
  "amp_volume": 11,
  "message": "Reticulating Splines",
  "geo_data": {
    <esi:include src="/geo_information" />
  }
}
```

**VCL synthetic response for geo data:**
```vcl
sub vcl_recv {
  if (req.url == "/geo_information") {
    error 900;
  }
}

sub vcl_error {
  if (obj.status == 900) {
    set obj.status = 200;
    set obj.response = "OK";
    synthetic {"\"city\": \""} geoip.city {"\""};
    return(deliver);
  }
}
```

## Performance Considerations

- ESI processing adds latency due to sequential sub-requests
- Each `<esi:include>` triggers a separate backend request
- Consider caching strategies for fragments
- Monitor cache hit ratios for both main content and fragments

## Migration from Full ESI

When migrating from platforms supporting full ESI specification:

1. Remove unsupported variables and conditionals
2. Implement logic in VCL instead of ESI expressions
3. Use Fastly's synthetic responses for dynamic content
4. Simplify error handling (no fallback URLs)

## References

- [Fastly ESI Documentation](https://www.fastly.com/documentation/reference/vcl/statements/esi/)
- [Using ESI Part 1: Simple Edge-Side Include](https://www.fastly.com/blog/using-esi-part-1-simple-edge-side-include)
- [Edge-side includes (ESI) Code Example](https://www.fastly.com/documentation/solutions/examples/edge-side-includes-esi/) 