package asor_examples

import (
	"fmt"
	"log"

	"sqloptimizer/asor"
)

func ExampleASOR() {
	// Create a new optimizer with default configuration
	optimizer := asor.NewOptimizer(nil)

	// Example query to optimize
	query := `
		SELECT *
		FROM customers
		WHERE region = 'NA'
		OR region = 'EU'
		OR region = 'APAC'
	`

	// Optimize for different dialects
	dialects := []string{"athena", "bigquery", "teradata", "postgresql"}

	for _, dialect := range dialects {
		optimized, err := optimizer.OptimizeQuery(query, dialect)
		if err != nil {
			log.Printf("Error optimizing for %s: %v", dialect, err)
			continue
		}
		fmt.Printf("\nDialect: %s\nOptimized Query:\n%s\n", dialect, optimized)
	}

	// Example with custom configuration
	customConfig := &asor.OptimizationConfig{
		EnableTableRouting:           true,
		EnableORToIN:                 false, // Disable OR to IN conversion
		DialectSpecificOptimizations: true,
	}

	customOptimizer := asor.NewOptimizer(customConfig)
	optimized, err := customOptimizer.OptimizeQuery(query, "postgresql")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nCustom Config (PostgreSQL):\n%s\n", optimized)
}
