package dialect

import (
	"fmt"
	"strings"
)

// AqferDialect defines the interface for dialect-specific Aqfer federated query provider functionality
type AqferDialect interface {
	// Basic SQL dialect functionality
	SQLDialect

	// IsFederatedTable checks if a table reference points to an Aqfer federated query provider
	// For example: "impressions.dsae" indicates this table is served by the Aqfer federated query provider
	IsFederatedTable(tableName string) bool

	// GetFederatedTableName extracts the actual table name that will be queried through the federated provider
	// For example: "impressions.dsae" -> "impressions"
	GetFederatedTableName(tableName string) string

	// GetFederatedDatasetType determines the type of dataset this federated table represents
	// This helps the provider optimize data access patterns
	GetFederatedDatasetType(tableName string) string

	// FormatFederatedFilter formats a filter condition for the federated query provider
	// This will be used to push down filters to the Aqfer backend
	FormatFederatedFilter(column string, operator string, values []string) string
}

// AthenaAqferDialect implements AqferDialect for AWS Athena's Aqfer federated query provider
type AthenaAqferDialect struct {
	*AthenaDialect
}

// NewAthenaAqferDialect creates a new Athena-specific Aqfer federated query provider dialect
func NewAthenaAqferDialect() *AthenaAqferDialect {
	return &AthenaAqferDialect{
		AthenaDialect: NewAthenaDialect(),
	}
}

// IsFederatedTable checks if a table name represents an Aqfer federated query provider table
func (d *AthenaAqferDialect) IsFederatedTable(tableName string) bool {
	// In Athena, Aqfer federated tables are identified by the .dsae suffix
	return strings.HasSuffix(tableName, ".dsae")
}

// GetFederatedTableName extracts the actual table name from an Aqfer federated table reference
func (d *AthenaAqferDialect) GetFederatedTableName(tableName string) string {
	// Remove the .dsae suffix to get the actual table name
	return strings.TrimSuffix(tableName, ".dsae")
}

// GetFederatedDatasetType determines the dataset type from the federated table name
func (d *AthenaAqferDialect) GetFederatedDatasetType(tableName string) string {
	// This would be based on naming conventions or metadata from the federated provider
	// For now, we'll assume it's encoded in the table name
	parts := strings.Split(tableName, "_")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// FormatFederatedFilter formats a filter condition for the federated query provider
func (d *AthenaAqferDialect) FormatFederatedFilter(column string, operator string, values []string) string {
	switch operator {
	case "IN":
		return fmt.Sprintf("%s IN (%s)",
			d.QuoteIdentifier(column),
			strings.Join(values, ","))
	case "=":
		if len(values) > 0 {
			return fmt.Sprintf("%s = %s",
				d.QuoteIdentifier(column),
				values[0])
		}
	}
	return ""
}

// IsAqferDialect checks if a dialect supports Aqfer functionality
func IsAqferDialect(d Dialect) bool {
	_, ok := d.(AqferDialect)
	return ok
}

// AsAqferDialect converts a Dialect to an AqferDialect if supported
func AsAqferDialect(d Dialect) (AqferDialect, bool) {
	if ad, ok := d.(AqferDialect); ok {
		return ad, true
	}
	return nil, false
}
