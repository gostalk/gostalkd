package utils

import (
	"regexp"
	"strings"
)

var (
	// replaceCharReg is the regular expression object for replacing chars in key.
	// It is used for function EqualFoldWithoutChars.
	replaceCharReg, _ = regexp.Compile(`[\-\.\_\s]+`)
)

// IsLetterUpper checks whether the given byte b is in upper case.
func IsLetterUpper(b byte) bool {
	if b >= byte('A') && b <= byte('Z') {
		return true
	}
	return false
}

// IsLetterLower checks whether the given byte b is in lower case.
func IsLetterLower(b byte) bool {
	if b >= byte('a') && b <= byte('z') {
		return true
	}
	return false
}

// IsLetter checks whether the given byte b is a letter.
func IsLetter(b byte) bool {
	return IsLetterUpper(b) || IsLetterLower(b)
}

// IsNumeric checks whether the given string s is numeric.
// Note that float string like "123.456" is also numeric.
func IsNumeric(s string) bool {
	length := len(s)
	if length == 0 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] == '-' && i == 0 {
			continue
		}
		if s[i] == '.' {
			if i > 0 && i < len(s)-1 {
				continue
			} else {
				return false
			}
		}
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// UcFirst returns a copy of the string s with the first letter mapped to its upper case.
func UcFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	if IsLetterLower(s[0]) {
		return string(s[0]-32) + s[1:]
	}
	return s
}

// ReplaceByMap returns a copy of <origin>,
// which is replaced by a map in unordered way, case-sensitively.
func ReplaceByMap(origin string, replaces map[string]string) string {
	for k, v := range replaces {
		origin = strings.Replace(origin, k, v, -1)
	}
	return origin
}

// EqualFoldWithoutChars checks string <s1> and <s2> equal case-insensitively,
// with/without chars '-'/'_'/'.'/' '.
func EqualFoldWithoutChars(s1, s2 string) bool {
	return strings.EqualFold(
		replaceCharReg.ReplaceAllString(s1, ""),
		replaceCharReg.ReplaceAllString(s2, ""),
	)
}
