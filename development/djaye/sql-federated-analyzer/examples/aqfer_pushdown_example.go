package examples

import (
	"fmt"
	"log"

	"sqloptimizer/dialect"
	"sqloptimizer/parser"
)

func ExampleFederatedQueryPushdown() {
	// Create an example query that joins a federated table with dimension tables
	// Note: The .dsae suffix indicates this table is served by the Aqfer federated query provider
	query := `
		SELECT 
			imp.impression_id,
			imp.timestamp,
			c.name as campaign_name,
			p.region as platform_region
		FROM 
			impressions.dsae imp  -- This is a federated table served by Aqfer
			JOIN campaigns c ON imp.campaign_id = c.id  -- This is a dimension table in the original query engine
			JOIN platforms p ON imp.platform_id = p.id  -- This is another dimension table
		WHERE 
			c.status = 'active'
			AND p.region = 'NA'
	`

	// Create the Athena dialect for the Aqfer federated query provider
	dialect := dialect.NewAthenaAqferDialect()

	// Create the analyzer
	analyzer := parser.NewFederatedQueryAnalyzer(dialect)

	// Perform the three-phase analysis
	analysis, err := analyzer.AnalyzeQuery(query)
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}

	// Print the analysis results
	fmt.Println("Phase 1: Original AST")
	fmt.Printf("Tables found: %d\n", len(analysis.OriginalAST.From.Tables))

	fmt.Println("\nPhase 2: Federated Table Analysis")
	fmt.Printf("Federated table: %s\n", analysis.FederatedTable.TableName)
	fmt.Printf("Dataset type: %s\n", analysis.FederatedTable.DatasetType)

	fmt.Println("\nPhase 3: Dimension Analysis")
	fmt.Println("Dimension Filters to Execute in Original Context:")
	for _, filter := range analysis.DimensionFilters {
		fmt.Printf("  Table: %s\n", filter.TableName)
		fmt.Printf("  SQL to fetch values: %s\n", filter.ValueFetchSQL)
	}

	fmt.Println("\nFinal Federated Subquery:")
	fmt.Printf("SELECT list length: %d\n", len(analysis.FederatedSubquery.SelectList))
}

/*
Expected output:

Phase 1: Original AST
Tables found: 3

Phase 2: Federated Table Analysis
Federated table: impressions
Dataset type: impressions

Phase 3: Dimension Analysis
Dimension Filters to Execute in Original Context:
  Table: campaigns
  SQL to fetch values: SELECT DISTINCT campaign_id FROM campaigns WHERE status = 'active'
  Table: platforms
  SQL to fetch values: SELECT DISTINCT platform_id FROM platforms WHERE region = 'NA'

Final Federated Subquery:
SELECT list length: 4
*/
