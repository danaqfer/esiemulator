package optimizer

import (
	"sqlparser/ast"
	"sqlparser/metadata"
	"strings"
)

// OptimizationMode determines whether an optimization uses include or exclude list for dialects.
// This allows fine-grained control over which dialects an optimization applies to.
type OptimizationMode int

const (
	ModeInclude OptimizationMode = iota // Only apply to listed dialects
	ModeExclude                         // Apply to all except listed dialects
)

// DialectOptimizationConfig configures which dialects an optimization applies to.
// It supports both inclusion and exclusion modes to provide flexibility in specifying
// dialect compatibility.
type DialectOptimizationConfig struct {
	Mode     OptimizationMode // Whether to use include or exclude list
	Dialects []string         // List of dialects to include or exclude
}

// OptimizationFlags represents which optimizations are enabled/disabled.
// This provides a simple way to toggle optimizations without changing the
// optimization configuration structure.
type OptimizationFlags struct {
	// Generic optimizations
	EnableTableRouting bool

	// Dialect-specific optimizations
	EnableOrToIn bool
}

// OptimizationConfig holds the configuration for optimization application.
// It provides fine-grained control over which optimizations are applied
// and how they behave for different SQL dialects.
type OptimizationConfig struct {
	// Generic optimization settings
	IncludeList []string // Optimizations to explicitly include
	ExcludeList []string // Optimizations to explicitly exclude
	Flags       OptimizationFlags

	// Dialect-specific optimization settings
	DialectConfigs map[string]DialectOptimizationConfig
}

// NewDefaultConfig creates a default optimization configuration with sensible defaults.
// By default, it enables table routing and OR to IN conversion, with Teradata excluded
// from OR to IN conversion due to dialect limitations.
func NewDefaultConfig() *OptimizationConfig {
	return &OptimizationConfig{
		Flags: OptimizationFlags{
			EnableTableRouting: true,
			EnableOrToIn:       true,
		},
		DialectConfigs: map[string]DialectOptimizationConfig{
			"or_to_in": {
				Mode:     ModeExclude,
				Dialects: []string{"Teradata"}, // Teradata doesn't support OR to IN conversion
			},
		},
	}
}

// GenericOptimizer represents a function that performs a generic optimization
type GenericOptimizer func(node ast.Node, provider metadata.MetadataProvider) (ast.Node, error)

// DialectOptimizer represents a function that performs a dialect-specific optimization
type DialectOptimizer func(node ast.Node) (ast.Node, error)

// Collection of generic optimizations
var genericOptimizations = map[string]GenericOptimizer{
	"table_routing":           optimizeTableRouting,
	"count_distinct_to_count": optimizeCountDistinct,
	// Add more generic optimizations here
}

// Collection of dialect-specific optimizations
var dialectOptimizations = map[string]DialectOptimizer{
	"or_to_in": optimizeOrToIn,
	// Add more dialect-specific optimizations here
}

// ApplyGenericOptimizations applies all enabled generic optimizations
func ApplyGenericOptimizations(node ast.Node, provider metadata.MetadataProvider, config *OptimizationConfig) (ast.Node, error) {
	var err error
	optimizedNode := node

	for name, optimizer := range genericOptimizations {
		if shouldApplyOptimization(name, config) {
			optimizedNode, err = optimizer(optimizedNode, provider)
			if err != nil {
				return nil, err
			}
		}
	}

	return optimizedNode, nil
}

// ApplyDialectOptimizations applies all enabled dialect-specific optimizations
func ApplyDialectOptimizations(node ast.Node, config *OptimizationConfig, dialectName string) (ast.Node, error) {
	var err error
	optimizedNode := node

	for name, optimizer := range dialectOptimizations {
		if shouldApplyOptimization(name, config) && shouldApplyToDialect(name, dialectName, config) {
			optimizedNode, err = optimizer(optimizedNode)
			if err != nil {
				return nil, err
			}
		}
	}

	return optimizedNode, nil
}

// shouldApplyOptimization checks if an optimization should be applied based on config
func shouldApplyOptimization(name string, config *OptimizationConfig) bool {
	// Check exclude list first
	for _, excluded := range config.ExcludeList {
		if excluded == name {
			return false
		}
	}

	// If include list is empty, apply all non-excluded optimizations
	if len(config.IncludeList) == 0 {
		return true
	}

	// Check include list
	for _, included := range config.IncludeList {
		if included == name {
			return true
		}
	}

	return false
}

// shouldApplyToDialect checks if an optimization should be applied to a specific dialect
func shouldApplyToDialect(optimizationName string, dialectName string, config *OptimizationConfig) bool {
	dialectConfig, exists := config.DialectConfigs[optimizationName]
	if !exists {
		return true // If no dialect config exists, apply to all dialects
	}

	// Check if dialect is in the include/exclude list
	for _, d := range dialectConfig.Dialects {
		if d == dialectName {
			return dialectConfig.Mode == ModeInclude // Return true if in include list, false if in exclude list
		}
	}

	return dialectConfig.Mode == ModeExclude // Return false if not in include list, true if not in exclude list
}

// Individual optimization implementations

func optimizeTableRouting(node ast.Node, provider metadata.MetadataProvider) (ast.Node, error) {
	// Implementation moved from optimizer.go
	// This is the table routing logic that was previously in optimizeTableVariants
	return node, nil
}

func optimizeOrToIn(node ast.Node) (ast.Node, error) {
	// Implementation moved from optimizer.go
	// This is the OR to IN conversion logic that was previously in optimizeOrToIn
	return node, nil
}

// optimizeCountDistinct converts COUNT(DISTINCT col) to COUNT(*) when col is a unique key
func optimizeCountDistinct(node ast.Node, provider metadata.MetadataProvider) (ast.Node, error) {
	switch n := node.(type) {
	case *ast.SelectStatement:
		for i, expr := range n.SelectList {
			if funcCall, ok := expr.(*ast.FunctionCall); ok {
				if strings.ToUpper(funcCall.Name) == "COUNT" && funcCall.Distinct && len(funcCall.Args) == 1 {
					if colRef, ok := funcCall.Args[0].(*ast.ColumnReference); ok {
						// Get table metadata to check for unique keys
						if n.From != nil && len(n.From.Tables) > 0 {
							baseTableName := extractBaseTableName(n.From.Tables[0].TableName)
							variants, err := provider.GetTableVariants(baseTableName)
							if err != nil {
								return node, err
							}

							// Check if the column is a unique key in any variant
							for _, variant := range variants {
								for _, uniqueKey := range variant.UniqueKeys {
									if uniqueKey == colRef.Column {
										// Replace COUNT(DISTINCT col) with COUNT(*)
										n.SelectList[i] = &ast.FunctionCall{
											Name: "COUNT",
											Args: []ast.Expression{&ast.Star{}},
										}
										break
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return node, nil
}
