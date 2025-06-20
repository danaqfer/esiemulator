# ESI Container Generator

A Go program that reads JSON configuration files and generates HTML files with ESI (Edge Side Include) embedded code for container tag functionality. This tool translates partner beacon configurations into fire-and-forget ESI includes with `MAXWAIT=0`.

## Overview

The ESI Container Generator takes a JSON configuration file that defines partner beacons (tracking pixels, analytics endpoints, etc.) and converts them into ESI include statements that can be embedded in HTML pages. Each beacon is converted to an `<esi:include>` element with `MAXWAIT=0` for fire-and-forget behavior.

## Features

- **JSON Configuration**: Define partner beacons in a structured JSON format
- **Fire-and-Forget**: All beacons use `MAXWAIT=0` for non-blocking execution
- **Macro Substitution**: Support for variable substitution in beacon URLs and parameters
- **Conditional Firing**: Support for conditional beacon firing based on various criteria
- **Multiple Categories**: Organize beacons by category (analytics, advertising, etc.)
- **Flexible Output**: Generate complete HTML documents or just ESI fragments
- **Verbose Logging**: Detailed output for debugging and monitoring

## Installation

### Prerequisites

- Go 1.21 or later
- Access to the edge-computing/emulator-suite repository

### Building

```bash
# From the repository root
go build -o bin/ESIcontainergenerator cmd/ESIcontainergenerator/main.go
```

### Using Makefile

```bash
# Build all binaries including ESIcontainergenerator
make build
```

## Usage

### Basic Usage

```bash
# Generate HTML from JSON configuration
ESIcontainergenerator -input config.json

# Specify output file
ESIcontainergenerator -input config.json -output container.html

# Enable verbose output
ESIcontainergenerator -input config.json -output container.html -verbose
```

### Command Line Options

- `-input string`: Input JSON configuration file (required)
- `-output string`: Output HTML file (optional, defaults to input filename with .html extension)
- `-verbose`: Enable verbose output
- `-help`: Show help information

### Example

```bash
# Generate HTML from example configuration
ESIcontainergenerator -input example-config.json -output demo-container.html -verbose
```

## JSON Configuration Format

### Basic Structure

```json
{
  "clientId": "client-123",
  "propertyId": "property-456",
  "environment": "production",
  "version": "1.0.0",
  "beacons": [
    {
      "id": "beacon1",
      "name": "Analytics Beacon",
      "url": "https://analytics.example.com/pixel",
      "method": "GET",
      "enabled": true,
      "category": "analytics",
      "parameters": {
        "user_id": "${USER_ID}",
        "site_id": "${SITE_ID}"
      }
    }
  ],
  "settings": {
    "defaultTimeout": 5000,
    "fireAndForget": true,
    "maxWait": 0,
    "enableLogging": true
  },
  "macros": {
    "USER_ID": "12345",
    "SITE_ID": "example.com"
  }
}
```

### Configuration Fields

#### Top-Level Fields

- `clientId` (string): Unique client identifier
- `propertyId` (string): Property or site identifier
- `environment` (string): Environment (dev, staging, prod)
- `version` (string): Configuration version
- `beacons` (array): Array of partner beacon configurations
- `settings` (object): Container-wide settings
- `macros` (object): Macro definitions for variable substitution

#### Beacon Configuration

- `id` (string): Unique beacon identifier
- `name` (string): Human-readable beacon name
- `url` (string): The beacon URL to fire
- `method` (string): HTTP method (GET, POST, etc.)
- `enabled` (boolean): Whether this beacon is enabled
- `category` (string): Beacon category (analytics, advertising, etc.)
- `description` (string): Description of the beacon
- `timeout` (integer): Custom timeout in milliseconds
- `parameters` (object): Query parameters or POST data
- `headers` (object): Additional HTTP headers
- `conditions` (object): Conditional firing rules
- `frequency` (string): Firing frequency (always, once, etc.)
- `priority` (integer): Priority for firing order

#### Settings Configuration

- `maxConcurrentBeacons` (integer): Maximum concurrent beacon fires
- `defaultTimeout` (integer): Default timeout in milliseconds
- `fireAndForget` (boolean): Whether to use fire-and-forget mode
- `maxWait` (integer): MAXWAIT value for ESI includes
- `enableLogging` (boolean): Whether to enable logging
- `enableErrorHandling` (boolean): Whether to handle errors
- `defaultMethod` (string): Default HTTP method

## Generated ESI Output

The tool generates HTML with ESI includes that look like this:

```html
<!DOCTYPE html>
<html>
<head>
    <title>Container Tag ESI</title>
    <meta charset="utf-8">
</head>
<body>
    <!-- Container Tag Generated Content -->
    <!--esi Container Tag Generated ESI -->
    <!--esi Client: demo-client-123 -->
    <!--esi Property: demo-property-456 -->
    <!--esi Environment: development -->
    <!--esi Generated: 2024-01-15T10:30:00Z -->

    <esi:include src="https://www.google-analytics.com/collect?v=1&tid=UA-123456789-1&cid=demo-user-123&t=pageview&dp=/demo-page&dt=Demo Page - Example.com&uip=192.168.1.100&ua=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" maxwait="0" timeout="5000" onerror="continue" alt="" /> <!-- Beacon: Google Analytics (google-analytics) -->

    <esi:include src="https://www.facebook.com/tr?id=123456789012345&ev=PageView&dl=https://example.com/demo-page&dt=Demo Page - Example.com&uip=192.168.1.100&ua=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" maxwait="0" timeout="5000" onerror="continue" alt="" /> <!-- Beacon: Facebook Pixel (facebook-pixel) -->

    <!-- End Container Tag Content -->
</body>
</html>
```

## Key Features

### Fire-and-Forget Behavior

All generated ESI includes use `MAXWAIT=0` and `onerror="continue"`, which means:
- Beacons fire asynchronously without blocking the page
- Failed beacons don't prevent other beacons from firing
- No tracking of beacon success/failure (as requested)

### Macro Substitution

The tool supports macro substitution in beacon URLs and parameters:

```json
{
  "macros": {
    "USER_ID": "12345",
    "SITE_ID": "example.com"
  },
  "beacons": [
    {
      "url": "https://analytics.example.com/pixel",
      "parameters": {
        "user_id": "${USER_ID}",
        "site_id": "${SITE_ID}"
      }
    }
  ]
}
```

### Built-in Macros

The tool automatically provides these built-in macros:
- `${CLIENT_ID}`: The client ID from configuration
- `${PROPERTY_ID}`: The property ID from configuration
- `${ENVIRONMENT}`: The environment from configuration
- `${TIMESTAMP}`: Current Unix timestamp
- `${RANDOM}`: Current Unix timestamp in nanoseconds

### Conditional Firing

Beacons can be conditionally enabled/disabled:

```json
{
  "beacons": [
    {
      "id": "geo-specific",
      "enabled": true,
      "conditions": {
        "country": "US",
        "consent": "required"
      }
    }
  ]
}
```

## Example Configuration

See `example-config.json` for a complete example configuration that includes:
- Google Analytics beacon
- Facebook Pixel
- Twitter Pixel
- LinkedIn Pixel (disabled)
- Custom analytics endpoint

## Integration with ESI Emulator

The generated HTML can be processed by the ESI emulator to test the beacon firing behavior:

```bash
# Start the ESI emulator
go run cmd/edge-emulator/main.go -mode akamai

# Process the generated HTML
curl -X POST http://localhost:3000/process \
  -H "Content-Type: application/json" \
  -d '{"html": "<esi:include src=\"https://example.com/beacon\" maxwait=\"0\" />"}'
```

## Error Handling

The tool provides comprehensive error handling:
- Validates JSON configuration format
- Checks for required fields
- Provides detailed error messages
- Graceful handling of missing optional fields

## Performance Considerations

- **Concurrent Processing**: All beacons fire concurrently due to `MAXWAIT=0`
- **No Batching**: Each beacon is a separate ESI include
- **No Pagination**: All enabled beacons are included in a single HTML file
- **Fire-and-Forget**: No waiting for beacon responses

## Limitations

- No tracking of beacon success/failure (by design)
- No retry logic for failed beacons
- No batching or pagination of beacons
- Limited to GET and POST methods
- Basic conditional logic support

## Contributing

To contribute to the ESI Container Generator:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## License

This project is part of the Edge Computing Emulator Suite and follows the same license terms. 