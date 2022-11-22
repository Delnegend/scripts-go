package main

import (
	"flag"
	"fmt"
	"libs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	LOW         *string
	HIGH        *string
	list_unique *bool
)

func init() {
	LOW = flag.String("l", "LOW", "low directory")
	HIGH = flag.String("h", "HIGH", "high directory")
	list_unique = flag.Bool("unique", false, "list unique m4a songs")
	flag.Parse()
	*LOW = filepath.Clean(*LOW)
	*HIGH = filepath.Clean(*HIGH)
}

func songName(path string) string {
	file_name := filepath.Base(path)
	no_ext := file_name[:len(file_name)-len(filepath.Ext(file_name))]
	no_bitrate := regexp.MustCompile(` *\[\d*\].*`).ReplaceAllString(no_ext, "")
	return filepath.Join(filepath.Dir(path), strings.Trim(no_bitrate, " "))
}

func inArr(elem string, arr []string) bool {
	for _, a := range arr {
		if len(elem) > len(a) {
			continue
		} else if a[:len(elem)] == elem {
			return true
		}
	}
	return false
}

func removeM4aInFlac() {
	for _, file := range libs.ListFiles(*HIGH, []string{".m4a"}, true, false) {
		m4a_has_flac_version := inArr(songName(file), libs.ListFiles(*HIGH, []string{".flac"}, true, false))
		m4a_exist_in_LOW := inArr(strings.Replace(songName(file), *HIGH, *LOW, 1), libs.ListFiles(*LOW, []string{".m4a"}, true, false))
		if m4a_has_flac_version || !m4a_exist_in_LOW {
			fmt.Println(file)
			os.Remove(file)
		}
	}
}

func removeFlacInLow() {
	for _, file := range libs.ListFiles(*LOW, []string{".flac"}, true, false) {
		fmt.Println(file)
		os.Remove(file)
	}
}

func copyUniqueM4aFromLowToHigh() {
	for _, file := range libs.ListFiles(*LOW, []string{".m4a"}, true, false) {
		m4a_exist_in_HIGH := inArr(strings.Replace(songName(file), *LOW, *HIGH, 1), libs.ListFiles(*HIGH, []string{".m4a"}, true, false))
		m4a_has_flac_version := inArr(strings.Replace(songName(file), *LOW, *HIGH, 1), libs.ListFiles(*HIGH, []string{".flac"}, true, false))
		if !m4a_exist_in_HIGH && !m4a_has_flac_version {
			fmt.Println(file)
			if err := libs.Copy(file, strings.Replace(file, *LOW, *HIGH, 1)); err != nil {
				panic(err)
			}
		}
	}
}

func missingM4aOfFlacInLOW() {
	for _, f := range libs.ListFiles(*HIGH, []string{".flac"}, true, false) {
		flac_has_m4a_version := inArr(strings.Replace(songName(f), *HIGH, *LOW, 1), libs.ListFiles(*LOW, []string{".m4a"}, true, false))
		if !flac_has_m4a_version {
			fmt.Println(f)
		}
	}
}

func listUnique() []string {
	var unique []string
	for _, file := range libs.ListFiles(*LOW, []string{".m4a"}, true, false) {
		if !inArr(strings.Replace(songName(file), *LOW, *HIGH, 1), libs.ListFiles(*HIGH, []string{".flac"}, true, false)) {
			unique = append(unique, strings.Replace(songName(file), *LOW, *HIGH, 1))
		}
	}
	return unique
}

func main() {

	if *list_unique {
		fmt.Println("These songs are don't have flac version:")
		for _, song := range listUnique() {
			fmt.Println(song)
		}
		os.Exit(0)
	}

	libs.PrintSign("Removing old unique m4a or m4a of flac in HIGH", "main")
	removeM4aInFlac()

	libs.PrintSign("Removing flac in LOW", "main")
	removeFlacInLow()

	libs.PrintSign("Copying unique m4a from LOW to HIGH", "main")
	copyUniqueM4aFromLowToHigh()

	libs.PrintSign("Missing m4a of flac in LOW", "main")
	missingM4aOfFlacInLOW()

}
