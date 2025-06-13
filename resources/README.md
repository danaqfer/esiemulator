# ESI Documentation Resources

This directory contains comprehensive documentation and specifications for Edge Side Includes (ESI) from various sources to support the development of the ESI emulator.

## Available Resources

### 1. W3C ESI Language Specification 1.0
**File:** `w3c-esi-specification-1.0.md`

The complete W3C specification for ESI 1.0, including:
- Complete ESI element reference
- ESI variables and expressions syntax
- Protocol considerations
- Implementation guidelines

**Key Features Covered:**
- `<esi:include>` with `src`, `alt`, and `onerror` attributes
- `<esi:inline>` for embedded fragments
- `<esi:choose>`, `<esi:when>`, `<esi:otherwise>` for conditionals
- `<esi:try>`, `<esi:attempt>`, `<esi:except>` for error handling
- `<esi:comment>`, `<esi:remove>`, `<esi:vars>`
- ESI variables (HTTP_COOKIE, HTTP_HOST, etc.)
- ESI expressions and operators

### 2. Fastly ESI Implementation
**File:** `fastly-esi-implementation.md`

Detailed documentation of Fastly's limited ESI implementation:
- Only supports `<esi:include>`, `<esi:comment>`, `<esi:remove>`
- VCL integration patterns
- Performance considerations and limitations
- Migration guidelines from full ESI

### 3. External Resources (To Be Downloaded)

#### Akamai ESI Developer's Guide
**URL:** https://www.akamai.com/site/zh/documents/technical-publication/akamai-esi-developers-guide-technical-publication.pdf

This comprehensive guide covers Akamai's extended ESI implementation beyond the W3C specification.

#### ESI Overview Document
**URL:** http://www.edge-delivery.org/dload/esi_overview.pdf

Additional overview documentation from the edge delivery organization.

## ESI Feature Comparison

| Feature | W3C Spec | Fastly | Akamai | Emulator Target |
|---------|----------|--------|---------|------------------|
| `<esi:include>` | ✅ | ✅ | ✅ | ✅ |
| `<esi:comment>` | ✅ | ✅ | ✅ | ✅ |
| `<esi:remove>` | ✅ | ✅ | ✅ | ✅ |
| `<esi:inline>` | ✅ (optional) | ❌ | ✅ | ✅ |
| `<esi:choose>` | ✅ | ❌ | ✅ | ✅ |
| `<esi:try>` | ✅ | ❌ | ✅ | ✅ |
| `<esi:vars>` | ✅ | ❌ | ✅ | ✅ |
| ESI Variables | ✅ | ❌ | ✅ | ✅ |
| ESI Expressions | ✅ | ❌ | ✅ | ✅ |
| `<!--esi ...-->` | ✅ | ❌ | ✅ | ✅ |
| Error handling | ✅ | Limited | ✅ | ✅ |

## Implementation Notes

### Akamai Extensions
Based on available information, Akamai likely provides:
- Full W3C ESI 1.0 compliance
- Additional ESI variables beyond the specification
- Enhanced caching directives
- Extended debugging capabilities
- Custom ESI functions

### Emulator Requirements
The ESI emulator should support:
1. **Fastly Mode:** Limited feature set for compatibility testing
2. **Akamai Mode:** Full feature set including extensions
3. **Strict W3C Mode:** Specification compliance testing
4. **Development Mode:** Enhanced debugging and validation

## Usage in Emulator Development

These resources will guide the implementation of:

1. **Parser Module:** XML/HTML parsing of ESI tags
2. **Expression Engine:** Variable substitution and expression evaluation
3. **Request Handler:** HTTP request management for includes
4. **Cache Emulator:** Fragment caching behavior
5. **Error Handler:** Exception and fallback processing
6. **Validator:** ESI markup validation and debugging

## Related Standards

- **XInclude:** W3C XML inclusion standard (similar but different scope)
- **Server Side Includes (SSI):** Traditional web server includes
- **XSLT:** XML transformation language (ESI borrows some concepts)
- **HTTP/1.1:** Caching and header specifications relevant to ESI

## Performance Considerations

Key performance aspects documented:
- Sequential vs. parallel processing models
- Fragment caching strategies
- Network request optimization
- Memory usage patterns
- Edge vs. origin processing trade-offs 