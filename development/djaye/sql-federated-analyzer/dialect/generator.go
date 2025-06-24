package dialect

import (
	"fmt"
	"sqlparser/ast"
	"strings"
)

// Generator is responsible for converting AST nodes back to SQL text
type Generator interface {
	// Generate SQL from an AST node
	GenerateSQL(node ast.Node) (string, error)
}

// BaseGenerator provides common SQL generation logic
type BaseGenerator struct {
	dialect Dialect
}

func NewBaseGenerator(d Dialect) *BaseGenerator {
	return &BaseGenerator{dialect: d}
}

func (g *BaseGenerator) GenerateSQL(node ast.Node) (string, error) {
	switch n := node.(type) {
	case *ast.SelectStatement:
		return g.generateSelect(n)
	case *ast.FromClause:
		return g.generateFrom(n)
	case *ast.WhereClause:
		return g.generateWhere(n)
	case *ast.BinaryExpression:
		return g.generateBinaryExpr(n)
	case *ast.ColumnReference:
		return g.generateColumnRef(n)
	case *ast.Literal:
		return g.generateLiteral(n)
	case *ast.InExpression:
		return g.generateInExpr(n)
	case *ast.FunctionCall:
		return g.generateFunctionCall(n)
	case *ast.Star:
		return "*", nil
	default:
		return "", fmt.Errorf("unsupported node type: %T", node)
	}
}

func (g *BaseGenerator) generateSelect(stmt *ast.SelectStatement) (string, error) {
	var parts []string

	// SELECT
	selectList := make([]string, len(stmt.SelectList))
	for i, expr := range stmt.SelectList {
		sql, err := g.GenerateSQL(expr)
		if err != nil {
			return "", err
		}
		selectList[i] = sql
	}
	parts = append(parts, fmt.Sprintf("SELECT %s", strings.Join(selectList, ", ")))

	// FROM
	if stmt.From != nil {
		fromSQL, err := g.GenerateSQL(stmt.From)
		if err != nil {
			return "", err
		}
		parts = append(parts, fmt.Sprintf("FROM %s", fromSQL))
	}

	// WHERE
	if stmt.Where != nil {
		whereSQL, err := g.GenerateSQL(stmt.Where)
		if err != nil {
			return "", err
		}
		parts = append(parts, whereSQL)
	}

	// GROUP BY
	if len(stmt.GroupBy) > 0 {
		groupByExprs := make([]string, len(stmt.GroupBy))
		for i, expr := range stmt.GroupBy {
			sql, err := g.GenerateSQL(expr)
			if err != nil {
				return "", err
			}
			groupByExprs[i] = sql
		}
		parts = append(parts, fmt.Sprintf("GROUP BY %s", strings.Join(groupByExprs, ", ")))
	}

	// HAVING
	if stmt.Having != nil {
		havingSQL, err := g.GenerateSQL(stmt.Having)
		if err != nil {
			return "", err
		}
		parts = append(parts, fmt.Sprintf("HAVING %s", havingSQL))
	}

	// ORDER BY
	if len(stmt.OrderBy) > 0 {
		orderByExprs := make([]string, len(stmt.OrderBy))
		for i, expr := range stmt.OrderBy {
			sql, err := g.GenerateSQL(expr.Expr)
			if err != nil {
				return "", err
			}
			direction := "ASC"
			if !expr.Ascending {
				direction = "DESC"
			}
			orderByExprs[i] = fmt.Sprintf("%s %s", sql, direction)
		}
		parts = append(parts, fmt.Sprintf("ORDER BY %s", strings.Join(orderByExprs, ", ")))
	}

	// LIMIT
	if stmt.Limit != nil {
		if g.dialect.Name() == "Teradata" {
			// Move TOP before SELECT for Teradata
			parts[0] = fmt.Sprintf("SELECT TOP %d", stmt.Limit.Count) + parts[0][6:]
		} else {
			limitSQL := fmt.Sprintf("LIMIT %d", stmt.Limit.Count)
			if stmt.Limit.Offset > 0 {
				limitSQL += fmt.Sprintf(" OFFSET %d", stmt.Limit.Offset)
			}
			parts = append(parts, limitSQL)
		}
	}

	return strings.Join(parts, " "), nil
}

func (g *BaseGenerator) generateFrom(from *ast.FromClause) (string, error) {
	tables := make([]string, len(from.Tables))
	for i, table := range from.Tables {
		name := table.TableName
		if table.Alias != "" {
			name = fmt.Sprintf("%s AS %s", name, table.Alias)
		}
		if table.JoinType != ast.INNER {
			joinType := ""
			switch table.JoinType {
			case ast.LEFT:
				joinType = "LEFT"
			case ast.RIGHT:
				joinType = "RIGHT"
			case ast.FULL:
				joinType = "FULL"
			case ast.CROSS:
				joinType = "CROSS"
			}
			name = fmt.Sprintf("%s JOIN %s", joinType, name)
			if table.JoinCond != nil {
				condSQL, err := g.GenerateSQL(table.JoinCond)
				if err != nil {
					return "", err
				}
				name = fmt.Sprintf("%s ON %s", name, condSQL)
			}
		}
		tables[i] = name
	}
	return strings.Join(tables, ", "), nil
}

func (g *BaseGenerator) generateWhere(where *ast.WhereClause) (string, error) {
	if where.Condition == nil {
		return "", nil
	}
	condSQL, err := g.GenerateSQL(where.Condition)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("WHERE %s", condSQL), nil
}

func (g *BaseGenerator) generateBinaryExpr(expr *ast.BinaryExpression) (string, error) {
	left, err := g.GenerateSQL(expr.Left)
	if err != nil {
		return "", err
	}
	right, err := g.GenerateSQL(expr.Right)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %s %s", left, expr.Operator, right), nil
}

func (g *BaseGenerator) generateColumnRef(ref *ast.ColumnReference) (string, error) {
	if ref.Table != "" {
		return fmt.Sprintf("%s%s%s", ref.Table, g.dialect.GetCatalogSeparator(), ref.Column), nil
	}
	return ref.Column, nil
}

func (g *BaseGenerator) generateLiteral(lit *ast.Literal) (string, error) {
	switch lit.Type {
	case ast.STRING:
		return fmt.Sprintf("%s%v%s", g.dialect.GetStringLiteralQuote(), lit.Value, g.dialect.GetStringLiteralQuote()), nil
	default:
		return fmt.Sprintf("%v", lit.Value), nil
	}
}

func (g *BaseGenerator) generateInExpr(expr *ast.InExpression) (string, error) {
	col, err := g.GenerateSQL(expr.Column)
	if err != nil {
		return "", err
	}

	values := make([]string, len(expr.Values))
	for i, val := range expr.Values {
		valSQL, err := g.GenerateSQL(val)
		if err != nil {
			return "", err
		}
		values[i] = valSQL
	}

	return fmt.Sprintf("%s IN (%s)", col, strings.Join(values, ", ")), nil
}

// generateFunctionCall generates SQL for a function call
func (g *BaseGenerator) generateFunctionCall(n *ast.FunctionCall) (string, error) {
	args := make([]string, len(n.Args))
	var err error
	for i, arg := range n.Args {
		args[i], err = g.GenerateSQL(arg)
		if err != nil {
			return "", err
		}
	}

	distinct := ""
	if n.Distinct {
		distinct = "DISTINCT "
	}

	return fmt.Sprintf("%s(%s%s)", n.Name, distinct, strings.Join(args, ", ")), nil
}
