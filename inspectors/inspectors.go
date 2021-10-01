package inspectors

// Inspector returns a function that can provide metadata about the schema we are building test data for.
type Inspector func() (map[string][]ColumnInfo, error)

// ColumnInfo tracks database column type information.
type ColumnInfo struct {
	Name       string `json:"name"`
	DataType   string `json:"data_type"`
	IsPrimary  bool   `json:"is_primary"`
	IsUnsigned bool   `json:"is_unsigned"`
	Length     int    `json:"length"`
	Type       string `json:"column_type"`
}
