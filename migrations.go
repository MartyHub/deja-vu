package dejavu

import (
	"fmt"
	"io/fs"
	"strings"
)

type Migrations interface {
	fmt.Stringer

	List() ([]string, error)
	Content(name string) (string, error)
}

type FsMigrations struct {
	fs fs.FS
}

func (m FsMigrations) List() ([]string, error) {
	result := make([]string, 0)
	err := fs.WalkDir(m.fs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			result = append(result, path)
		}

		return nil
	})

	return result, err
}

func (m FsMigrations) Content(name string) (string, error) {
	var err error

	fsys := m.fs
	paths := strings.Split(name, "/")

	for _, path := range paths[:len(paths)-1] {
		fsys, err = fs.Sub(fsys, path)
		if err != nil {
			return "", err
		}
	}

	data, err := fs.ReadFile(fsys, paths[len(paths)-1])
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (m FsMigrations) String() string {
	return fmt.Sprintf("%v", m.fs)
}
