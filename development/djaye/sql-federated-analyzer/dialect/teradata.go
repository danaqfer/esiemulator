package dialect

// TeradataDialect implements the Dialect interface for Teradata
type TeradataDialect struct{}

func NewTeradataDialect() *TeradataDialect {
	return &TeradataDialect{}
}

func (d *TeradataDialect) Name() string {
	return "Teradata"
}

func (d *TeradataDialect) ValidateIdentifier(identifier string) bool {
	// Teradata identifiers must start with a letter
	// and can contain letters, numbers, and underscores
	if len(identifier) == 0 {
		return false
	}
	first := identifier[0]
	if !(first >= 'a' && first <= 'z' || first >= 'A' && first <= 'Z') {
		return false
	}
	for _, c := range identifier {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_') {
			return false
		}
	}
	return true
}

func (d *TeradataDialect) GetQuoteCharacter() string {
	return "\""
}

func (d *TeradataDialect) GetStringLiteralQuote() string {
	return "'"
}

func (d *TeradataDialect) GetCatalogSeparator() string {
	return "."
}

func (d *TeradataDialect) SupportsArrayType() bool {
	return false
}

func (d *TeradataDialect) SupportsWindowFunctions() bool {
	return true
}

func (d *TeradataDialect) GetReservedKeywords() map[string]bool {
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
		"TOP":       true,
		"SAMPLE":    true,
		"QUALIFY":   true,
		"UNION":     true,
		"ALL":       true,
		"EXCEPT":    true,
		"INTERSECT": true,
		"DATABASE":  true,
		"USER":      true,
		"ROLE":      true,
	}
}

func (d *TeradataDialect) GetSpecialFunctions() map[string]bool {
	return map[string]bool{
		"QUALIFY":     true,
		"RANK":        true,
		"DENSE_RANK":  true,
		"ROW_NUMBER":  true,
		"COLLECT":     true,
		"PERIOD":      true,
		"EXPAND":      true,
		"NORMALIZE":   true,
		"TD_SYSFNLIB": true,
		"HASHROW":     true,
		"HASHBUCKET":  true,
	}
}
