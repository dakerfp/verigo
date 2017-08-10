package comp

import (
	"testing"
)

func TestParse(t *testing.T) {
	_, _, err := parse("testdata/func.go")
	if err != nil {
		t.Fatal(err)
	}
}
