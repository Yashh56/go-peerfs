package file

import (
	"os"
	"testing"
)

func TestChunker(t *testing.T) {

	f, err := os.Open("testdata/sample.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	chunks, err := Chunk(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(chunks) == 0 {
		t.Error("Expected Chunks, Got None")
	}
	all := []byte{}
	for _, c := range chunks {
		all = append(all, c...)
	}
	original, _ := os.ReadFile("testdata/sample.txt")
	if string(all) != string(original) {
		t.Error("Reassembled Chunks do not match original File")
	}

}
