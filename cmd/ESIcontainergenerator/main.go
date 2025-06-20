package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/edge-computing/emulator-suite/pkg/esi"
)

// Command line flags
var (
	inputFile   = flag.String("input", "", "Input JSON configuration file (required)")
	outputFile  = flag.String("output", "", "Output HTML file (optional, defaults to input filename with .html extension)")
	verbose     = flag.Bool("verbose", false, "Enable verbose output")
	browserVars = flag.Bool("browser-vars", false, "Use browser-like ESI variable substitution (default: static macros only)")
	help        = flag.Bool("help", false, "Show help information")
)

// Example configuration structure for demonstration
type ExampleConfig struct {
	ClientID    string                `json:"clientId"`
	PropertyID  string                `json:"propertyId"`
	Environment string                `json:"environment"`
	Version     string                `json:"version"`
	Beacons     []esi.PartnerBeacon   `json:"beacons"`
	Settings    esi.ContainerSettings `json:"settings"`
	Macros      map[string]string     `json:"macros"`
}

func main() {
	flag.Parse()

	if *help {
		printHelp()
		return
	}

	if *inputFile == "" {
		log.Fatal("Error: Input file is required. Use -input flag to specify the JSON configuration file.")
	}

	// Read and parse the JSON configuration file
	config, err := readConfigFile(*inputFile)
	if err != nil {
		log.Fatalf("Error reading configuration file: %v", err)
	}

	// Determine output filename
	outputFilename := *outputFile
	if outputFilename == "" {
		// Generate output filename based on input filename
		baseName := strings.TrimSuffix(filepath.Base(*inputFile), filepath.Ext(*inputFile))
		outputFilename = baseName + ".html"
	}

	// Generate the ESI HTML content
	htmlContent, err := generateESIHTML(config)
	if err != nil {
		log.Fatalf("Error generating ESI HTML: %v", err)
	}

	// Write the output file
	err = writeOutputFile(outputFilename, htmlContent)
	if err != nil {
		log.Fatalf("Error writing output file: %v", err)
	}

	if *verbose {
		printStats(config)
	}

	fmt.Printf("✅ Successfully generated ESI container HTML: %s\n", outputFilename)
	fmt.Printf("📊 Processed %d beacons (%d enabled)\n", len(config.Beacons), countEnabledBeacons(config.Beacons))
}

func readConfigFile(filename string) (*esi.ContainerTagConfig, error) {
	if *verbose {
		fmt.Printf("📖 Reading configuration file: %s\n", filename)
	}

	// Read the file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Try to parse as ContainerTagConfig first
	var config esi.ContainerTagConfig
	err = json.Unmarshal(data, &config)
	if err == nil {
		// Successfully parsed as ContainerTagConfig
		setDefaults(&config)
		return &config, nil
	}

	// If that fails, try parsing as ExampleConfig and convert
	if *verbose {
		fmt.Printf("⚠️  Failed to parse as ContainerTagConfig, trying ExampleConfig format...\n")
	}

	var exampleConfig ExampleConfig
	err = json.Unmarshal(data, &exampleConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Convert ExampleConfig to ContainerTagConfig
	config = esi.ContainerTagConfig{
		ClientID:    exampleConfig.ClientID,
		PropertyID:  exampleConfig.PropertyID,
		Environment: exampleConfig.Environment,
		Version:     exampleConfig.Version,
		CreatedAt:   time.Now(),
		Beacons:     exampleConfig.Beacons,
		Settings:    exampleConfig.Settings,
		Macros:      exampleConfig.Macros,
	}

	setDefaults(&config)
	return &config, nil
}

func setDefaults(config *esi.ContainerTagConfig) {
	// Set default values if not provided
	if config.Settings.DefaultTimeout == 0 {
		config.Settings.DefaultTimeout = 5000 // 5 seconds
	}
	if config.Settings.MaxWait == 0 {
		config.Settings.MaxWait = 0 // Fire and forget
	}
	if config.Settings.DefaultMethod == "" {
		config.Settings.DefaultMethod = "GET"
	}
	if config.Settings.MaxConcurrentBeacons == 0 {
		config.Settings.MaxConcurrentBeacons = 10
	}
	if config.CreatedAt.IsZero() {
		config.CreatedAt = time.Now()
	}
}

func generateESIHTML(config *esi.ContainerTagConfig) (string, error) {
	if *verbose {
		fmt.Printf("🔧 Generating ESI HTML for client: %s, property: %s\n", config.ClientID, config.PropertyID)
		if *browserVars {
			fmt.Printf("🌐 Using browser-like ESI variable substitution\n")
		} else {
			fmt.Printf("📝 Using static macro substitution only\n")
		}
	}

	// Create ESI processor
	esiProcessor := esi.NewProcessor(esi.Config{
		Mode:        "akamai",
		Debug:       *verbose,
		MaxIncludes: 256,
		MaxDepth:    5,
		Cache: esi.CacheConfig{
			Enabled: true,
			TTL:     300,
		},
	})

	// Create container tag processor
	ctp := esi.NewContainerTagProcessor(*config, esiProcessor)

	var htmlContent string
	var err error

	// Choose generation method based on browser-vars flag
	if *browserVars {
		// Generate HTML with ESI variable expressions that will be resolved at runtime
		htmlContent, err = ctp.GenerateCompleteESIHTMLWithBrowserVariables()
	} else {
		// Generate HTML with static macro substitution only
		htmlContent, err = ctp.GenerateCompleteESIHTML()
	}

	if err != nil {
		return "", fmt.Errorf("failed to generate ESI HTML: %w", err)
	}

	return htmlContent, nil
}

func writeOutputFile(filename string, content string) error {
	if *verbose {
		fmt.Printf("💾 Writing output file: %s\n", filename)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if dir != "." && dir != "" {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write the file
	err := ioutil.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	return nil
}

func countEnabledBeacons(beacons []esi.PartnerBeacon) int {
	count := 0
	for _, beacon := range beacons {
		if beacon.Enabled {
			count++
		}
	}
	return count
}

func printStats(config *esi.ContainerTagConfig) {
	fmt.Println("\n📊 Configuration Statistics:")
	fmt.Printf("   Client ID: %s\n", config.ClientID)
	fmt.Printf("   Property ID: %s\n", config.PropertyID)
	fmt.Printf("   Environment: %s\n", config.Environment)
	fmt.Printf("   Version: %s\n", config.Version)
	fmt.Printf("   Total Beacons: %d\n", len(config.Beacons))
	fmt.Printf("   Enabled Beacons: %d\n", countEnabledBeacons(config.Beacons))
	fmt.Printf("   Default Timeout: %dms\n", config.Settings.DefaultTimeout)
	fmt.Printf("   Max Wait: %dms\n", config.Settings.MaxWait)
	fmt.Printf("   Fire and Forget: %t\n", config.Settings.FireAndForget)
	fmt.Printf("   Macros Defined: %d\n", len(config.Macros))

	// Count beacons by category
	categories := make(map[string]int)
	for _, beacon := range config.Beacons {
		if beacon.Category != "" {
			categories[beacon.Category]++
		}
	}

	if len(categories) > 0 {
		fmt.Println("   Beacon Categories:")
		for category, count := range categories {
			fmt.Printf("     %s: %d\n", category, count)
		}
	}

	if len(config.Macros) > 0 {
		fmt.Println("   Macros:")
		for macro, value := range config.Macros {
			fmt.Printf("     ${%s}: %s\n", macro, value)
		}
	}
}

func printHelp() {
	fmt.Println("ESI Container Generator")
	fmt.Println("=======================")
	fmt.Println()
	fmt.Println("Generates HTML files with ESI embedded code from JSON configuration files.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ESIcontainergenerator -input <config.json> [-output <output.html>] [-verbose] [-browser-vars]")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -input string")
	fmt.Println("        Input JSON configuration file (required)")
	fmt.Println("  -output string")
	fmt.Println("        Output HTML file (optional, defaults to input filename with .html extension)")
	fmt.Println("  -verbose")
	fmt.Println("        Enable verbose output")
	fmt.Println("  -browser-vars")
	fmt.Println("        Use browser-like ESI variable substitution (default: static macros only)")
	fmt.Println("        When enabled, generates ESI expressions like $(HTTP_USER_AGENT) that are")
	fmt.Println("        resolved at runtime by the ESI processor for each browser request.")
	fmt.Println("  -help")
	fmt.Println("        Show this help information")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ESIcontainergenerator -input beacon-config.json -output container.html -verbose")
	fmt.Println("  ESIcontainergenerator -input beacon-config.json -browser-vars -output dynamic-container.html")
	fmt.Println()
	fmt.Println("Variable Substitution Modes:")
	fmt.Println("  Static Mode (default):")
	fmt.Println("    - Substitutes only static macros from JSON configuration")
	fmt.Println("    - All values are resolved at generation time")
	fmt.Println("    - Example: ${USER_ID} -> 12345")
	fmt.Println()
	fmt.Println("  Browser Variables Mode (-browser-vars):")
	fmt.Println("    - Generates ESI expressions for dynamic browser data")
	fmt.Println("    - Variables are resolved at runtime by ESI processor")
	fmt.Println("    - Example: ${USER_AGENT} -> $(HTTP_USER_AGENT)")
	fmt.Println("    - Example: ${CLIENT_IP} -> $(CLIENT_IP)")
	fmt.Println("    - Example: ${PAGE_URL} -> $(REQUEST_URI)")
	fmt.Println()
	fmt.Println("JSON Configuration Format:")
	fmt.Println("  {")
	fmt.Println("    \"clientId\": \"client-123\",")
	fmt.Println("    \"propertyId\": \"property-456\",")
	fmt.Println("    \"environment\": \"production\",")
	fmt.Println("    \"version\": \"1.0.0\",")
	fmt.Println("    \"beacons\": [")
	fmt.Println("      {")
	fmt.Println("        \"id\": \"beacon1\",")
	fmt.Println("        \"name\": \"Analytics Beacon\",")
	fmt.Println("        \"url\": \"https://analytics.example.com/pixel\",")
	fmt.Println("        \"method\": \"GET\",")
	fmt.Println("        \"enabled\": true,")
	fmt.Println("        \"category\": \"analytics\",")
	fmt.Println("        \"parameters\": {")
	fmt.Println("          \"user_id\": \"${USER_ID}\",")
	fmt.Println("          \"site_id\": \"${SITE_ID}\",")
	fmt.Println("          \"user_agent\": \"${USER_AGENT}\",")
	fmt.Println("          \"client_ip\": \"${CLIENT_IP}\"")
	fmt.Println("        }")
	fmt.Println("      }")
	fmt.Println("    ],")
	fmt.Println("    \"settings\": {")
	fmt.Println("      \"defaultTimeout\": 5000,")
	fmt.Println("      \"fireAndForget\": true,")
	fmt.Println("      \"maxWait\": 0,")
	fmt.Println("      \"enableLogging\": true")
	fmt.Println("    },")
	fmt.Println("    \"macros\": {")
	fmt.Println("      \"USER_ID\": \"12345\",")
	fmt.Println("      \"SITE_ID\": \"example.com\"")
	fmt.Println("    }")
	fmt.Println("  }")
}
