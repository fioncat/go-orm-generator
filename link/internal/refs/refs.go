package refs

import (
	"path/filepath"

	"github.com/fioncat/go-gendb/compile/sql"
)

var refs = make(map[string]interface{})

func Import(oriPath, path, mode string) (interface{}, error) {
	dir := filepath.Dir(oriPath)
	path = filepath.Join(dir, path)
	v := refs[path]
	if v != nil {
		return v, nil
	}
	switch mode {
	case "sql":
		file, err := sql.ReadFile(path)
		if err != nil {
			return nil, err
		}
		v = file

	case "go":
		file, err := sql.ReadFile(path)
		if err != nil {
			return nil, err
		}
		v = file

	default:
		panic("unknown link mode: " + mode)
	}
	refs[path] = v

	return v, nil
}
