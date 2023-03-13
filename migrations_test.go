package dejavu

import (
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMigrations(t *testing.T) FsMigrations {
	t.Helper()

	root, err := fs.Sub(os.DirFS("testdata"), "db")
	require.NoError(t, err)

	return FsMigrations{fs: root}
}

func TestNewMigrations(t *testing.T) {
	migs := newTestMigrations(t)

	assert.NotNil(t, migs)
}

func Test_fsMigrations_List(t *testing.T) {
	migs := newTestMigrations(t)
	names, err := migs.List()

	require.NoError(t, err)
	assert.Equal(t, []string{"2023-01-01/01_create_country_table.sql", "2023-01-01/02_populate_country_table.sql"}, names)
}

func Test_fsMigrations_Content(t *testing.T) {
	migs := newTestMigrations(t)

	type args struct {
		name string
	}

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "error",
			args:    args{name: "unknown"},
			wantErr: assert.Error,
		},
		{
			name:    "ok",
			args:    args{name: "2023-01-01/01_create_country_table.sql"},
			wantErr: assert.NoError,
			want:    "create table country\n(\n    name                     varchar(64) not null,\n    alpha2                   char(2)     not null,\n    alpha3                   char(3)     not null,\n    country_code             char(3)     not null,\n    iso_3166_2               char(13)    not null,\n    region                   varchar(8),\n    sub_region               varchar(32),\n    intermediate_region      varchar(16),\n    region_code              varchar(3),\n    sub_region_code          varchar(3),\n    intermediate_region_code varchar(3),\n    constraint country_pk primary key (name)\n);\n", //nolint:lll
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := migs.Content(tt.args.name)
			if !tt.wantErr(t, err) {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
