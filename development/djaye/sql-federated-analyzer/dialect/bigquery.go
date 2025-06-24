package dialect

// BigQueryDialect implements the Dialect interface for Google BigQuery
type BigQueryDialect struct{}

func NewBigQueryDialect() *BigQueryDialect {
	return &BigQueryDialect{}
}

func (d *BigQueryDialect) Name() string {
	return "BigQuery"
}

func (d *BigQueryDialect) ValidateIdentifier(identifier string) bool {
	// BigQuery identifiers must start with a letter or underscore
	// and can contain letters, numbers, and underscores
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

func (d *BigQueryDialect) GetQuoteCharacter() string {
	return "`"
}

func (d *BigQueryDialect) GetStringLiteralQuote() string {
	return "'"
}

func (d *BigQueryDialect) GetCatalogSeparator() string {
	return "."
}

func (d *BigQueryDialect) SupportsArrayType() bool {
	return true
}

func (d *BigQueryDialect) SupportsWindowFunctions() bool {
	return true
}

func (d *BigQueryDialect) GetReservedKeywords() map[string]bool {
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
		"UNION":     true,
		"ALL":       true,
		"EXCEPT":    true,
		"STRUCT":    true,
		"ARRAY":     true,
		"UNNEST":    true,
		"PARTITION": true,
		"CLUSTER":   true,
	}
}

func (d *BigQueryDialect) GetSpecialFunctions() map[string]bool {
	return map[string]bool{
		"ARRAY_AGG":           true,
		"ANY_VALUE":           true,
		"ARRAY_CONCAT":        true,
		"GENERATE_ARRAY":      true,
		"GENERATE_DATE_ARRAY": true,
		"ST_GEOGPOINT":        true,
		"ST_DISTANCE":         true,
		"ML.PREDICT":          true,
		"ML.EVALUATE":         true,
	}
}
