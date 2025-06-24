package dialect

// AthenaDialect implements the Dialect interface for Amazon Athena
type AthenaDialect struct{}

func NewAthenaDialect() *AthenaDialect {
	return &AthenaDialect{}
}

func (d *AthenaDialect) Name() string {
	return "Athena"
}

func (d *AthenaDialect) ValidateIdentifier(identifier string) bool {
	// Athena identifiers must start with a letter or underscore
	// and can contain only alphanumeric characters and underscores
	if len(identifier) == 0 {
		return false
	}
	first := identifier[0]
	if !(first >= 'a' && first <= 'z' || first >= 'A' && first <= 'Z' || first == '_') {
		return false
	}
	for _, c := range identifier {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_') {
			return false
		}
	}
	return true
}

func (d *AthenaDialect) GetQuoteCharacter() string {
	return "`"
}

func (d *AthenaDialect) GetStringLiteralQuote() string {
	return "'"
}

func (d *AthenaDialect) GetCatalogSeparator() string {
	return "."
}

func (d *AthenaDialect) SupportsArrayType() bool {
	return true
}

func (d *AthenaDialect) SupportsWindowFunctions() bool {
	return true
}

func (d *AthenaDialect) GetReservedKeywords() map[string]bool {
	return map[string]bool{
		"SELECT": true,
		"FROM":   true,
		"WHERE":  true,
		"AND":    true,
		"OR":     true,
		"NOT":    true,
		"IN":     true,
		"EXISTS": true,
		"GROUP":  true,
		"BY":     true,
		"HAVING": true,
		"ORDER":  true,
		"LIMIT":  true,
		"UNION":  true,
		"ALL":    true,
		"EXCEPT": true,
	}
}

func (d *AthenaDialect) GetSpecialFunctions() map[string]bool {
	return map[string]bool{
		"date_add":          true,
		"date_sub":          true,
		"date_trunc":        true,
		"from_unixtime":     true,
		"array_agg":         true,
		"map_agg":           true,
		"approx_distinct":   true,
		"approx_percentile": true,
	}
}
