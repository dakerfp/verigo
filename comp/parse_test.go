package comp

import (
	"testing"
)

func TestParse(t *testing.T) {
	modules, funcs, err := parse("testdata/func.go")
	if err != nil {
		t.Fatal(err)
	}
	if len(modules) != 0 {
		t.Fatal(len(modules))
	}
	if len(funcs) != 1 {
		t.Fatal(len(funcs))
	}
}
