package comp

import (
	"testing"
)

func TestParse(t *testing.T) {
	err := parse("testdata/sample.go")
	if err != nil {
		t.Fatal(err)
	}
	t.Fail()
}
