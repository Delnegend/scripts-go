package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type List struct {
	FILE  []string `json:"FILE"`
	FOLDER []string `json:"FOLDER"`
}

type Config struct {
	HIGH_PATH  string `json:"HIGH_PATH"`
	LOW_PATH   string `json:"LOW_PATH"`
	TRASH_PATH string `json:"TRASH_PATH"`
	NO_LRC     List   `json:"NO_LRC"`
	NO_FLAC    List   `json:"NO_FLAC"`
}

func normpath(path string) string {
	return strings.Replace(filepath.Clean(path), "\\", "/", -1)
}

func listFiles(path string, ext []string) []string {
	var files []string
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}
		for _, e := range ext {
			if strings.HasSuffix(f.Name(), e) {
				files = append(files, path)
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files
}

func loadConfig(path string) Config {
	// parse json file to struct and return
	content, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer content.Close()
	byteValue, _ := ioutil.ReadAll(content)
	var config Config
	json.Unmarshal(byteValue, &config)
	return config
}

func cleanFiles(c Config, FLAC_IN_HIGH []string) {
	// Create an exact copy of the FLAC_IN_HIGH list except replace .flac with .m4a
	m4a_in_high := make([]string, len(FLAC_IN_HIGH))
	for i, f := range FLAC_IN_HIGH {
		m4a_in_high[i] = strings.Replace(f, ".flac", ".m4a", 1)
	}
	for _, f := range m4a_in_high {
		// if f is exists then remove it
		if _, err := os.Stat(f); !os.IsNotExist(err) {
			fmt.Printf("- Removing m4a version of flac: %s", f)
			os.Rename(f, c.TRASH_PATH + "/" + filepath.Base(f))
		}
	}

	// Remove flac in LOW folder
	for _, f := range listFiles(c.LOW_PATH, []string{".flac"}) {
		fmt.Printf("- Removing flac in LOW folder: %s", f)
		os.Rename(f, c.TRASH_PATH + "/" + filepath.Base(f))
	}

	// Remove m4a in LOW not in HIGH
	m4a_in_high = listFiles(c.HIGH_PATH, []string{".m4a"})
	for _, f := range listFiles(c.LOW_PATH, []string{".m4a"}) {
		have_m4a := false
		have_flac := false
		for _, m := range m4a_in_high {
			if f == m {
				have_m4a = true
			}
		}
		for _, flac := range FLAC_IN_HIGH {
			if strings.Replace(f, ".m4a", ".flac", 1) == flac {
				have_flac = true
			}
		}
		if !have_m4a && !have_flac {
			fmt.Printf("- Removing m4a in LOW not in HIGH or having flac version: %s", f)
			os.Rename(f, c.TRASH_PATH + "/" + filepath.Base(f))
		}
	}

func main() {
	config := loadConfig("config.json")
}