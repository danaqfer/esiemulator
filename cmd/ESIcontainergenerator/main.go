package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/edge-computing/emulator-suite/pkg/esi"
)

func main() {
	// Define command line flags
	inputFile := flag.String("input", "", "Input JSON configuration file")
	outputFile := flag.String("output", "", "Output HTML file (default: input_name.html)")
	browserVars := flag.Bool("browser-vars", false, "Use browser-like ESI variable substitution")
	maxWait := flag.Int("maxwait", 0, "Maximum wait time for ESI includes (default: 0 for fire-and-forget)")
	outputJSON := flag.String("output-json", "", "Output JSON file for browser-executed pixels (frm/script types)")
	showHelp := flag.Bool("help", false, "Show help information")

	flag.Parse()

	if *showHelp {
		printHelp()
		return
	}

	// Validate required input file
	if *inputFile == "" {
		log.Fatal("Error: Input file is required. Use -input flag to specify the JSON configuration file.")
	}

	// Read input JSON file
	inputData, err := ioutil.ReadFile(*inputFile)
	if err != nil {
		log.Fatalf("Error reading input file: %v", err)
	}

	// Parse JSON configuration
	var config esi.ContainerConfig
	if err := json.Unmarshal(inputData, &config); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	// Create ESI configuration
	esiConfig := esi.ESIConfig{
		BrowserVars: *browserVars,
		MaxWait:     *maxWait,
	}

	// Process the configuration
	esiContent, browserConfig, err := esi.ProcessContainerConfig(config, esiConfig)
	if err != nil {
		log.Fatalf("Error processing configuration: %v", err)
	}

	// Generate output filename if not provided
	if *outputFile == "" {
		baseName := strings.TrimSuffix(filepath.Base(*inputFile), filepath.Ext(*inputFile))
		*outputFile = baseName + ".html"
	}

	// Generate the complete HTML content
	htmlContent := generateHTMLContent(esiContent, esiConfig)

	// Write HTML output
	if err := ioutil.WriteFile(*outputFile, []byte(htmlContent), 0644); err != nil {
		log.Fatalf("Error writing HTML output file: %v", err)
	}

	fmt.Printf("✅ Generated HTML file: %s\n", *outputFile)
	fmt.Printf("📊 Processed %d pixels:\n", len(config.Pixels))

	// Count pixel types
	dirCount := 0
	frmCount := 0
	scriptCount := 0
	for _, pixel := range config.Pixels {
		switch pixel.TYPE {
		case "dir":
			dirCount++
		case "frm":
			frmCount++
		case "script":
			scriptCount++
		}
	}

	fmt.Printf("   - %d 'dir' pixels → ESI includes\n", dirCount)
	fmt.Printf("   - %d 'frm' pixels → Browser execution\n", frmCount)
	fmt.Printf("   - %d 'script' pixels → Browser execution\n", scriptCount)

	// Generate browser JSON file if requested
	if *outputJSON != "" {
		if err := generateBrowserJSON(browserConfig, *outputJSON); err != nil {
			log.Fatalf("Error writing browser JSON file: %v", err)
		}
		fmt.Printf("✅ Generated browser JSON file: %s\n", *outputJSON)
		fmt.Printf("📋 Browser JSON contains %d pixels for client-side execution\n", len(browserConfig.Pixels))
	}

	// Show configuration details
	fmt.Printf("\n🔧 Configuration:\n")
	fmt.Printf("   - Browser variables: %t\n", esiConfig.BrowserVars)
	fmt.Printf("   - Max wait time: %d\n", esiConfig.MaxWait)
	fmt.Printf("   - Fire-and-forget: %t\n", esiConfig.MaxWait == 0)
}

func generateHTMLContent(esiContent string, config esi.ESIConfig) string {
	var html strings.Builder

	html.WriteString("<!DOCTYPE html>\n")
	html.WriteString("<html>\n")
	html.WriteString("<head>\n")
	html.WriteString("    <meta charset=\"UTF-8\">\n")
	html.WriteString("    <title>ESI Container Generated Content</title>\n")
	html.WriteString("</head>\n")
	html.WriteString("<body>\n")
	html.WriteString("    <!-- ESI Functions for Advanced Macro Processing -->\n")
	html.WriteString(esi.GenerateESIFunctions())
	html.WriteString("\n\n")
	html.WriteString("    <!-- Generated ESI Content -->\n")
	html.WriteString(esiContent)
	html.WriteString("\n</body>\n")
	html.WriteString("</html>\n")

	return html.String()
}

func generateBrowserJSON(config esi.ContainerConfig, outputFile string) error {
	// Convert to JSON
	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling browser config to JSON: %w", err)
	}

	// Write to file
	if err := ioutil.WriteFile(outputFile, jsonData, 0644); err != nil {
		return fmt.Errorf("error writing browser JSON file: %w", err)
	}

	return nil
}

func printHelp() {
	fmt.Println("ESI Container Generator")
	fmt.Println("=======================")
	fmt.Println()
	fmt.Println("Converts JSON partner beacon configurations into ESI includes for server-side execution.")
	fmt.Println("Filters 'frm' and 'script' type pixels for browser execution.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ESIcontainergenerator -input config.json [options]")
	fmt.Println()
	fmt.Println("Required Flags:")
	fmt.Println("  -input string")
	fmt.Println("        Input JSON configuration file")
	fmt.Println()
	fmt.Println("Optional Flags:")
	fmt.Println("  -output string")
	fmt.Println("        Output HTML file (default: input_name.html)")
	fmt.Println("  -output-json string")
	fmt.Println("        Output JSON file for browser-executed pixels (frm/script types)")
	fmt.Println("  -browser-vars")
	fmt.Println("        Use browser-like ESI variable substitution")
	fmt.Println("  -maxwait int")
	fmt.Println("        Maximum wait time for ESI includes (default: 0 for fire-and-forget)")
	fmt.Println("  -help")
	fmt.Println("        Show this help information")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Basic conversion")
	fmt.Println("  ESIcontainergenerator -input partner_beacons.json")
	fmt.Println()
	fmt.Println("  # With browser JSON output")
	fmt.Println("  ESIcontainergenerator -input partner_beacons.json -output-json browser_pixels.json")
	fmt.Println()
	fmt.Println("  # With browser variables")
	fmt.Println("  ESIcontainergenerator -input partner_beacons.json -browser-vars")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  ✅ Converts 'dir' type pixels to ESI includes")
	fmt.Println("  ✅ Filters 'frm' and 'script' pixels for browser execution")
	fmt.Println("  ✅ Supports advanced macro substitution")
	fmt.Println("  ✅ Generates fingerprint IDs (suu)")
	fmt.Println("  ✅ Handles cookie hashing (hpr/hpo)")
	fmt.Println("  ✅ URL decoding support")
	fmt.Println("  ✅ Fire-and-forget execution (MAXWAIT=0)")
}
