package libs

import (
	"fmt"
)

func ReportResult(files_len int, orig_size int, new_size int, start_time int64, end_time int64) string {
	h_size_orig, h_size_new := Hsize(orig_size), Hsize(new_size)
	ratio := fmt.Sprintf("%.2f%%", float64(new_size)/float64(orig_size)*100)
	time_taken := Htime(float64(end_time-start_time))
	var report string
	if orig_size > 0 && new_size > 0 {
		report = fmt.Sprintf("%d file(s) processed in %s | %s -> %s (%s)", files_len, time_taken, h_size_orig, h_size_new, ratio)
	} else {
		report = fmt.Sprintf("%d file processed", files_len)
	}
	return report
}