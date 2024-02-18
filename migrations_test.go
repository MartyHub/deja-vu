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

	return FsMigrations{
		fs: root,
	}
}

func TestNewMigrations(t *testing.T) {
	migs := newTestMigrations(t)

	assert.NotNil(t, migs)
}

func Test_fsMigrations_List(t *testing.T) {
	migs := newTestMigrations(t)
	mds, err := migs.List("postgresql")

	require.NoError(t, err)
	assert.Len(t, mds, 3)
	assert.Equal(t, "2023-01-01/01_create_country_table.postgresql.sql", mds[0])
	assert.Equal(t, "2023-01-01/02_create_country_index.sql", mds[1])
	assert.Equal(t, "2023-01-01/03_populate_country_table.sql", mds[2])
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
			args:    args{name: "2023-01-01/01_create_country_table.mysql.sql"},
			wantErr: assert.NoError,
			want: `create table country
(
    name                     nvarchar(64) not null,
    alpha2                   char(2)  not null,
    alpha3                   char(3)  not null,
    country_code             char(3)  not null,
    iso_3166_2               char(13) not null,
    region                   varchar(8),
    sub_region               varchar(32),
    intermediate_region      varchar(16),
    region_code              varchar(3),
    sub_region_code          varchar(3),
    intermediate_region_code varchar(3),
    constraint country_pk primary key (name)
);
`,
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
