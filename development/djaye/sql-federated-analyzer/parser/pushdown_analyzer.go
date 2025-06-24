package parser

import (
	"fmt"
	"strings"

	"sqloptimizer/ast"
)

// PushdownAnalyzer analyzes queries to determine what can be pushed down to Aqfer
type PushdownAnalyzer struct {
	aqferParser *AqferParser
}

// NewPushdownAnalyzer creates a new PushdownAnalyzer
func NewPushdownAnalyzer(parser *AqferParser) *PushdownAnalyzer {
	return &PushdownAnalyzer{
		aqferParser: parser,
	}
}

// AnalyzeQuery analyzes a query to determine what can be pushed down to Aqfer
func (a *PushdownAnalyzer) AnalyzeQuery(query string) (*ast.AqferPushdownInfo, error) {
	// First parse the query normally
	stmt, aqferRef, err := a.aqferParser.ParseAqferQuery()
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %v", err)
	}

	// Extract pushdown information
	info, err := ast.ExtractPushdownInfo(stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to extract pushdown info: %v", err)
	}

	// Set the Aqfer table reference
	info.AqferTable = aqferRef

	// Build the federated subquery
	if err := a.buildFederatedSubquery(info); err != nil {
		return nil, fmt.Errorf("failed to build federated subquery: %v", err)
	}

	return info, nil
}

// buildFederatedSubquery constructs the subquery that will be passed to the federated query provider
func (a *PushdownAnalyzer) buildFederatedSubquery(info *ast.AqferPushdownInfo) error {
	subquery := info.FederatedSubquery

	// Build the WHERE clause combining pushable filters and join conditions
	var conditions []ast.Expression

	// Add pushable filters
	conditions = append(conditions, info.PushableFilters...)

	// Add join conditions for each pushable join
	for _, join := range info.PushableJoins {
		for _, cond := range join.Conditions {
			conditions = append(conditions, &ast.BinaryExpression{
				Left: &ast.ColumnReference{
					Table:  info.AqferTable.TableName,
					Column: cond.AqferColumn,
				},
				Operator: cond.Operator,
				Right: &ast.ColumnReference{
					Table:  join.DimensionTable,
					Column: cond.DimensionColumn,
				},
			})
		}
	}

	// Combine all conditions with AND
	if len(conditions) > 0 {
		var whereCondition ast.Expression = conditions[0]
		for _, cond := range conditions[1:] {
			whereCondition = &ast.BinaryExpression{
				Left:     whereCondition,
				Operator: "AND",
				Right:    cond,
			}
		}
		subquery.Where.Condition = whereCondition
	}

	return nil
}

// GeneratePushdownSQL generates the SQL for the federated subquery
func (a *PushdownAnalyzer) GeneratePushdownSQL(info *ast.AqferPushdownInfo) (string, error) {
	var parts []string

	// SELECT clause
	selectItems := make([]string, len(info.FederatedSubquery.SelectList))
	for i, expr := range info.FederatedSubquery.SelectList {
		if colRef, ok := expr.(*ast.ColumnReference); ok {
			if colRef.Table != "" {
				selectItems[i] = fmt.Sprintf("%s.%s", colRef.Table, colRef.Column)
			} else {
				selectItems[i] = colRef.Column
			}
		}
	}
	parts = append(parts, fmt.Sprintf("SELECT %s", strings.Join(selectItems, ", ")))

	// FROM clause
	parts = append(parts, fmt.Sprintf("FROM %s", info.AqferTable.TableName))

	// WHERE clause
	if info.FederatedSubquery.Where != nil && info.FederatedSubquery.Where.Condition != nil {
		whereSQL, err := a.generateWhereSQL(info.FederatedSubquery.Where.Condition)
		if err != nil {
			return "", err
		}
		parts = append(parts, fmt.Sprintf("WHERE %s", whereSQL))
	}

	return strings.Join(parts, "\n"), nil
}

// generateWhereSQL generates SQL for WHERE conditions
func (a *PushdownAnalyzer) generateWhereSQL(expr ast.Expression) (string, error) {
	switch e := expr.(type) {
	case *ast.BinaryExpression:
		left, err := a.generateWhereSQL(e.Left)
		if err != nil {
			return "", err
		}
		right, err := a.generateWhereSQL(e.Right)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s %s %s)", left, e.Operator, right), nil

	case *ast.ColumnReference:
		if e.Table != "" {
			return fmt.Sprintf("%s.%s", e.Table, e.Column), nil
		}
		return e.Column, nil

	case *ast.Literal:
		switch e.Type {
		case ast.STRING:
			return fmt.Sprintf("'%v'", e.Value), nil
		default:
			return fmt.Sprintf("%v", e.Value), nil
		}
	}

	return "", fmt.Errorf("unsupported expression type: %T", expr)
}
