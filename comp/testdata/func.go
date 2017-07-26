package testdata

func fullAdder(a, b, cin bool) (r, cout bool) {
	if a && b {
		return true, cin
	} else if a || b {
		return !cin, cin
	} else {
		return cin, false
	}
}
