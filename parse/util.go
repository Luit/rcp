package parse

import "errors"

// itoa fills dst with ASCII representation of val, and returns the number of
// bytes used. dst should have enough room for a full int64 (20 byte len).
// itoa can leave dst dirty beyond the number of bytes used for the final
// value.
func itoa(dst []byte, val int64) int {
	if val == 0 {
		dst[0] = '0'
		return 1
	}
	i := 0
	if val < 0 {
		dst[i] = '-'
		val = -val
		i++
	}
	l := 0
	for val > 0 {
		dst[19-l], val = '0'+byte(val%10), val/10
		l++
	}
	copy(dst[i:i+l], dst[20-l:20])
	return i + l
}

// atoi parses an integer from an ASCII representation in src, returning an
// error if src contains anything other than a minus in the first position, or
// digits, or if src is empty.
func atoi(src []byte) (v int64, err error) {
	if src == nil || len(src) < 1 {
		return 0, errors.New("parse.atoi: no data")
	}
	neg := false
	for i := 0; i < len(src); i += 1 {
		switch {
		case i == 0 && src[i] == '-':
			neg = true
		case '0' <= src[i] && src[i] <= '9':
			v = v*10 + int64(src[i]-'0')
		default:
			return v, errors.New("parse.atoi: invalid number")
		}
	}
	if neg {
		v = -v
	}
	return
}
