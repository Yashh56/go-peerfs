package file

import "strings"

func SearchLocal(files []FileMeta, query string) []FileMeta {
	var results []FileMeta
	for _, f := range files {
		if containsIgnoreCase(f.Name, query) {
			results = append(results, f)
		}
	}
	return results
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
