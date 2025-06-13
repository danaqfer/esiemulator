package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"esi-simulator/pkg/esi"
	"esi-simulator/pkg/server"
)

func main() {
	// Command line flags
	var (
		mode    = flag.String("mode", getEnv("ESI_MODE", "akamai"), "ESI processing mode: fastly, akamai, w3c, development")
		port    = flag.String("port", getEnv("PORT", "3000"), "Server port")
		debug   = flag.Bool("debug", getEnv("DEBUG", "false") == "true", "Enable debug logging")
		help    = flag.Bool("help", false, "Show help")
		version = flag.Bool("version", false, "Show version")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	if *version {
		showVersion()
		return
	}

	fmt.Println("🚀 ESI Simulator starting...")

	// Create ESI processor configuration
	config := esi.Config{
		Mode:        *mode,
		Debug:       *debug,
		MaxIncludes: 256,
		MaxDepth:    5,
		Cache: esi.CacheConfig{
			Enabled: true,
			TTL:     300, // 5 minutes
		},
	}

	// Validate mode
	if !isValidMode(*mode) {
		log.Fatalf("❌ Invalid mode: %s. Valid modes: fastly, akamai, w3c, development", *mode)
	}

	// Parse port
	portNum, err := strconv.Atoi(*port)
	if err != nil {
		log.Fatalf("❌ Invalid port: %s", *port)
	}

	fmt.Printf("📋 Configuration:\n")
	fmt.Printf("  - Mode: %s\n", *mode)
	fmt.Printf("  - Port: %s\n", *port)
	fmt.Printf("  - Debug: %t\n", *debug)
	fmt.Printf("  - Cache: %s\n", func() string {
		if config.Cache.Enabled {
			return "enabled"
		}
		return "disabled"
	}())

	// Create ESI processor
	processor := esi.NewProcessor(config)

	// Create and start server
	srv := server.New(processor, server.Config{
		Port:  portNum,
		Debug: *debug,
		Mode:  *mode,
	})

	// Handle graceful shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		fmt.Println("\n🛑 ESI Simulator shutting down...")
		srv.Shutdown()
		os.Exit(0)
	}()

	// Start server
	fmt.Printf("✅ ESI Simulator ready at http://localhost:%s\n", *port)
	fmt.Printf("🎯 Mode: %s\n", *mode)

	if err := srv.Start(); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

func isValidMode(mode string) bool {
	validModes := []string{"fastly", "akamai", "w3c", "development"}
	for _, valid := range validModes {
		if mode == valid {
			return true
		}
	}
	return false
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func showHelp() {
	fmt.Println("ESI Simulator - A comprehensive Edge Side Include processor")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  esi-simulator [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -mode string")
	fmt.Println("        ESI processing mode: fastly, akamai, w3c, development (default \"akamai\")")
	fmt.Println("  -port string")
	fmt.Println("        Server port (default \"3000\")")
	fmt.Println("  -debug")
	fmt.Println("        Enable debug logging")
	fmt.Println("  -help")
	fmt.Println("        Show this help")
	fmt.Println("  -version")
	fmt.Println("        Show version")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  ESI_MODE    ESI processing mode")
	fmt.Println("  PORT        Server port")
	fmt.Println("  DEBUG       Enable debug logging (true/false)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  esi-simulator                           # Start with default settings")
	fmt.Println("  esi-simulator -mode fastly -debug      # Fastly mode with debug")
	fmt.Println("  ESI_MODE=w3c PORT=8080 esi-simulator   # W3C mode on port 8080")
}

func showVersion() {
	fmt.Println("ESI Simulator v0.1.0")
	fmt.Println("A comprehensive Edge Side Include processor supporting Fastly and Akamai implementations")
}
