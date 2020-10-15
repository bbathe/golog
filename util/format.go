package util

import (
	"strings"
	"unicode/utf8"
)

// FormatFrequency returns a string with frequency formatted like on my IC-7300
func FormatFrequency(freq string) string {
	// split on current decimal point
	parts := strings.Split(freq, ".")
	s1 := parts[0]

	// pad last part out to 2 digits
	s2 := parts[1] + strings.Repeat("0", 2-len(parts[1]))

	// figure out if we need to do anything (first part more than 3 digits)
	startOffset := 0
	const groupLen = 3
	groups := (len(s1) - startOffset - 1) / groupLen

	if groups == 0 {
		// recombine with formatted second part
		return s1 + "." + s2
	}

	sep := '.'
	sepLen := utf8.RuneLen(sep)
	sepBytes := make([]byte, sepLen)
	_ = utf8.EncodeRune(sepBytes, sep)

	buf := make([]byte, groups*(groupLen+sepLen)+len(s1)-(groups*groupLen))

	// move over in groups of 3, adding seperator
	startOffset += groupLen
	p := len(s1)
	q := len(buf)
	for p > startOffset {
		p -= groupLen
		q -= groupLen
		copy(buf[q:q+groupLen], s1[p:])
		q -= sepLen
		copy(buf[q:], sepBytes)
	}
	if q > 0 {
		copy(buf[:q], s1)
	}

	// recombine with formatted second part
	return string(buf) + "." + s2
}
