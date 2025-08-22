package file

import "testing"

func TestSearchLocal(t *testing.T) {
	files := []FileMeta{
		{Name: "Movie.mp4"},
		{Name: "Sample1.txt"},
	}

	results := SearchLocal(files, "Movie")
	if len(results) != 1 {
		t.Errorf("Expected 1 results, got %d", len(results))
	}
	if results[0].Name != "Movie.mp4" {
		t.Errorf("Expected Movie.mp4, got %s", results[0].Name)
	}
}
