package comp

import (
	"testing"
)

func TestParse(t *testing.T) {
	mods, funcs, err := parse("testdata/func.go")
	if err != nil {
		t.Fatal(err)
	}
	t.Fatal(funcs)
	t.Fatal(mods)
}
