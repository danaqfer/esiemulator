package testutil

import "sqlparser/metadata"

// MockMetadataProvider implements a simple metadata provider for testing and examples
type MockMetadataProvider struct{}

func (m *MockMetadataProvider) GetTableVariants(baseTableName string) ([]metadata.TableMetadata, error) {
	switch baseTableName {
	case "customers":
		return []metadata.TableMetadata{
			{
				BaseTableName: "customers",
				Suffix:        "a",
				Format:        "parquet",
				Dimensions: []metadata.Dimension{
					{ColumnName: "region", IsPartitionKey: true},
					{ColumnName: "status", IsOrganizedBy: true},
				},
				UniqueKeys: []string{"id", "email"},
			},
			{
				BaseTableName: "customers",
				Suffix:        "b",
				Format:        "parquet",
				Dimensions: []metadata.Dimension{
					{ColumnName: "status", IsPartitionKey: true},
					{ColumnName: "created_date", IsOrganizedBy: true},
				},
				UniqueKeys: []string{"id", "email"},
			},
		}, nil
	default:
		return nil, nil
	}
}
