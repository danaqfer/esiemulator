// Package optimizer provides SQL query optimization capabilities through a configurable
// optimization pipeline. It supports both generic optimizations that apply to all SQL
// dialects and dialect-specific optimizations that are only applied to particular SQL
// variants.
package optimizer

import (
	"sqlparser/ast"
	"sqlparser/dialect"
	"sqlparser/metadata"
	"strings"
)

// OptimizationPhase represents when an optimization should be applied
type OptimizationPhase int

const (
	PhaseGeneric         OptimizationPhase = iota // Applied during initial parsing
	PhaseDialectSpecific                          // Applied during SQL generation for specific dialects
)

// Optimization represents a single optimization rule that can be applied to an AST node.
// Each optimization has a name for identification, a phase indicating when it should be
// applied, and an Apply function that implements the actual optimization logic.
type Optimization struct {
	Name  string
	Phase OptimizationPhase
	Apply func(node ast.Node, ctx *OptimizationContext) (ast.Node, error)
}

// OptimizationContext provides context for optimizations including the current SQL
// dialect and metadata provider. This context is passed to each optimization during
// execution.
type OptimizationContext struct {
	Dialect          dialect.Dialect
	MetadataProvider metadata.MetadataProvider
}

// Optimizer applies optimization rules to the AST. It maintains a list of registered
// optimizations and applies them in the appropriate phase based on the optimization
// configuration.
type Optimizer struct {
	optimizations []Optimization
	ctx           *OptimizationContext
}

// NewOptimizer creates a new Optimizer instance
func NewOptimizer(metadataProvider metadata.MetadataProvider) *Optimizer {
	ctx := &OptimizationContext{
		MetadataProvider: metadataProvider,
	}

	opt := &Optimizer{
		ctx: ctx,
	}

	// Register optimizations
	opt.registerOptimizations()

	return opt
}

// Optimize applies optimization rules to the AST
func (o *Optimizer) Optimize(node ast.Node, phase OptimizationPhase, d dialect.Dialect) (ast.Node, error) {
	o.ctx.Dialect = d

	var err error
	for _, optimization := range o.optimizations {
		if optimization.Phase == phase {
			node, err = optimization.Apply(node, o.ctx)
			if err != nil {
				return nil, err
			}
		}
	}

	return node, nil
}

// registerOptimizations registers all available optimizations
func (o *Optimizer) registerOptimizations() {
	// Generic optimizations (applied during parsing)
	o.optimizations = append(o.optimizations, Optimization{
		Name:  "TableVariantSelection",
		Phase: PhaseGeneric,
		Apply: o.optimizeTableVariants,
	})

	// Dialect-specific optimizations (applied during SQL generation)
	o.optimizations = append(o.optimizations, Optimization{
		Name:  "OrToIn",
		Phase: PhaseDialectSpecific,
		Apply: func(node ast.Node, ctx *OptimizationContext) (ast.Node, error) {
			// Skip OR to IN optimization for Teradata
			if ctx.Dialect.Name() == "Teradata" {
				return node, nil
			}
			return o.optimizeOrToIn(node), nil
		},
	})
}

// optimizeTableVariants optimizes table selection based on partitioning and organization
func (o *Optimizer) optimizeTableVariants(node ast.Node, ctx *OptimizationContext) (ast.Node, error) {
	switch n := node.(type) {
	case *ast.SelectStatement:
		if n.From != nil {
			for i, table := range n.From.Tables {
				baseTableName := extractBaseTableName(table.TableName)
				if baseTableName != table.TableName {
					// This is a variant table, try to find the best variant
					variants, err := ctx.MetadataProvider.GetTableVariants(baseTableName)
					if err != nil {
						return node, err
					}

					// Get dimensions used in WHERE clause
					whereDimensions := extractWhereDimensions(n.Where)

					// Find best matching variant
					bestVariant := findBestTableVariant(variants, whereDimensions)
					if bestVariant != nil {
						n.From.Tables[i].TableName = baseTableName + "_" + bestVariant.Suffix
					}
				}
			}
		}
		return n, nil
	default:
		return node, nil
	}
}

// optimizeOrToIn converts OR conditions to IN clauses
func (o *Optimizer) optimizeOrToIn(node ast.Node) ast.Node {
	switch n := node.(type) {
	case *ast.SelectStatement:
		if n.Where != nil {
			n.Where = o.optimizeOrToIn(n.Where).(*ast.WhereClause)
		}
		return n
	case *ast.WhereClause:
		if n.Condition != nil {
			n.Condition = o.optimizeOrToIn(n.Condition).(ast.Expression)
		}
		return n
	case *ast.BinaryExpression:
		if n.Operator != "OR" {
			return n
		}

		orConditions := o.collectOrConditions(n)
		if len(orConditions) < 3 {
			return n
		}

		columnGroups := make(map[string][]ast.Expression)
		for _, cond := range orConditions {
			if binExpr, ok := cond.(*ast.BinaryExpression); ok && binExpr.Operator == "=" {
				if colRef, ok := binExpr.Left.(*ast.ColumnReference); ok {
					key := colRef.Column
					if colRef.Table != "" {
						key = colRef.Table + "." + key
					}
					columnGroups[key] = append(columnGroups[key], binExpr.Right)
				}
			}
		}

		for col, values := range columnGroups {
			if len(values) >= 3 {
				parts := splitColumnName(col)
				return &ast.InExpression{
					Column: &ast.ColumnReference{
						Table:  parts[0],
						Column: parts[1],
					},
					Values: values,
				}
			}
		}
	}
	return node
}

// Helper functions

func extractBaseTableName(tableName string) string {
	parts := strings.Split(tableName, "_")
	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], "_")
	}
	return tableName
}

func extractWhereDimensions(where *ast.WhereClause) map[string]bool {
	dimensions := make(map[string]bool)
	if where == nil || where.Condition == nil {
		return dimensions
	}

	var extract func(expr ast.Expression)
	extract = func(expr ast.Expression) {
		switch e := expr.(type) {
		case *ast.BinaryExpression:
			if e.Left != nil {
				if colRef, ok := e.Left.(*ast.ColumnReference); ok {
					dimensions[colRef.Column] = true
				}
			}
			if e.Right != nil {
				extract(e.Right)
			}
		}
	}

	extract(where.Condition)
	return dimensions
}

func findBestTableVariant(variants []metadata.TableMetadata, whereDimensions map[string]bool) *metadata.TableMetadata {
	var bestMatch *metadata.TableMetadata
	var maxScore int

	for i, variant := range variants {
		score := 0
		for _, dim := range variant.Dimensions {
			if whereDimensions[dim.ColumnName] {
				if dim.IsPartitionKey {
					score += 2 // Partition keys are more valuable
				}
				if dim.IsOrganizedBy {
					score += 1
				}
			}
		}
		if score > maxScore {
			maxScore = score
			bestMatch = &variants[i]
		}
	}

	return bestMatch
}

func (o *Optimizer) collectOrConditions(expr ast.Expression) []ast.Expression {
	if binExpr, ok := expr.(*ast.BinaryExpression); ok && binExpr.Operator == "OR" {
		return append(
			o.collectOrConditions(binExpr.Left),
			o.collectOrConditions(binExpr.Right)...,
		)
	}
	return []ast.Expression{expr}
}

func splitColumnName(col string) [2]string {
	var parts [2]string
	for i := len(col) - 1; i >= 0; i-- {
		if col[i] == '.' {
			parts[0] = col[:i]
			parts[1] = col[i+1:]
			return parts
		}
	}
	parts[1] = col
	return parts
}
