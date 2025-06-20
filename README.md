# Edge Computing Emulator Suite

A comprehensive suite of edge computing emulators written in Go that provides development and testing capabilities for modern edge computing platforms.

## Overview

This project provides standalone, production-ready emulators for major edge computing technologies:

- **ESI Emulator** - Edge Side Include processor supporting Fastly and Akamai implementations
- **Property Manager Emulator** - Akamai Property Manager for traffic management and content delivery
- **HTTP Server** - Common HTTP server infrastructure for both emulators

## Components

### 🚀 [ESI Emulator](./pkg/esi/README.md)

A comprehensive Edge Side Include (ESI) emulator that supports both Fastly's limited ESI functionality and Akamai's extended ESI capabilities.

**Key Features:**
- Full ESI 1.0 specification compliance
- Fastly and Akamai implementation modes
- Advanced variable substitution and conditional processing
- Error handling with try/catch blocks
- Comment block processing for graceful degradation
- Akamai-specific extensions (assign, eval, functions, debug)

**Use Cases:**
- Content assembly at the edge
- Dynamic content composition
- A/B testing and personalization
- Performance optimization through selective caching

[📖 Read ESI Documentation](./pkg/esi/README.md)

### 🎯 [Property Manager Emulator](./pkg/propertymanager/README.md)

A comprehensive Akamai Property Manager emulator that simulates Akamai's edge computing platform for traffic management and content delivery.

**Key Features:**
- Hierarchical rule-based traffic management
- Comprehensive criteria evaluation (path, headers, cookies, etc.)
- Complete behavior library (caching, security, performance, content)
- HTTP context processing and simulation
- Thread-safe concurrent processing
- Performance optimization and monitoring

**Use Cases:**
- Traffic management and routing
- Content delivery optimization
- Security policy enforcement
- Performance monitoring and analytics

[📖 Read Property Manager Documentation](./pkg/propertymanager/README.md)

### 🌐 HTTP Server

A common HTTP server infrastructure that provides RESTful APIs for both ESI and Property Manager emulators.

**Key Features:**
- RESTful API endpoints for processing requests
- Built-in examples and test fragments
- Statistics and monitoring endpoints
- Health checks and cache management
- CORS support and error handling

## Quick Start

### Installation

1. **Clone this repository**
2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

### Building

**Option A: Using PowerShell (Windows):**
```powershell
.\build.ps1 build
```

**Option B: Using Make (Linux/macOS):**
```bash
make build
```

**Option C: Direct Go build:**
```bash
go build -o edge-emulator
```

### Running

#### ESI Emulator

**PowerShell (Windows):**
```powershell
# Akamai mode (default) - full ESI features
.\build.ps1 run

# Fastly mode - limited ESI features
.\build.ps1 run-fastly

# W3C specification mode
.\build.ps1 run-w3c
```

**Linux/macOS:**
```bash
# Akamai mode (default) - full ESI features
make run

# Fastly mode - limited ESI features
make run-fastly

# W3C specification mode
make run-w3c
```

#### Property Manager Emulator

**PowerShell (Windows):**
```powershell
# Property Manager mode
$env:EMULATOR_MODE="property-manager"; $env:GOWORK="off"; go run main.go
```

**Linux/macOS:**
```bash
# Property Manager mode
EMULATOR_MODE=property-manager go run main.go
```

### API Usage

#### ESI Processing

```bash
curl -X POST http://localhost:3000/process \
  -H "Content-Type: application/json" \
  -d '{"html": "<esi:include src=\"/fragments/header\" />Hello World!"}'
```

#### Property Manager Processing

```bash
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

## Architecture

```
edge-computing-emulator/
├── pkg/
│   ├── esi/                    # ESI Emulator
│   │   ├── processor.go        # Core ESI processing engine
│   │   ├── akamai_extensions.go # Akamai-specific extensions
│   │   ├── processor_test.go   # Comprehensive test suite
│   │   └── README.md          # ESI documentation
│   ├── propertymanager/        # Property Manager Emulator
│   │   ├── processor.go        # Rule processing engine
│   │   ├── behaviors.go        # Behavior library
│   │   ├── types.go           # Type definitions
│   │   ├── propertymanager_test.go # Test suite
│   │   └── README.md          # Property Manager documentation
│   └── server/                # HTTP Server
│       └── server.go          # Common server infrastructure
├── main.go                    # Application entry point
├── build.ps1                  # PowerShell build script
├── Makefile                   # Make build script
└── README.md                  # This file
```

## Development

### Running Tests

**PowerShell (Windows):**
```powershell
.\build.ps1 test
```

**Linux/macOS:**
```bash
make test
```

### Available Commands

**PowerShell (Windows):**
```powershell
.\build.ps1 help          # Show all available commands
.\build.ps1 build         # Build the application
.\build.ps1 test          # Run tests
.\build.ps1 clean         # Clean build artifacts
.\build.ps1 run           # Run ESI emulator (Akamai mode)
.\build.ps1 run-fastly    # Run ESI emulator (Fastly mode)
.\build.ps1 run-w3c       # Run ESI emulator (W3C mode)
.\build.ps1 examples      # Run example programs
```

**Make (Linux/macOS):**
```bash
make help                 # Show all available commands
make build                # Build the application  
make test                 # Run tests
make clean                # Clean build artifacts
make run                  # Run ESI emulator (Akamai mode)
make run-fastly           # Run ESI emulator (Fastly mode)
make run-w3c              # Run ESI emulator (W3C mode)
make examples             # Run example programs
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `3000` |
| `EMULATOR_MODE` | Emulator mode (`esi`, `property-manager`) | `esi` |
| `ESI_MODE` | ESI mode (`fastly`, `akamai`, `w3c`, `development`) | `akamai` |
| `DEBUG` | Enable debug mode | `false` |

### Command Line Flags

```bash
go run main.go -help
```

Available flags:
- `-port` - Port to run the server on (default: 3000)
- `-mode` - Emulator mode: esi, property-manager (default: esi)
- `-esi-mode` - ESI mode: fastly, akamai, w3c, development (default: akamai)
- `-debug` - Enable debug mode
- `-help` - Show help information
- `-version` - Show version

## Current Status

### ✅ Fully Implemented

**ESI Emulator:**
- Complete ESI 1.0 specification compliance
- Fastly and Akamai implementation modes
- Advanced variable substitution and conditional processing
- Error handling with try/catch blocks
- Comment block processing
- Akamai-specific extensions
- 140+ comprehensive tests

**Property Manager Emulator:**
- Hierarchical rule-based traffic management
- Comprehensive criteria evaluation
- Complete behavior library
- HTTP context processing
- Thread-safe concurrent processing
- 700+ comprehensive tests

**HTTP Server:**
- RESTful API endpoints
- Built-in examples and test fragments
- Statistics and monitoring
- Health checks and cache management
- CORS support and error handling

### 📋 Future Enhancements

- **Streaming Processing**: Stream-based processing for large documents
- **Advanced Caching**: Redis/Memcached backends for distributed caching
- **Web-based Interfaces**: Browser-based configuration and testing
- **Load Testing**: Concurrent request testing and stress testing
- **Integration Examples**: Docker, Kubernetes, reverse proxy configurations
- **Performance Profiling**: Detailed performance analysis and optimization

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## References

- [W3C ESI Language Specification 1.0](https://www.w3.org/TR/esi-lang/)
- [Fastly ESI Documentation](https://www.fastly.com/documentation/reference/vcl/statements/esi/)
- [Akamai ESI Developer's Guide](https://www.akamai.com/site/zh/documents/technical-publication/akamai-esi-developers-guide-technical-publication.pdf)
- [Akamai Property Manager Documentation](https://techdocs.akamai.com/property-manager/docs) 