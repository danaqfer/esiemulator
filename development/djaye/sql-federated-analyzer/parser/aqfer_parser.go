package parser

import (
	"fmt"

	"sqloptimizer/ast"
	"sqloptimizer/dialect"
)

// AqferParser extends the base Parser to handle Aqfer federated queries
type AqferParser struct {
	*Parser
	aqferDialect dialect.AqferDialect
}

// NewAqferParser creates a new AqferParser instance
func NewAqferParser(query string, d dialect.Dialect) (*AqferParser, error) {
	// Check if the dialect supports Aqfer functionality
	aqferDialect, ok := dialect.AsAqferDialect(d)
	if !ok {
		return nil, fmt.Errorf("dialect %s does not support Aqfer functionality", d.Name())
	}

	return &AqferParser{
		Parser:       NewParser(query, d),
		aqferDialect: aqferDialect,
	}, nil
}

// ParseAqferQuery parses a query and extracts Aqfer-specific components
func (p *AqferParser) ParseAqferQuery() (*ast.SelectStatement, *ast.AqferTableReference, error) {
	// Parse the full query first
	stmt, err := p.Parse()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse query: %v", err)
	}

	// Extract Aqfer table reference and its joins
	aqferRef, err := p.extractAqferTable(stmt)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to extract Aqfer table: %v", err)
	}

	return stmt, aqferRef, nil
}

// extractAqferTable finds and extracts the Aqfer table reference from the query
func (p *AqferParser) extractAqferTable(stmt *ast.SelectStatement) (*ast.AqferTableReference, error) {
	if stmt.From == nil || len(stmt.From.Tables) == 0 {
		return nil, fmt.Errorf("no tables found in query")
	}

	// Find the Aqfer table reference
	var aqferTable *ast.TableReference
	for _, table := range stmt.From.Tables {
		if p.aqferDialect.IsAqferTable(table.TableName) {
			aqferTable = &table
			break
		}
	}

	if aqferTable == nil {
		return nil, fmt.Errorf("no Aqfer table found in query")
	}

	// Validate the Aqfer table name
	if !p.aqferDialect.ValidateAqferTableName(aqferTable.TableName) {
		return nil, fmt.Errorf("invalid Aqfer table name: %s", aqferTable.TableName)
	}

	// Create AqferTableReference
	aqferRef := &ast.AqferTableReference{
		TableName:   p.aqferDialect.GetAqferTableName(aqferTable.TableName),
		DatasetName: p.aqferDialect.GetAqferDatasetType(aqferTable.TableName),
	}

	// Extract dimensions and metrics from the SELECT list
	p.extractColumnsFromSelect(stmt, aqferRef)

	// Extract joined tables and their conditions
	if err := p.extractJoinedTables(stmt, aqferRef); err != nil {
		return nil, err
	}

	return aqferRef, nil
}

// extractColumnsFromSelect analyzes the SELECT list to identify dimensions and metrics
func (p *AqferParser) extractColumnsFromSelect(stmt *ast.SelectStatement, aqferRef *ast.AqferTableReference) {
	for _, expr := range stmt.SelectList {
		if colRef, ok := expr.(*ast.ColumnReference); ok {
			// For now, we'll consider any column from the Aqfer table as a metric
			// This can be enhanced with actual metadata about dimensions vs metrics
			if colRef.Table == aqferRef.TableName {
				aqferRef.Metrics = append(aqferRef.Metrics, colRef.Column)
			}
		}
	}
}

// extractJoinedTables analyzes the FROM clause to find tables joined with the Aqfer table
func (p *AqferParser) extractJoinedTables(stmt *ast.SelectStatement, aqferRef *ast.AqferTableReference) error {
	for _, table := range stmt.From.Tables {
		if table.JoinType != ast.NONE && table.JoinCond != nil {
			joinedTable := &ast.JoinedTable{
				TableName: table.TableName,
				JoinType:  table.JoinType,
			}

			// Extract join keys from the join condition
			if err := p.extractJoinKeys(table.JoinCond, joinedTable); err != nil {
				return err
			}

			// Check if this is a dimension table by looking at the join keys
			joinedTable.IsDimension = p.isDimensionTable(joinedTable, aqferRef)

			aqferRef.JoinedTables = append(aqferRef.JoinedTables, joinedTable)
		}
	}
	return nil
}

// extractJoinKeys extracts the join conditions between tables
func (p *AqferParser) extractJoinKeys(expr ast.Expression, joinedTable *ast.JoinedTable) error {
	switch e := expr.(type) {
	case *ast.BinaryExpression:
		if e.Operator == "=" {
			if left, ok := e.Left.(*ast.ColumnReference); ok {
				if right, ok := e.Right.(*ast.ColumnReference); ok {
					joinedTable.JoinKeys = append(joinedTable.JoinKeys, ast.JoinKey{
						LeftColumn:  left.Column,
						RightColumn: right.Column,
					})
				}
			}
		}
	}
	return nil
}

// isDimensionTable checks if a joined table represents an Aqfer dimension
func (p *AqferParser) isDimensionTable(joinedTable *ast.JoinedTable, aqferRef *ast.AqferTableReference) bool {
	// A table is considered a dimension if it's joined on a column that's
	// marked as a dimension in the Aqfer table metadata
	for _, key := range joinedTable.JoinKeys {
		for _, dim := range aqferRef.Dimensions {
			if key.LeftColumn == dim || key.RightColumn == dim {
				return true
			}
		}
	}
	return false
}
