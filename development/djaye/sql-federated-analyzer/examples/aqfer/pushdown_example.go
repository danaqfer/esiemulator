package aqfer_examples

import (
	"fmt"
	"log"

	"sqloptimizer/ast"
	"sqloptimizer/dialect"
	"sqloptimizer/parser"
)

func ExamplePushdownAnalyzer() {
	// Example query with multiple joins and conditions
	query := `
		SELECT 
			a.campaign_id,
			c.campaign_name,
			p.platform_name,
			a.impressions,
			a.clicks,
			a.conversions,
			a.spend
		FROM my_aqfer_table.dsae a
		INNER JOIN campaigns c ON a.campaign_id = c.id
		INNER JOIN platforms p ON a.platform_id = p.id
		WHERE a.date_key >= '2024-01-01'
		  AND a.date_key < '2024-02-01'
		  AND c.status = 'active'
		  AND p.region = 'NA'
		GROUP BY a.campaign_id, c.campaign_name, p.platform_name
	`

	// Create an Athena dialect with Aqfer support
	athenaAqfer := dialect.NewAthenaAqferDialect()

	// Create a new Aqfer parser with Athena dialect
	aqferParser, err := parser.NewAqferParser(query, athenaAqfer)
	if err != nil {
		log.Fatalf("Failed to create parser: %v", err)
	}

	// Create a pushdown analyzer
	analyzer := parser.NewPushdownAnalyzer(aqferParser)

	// Analyze the query for pushdown opportunities
	info, err := analyzer.AnalyzeQuery(query)
	if err != nil {
		log.Fatalf("Failed to analyze query: %v", err)
	}

	// Print the analysis results
	fmt.Println("Pushdown Analysis Results:")
	fmt.Printf("\nAqfer Table: %s\n", info.AqferTable.TableName)
	fmt.Printf("Dataset Type: %s\n", info.AqferTable.DatasetName)

	fmt.Println("\nPushable Joins:")
	for _, join := range info.PushableJoins {
		fmt.Printf("\n- Table: %s\n", join.DimensionTable)
		fmt.Printf("  Required: %v\n", join.IsRequired)
		fmt.Printf("  Required Columns: %v\n", join.RequiredColumns)
		fmt.Println("  Join Conditions:")
		for _, cond := range join.Conditions {
			fmt.Printf("    %s.%s %s %s.%s\n",
				info.AqferTable.TableName, cond.AqferColumn,
				cond.Operator,
				join.DimensionTable, cond.DimensionColumn)
		}
	}

	fmt.Println("\nPushable Filters:")
	for _, filter := range info.PushableFilters {
		if binExpr, ok := filter.(*ast.BinaryExpression); ok {
			fmt.Printf("- %v %s %v\n",
				binExpr.Left, binExpr.Operator, binExpr.Right)
		}
	}

	// Generate and print the federated subquery
	subquery, err := analyzer.GeneratePushdownSQL(info)
	if err != nil {
		log.Fatalf("Failed to generate pushdown SQL: %v", err)
	}
	fmt.Printf("\nFederated Subquery:\n%s\n", subquery)
}
