package main

import (
	"fmt"
	"log"

	"sqlparser/dialect"
	"sqlparser/optimizer"
	"sqlparser/parser"
	"sqlparser/testutil"
)

func main() {
	// Example query demonstrating table routing and OR to IN optimization
	query := `
		SELECT 
			id,
			name,
			email,
			status
		FROM customers
		WHERE region = 'NA'  -- Should route to customers_a (partitioned by region)
		AND (
			status = 'ACTIVE' OR 
			status = 'PENDING' OR 
			status = 'NEW'
		)
	`

	// Create metadata provider
	metadataProvider := &testutil.MockMetadataProvider{}

	// Process query with different dialects
	dialects := map[string]dialect.Dialect{
		"Athena":     dialect.NewAthenaDialect(),
		"BigQuery":   dialect.NewBigQueryDialect(),
		"PostgreSQL": dialect.NewPostgresDialect(),
		"Teradata":   dialect.NewTeradataDialect(),
	}

	for name, d := range dialects {
		fmt.Printf("\nProcessing with %s dialect:\n", name)
		fmt.Printf("Original query:\n%s\n", query)

		// Create optimization config
		config := optimizer.NewDefaultConfig()

		// Parse query
		p := parser.NewParser(query, d)
		astNode, err := p.Parse()
		if err != nil {
			log.Printf("Failed to parse query for %s: %v\n", name, err)
			continue
		}

		// Apply generic optimizations (during parse phase)
		fmt.Printf("\nApplying generic optimizations...\n")
		optimizedNode, err := optimizer.ApplyGenericOptimizations(astNode, metadataProvider, config)
		if err != nil {
			log.Printf("Failed to apply generic optimizations for %s: %v\n", name, err)
			continue
		}

		// Show intermediate state
		generator := dialect.NewBaseGenerator(d)
		intermediateSQL, err := generator.GenerateSQL(optimizedNode)
		if err != nil {
			log.Printf("Failed to generate intermediate SQL for %s: %v\n", name, err)
			continue
		}
		fmt.Printf("\nAfter generic optimizations:\n%s\n", intermediateSQL)

		// Apply dialect-specific optimizations
		fmt.Printf("\nApplying dialect-specific optimizations...\n")
		finalNode, err := optimizer.ApplyDialectOptimizations(optimizedNode, config, name)
		if err != nil {
			log.Printf("Failed to apply dialect-specific optimizations for %s: %v\n", name, err)
			continue
		}

		// Generate final SQL
		finalSQL, err := generator.GenerateSQL(finalNode)
		if err != nil {
			log.Printf("Failed to generate final SQL for %s: %v\n", name, err)
			continue
		}
		fmt.Printf("\nFinal optimized SQL:\n%s\n", finalSQL)
	}
}
