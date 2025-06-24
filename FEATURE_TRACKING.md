# Feature Tracking: ESI Container Generator vs JavaScript Container

## Overview
This document tracks the features supported by the existing JavaScript container code and their implementation status in our Go ESI Container Generator.

## Core Pixel Types

| Feature | JavaScript | ESI Generator | Status | Notes |
|---------|------------|---------------|--------|-------|
| `dir` (Direct 1x1 pixel) | ✅ | ✅ | **IMPLEMENTED** | Converted to ESI includes |
| `frm` (Iframe-based) | ✅ | ❌ | **TODO** | Will remain in JSON for browser execution |
| `script` (Script-based) | ✅ | ❌ | **TODO** | Will remain in JSON for browser execution |

## Macro Substitution System

### Basic Macros

| Macro | JavaScript | ESI Generator | Status | Notes |
|-------|------------|---------------|--------|-------|
| `~~r~~` (Timestamp) | ✅ | ✅ | **IMPLEMENTED** | `$(TIME)` |
| `~~evid~~` (Event ID) | ✅ | ✅ | **IMPLEMENTED** | `$(PMUSER_EVID)` |
| `~~cs~~` (Consent string) | ✅ | ✅ | **IMPLEMENTED** | `$(HTTP_COOKIE{consent})` |
| `~~cc~~` (Country code) | ✅ | ✅ | **IMPLEMENTED** | `$(GEO_COUNTRY)` |
| `~~uu~~` (User ID) | ✅ | ✅ | **IMPLEMENTED** | `$(PMUSER_UU)` |
| `~~suu~~` (Fingerprint ID) | ✅ | ❌ | **TODO** | Hash of (IP + Accept + User-Agent) |

### Advanced Macros

| Macro | JavaScript | ESI Generator | Status | Notes |
|-------|------------|---------------|--------|-------|
| `~~c~cookieName~~` (Cookie value) | ✅ | ✅ | **IMPLEMENTED** | `$(HTTP_COOKIE{cookieName})` |
| `~~c~cookieName~hpr~salt~~` (Cookie hash path+cookie) | ✅ | ❌ | **TODO** | MD5 hash implementation |
| `~~c~cookieName~hpo~salt~~` (Cookie hash cookie+path) | ✅ | ❌ | **TODO** | MD5 hash implementation |
| `~~dl:qs~~` (Query string decode) | ✅ | ❌ | **TODO** | URL decode implementation |
| `~~dl:qs~paramName~~` (Query param decode) | ✅ | ❌ | **TODO** | URL decode implementation |
| `~~m~mapName~param1~param2~~` (Mapping function) | ✅ | ❌ | **SKIP** | Complex mapping - skip for now |
| Default values (`@` prefix) | ✅ | ❌ | **TODO** | Default value handling |
| Encoding levels (`^` prefix) | ✅ | ❌ | **SKIP** | Skip encode/decode levels |

## Pixel Configuration Properties

| Property | JavaScript | ESI Generator | Status | Notes |
|----------|------------|---------------|--------|-------|
| `ID` | ✅ | ✅ | **IMPLEMENTED** | Unique identifier |
| `URL` | ✅ | ✅ | **IMPLEMENTED** | Target URL with macro support |
| `TYPE` | ✅ | ✅ | **IMPLEMENTED** | Pixel type filtering |
| `REQ` | ✅ | ❌ | **TODO** | Required flag handling |
| `PCT` | ✅ | ❌ | **TODO** | Percentage chance to fire |
| `CAP` | ✅ | ❌ | **TODO** | Capacity limit |
| `RC` | ✅ | ❌ | **TODO** | Rotation category |
| `CONTINENT_FREQ` | ✅ | ❌ | **TODO** | Continent-specific frequency |
| `FIRE_EXPR` | ✅ | ❌ | **TODO** | Conditional firing expression |
| `SCRIPT` | ✅ | ❌ | **TODO** | Script content (for script type) |

## Advanced Features

| Feature | JavaScript | ESI Generator | Status | Notes |
|---------|------------|---------------|--------|-------|
| Batching (`pixelBatchSize`, `pixelBatchPeriod`) | ✅ | ❌ | **SKIP** | Not needed for fire-and-forget |
| Rotation (Cookie-based) | ✅ | ❌ | **SKIP** | Not needed for fire-and-forget |
| Logging beacon | ✅ | ❌ | **SKIP** | Not needed for fire-and-forget |
| GDPR consent checking | ✅ | ❌ | **TODO** | Privacy compliance |
| GPP support | ✅ | ❌ | **TODO** | Global Privacy Platform |
| Performance tracking | ✅ | ❌ | **SKIP** | Not needed for fire-and-forget |
| Maps (Complex mapping) | ✅ | ❌ | **SKIP** | Skip complex mapping functions |
| Encoding/Decoding | ✅ | ❌ | **SKIP** | Skip encode/decode levels |
| Hashing (MD5) | ✅ | ❌ | **TODO** | For cookie hashing |
| Fingerprinting | ✅ | ❌ | **TODO** | suu implementation |
| 3PC detection | ✅ | ❌ | **SKIP** | Browser-specific feature |
| Cross-browser compatibility | ✅ | ❌ | **SKIP** | Server-side only |

## Implementation Priority

### Phase 1 (High Priority)
- [ ] `~~suu~~` fingerprint ID implementation
- [ ] `~~c~cookieName~hpr~salt~~` and `~~c~cookieName~hpo~salt~~` cookie hashing
- [ ] `~~dl:qs~~` and `~~dl:qs~paramName~~` query string decoding
- [ ] Default value handling (`@` prefix)
- [ ] GDPR consent checking
- [ ] GPP support

### Phase 2 (Medium Priority)
- [ ] `REQ`, `PCT`, `CAP` pixel properties
- [ ] `CONTINENT_FREQ` continent-specific frequency
- [ ] `FIRE_EXPR` conditional firing
- [ ] `frm` and `script` type filtering (keep in JSON)

### Phase 3 (Low Priority - Skip)
- [ ] Complex mapping functions (`~~m~mapName~~`)
- [ ] Encoding/decoding levels (`^` prefix)
- [ ] Batching and rotation systems
- [ ] Performance tracking
- [ ] Browser-specific features

## Current Implementation Status

### ✅ Implemented
- Basic ESI variable substitution (`$(HTTP_USER_AGENT)`, etc.)
- Simple macro substitution (`~~r~~`, `~~evid~~`, `~~cs~~`, `~~cc~~`, `~~uu~~`)
- Basic cookie substitution (`~~c~cookieName~~`)
- `dir` type pixel conversion to ESI includes
- Fire-and-forget approach with `MAXWAIT=0`

### ❌ TODO (Phase 1)
- Fingerprint ID generation (suu)
- Advanced cookie hashing macros
- Query string decoding macros
- Default value handling
- Privacy compliance features

### ❌ TODO (Phase 2)
- Pixel property filtering and validation
- Continent-specific frequency handling
- Conditional firing expressions
- Type filtering for browser-executed pixels

### ❌ SKIP
- Complex mapping functions
- Encoding/decoding levels
- Batching and rotation
- Performance tracking
- Browser-specific features 