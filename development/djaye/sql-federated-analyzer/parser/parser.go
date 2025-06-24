package parser

import (
	"fmt"
	"strings"

	"sqlparser/ast"
	"sqlparser/dialect"
)

// Parser represents a SQL parser
type Parser struct {
	query   string
	pos     int
	tokens  []string
	dialect dialect.Dialect
}

// NewParser creates a new Parser instance
func NewParser(query string, d dialect.Dialect) *Parser {
	return &Parser{
		query:   query,
		pos:     0,
		tokens:  tokenize(query),
		dialect: d,
	}
}

// Parse parses a SQL query and returns an AST
func (p *Parser) Parse() (*ast.SelectStatement, error) {
	// Skip any leading whitespace
	p.skipWhitespace()

	// Expect SELECT keyword
	if !p.expectKeyword("SELECT") {
		return nil, fmt.Errorf("expected SELECT keyword at position %d", p.pos)
	}

	stmt := &ast.SelectStatement{}

	// Parse SELECT list
	selectList, err := p.parseSelectList()
	if err != nil {
		return nil, err
	}
	stmt.SelectList = selectList

	// Parse FROM clause
	if p.expectKeyword("FROM") {
		fromClause, err := p.parseFromClause()
		if err != nil {
			return nil, err
		}
		stmt.From = fromClause
	}

	// Parse WHERE clause
	if p.expectKeyword("WHERE") {
		whereClause, err := p.parseWhereClause()
		if err != nil {
			return nil, err
		}
		stmt.Where = whereClause
	}

	// Parse GROUP BY clause
	if p.expectKeyword("GROUP") && p.expectKeyword("BY") {
		groupBy, err := p.parseExpressionList()
		if err != nil {
			return nil, err
		}
		stmt.GroupBy = groupBy
	}

	// Parse HAVING clause
	if p.expectKeyword("HAVING") {
		having, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		stmt.Having = having
	}

	// Parse ORDER BY clause
	if p.expectKeyword("ORDER") && p.expectKeyword("BY") {
		orderBy, err := p.parseOrderByClause()
		if err != nil {
			return nil, err
		}
		stmt.OrderBy = orderBy
	}

	// Parse LIMIT clause (except for Teradata which uses TOP)
	if p.dialect.Name() != "Teradata" && p.expectKeyword("LIMIT") {
		limit, err := p.parseLimitClause()
		if err != nil {
			return nil, err
		}
		stmt.Limit = limit
	}

	// Parse Teradata-specific TOP clause
	if p.dialect.Name() == "Teradata" && p.expectKeyword("TOP") {
		limit, err := p.parseTeradataTopClause()
		if err != nil {
			return nil, err
		}
		stmt.Limit = limit
	}

	return stmt, nil
}

func (p *Parser) parseSelectList() ([]ast.Expression, error) {
	var exprs []ast.Expression

	for {
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)

		if !p.expectToken(",") {
			break
		}
	}

	return exprs, nil
}

func (p *Parser) parseFromClause() (*ast.FromClause, error) {
	fromClause := &ast.FromClause{}

	for {
		table, err := p.parseTableReference()
		if err != nil {
			return nil, err
		}
		fromClause.Tables = append(fromClause.Tables, *table)

		if !p.expectToken(",") {
			break
		}
	}

	return fromClause, nil
}

func (p *Parser) parseTableReference() (*ast.TableReference, error) {
	tableName, err := p.parseIdentifier()
	if err != nil {
		return nil, err
	}

	ref := &ast.TableReference{
		TableName: tableName,
	}

	// Check for alias
	if p.expectKeyword("AS") {
		alias, err := p.parseIdentifier()
		if err != nil {
			return nil, err
		}
		ref.Alias = alias
	}

	// Check for JOIN
	if p.isJoinKeyword() {
		joinType, err := p.parseJoinType()
		if err != nil {
			return nil, err
		}
		ref.JoinType = joinType

		// Parse join condition
		if p.expectKeyword("ON") {
			joinCond, err := p.parseExpression()
			if err != nil {
				return nil, err
			}
			ref.JoinCond = joinCond
		}
	}

	return ref, nil
}

func (p *Parser) parseWhereClause() (*ast.WhereClause, error) {
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &ast.WhereClause{Condition: expr}, nil
}

func (p *Parser) parseExpression() (ast.Expression, error) {
	// This is a simplified expression parser
	// In a real implementation, you would need to handle operator precedence,
	// parentheses, functions, etc.

	left, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	// Look for binary operators
	if p.isOperator() {
		op := p.currentToken()
		p.advance()

		right, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		return &ast.BinaryExpression{
			Left:     left,
			Operator: op,
			Right:    right,
		}, nil
	}

	return left, nil
}

func (p *Parser) parsePrimary() (ast.Expression, error) {
	token := p.currentToken()
	p.advance()

	// Check if it's a literal
	if isStringLiteral(token) {
		return &ast.Literal{
			Value: strings.Trim(token, "'\""),
			Type:  ast.STRING,
		}, nil
	}

	if isNumericLiteral(token) {
		return &ast.Literal{
			Value: token,
			Type:  ast.NUMBER,
		}, nil
	}

	// Assume it's a column reference
	parts := strings.Split(token, ".")
	if len(parts) == 2 {
		return &ast.ColumnReference{
			Table:  parts[0],
			Column: parts[1],
		}, nil
	}

	return &ast.ColumnReference{
		Column: token,
	}, nil
}

// Helper functions
func (p *Parser) skipWhitespace() {
	for p.pos < len(p.tokens) && isWhitespace(p.tokens[p.pos]) {
		p.pos++
	}
}

func (p *Parser) expectKeyword(keyword string) bool {
	if p.pos >= len(p.tokens) {
		return false
	}

	if strings.ToUpper(p.tokens[p.pos]) == strings.ToUpper(keyword) {
		p.pos++
		return true
	}
	return false
}

func (p *Parser) expectToken(token string) bool {
	if p.pos >= len(p.tokens) {
		return false
	}

	if p.tokens[p.pos] == token {
		p.pos++
		return true
	}
	return false
}

func (p *Parser) currentToken() string {
	if p.pos >= len(p.tokens) {
		return ""
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() {
	p.pos++
}

func (p *Parser) isOperator() bool {
	ops := []string{"=", "<", ">", "<=", ">=", "<>", "AND", "OR", "+", "-", "*", "/"}
	token := strings.ToUpper(p.currentToken())
	for _, op := range ops {
		if token == op {
			return true
		}
	}
	return false
}

func (p *Parser) isJoinKeyword() bool {
	joins := []string{"JOIN", "INNER JOIN", "LEFT JOIN", "RIGHT JOIN", "FULL JOIN", "CROSS JOIN"}
	token := strings.ToUpper(p.currentToken())
	for _, join := range joins {
		if token == join {
			return true
		}
	}
	return false
}

func tokenize(query string) []string {
	// This is a simplified tokenizer
	// In a real implementation, you would need a more sophisticated tokenizer
	// that handles quoted strings, comments, etc.
	return strings.Fields(query)
}

func isWhitespace(s string) bool {
	return strings.TrimSpace(s) == ""
}

func isStringLiteral(s string) bool {
	return (strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) ||
		(strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\""))
}

func isNumericLiteral(s string) bool {
	// This is a simplified check
	// In a real implementation, you would need to handle different numeric formats
	_, err := fmt.Sscanf(s, "%f", new(float64))
	return err == nil
}

func (p *Parser) parseIdentifier() (string, error) {
	token := p.currentToken()
	if token == "" {
		return "", fmt.Errorf("unexpected end of input while parsing identifier")
	}

	// Check if the identifier needs to be quoted
	if strings.HasPrefix(token, p.dialect.GetQuoteCharacter()) && strings.HasSuffix(token, p.dialect.GetQuoteCharacter()) {
		// Remove quotes and return
		return token[1 : len(token)-1], nil
	}

	// For unquoted identifiers, validate according to dialect rules
	if !p.dialect.ValidateIdentifier(token) {
		return "", fmt.Errorf("invalid identifier '%s' for %s dialect", token, p.dialect.Name())
	}

	p.advance()
	return token, nil
}

func (p *Parser) parseExpressionList() ([]ast.Expression, error) {
	var exprs []ast.Expression

	for {
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)

		if !p.expectToken(",") {
			break
		}
	}

	return exprs, nil
}

func (p *Parser) parseOrderByClause() ([]ast.OrderByExpr, error) {
	var orderBy []ast.OrderByExpr

	for {
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		// Check for ASC/DESC
		ascending := true
		if p.expectKeyword("DESC") {
			ascending = false
		} else {
			p.expectKeyword("ASC") // ASC is optional
		}

		orderBy = append(orderBy, ast.OrderByExpr{
			Expr:      expr,
			Ascending: ascending,
		})

		if !p.expectToken(",") {
			break
		}
	}

	return orderBy, nil
}

func (p *Parser) parseLimitClause() (*ast.LimitClause, error) {
	// Parse the LIMIT value
	token := p.currentToken()
	p.advance()

	var count int64
	if _, err := fmt.Sscanf(token, "%d", &count); err != nil {
		return nil, fmt.Errorf("invalid LIMIT value: %s", token)
	}

	// Check for optional OFFSET
	var offset int64
	if p.expectKeyword("OFFSET") {
		token = p.currentToken()
		p.advance()

		if _, err := fmt.Sscanf(token, "%d", &offset); err != nil {
			return nil, fmt.Errorf("invalid OFFSET value: %s", token)
		}
	}

	return &ast.LimitClause{
		Count:  count,
		Offset: offset,
	}, nil
}

func (p *Parser) parseJoinType() (ast.JoinType, error) {
	token := strings.ToUpper(p.currentToken())
	p.advance()

	switch token {
	case "JOIN", "INNER JOIN":
		return ast.INNER, nil
	case "LEFT JOIN":
		return ast.LEFT, nil
	case "RIGHT JOIN":
		return ast.RIGHT, nil
	case "FULL JOIN":
		return ast.FULL, nil
	case "CROSS JOIN":
		return ast.CROSS, nil
	default:
		return ast.INNER, fmt.Errorf("unknown join type: %s", token)
	}
}

func (p *Parser) parseTeradataTopClause() (*ast.LimitClause, error) {
	// Parse the number after TOP
	token := p.currentToken()
	p.advance()

	var count int64
	if _, err := fmt.Sscanf(token, "%d", &count); err != nil {
		return nil, fmt.Errorf("invalid TOP value: %s", token)
	}

	return &ast.LimitClause{
		Count: count,
	}, nil
}

func (p *Parser) isKeyword(word string) bool {
	return p.dialect.GetReservedKeywords()[strings.ToUpper(word)]
}

func (p *Parser) isSpecialFunction(word string) bool {
	return p.dialect.GetSpecialFunctions()[strings.ToUpper(word)]
}
