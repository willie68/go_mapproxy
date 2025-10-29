package extstrgutils

import "strings"

// SplitMultiValueParam splits a string into multiple values using space, comma or semicolon as separator
func SplitMultiValueParam(value string) []string {
	return strings.FieldsFunc(value, func(r rune) bool {
		return r == ' ' || r == ',' || r == ';'
	})
}
