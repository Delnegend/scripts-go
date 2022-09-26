package libs

import (
	"os"
	"path/filepath"
)

func Rel(path string) string {
	abs_path, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	cwd, _ := os.Getwd()
	rel_path, err2 := filepath.Rel(cwd, abs_path)
	if err2 != nil {
		panic(err2)
	}
	return filepath.Clean(rel_path)
}