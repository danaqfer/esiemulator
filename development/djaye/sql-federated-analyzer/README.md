# Aqfer SQL Query Analyzer

A SQL query analyzer specifically designed for optimizing queries that use the Aqfer federated query provider in AWS Athena. This tool analyzes SQL queries and optimizes them by pushing down appropriate filters to the Aqfer backend.

## Features

- Dialect-aware SQL parsing supporting Athena, BigQuery, Teradata, and PostgreSQL
- Automatic detection of Aqfer federated query provider tables
- Analysis of join conditions between federated and dimension tables
- Pushdown optimization for dimension table filters
- Generation of optimized federated subqueries

## How It Works

The analyzer follows a three-phase process:

1. **SQL Parsing**: Parses the original SQL statement into a dialect-aware AST
2. **Federated Table Detection**: Identifies tables served by the Aqfer federated query provider
3. **Join Analysis**: Determines which filters can be pushed down to the Aqfer backend

### Example

```sql
SELECT 
    imp.impression_id,
    imp.timestamp,
    c.name as campaign_name,
    p.region as platform_region
FROM 
    impressions.dsae imp  -- Federated table served by Aqfer
    JOIN campaigns c ON imp.campaign_id = c.id
    JOIN platforms p ON imp.platform_id = p.id
WHERE 
    c.status = 'active'
    AND p.region = 'NA'
```

The analyzer will:
1. Execute dimension table queries in the original context
2. Pass the results to the Aqfer federated query provider
3. Generate an optimized federated subquery

## Project Structure

```
.
├── ast/                    # Abstract Syntax Tree definitions
│   ├── ast.go             # Core AST structures
│   ├── aqfer_query.go     # Federated query analysis structures
│   └── aqfer_pushdown.go  # Pushdown optimization structures
├── dialect/               # SQL dialect implementations
│   └── aqfer_dialect.go  # Aqfer-specific dialect support
├── parser/               # SQL parsing and analysis
│   ├── parser.go         # Core SQL parser
│   ├── aqfer_parser.go   # Federated query parsing
│   └── aqfer_analyzer.go # Query analysis and optimization
└── examples/             # Example implementations
    └── aqfer_pushdown_example.go
```

## Usage

```go
// Create the Athena dialect for the Aqfer federated query provider
dialect := dialect.NewAthenaAqferDialect()

// Create the analyzer
analyzer := parser.NewFederatedQueryAnalyzer(dialect)

// Analyze a query
analysis, err := analyzer.AnalyzeQuery(query)
if err != nil {
    log.Fatalf("Analysis failed: %v", err)
}

// Access the analysis results
fmt.Printf("Federated table: %s\n", analysis.FederatedTable.TableName)
for _, filter := range analysis.DimensionFilters {
    fmt.Printf("Filter SQL: %s\n", filter.ValueFetchSQL)
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 