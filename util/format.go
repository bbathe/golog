package util

import (
	"strings"
	"unicode/utf8"
)

// FormatFrequency returns a string with frequency formatted like 999.999.99
func FormatFrequency(freq string) string {
	// split on current decimal point
	parts := strings.SplitN(freq, ".", 2)
	s1 := parts[0]

	// keep two fractional digits; a missing fraction is zero
	frac := ""
	if len(parts) == 2 {
		frac = parts[1]
	}
	if len(frac) > 2 {
		frac = frac[:2]
	}
	s2 := frac + strings.Repeat("0", 2-len(frac))

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
