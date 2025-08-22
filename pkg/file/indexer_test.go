package file

import "testing"

func TestIndexDirectory(t *testing.T) {
	files, err := IndexDirectory("testdata")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("Expected Files, Got None")
	}
	first := files[0]
	if first.FileHash == "" {
		t.Error("Expected File Hash, Got Empty")
	}
	if len(first.ChunkHash) == 0 {
		t.Error("Expected Chunk Hashes, Got None")
	}
}
