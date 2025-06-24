package dialect

// PostgresDialect implements the Dialect interface for PostgreSQL
type PostgresDialect struct{}

func NewPostgresDialect() *PostgresDialect {
	return &PostgresDialect{}
}

func (d *PostgresDialect) Name() string {
	return "PostgreSQL"
}

func (d *PostgresDialect) ValidateIdentifier(identifier string) bool {
	// PostgreSQL identifiers must start with a letter or underscore
	// and can contain letters, numbers, underscores, and dollar signs
	if len(identifier) == 0 {
		return false
	}
	first := identifier[0]
	if !(first >= 'a' && first <= 'z' || first >= 'A' && first <= 'Z' || first == '_') {
		return false
	}
	for _, c := range identifier {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '$') {
			return false
		}
	}
	return true
}

func (d *PostgresDialect) GetQuoteCharacter() string {
	return "\""
}

func (d *PostgresDialect) GetStringLiteralQuote() string {
	return "'"
}

func (d *PostgresDialect) GetCatalogSeparator() string {
	return "."
}

func (d *PostgresDialect) SupportsArrayType() bool {
	return true
}

func (d *PostgresDialect) SupportsWindowFunctions() bool {
	return true
}

func (d *PostgresDialect) GetReservedKeywords() map[string]bool {
	return map[string]bool{
		"SELECT":    true,
		"FROM":      true,
		"WHERE":     true,
		"AND":       true,
		"OR":        true,
		"NOT":       true,
		"IN":        true,
		"EXISTS":    true,
		"GROUP":     true,
		"BY":        true,
		"HAVING":    true,
		"ORDER":     true,
		"LIMIT":     true,
		"OFFSET":    true,
		"UNION":     true,
		"ALL":       true,
		"EXCEPT":    true,
		"INTERSECT": true,
		"RETURNING": true,
		"INTO":      true,
		"WITH":      true,
		"RECURSIVE": true,
	}
}

func (d *PostgresDialect) GetSpecialFunctions() map[string]bool {
	return map[string]bool{
		"array_agg":       true,
		"string_agg":      true,
		"json_agg":        true,
		"jsonb_agg":       true,
		"to_tsvector":     true,
		"to_tsquery":      true,
		"generate_series": true,
		"unnest":          true,
		"array_remove":    true,
		"array_replace":   true,
	}
}
