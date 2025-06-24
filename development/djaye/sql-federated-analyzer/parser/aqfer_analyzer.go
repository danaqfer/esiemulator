package parser

import (
	"fmt"

	"sqloptimizer/ast"
	"sqloptimizer/dialect"
)

// FederatedQueryAnalyzer handles the analysis of queries containing Aqfer federated provider tables
type FederatedQueryAnalyzer struct {
	dialect dialect.AqferDialect
}

// NewFederatedQueryAnalyzer creates a new analyzer for queries with federated tables
func NewFederatedQueryAnalyzer(d dialect.AqferDialect) *FederatedQueryAnalyzer {
	return &FederatedQueryAnalyzer{
		dialect: d,
	}
}

// AnalyzeQuery performs the three-phase analysis of a query
func (a *FederatedQueryAnalyzer) AnalyzeQuery(query string) (*ast.FederatedQueryAnalysis, error) {
	// Phase 1: Parse the query into an AST
	parser := NewParser(query, a.dialect)
	stmt, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("phase 1 - parsing failed: %v", err)
	}

	analysis := &ast.FederatedQueryAnalysis{
		OriginalAST: stmt,
	}

	// Phase 2: Identify and analyze federated tables
	fedTable, err := a.identifyFederatedTable(stmt)
	if err != nil {
		return nil, fmt.Errorf("phase 2 - federated table analysis failed: %v", err)
	}
	analysis.FederatedTable = fedTable

	// Phase 3: Analyze joins and prepare dimension filters
	if err := a.analyzeDimensions(stmt, analysis); err != nil {
		return nil, fmt.Errorf("phase 3 - dimension analysis failed: %v", err)
	}

	return analysis, nil
}

// Phase 2: Identify and analyze federated tables
func (a *FederatedQueryAnalyzer) identifyFederatedTable(stmt *ast.SelectStatement) (*ast.FederatedTableReference, error) {
	if stmt.From == nil {
		return nil, fmt.Errorf("no FROM clause found")
	}

	var fedTable *ast.TableReference
	var position int
	for i, table := range stmt.From.Tables {
		if a.dialect.IsFederatedTable(table.TableName) {
			fedTable = &table
			position = i
			break
		}
	}

	if fedTable == nil {
		return nil, fmt.Errorf("no federated table found in query")
	}

	return &ast.FederatedTableReference{
		TableName:   a.dialect.GetFederatedTableName(fedTable.TableName),
		DatasetType: a.dialect.GetFederatedDatasetType(fedTable.TableName),
		Alias:       fedTable.Alias,
		Position:    position,
	}, nil
}

// Phase 3: Analyze dimensions and prepare filters
func (a *FederatedQueryAnalyzer) analyzeDimensions(stmt *ast.SelectStatement, analysis *ast.FederatedQueryAnalysis) error {
	// First, identify all joins to the federated table
	for _, table := range stmt.From.Tables {
		if table.JoinType == ast.NONE {
			continue
		}

		dimension, filter, err := a.analyzeDimensionJoin(&table, analysis.FederatedTable)
		if err != nil {
			return err
		}

		if dimension != nil {
			analysis.FederatedTable.Dimensions = append(analysis.FederatedTable.Dimensions, *dimension)
		}
		if filter != nil {
			analysis.DimensionFilters = append(analysis.DimensionFilters, *filter)
		}
	}

	// Then, look for additional filters in the WHERE clause
	if stmt.Where != nil {
		if err := a.analyzeWhereClause(stmt.Where, analysis); err != nil {
			return err
		}
	}

	// Finally, prepare the federated subquery
	return a.prepareFederatedSubquery(stmt, analysis)
}

func (a *FederatedQueryAnalyzer) analyzeDimensionJoin(table *ast.TableReference, fedTable *ast.FederatedTableReference) (*ast.DimensionJoin, *ast.DimensionFilter, error) {
	if table.JoinCond == nil {
		return nil, nil, nil
	}

	// Extract join conditions
	binExpr, ok := table.JoinCond.(*ast.BinaryExpression)
	if !ok || binExpr.Operator != "=" {
		return nil, nil, nil // Not a simple equality join
	}

	left, lok := binExpr.Left.(*ast.ColumnReference)
	right, rok := binExpr.Right.(*ast.ColumnReference)
	if !lok || !rok {
		return nil, nil, nil // Not a column-to-column join
	}

	// Create dimension and filter info
	dimension := &ast.DimensionJoin{
		TableName:        table.TableName,
		RequiresCallback: true, // We'll assume all dimensions need callbacks for now
	}

	// Figure out which side is the federated column
	if left.Table == fedTable.TableName || left.Table == fedTable.Alias {
		dimension.JoinColumn = left.Column
		dimension.TableColumn = right.Column
	} else {
		dimension.JoinColumn = right.Column
		dimension.TableColumn = left.Column
	}

	filter := &ast.DimensionFilter{
		TableName:       table.TableName,
		FederatedColumn: dimension.JoinColumn,
	}

	return dimension, filter, nil
}

func (a *FederatedQueryAnalyzer) analyzeWhereClause(where *ast.WhereClause, analysis *ast.FederatedQueryAnalysis) error {
	// Extract conditions that apply to dimension tables
	conditions := extractConditions(where.Condition)

	// Group conditions by table
	tableConditions := make(map[string][]ast.DimensionCondition)

	for _, cond := range conditions {
		if binExpr, ok := cond.(*ast.BinaryExpression); ok {
			if col, ok := binExpr.Left.(*ast.ColumnReference); ok {
				if col.Table != "" && col.Table != analysis.FederatedTable.TableName {
					tableCond := ast.DimensionCondition{
						Column:   col.Column,
						Operator: binExpr.Operator,
						Value:    binExpr.Right,
					}
					tableConditions[col.Table] = append(tableConditions[col.Table], tableCond)
				}
			}
		}
	}

	// Update dimension filters with conditions
	for tableName, conditions := range tableConditions {
		for i, filter := range analysis.DimensionFilters {
			if filter.TableName == tableName {
				analysis.DimensionFilters[i].Conditions = append(filter.Conditions, conditions...)
				// Generate the SQL to fetch dimension values
				analysis.DimensionFilters[i].ValueFetchSQL = filter.GenerateValueFetchSQL()
			}
		}
	}

	return nil
}

func (a *FederatedQueryAnalyzer) prepareFederatedSubquery(stmt *ast.SelectStatement, analysis *ast.FederatedQueryAnalysis) error {
	// Create a new SELECT statement for the federated query
	federated := &ast.SelectStatement{
		SelectList: make([]ast.Expression, 0),
		From: &ast.FromClause{
			Tables: []ast.TableReference{
				{
					TableName: analysis.FederatedTable.TableName,
					Alias:     analysis.FederatedTable.Alias,
				},
			},
		},
	}

	// Add required columns to the SELECT list
	for _, expr := range stmt.SelectList {
		if col, ok := expr.(*ast.ColumnReference); ok {
			if col.Table == analysis.FederatedTable.TableName || col.Table == analysis.FederatedTable.Alias {
				federated.SelectList = append(federated.SelectList, col)
			}
		}
	}

	// Add join columns
	for _, dim := range analysis.FederatedTable.Dimensions {
		federated.SelectList = append(federated.SelectList, &ast.ColumnReference{
			Table:  analysis.FederatedTable.Alias,
			Column: dim.JoinColumn,
		})
	}

	analysis.FederatedSubquery = federated
	return nil
}

// Helper function to extract conditions from a WHERE clause
func extractConditions(expr ast.Expression) []ast.Expression {
	var conditions []ast.Expression

	switch e := expr.(type) {
	case *ast.BinaryExpression:
		if e.Operator == "AND" {
			conditions = append(conditions, extractConditions(e.Left)...)
			conditions = append(conditions, extractConditions(e.Right)...)
		} else {
			conditions = append(conditions, e)
		}
	}

	return conditions
}
