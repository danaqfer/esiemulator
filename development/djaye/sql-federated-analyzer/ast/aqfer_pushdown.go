package ast

// AqferPushdownInfo represents information about what can be pushed down to Aqfer
type AqferPushdownInfo struct {
	// The original Aqfer table reference
	AqferTable *AqferTableReference

	// Join predicates that can be pushed down
	PushableJoins []PushableJoin

	// Filters that can be pushed down
	PushableFilters []Expression

	// The subquery that would be passed to the federated query provider
	FederatedSubquery *SelectStatement
}

// PushableJoin represents a join that can be pushed down to Aqfer
type PushableJoin struct {
	// The dimension table being joined
	DimensionTable string

	// The join conditions that can be pushed down
	Conditions []JoinCondition

	// The columns from the dimension table that are used in the query
	RequiredColumns []string

	// Whether this join is required for the final result
	IsRequired bool
}

// JoinCondition represents a single join condition that can be pushed down
type JoinCondition struct {
	// The column from the Aqfer table
	AqferColumn string

	// The column from the dimension table
	DimensionColumn string

	// The operator (usually "=")
	Operator string
}

// ExtractPushdownInfo analyzes a SelectStatement to determine what can be pushed down
func ExtractPushdownInfo(stmt *SelectStatement) (*AqferPushdownInfo, error) {
	info := &AqferPushdownInfo{
		FederatedSubquery: &SelectStatement{
			SelectList: make([]Expression, 0),
			Where:      &WhereClause{},
		},
	}

	// Copy the original select list and add any columns needed for joins
	requiredColumns := make(map[string]bool)
	for _, expr := range stmt.SelectList {
		if colRef, ok := expr.(*ColumnReference); ok {
			requiredColumns[colRef.Column] = true
		}
	}

	// Analyze each join to determine if it can be pushed down
	if stmt.From != nil {
		for _, table := range stmt.From.Tables {
			if join, ok := analyzePushableJoin(&table, requiredColumns); ok {
				info.PushableJoins = append(info.PushableJoins, join)

				// Add join columns to required columns
				for _, cond := range join.Conditions {
					requiredColumns[cond.AqferColumn] = true
				}
			}
		}
	}

	// Add all required columns to the federated subquery
	for col := range requiredColumns {
		info.FederatedSubquery.SelectList = append(
			info.FederatedSubquery.SelectList,
			&ColumnReference{Column: col},
		)
	}

	// Analyze WHERE clause for pushable filters
	if stmt.Where != nil {
		info.PushableFilters = analyzePushableFilters(stmt.Where.Condition)
	}

	return info, nil
}

// analyzePushableJoin determines if a join can be pushed down to Aqfer
func analyzePushableJoin(table *TableReference, requiredColumns map[string]bool) (PushableJoin, bool) {
	join := PushableJoin{
		DimensionTable: table.TableName,
	}

	// Only consider joins with conditions
	if table.JoinCond == nil {
		return join, false
	}

	// Extract join conditions
	conditions := extractJoinConditions(table.JoinCond)
	if len(conditions) == 0 {
		return join, false
	}

	// Check if all join columns are available
	for _, cond := range conditions {
		if !requiredColumns[cond.DimensionColumn] {
			join.RequiredColumns = append(join.RequiredColumns, cond.DimensionColumn)
		}
	}

	join.Conditions = conditions
	join.IsRequired = isJoinRequired(table, requiredColumns)

	return join, true
}

// extractJoinConditions extracts join conditions from an expression
func extractJoinConditions(expr Expression) []JoinCondition {
	var conditions []JoinCondition

	switch e := expr.(type) {
	case *BinaryExpression:
		if e.Operator == "=" {
			if left, ok := e.Left.(*ColumnReference); ok {
				if right, ok := e.Right.(*ColumnReference); ok {
					conditions = append(conditions, JoinCondition{
						AqferColumn:     left.Column,
						DimensionColumn: right.Column,
						Operator:        "=",
					})
				}
			}
		} else if e.Operator == "AND" {
			conditions = append(conditions,
				extractJoinConditions(e.Left)...,
			)
			conditions = append(conditions,
				extractJoinConditions(e.Right)...,
			)
		}
	}

	return conditions
}

// analyzePushableFilters identifies filters that can be pushed down
func analyzePushableFilters(expr Expression) []Expression {
	var filters []Expression

	switch e := expr.(type) {
	case *BinaryExpression:
		if e.Operator == "AND" {
			filters = append(filters, analyzePushableFilters(e.Left)...)
			filters = append(filters, analyzePushableFilters(e.Right)...)
		} else if isPushableOperator(e.Operator) {
			if isAqferColumn(e.Left) || isAqferColumn(e.Right) {
				filters = append(filters, e)
			}
		}
	}

	return filters
}

// isJoinRequired checks if a join is required based on the columns used
func isJoinRequired(table *TableReference, requiredColumns map[string]bool) bool {
	// A join is required if any of its columns are used in the final result
	if table.JoinCond != nil {
		switch e := table.JoinCond.(type) {
		case *BinaryExpression:
			if right, ok := e.Right.(*ColumnReference); ok {
				if requiredColumns[right.Column] {
					return true
				}
			}
		}
	}
	return false
}

// Helper functions
func isPushableOperator(op string) bool {
	switch op {
	case "=", "<", ">", "<=", ">=", "<>", "IN":
		return true
	default:
		return false
	}
}

func isAqferColumn(expr Expression) bool {
	if col, ok := expr.(*ColumnReference); ok {
		// This should be enhanced to properly check if the column is from the Aqfer table
		return col.Table != ""
	}
	return false
}
