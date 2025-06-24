# ESI Container Generator

A Go tool that converts JSON partner beacon configurations into ESI (Edge Side Includes) for server-side execution. This tool filters pixel types and generates separate outputs for server-side and browser-executed pixels.

## Features

### ✅ Implemented Features

#### Core Functionality
- **Pixel Type Filtering**: Converts `dir` type pixels to ESI includes, keeps `frm` and `script` types for browser execution
- **Fire-and-Forget Execution**: Uses `MAXWAIT=0` for non-blocking pixel firing
- **Dual Output**: Generates both HTML with ESI includes and JSON for browser-executed pixels

#### Macro Substitution System
- **Basic Macros**:
  - `~~r~~` → `$(TIME)` (timestamp)
  - `~~evid~~` → `$(PMUSER_EVID)` (event ID)
  - `~~cs~~` → `$(HTTP_COOKIE{consent})` (consent string)
  - `~~cc~~` → `$(GEO_COUNTRY)` (country code)
  - `~~uu~~` → `$(PMUSER_UU)` (user ID)
  - `~~suu~~` → `$(PMUSER_SUU)` (fingerprint ID)

- **Advanced Cookie Macros**:
  - `~~c~cookieName~~` → `$(HTTP_COOKIE{cookieName})` (simple cookie)
  - `~~c~cookieName~hpr~salt~~` → `$(PMUSER_COOKIE_HASH_PR_{cookieName}_{salt})` (path + cookie hash)
  - `~~c~cookieName~hpo~salt~~` → `$(PMUSER_COOKIE_HASH_PO_{cookieName}_{salt})` (cookie + path hash)

- **Decode Macros**:
  - `~~dl:qs~~` → `$(PMUSER_DECODED_QUERY_STRING)` (full query string decode)
  - `~~dl:qs~paramName~~` → `$(PMUSER_DECODED_QS_{paramName})` (specific parameter decode)

- **User Variables**:
  - `~~u1~~` → `$(PMUSER_V1)` (user variable 1)
  - `~~u2~~` → `$(PMUSER_V2)` (user variable 2)
  - `~~customvar~~` → `$(PMUSER_CUSTOMVAR)` (custom variable)

#### Advanced Features
- **Fingerprint Generation**: Creates unique fingerprint IDs based on IP + Accept headers + User-Agent
- **MD5 Hashing**: Supports cookie value hashing with salt
- **URL Decoding**: Handles URL-encoded query parameters
- **ESI Functions**: Generates ESI functions for advanced macro processing

### 🔄 TODO Features (Phase 1)

- [ ] Default value handling (`@` prefix)
- [ ] GDPR consent checking
- [ ] GPP (Global Privacy Platform) support
- [ ] Enhanced error handling for malformed macros

### 🔄 TODO Features (Phase 2)

- [ ] `REQ`, `PCT`, `CAP` pixel property filtering
- [ ] `CONTINENT_FREQ` continent-specific frequency handling
- [ ] `FIRE_EXPR` conditional firing expressions
- [ ] Enhanced pixel validation

### ❌ Skipped Features

- Complex mapping functions (`~~m~mapName~~`)
- Encoding/decoding levels (`^` prefix)
- Batching and rotation systems
- Performance tracking
- Browser-specific features (3PC detection, etc.)

## Installation

```bash
# Build the tool
make build

# Or build directly
go build -o bin/ESIcontainergenerator cmd/ESIcontainergenerator/main.go
```

## Usage

### Basic Usage

```bash
# Convert JSON to HTML with ESI includes
./bin/ESIcontainergenerator -input partner_beacons.json

# Specify output file
./bin/ESIcontainergenerator -input partner_beacons.json -output container.html
```

### Advanced Usage

```bash
# Generate both HTML and browser JSON
./bin/ESIcontainergenerator -input partner_beacons.json \
  -output container.html \
  -output-json browser_pixels.json

# Use browser-like ESI variable substitution
./bin/ESIcontainergenerator -input partner_beacons.json -browser-vars

# Set custom max wait time (default: 0 for fire-and-forget)
./bin/ESIcontainergenerator -input partner_beacons.json -maxwait 5
```

### Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-input` | Input JSON configuration file | (required) |
| `-output` | Output HTML file | `input_name.html` |
| `-output-json` | Output JSON file for browser pixels | (none) |
| `-browser-vars` | Use browser-like ESI variable substitution | `false` |
| `-maxwait` | Maximum wait time for ESI includes | `0` |
| `-help` | Show help information | `false` |

## JSON Configuration Format

### Pixel Properties

```json
{
  "pixels": [
    {
      "ID": "unique_pixel_id",
      "URL": "https://partner.com/pixel.gif?param=~~macro~~",
      "TYPE": "dir|frm|script",
      "REQ": true,
      "PCT": 100,
      "CAP": 1,
      "RC": "default",
      "CONTINENT_FREQ": {
        "NA": 90,
        "EU": 85
      },
      "FIRE_EXPR": "country == 'US'",
      "SCRIPT": "console.log('script content');"
    }
  ]
}
```

### Property Descriptions

| Property | Type | Description | Default |
|----------|------|-------------|---------|
| `ID` | string | Unique pixel identifier | (required) |
| `URL` | string | Target URL with macro support | (required) |
| `TYPE` | string | Pixel type: `dir`, `frm`, or `script` | `dir` |
| `REQ` | boolean | Required flag | `true` |
| `PCT` | integer | Percentage chance to fire (1-100) | `100` |
| `CAP` | integer | Capacity limit | `1` |
| `RC` | string | Rotation category | `default` |
| `CONTINENT_FREQ` | object | Continent-specific frequency mapping | (none) |
| `FIRE_EXPR` | string | Conditional firing expression | (none) |
| `SCRIPT` | string | Script content (for script type) | (none) |

### Pixel Type Behavior

- **`dir`**: Converted to ESI includes for server-side execution
- **`frm`**: Kept in JSON for browser iframe execution
- **`script`**: Kept in JSON for browser script execution

## Macro Examples

### Basic Macros

```json
{
  "URL": "https://example.com/pixel.gif?evid=~~evid~~&time=~~r~~&country=~~cc~~&user=~~uu~~&fingerprint=~~suu~~"
}
```

Generates:
```
https://example.com/pixel.gif?evid=$(PMUSER_EVID)&time=$(TIME)&country=$(GEO_COUNTRY)&user=$(PMUSER_UU)&fingerprint=$(PMUSER_SUU)
```

### Cookie Macros

```json
{
  "URL": "https://example.com/track?cookie=~~c~userid~~&hash=~~c~session~hpr~path~~"
}
```

Generates:
```
https://example.com/track?cookie=$(HTTP_COOKIE{userid})&hash=$(PMUSER_COOKIE_HASH_PR_{session}_{path})
```

### Decode Macros

```json
{
  "URL": "https://example.com/beacon?full_qs=~~dl:qs~~&campaign=~~dl:qs~utm_campaign~~"
}
```

Generates:
```
https://example.com/beacon?full_qs=$(PMUSER_DECODED_QUERY_STRING)&campaign=$(PMUSER_DECODED_QS_{utm_campaign})
```

## Output Files

### HTML Output

The generated HTML file contains:
- ESI functions for advanced macro processing
- ESI includes for each `dir` type pixel
- Fire-and-forget execution with `MAXWAIT=0`

Example:
```html
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>ESI Container Generated Content</title>
</head>
<body>
    <!-- ESI Functions for Advanced Macro Processing -->
    <esi:function name="generate_suu">
        <!-- ... -->
    </esi:function>
    
    <!-- Generated ESI Content -->
    <esi:include src="https://partner1.com/pixel.gif?evid=$(PMUSER_EVID)&time=$(TIME)" maxwait="0" />
    <esi:include src="https://partner2.com/track?cookie=$(HTTP_COOKIE{userid})" maxwait="0" />
</body>
</html>
```

### Browser JSON Output

The browser JSON file contains only `frm` and `script` type pixels for client-side execution:

```json
{
  "pixels": [
    {
      "ID": "partner6_iframe",
      "URL": "https://partner6.com/iframe.html?user=~~uu~~&time=~~r~~",
      "TYPE": "frm",
      "REQ": true,
      "PCT": 100,
      "CAP": 1,
      "RC": "default"
    },
    {
      "ID": "partner7_script",
      "URL": "https://partner7.com/script.js",
      "TYPE": "script",
      "SCRIPT": "console.log('Partner 7 script loaded'); window.partner7Track('~~evid~~', '~~cc~~');",
      "REQ": true,
      "PCT": 100,
      "CAP": 1,
      "RC": "default"
    }
  ]
}
```

## Testing

Run the test suite:

```bash
go test -v ./container_tag.go ./container_tag_test.go
```

Test with example files:

```bash
# Test with basic example
./bin/ESIcontainergenerator -input example.json

# Test with advanced example
./bin/ESIcontainergenerator -input example_advanced.json -output-json browser.json
```

## Examples

See the `example.json` and `example_advanced.json` files for complete examples of the JSON configuration format.

## Architecture

The tool follows a modular design:

1. **JSON Parsing**: Reads and validates partner beacon configurations
2. **Pixel Filtering**: Separates `dir` pixels from `frm`/`script` pixels
3. **Macro Processing**: Converts `~~macro~~` patterns to ESI variables
4. **ESI Generation**: Creates ESI includes with proper syntax
5. **Output Generation**: Produces HTML and optional JSON files

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Update documentation
5. Submit a pull request

## License

This project is part of the Edge Computing Emulator Suite. 