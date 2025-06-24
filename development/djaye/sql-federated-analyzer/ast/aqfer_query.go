package ast

import (
	"fmt"
	"strings"
)

// FederatedQueryAnalysis represents the complete analysis of a query containing Aqfer federated provider tables
type FederatedQueryAnalysis struct {
	// The original parsed AST
	OriginalAST *SelectStatement

	// Information about the federated table reference
	FederatedTable *FederatedTableReference

	// Analysis of dimension filters that can be pushed down
	DimensionFilters []DimensionFilter

	// The final subquery structure that will be sent to the federated provider
	FederatedSubquery *SelectStatement
}

// DimensionFilter represents a filter on a dimension table that will be used
// to fetch values from the original query context
type DimensionFilter struct {
	// The dimension table being filtered
	TableName string

	// The column in the federated table that references this dimension
	FederatedColumn string

	// The join and filter conditions that will be used to fetch values
	Conditions []DimensionCondition

	// The SQL to execute in the original context to get dimension values
	ValueFetchSQL string
}

// DimensionCondition represents a condition on a dimension table
type DimensionCondition struct {
	// The column being filtered
	Column string

	// The operator (e.g., =, IN, etc.)
	Operator string

	// The value or expression being compared against
	Value Expression
}

// Phase 1: Initial AST population
func (s *SelectStatement) ContainsAqferTable() bool {
	if s.From == nil {
		return false
	}

	for _, table := range s.From.Tables {
		// This will be replaced by dialect-specific logic
		if strings.HasSuffix(table.TableName, ".dsae") {
			return true
		}
	}
	return false
}

// Phase 2: Aqfer Table Recognition
type FederatedTableReference struct {
	// Basic table information
	TableName   string // The name of the table in the federated provider
	DatasetType string // The type of dataset this table represents
	Alias       string // SQL alias for the table

	// Join information with dimension tables
	Dimensions []DimensionJoin

	// Original position in the FROM clause
	Position int
}

// DimensionJoin represents a join between a federated table and a dimension table
type DimensionJoin struct {
	// The dimension table information
	TableName    string
	JoinColumn   string // Column in the federated table
	TableColumn  string // Column in the dimension table
	FilterColumn string // Column used in WHERE clause

	// The filter condition on this dimension
	FilterOperator string
	FilterValue    Expression

	// Whether this dimension's values should be fetched from the original context
	RequiresCallback bool
}

// Phase 3: Join Analysis
type JoinAnalysis struct {
	// Maps dimension table names to their filter conditions
	DimensionFilters map[string]*DimensionFilter

	// The columns that need to be included in the federated subquery
	RequiredColumns []string

	// The final WHERE clause for the federated query
	FederatedWhere Expression
}

// Helper method to identify pushable join conditions
func (j *JoinAnalysis) IsPushableJoin(join *JoinClause) bool {
	if join == nil || join.Conditions == nil {
		return false
	}

	// A join is pushable if:
	// 1. It's to a dimension table
	// 2. The join condition uses the dimension's key
	// 3. Any filters on the dimension can be evaluated in the original context
	for _, cond := range join.Conditions {
		if !j.isPushableJoinCondition(cond) {
			return false
		}
	}
	return true
}

func (j *JoinAnalysis) isPushableJoinCondition(cond Expression) bool {
	// For now, we only support simple equality joins
	binExpr, ok := cond.(*BinaryExpression)
	if !ok || binExpr.Operator != "=" {
		return false
	}

	// One side must be from the Aqfer table, the other from a dimension
	left, lok := binExpr.Left.(*ColumnReference)
	right, rok := binExpr.Right.(*ColumnReference)
	if !lok || !rok {
		return false
	}

	// At least one side must be from the Aqfer table
	return left.Table != "" && right.Table != ""
}

// Helper method to generate the SQL to fetch dimension values
func (d *DimensionFilter) GenerateValueFetchSQL() string {
	var conditions []string
	for _, cond := range d.Conditions {
		conditions = append(conditions, fmt.Sprintf("%s %s %v",
			cond.Column, cond.Operator, cond.Value))
	}

	return fmt.Sprintf(
		"SELECT DISTINCT %s FROM %s WHERE %s",
		d.FederatedColumn,
		d.TableName,
		strings.Join(conditions, " AND "),
	)
}
