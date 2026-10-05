package main

func rotateRunes(s string, shift int) string {
	r := []rune(s)
	n := len(r)
	if n <= 1 || shift == 0 {
		return string(r)
	}

	res := make([]rune, len(r))
	shift = shift % n
	if shift < 0 {
		shift += n
	}

	copy(res, r[shift:])
	copy(res[n-shift:], r[:shift])
	return string(res)
}
