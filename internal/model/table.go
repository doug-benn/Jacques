package model

type TableDef struct {
	Schema            string
	Name              string
	RawHeader         string
	Inherits          string
	PartitionBy       string
	IsPartition       bool
	Columns           []*ColumnDef
	TableConstraints  []string
	TableLevelPK      string
	TableLevelUniques []string
	TableLevelFKs     []string
	TableExclusions   []string
}

type ColumnDef struct {
	Name              string
	RawDef            string
	Default           string
	IsPrimaryKey      bool
	IsUnique          bool
	IsSerial          bool
	SequenceName      string
	References        string
	OnDelete          string
	OnUpdate          string
	Match             string
	IsDeferrable      bool
	InitiallyDeferred bool
	IndexMethod       string
}

func (t *TableDef) QualifiedName() string {
	if t.Schema != "" && t.Schema != "public" {
		return t.Schema + "." + t.Name
	}
	return t.Name
}
