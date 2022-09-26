package libs

import (
	"os"
	"path/filepath"
	"strings"
)

func ReplaceIO(orig_path, in_path, out_path string) string {
	orig_path = Rel(orig_path)
	in_path = Rel(in_path)
	out_path = Rel(out_path)
	cwd, _ := os.Getwd()
	if Rel(cwd) == Rel(in_path) {
		return filepath.Join(out_path, orig_path)
	} else {
		return strings.Replace(orig_path, in_path, out_path, 1)
	}
}
