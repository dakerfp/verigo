package meta

import (
	"testing"
)

func TestTrue(t *testing.T) {
	if !True(true) {
		t.Fatal(true)
	}

	if True(false) {
		t.Fatal(false)
	}

	if !True(1) {
		t.Fatal(1)
	}

	if !True(-1) {
		t.Fatal(-1)
	}

	if True(0) {
		t.Fatal(0)
	}
}

func TestWidth(t *testing.T) {
	if w := Width(true); w != 1 {
		t.Fatal(w)
	}

	if w := Width(int(42)); w != 64 {
		t.Fatal(w)
	}
}

func TestCat(t *testing.T) {
	if l := Cat(true, false, true, true); len(l) != 4 {
		t.Fatal(l)
	}

	if l := Cat(); len(l) != 0 {
		t.Fatal(l)
	}

	if l := Cat(42, true, true); len(l) != 64+2 {
		t.Fatal(l)
	}
}
