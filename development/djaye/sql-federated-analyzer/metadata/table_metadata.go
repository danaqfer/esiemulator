// Package metadata provides interfaces and types for accessing and managing table
// metadata information. This includes table variants, partitioning schemes, and
// storage formats across different metadata providers like VTMS, Iceberg, and AWS Glue.
package metadata

// Dimension represents a column used for partitioning or organization in a table.
// This information is used to optimize query routing and execution by leveraging
// the physical organization of data.
type Dimension struct {
	ColumnName     string // Name of the column
	IsPartitionKey bool   // Whether this column is used as a partition key
	IsOrganizedBy  bool   // Whether data is organized/clustered by this column
}

// TableMetadata represents the organization and partitioning information for a table.
// This includes both logical information (base table name) and physical organization
// details (format, partitioning, etc.).
type TableMetadata struct {
	BaseTableName string      // e.g., "customers" for "customers_a"
	Suffix        string      // e.g., "a" for "customers_a"
	Format        string      // e.g., "parquet", "iceberg"
	Dimensions    []Dimension // Columns used for partitioning and organization
	UniqueKeys    []string    // List of columns that are unique keys (including primary key)
}

// MetadataProvider defines the interface for retrieving table metadata.
// Different implementations can support various metadata storage systems
// like VTMS, Iceberg tables, or AWS Glue catalog.
type MetadataProvider interface {
	// GetTableVariants returns all variants of a base table name.
	// A variant is a physical table that shares the same logical schema
	// but may have different partitioning or organization.
	GetTableVariants(baseTableName string) ([]TableMetadata, error)
}

// VTMSMetadataProvider implements MetadataProvider for VTMS (Variant Table Management System).
// VTMS is a custom system for managing table variants with different physical organizations.
type VTMSMetadataProvider struct {
	// Add VTMS client configuration here
}

// NewVTMSMetadataProvider creates a new VTMS metadata provider instance.
func NewVTMSMetadataProvider() *VTMSMetadataProvider {
	return &VTMSMetadataProvider{}
}

// GetTableVariants implements the MetadataProvider interface for VTMS.
func (p *VTMSMetadataProvider) GetTableVariants(baseTableName string) ([]TableMetadata, error) {
	// Implement VTMS metadata retrieval
	// This is a placeholder implementation
	return nil, nil
}

// IcebergMetadataProvider implements MetadataProvider for Apache Iceberg tables.
// Iceberg provides its own metadata and manifest management system.
type IcebergMetadataProvider struct {
	// Add Iceberg client configuration here
}

// NewIcebergMetadataProvider creates a new Iceberg metadata provider instance.
func NewIcebergMetadataProvider() *IcebergMetadataProvider {
	return &IcebergMetadataProvider{}
}

// GetTableVariants implements the MetadataProvider interface for Iceberg.
func (p *IcebergMetadataProvider) GetTableVariants(baseTableName string) ([]TableMetadata, error) {
	// Implement Iceberg metadata retrieval
	// This is a placeholder implementation
	return nil, nil
}

// GlueMetadataProvider implements MetadataProvider for AWS Glue Data Catalog.
// AWS Glue provides a centralized metadata repository for AWS data services.
type GlueMetadataProvider struct {
	// Add AWS Glue client configuration here
}

// NewGlueMetadataProvider creates a new AWS Glue metadata provider instance.
func NewGlueMetadataProvider() *GlueMetadataProvider {
	return &GlueMetadataProvider{}
}

// GetTableVariants implements the MetadataProvider interface for AWS Glue.
func (p *GlueMetadataProvider) GetTableVariants(baseTableName string) ([]TableMetadata, error) {
	// Implement AWS Glue metadata retrieval
	// This is a placeholder implementation
	return nil, nil
}
