package libs

func InArr(elem string, arr []string) string {
	for _, v := range arr {
		if len(elem) < len(v) {
			if elem == v[:len(elem)] {
				return v
			}
		} else {
			if elem == v {
				return v
			}
		}
	}
	return ""
}
