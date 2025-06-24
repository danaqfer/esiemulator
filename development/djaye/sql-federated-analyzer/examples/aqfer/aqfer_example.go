package aqfer_examples

import (
	"fmt"
	"log"

	"sqloptimizer/dialect"
	"sqloptimizer/parser"
)

func ExampleAqferParser() {
	// Example query with an Aqfer table and dimension joins
	query := `
		SELECT 
			a.campaign_id,
			c.campaign_name,
			a.impressions,
			a.clicks,
			a.conversions
		FROM my_aqfer_table.dsae a
		INNER JOIN campaigns c ON a.campaign_id = c.id
		WHERE a.date_key >= '2024-01-01'
		GROUP BY a.campaign_id, c.campaign_name
	`

	// Create an Athena dialect with Aqfer support
	athenaAqfer := dialect.NewAthenaAqferDialect()

	// Create a new Aqfer parser with Athena dialect
	aqferParser, err := parser.NewAqferParser(query, athenaAqfer)
	if err != nil {
		log.Fatalf("Failed to create parser: %v", err)
	}

	// Parse the query and extract Aqfer-specific components
	stmt, aqferRef, err := aqferParser.ParseAqferQuery()
	if err != nil {
		log.Fatalf("Failed to parse query: %v", err)
	}

	// Print the full query structure
	fmt.Println("Full Query Structure:")
	fmt.Printf("Number of selected columns: %d\n", len(stmt.SelectList))
	if stmt.Where != nil {
		fmt.Println("Has WHERE clause: yes")
	}
	if len(stmt.GroupBy) > 0 {
		fmt.Printf("Number of GROUP BY columns: %d\n", len(stmt.GroupBy))
	}
	fmt.Println()

	// Print the Aqfer-specific information
	fmt.Printf("Aqfer Table: %s\n", aqferRef.TableName)
	fmt.Printf("Dataset: %s\n", aqferRef.DatasetName)

	fmt.Println("\nMetrics:")
	for _, metric := range aqferRef.Metrics {
		fmt.Printf("- %s\n", metric)
	}

	fmt.Println("\nJoined Tables:")
	for _, joinedTable := range aqferRef.JoinedTables {
		fmt.Printf("- Table: %s\n", joinedTable.TableName)
		fmt.Printf("  Join Type: %v\n", joinedTable.JoinType)
		fmt.Printf("  Is Dimension: %v\n", joinedTable.IsDimension)
		fmt.Println("  Join Keys:")
		for _, key := range joinedTable.JoinKeys {
			fmt.Printf("    %s = %s\n", key.LeftColumn, key.RightColumn)
		}
	}

	// Example showing that BigQuery dialect doesn't support Aqfer
	bigQuery := dialect.NewBigQueryDialect()
	_, err = parser.NewAqferParser(query, bigQuery)
	if err != nil {
		fmt.Printf("\nExpected error for BigQuery: %v\n", err)
	}
}
