package model

type TableDef struct {
	Schema            string
	Name              string
	RawHeader         string
	Columns           []*ColumnDef
	TableConstraints  []string
	TableLevelPK      string
	TableLevelUniques []string
	TableLevelFKs     []string
}

type ColumnDef struct {
	Name         string
	RawDef       string
	Default      string
	IsPrimaryKey bool
	IsUnique     bool
	IsSerial     bool
	SequenceName string
	References   string
	OnDelete     string
	OnUpdate     string
	Match        string
}

func (t *TableDef) QualifiedName() string {
	if t.Schema != "" {
		return t.Schema + "." + t.Name
	}
	return t.Name
}
