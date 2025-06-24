// Package ast provides types and interfaces for representing SQL queries as an Abstract
// Syntax Tree (AST). The AST structure allows for easy traversal, analysis, and
// transformation of SQL queries during the optimization process.
package ast

// Node represents a generic AST node that can be visited using the visitor pattern.
// All AST types must implement this interface to support traversal and transformation.
type Node interface {
	// Accept implements the visitor pattern, allowing operations to be performed
	// on the node during tree traversal.
	Accept(visitor Visitor) interface{}
}

// Visitor interface for traversing the AST. Each method corresponds to a different
// type of node in the tree and is called when that node type is encountered.
type Visitor interface {
	VisitSelect(node *SelectStatement) interface{}
	VisitFrom(node *FromClause) interface{}
	VisitWhere(node *WhereClause) interface{}
	VisitExpression(node Expression) interface{}
	VisitAqferTable(node *AqferTableReference) interface{}
}

// Expression represents any SQL expression that can appear in a query.
// This includes column references, literals, function calls, and operators.
type Expression interface {
	Node
	expressionNode()
}

// SelectStatement represents a SELECT query with all its components.
// This is typically the root node of the AST for a SELECT query.
type SelectStatement struct {
	SelectList []Expression  `json:"select_list"`        // List of expressions to select
	From       *FromClause   `json:"from,omitempty"`     // Source tables and joins
	Where      *WhereClause  `json:"where,omitempty"`    // Filter conditions
	GroupBy    []Expression  `json:"group_by,omitempty"` // Grouping expressions
	Having     Expression    `json:"having,omitempty"`   // Post-grouping filter
	OrderBy    []OrderByExpr `json:"order_by,omitempty"` // Sort specifications
	Limit      *LimitClause  `json:"limit,omitempty"`    // Row count limit
}

func (s *SelectStatement) Accept(v Visitor) interface{} {
	return v.VisitSelect(s)
}

// FromClause represents the FROM part of a query, containing the source tables
// and any joins between them.
type FromClause struct {
	Tables []TableReference `json:"tables"` // List of tables and their joins
}

func (f *FromClause) Accept(v Visitor) interface{} {
	return v.VisitFrom(f)
}

// TableReference represents a table in the FROM clause, including any alias
// and join information if applicable.
type TableReference struct {
	TableName string     `json:"table_name"`               // Name of the table
	Alias     string     `json:"alias,omitempty"`          // Optional table alias
	JoinType  JoinType   `json:"join_type,omitempty"`      // Type of join if this is a joined table
	JoinCond  Expression `json:"join_condition,omitempty"` // Join condition if this is a joined table
}

// JoinType represents different types of SQL JOINs supported by the parser.
type JoinType int

const (
	NONE  JoinType = iota // No join (base table)
	INNER                 // INNER JOIN or just JOIN
	LEFT                  // LEFT [OUTER] JOIN
	RIGHT                 // RIGHT [OUTER] JOIN
	FULL                  // FULL [OUTER] JOIN
	CROSS                 // CROSS JOIN
)

// WhereClause represents the WHERE part of a query, containing the filter
// conditions to be applied to the result set.
type WhereClause struct {
	Condition Expression `json:"condition"` // The boolean expression for filtering
}

func (w *WhereClause) Accept(v Visitor) interface{} {
	return v.VisitWhere(w)
}

// OrderByExpr represents an ORDER BY expression with its sort direction.
type OrderByExpr struct {
	Expr      Expression `json:"expression"` // The expression to sort by
	Ascending bool       `json:"ascending"`  // True for ASC, false for DESC
}

// LimitClause represents the LIMIT part of a query, optionally including
// an OFFSET clause.
type LimitClause struct {
	Count  int64 `json:"count"`            // Maximum number of rows to return
	Offset int64 `json:"offset,omitempty"` // Number of rows to skip
}

// BinaryExpression represents operations like AND, OR, =, >, etc. that
// operate on two expressions.
type BinaryExpression struct {
	Left     Expression `json:"left"`     // Left operand
	Operator string     `json:"operator"` // Operator symbol
	Right    Expression `json:"right"`    // Right operand
}

func (b *BinaryExpression) Accept(v Visitor) interface{} {
	return v.VisitExpression(b)
}

func (b *BinaryExpression) expressionNode() {}

// ColumnReference represents a reference to a column, optionally qualified
// with a table name.
type ColumnReference struct {
	Table  string `json:"table,omitempty"` // Optional table qualifier
	Column string `json:"column"`          // Column name
}

func (c *ColumnReference) Accept(v Visitor) interface{} {
	return v.VisitExpression(c)
}

func (c *ColumnReference) expressionNode() {}

// Literal represents a literal value in the query, such as strings,
// numbers, booleans, etc.
type Literal struct {
	Value interface{} `json:"value"` // The literal value
	Type  LiteralType `json:"type"`  // The type of the literal
}

func (l *Literal) Accept(v Visitor) interface{} {
	return v.VisitExpression(l)
}

func (l *Literal) expressionNode() {}

// LiteralType represents different types of literals that can appear
// in a SQL query.
type LiteralType int

const (
	STRING    LiteralType = iota // String literal
	NUMBER                       // Numeric literal
	BOOLEAN                      // Boolean literal
	NULL                         // NULL literal
	DATE                         // Date literal
	TIMESTAMP                    // Timestamp literal
)

// InExpression represents an IN clause, which tests if a column's value
// is in a set of values.
type InExpression struct {
	Column *ColumnReference `json:"column"` // The column to test
	Values []Expression     `json:"values"` // The set of values to test against
}

func (i *InExpression) Accept(v Visitor) interface{} {
	return v.VisitExpression(i)
}

func (i *InExpression) expressionNode() {}

// FunctionCall represents a SQL function call with its arguments
type FunctionCall struct {
	Name     string       `json:"name"`     // Function name
	Args     []Expression `json:"args"`     // Function arguments
	Distinct bool         `json:"distinct"` // Whether DISTINCT is specified
}

func (f *FunctionCall) Accept(v Visitor) interface{} {
	return v.VisitExpression(f)
}

func (f *FunctionCall) expressionNode() {}

// Star represents the * in SELECT * or COUNT(*)
type Star struct{}

func (s *Star) Accept(v Visitor) interface{} {
	return v.VisitExpression(s)
}

func (s *Star) expressionNode() {}

// AqferTableReference represents an Aqfer federated table in Athena
type AqferTableReference struct {
	TableName    string         // The original table name (e.g., my_aqfer_table)
	DatasetName  string         // The dataset name (e.g., dsae)
	Dimensions   []string       // List of dimension columns
	Metrics      []string       // List of metric columns
	JoinedTables []*JoinedTable // Tables joined with this Aqfer table
}

func (a *AqferTableReference) Accept(v Visitor) interface{} {
	return v.VisitAqferTable(a)
}

// JoinedTable represents a table joined with an Aqfer table
type JoinedTable struct {
	TableName   string    // Name of the joined table
	JoinType    JoinType  // Type of join (INNER, LEFT, etc.)
	JoinKeys    []JoinKey // Join conditions
	IsDimension bool      // Whether this table represents an Aqfer dimension
}

// JoinKey represents a join condition between tables
type JoinKey struct {
	LeftColumn  string // Column from the left table
	RightColumn string // Column from the right table
}
