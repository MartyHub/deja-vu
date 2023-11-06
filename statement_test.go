package dejavu

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStatement(t *testing.T) {
	qb := NewStatement("select * from test_table")

	require.NotNil(t, qb)
	assert.Equal(t, "select * from test_table", qb.sql)
	assert.Empty(t, qb.args)
}

func TestStatement_Arg(t *testing.T) {
	qb := NewStatement("select * from test_table").Arg("name", "value")

	require.NotNil(t, qb)
	assert.Len(t, qb.args, 1)
	assert.Equal(t, "name", qb.args[0].Name)
	assert.Equal(t, "value", qb.args[0].Value)
}

func TestStatement_WithSyntax_NoArg(t *testing.T) {
	tests := []struct {
		name         string
		placeholders Placeholders
		want         string
		want1        []any
	}{
		{
			name:         "index",
			placeholders: PlaceholdersIndexed(":"),
			want:         "select * from test_table",
			want1:        []any{},
		},
		{
			name:         "name",
			placeholders: PlaceholdersNamed(":"),
			want:         "select * from test_table",
			want1:        []any{},
		},
		{
			name:         "question mark",
			placeholders: PlaceholdersQuestionMark(),
			want:         "select * from test_table",
			want1:        []any{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := NewStatement("select * from test_table").WithPlaceholders(tt.placeholders)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
		})
	}
}

func TestStatement_WithSyntax(t *testing.T) {
	tests := []struct {
		name         string
		placeholders Placeholders
		want         string
		want1        []any
	}{
		{
			name:         "index",
			placeholders: PlaceholdersIndexed("$"),
			want:         "select * from test_table where field1 = $1 and field2 = $2 and field3 = $1",
			want1:        []any{"value1", "value2"},
		},
		{
			name:         "name",
			placeholders: PlaceholdersNamed(":"),
			want:         "select * from test_table where field1 = :arg1 and field2 = :arg2 and field3 = :arg1",
			want1: []any{
				sql.Named("arg1", "value1"),
				sql.Named("arg2", "value2"),
			},
		},
		{
			name:         "question mark",
			placeholders: PlaceholdersQuestionMark(),
			want:         "select * from test_table where field1 = ? and field2 = ? and field3 = ?",
			want1:        []any{"value1", "value2", "value1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := NewStatement("select * from test_table where field1 = :arg1 and field2 = :arg2 and field3 = :arg1").
				Arg("arg1", "value1").
				Arg("arg2", "value2").
				WithPlaceholders(tt.placeholders)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
		})
	}
}

func Test_allIndexes(t *testing.T) {
	type args struct {
		s      string
		substr string
	}

	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "not found",
			args: args{
				s:      "select * from test_table",
				substr: ":arg",
			},
			want: []int{},
		},
		{
			name: "found",
			args: args{
				s:      "select * from test_table where field1 = :arg1 and field2 = :arg2",
				substr: ":arg1",
			},
			want: []int{40},
		},
		{
			name: "multiple indexes",
			args: args{
				s:      "select * from test_table where field1 = :arg1 and field2 = :arg1",
				substr: ":arg1",
			},
			want: []int{40, 59},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, allIndexes(tt.args.s, tt.args.substr))
		})
	}
}
