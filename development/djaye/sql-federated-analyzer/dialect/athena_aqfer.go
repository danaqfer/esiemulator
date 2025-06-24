package dialect

import (
	"strings"
)

// AthenaAqferDialect extends AthenaDialect with Aqfer-specific functionality
type AthenaAqferDialect struct {
	*AthenaDialect
}

// NewAthenaAqferDialect creates a new AthenaAqferDialect instance
func NewAthenaAqferDialect() *AthenaAqferDialect {
	return &AthenaAqferDialect{
		AthenaDialect: NewAthenaDialect(),
	}
}

// IsAqferTable checks if a table reference is an Aqfer federated table
func (d *AthenaAqferDialect) IsAqferTable(tableName string) bool {
	// In Athena, Aqfer tables use the .dsae suffix
	return strings.HasSuffix(tableName, ".dsae")
}

// GetAqferDatasetType extracts the dataset type from an Aqfer table name
func (d *AthenaAqferDialect) GetAqferDatasetType(tableName string) string {
	if d.IsAqferTable(tableName) {
		return "dsae"
	}
	return ""
}

// GetAqferTableName extracts the base table name from an Aqfer table reference
func (d *AthenaAqferDialect) GetAqferTableName(tableName string) string {
	if d.IsAqferTable(tableName) {
		return strings.TrimSuffix(tableName, ".dsae")
	}
	return tableName
}

// ValidateAqferTableName checks if an Aqfer table name is valid
func (d *AthenaAqferDialect) ValidateAqferTableName(tableName string) bool {
	// Athena Aqfer tables must:
	// 1. End with .dsae
	// 2. Have a valid base table name
	if !d.IsAqferTable(tableName) {
		return false
	}
	baseName := d.GetAqferTableName(tableName)
	return d.ValidateIdentifier(baseName)
}

// GetAqferCatalogFormat returns the format for Aqfer catalog references
func (d *AthenaAqferDialect) GetAqferCatalogFormat() string {
	return "%s.dsae" // Athena uses table_name.dsae format
}

// GetAqferSpecialFunctions returns Aqfer-specific functions available in Athena
func (d *AthenaAqferDialect) GetAqferSpecialFunctions() map[string]bool {
	// Add any Aqfer-specific functions that are only available in Athena
	return map[string]bool{
		"aqfer_dimension": true,
		"aqfer_metric":    true,
	}
}
