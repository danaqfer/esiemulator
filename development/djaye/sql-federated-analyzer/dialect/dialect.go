// Package dialect provides interfaces and implementations for handling different SQL
// dialects. Each dialect implementation encapsulates the specific syntax rules,
// identifier conventions, and special functions supported by a particular SQL variant.
package dialect

// Dialect represents a specific SQL dialect's parsing rules and behaviors.
// Each implementation of this interface handles the unique characteristics of
// a particular SQL variant (e.g., Athena, BigQuery, Teradata, PostgreSQL).
type Dialect interface {
	// Name returns the name of the dialect (e.g., "Athena", "BigQuery", "Teradata", "PostgreSQL").
	// This is used for identification and configuration purposes.
	Name() string

	// ValidateIdentifier checks if an identifier is valid in this dialect.
	// Different dialects have different rules for valid identifier characters
	// and length limitations.
	ValidateIdentifier(identifier string) bool

	// GetQuoteCharacter returns the character used for quoting identifiers.
	// For example: backtick (`) in BigQuery, double quote (") in PostgreSQL.
	GetQuoteCharacter() string

	// GetStringLiteralQuote returns the character used for string literals.
	// Most dialects use single quote ('), but some may have alternatives.
	GetStringLiteralQuote() string

	// GetCatalogSeparator returns the character that separates catalog/schema/table.
	// Most dialects use period (.), but some may have different conventions.
	GetCatalogSeparator() string

	// SupportsArrayType returns whether the dialect supports array types.
	// Some dialects like PostgreSQL and BigQuery support arrays, while others don't.
	SupportsArrayType() bool

	// SupportsWindowFunctions returns whether the dialect supports window functions.
	// Most modern SQL dialects support window functions, but some older ones may not.
	SupportsWindowFunctions() bool

	// GetReservedKeywords returns a set of keywords that are reserved in this dialect.
	// These keywords cannot be used as identifiers without proper quoting.
	GetReservedKeywords() map[string]bool

	// GetSpecialFunctions returns a map of special functions specific to this dialect.
	// These are functions that may have unique syntax or behavior in this dialect.
	GetSpecialFunctions() map[string]bool
}
