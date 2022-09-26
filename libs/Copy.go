package libs

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func Copy(src, dst string) error {
	// create dst dir if not exists
	dst_dir := filepath.Dir(dst)
	if _, err := os.Stat(dst_dir); os.IsNotExist(err) {
		os.MkdirAll(dst_dir, 0755)
	}
	bytesRead, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(dst, bytesRead, 0644)
	if err != nil {
		return err
	}
	return nil
}
