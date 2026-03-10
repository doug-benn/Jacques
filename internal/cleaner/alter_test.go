package cleaner

import (
	"testing"

	"github.com/doug-benn/Jacques/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestNewTable(t *testing.T) {
	cols := []*model.ColumnDef{{Name: "id", RawDef: "int"}}
	td := newTable("foo", cols)
	assert.Equal(t, "foo", td.Name)
	assert.Equal(t, cols, td.Columns)
}

func TestNewTableWithConstraints(t *testing.T) {
	cols := []*model.ColumnDef{{Name: "id", RawDef: "int"}}
	constraints := []string{"CHECK (id > 0)"}
	td := newTableWithConstraints("foo", cols, constraints)
	assert.Equal(t, "foo", td.Name)
	assert.Equal(t, cols, td.Columns)
	assert.Equal(t, constraints, td.TableConstraints)
}
