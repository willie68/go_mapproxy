package extstrgutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitMultiValueParam(t *testing.T) {
	ast := assert.New(t)
	tt := []struct {
		name string
		line string
		exp  []string
	}{
		{"single", "value", []string{"value"}},
		{"space", "value1 value2", []string{"value1", "value2"}},
		{"comma", "value1,value2", []string{"value1", "value2"}},
		{"semicolon", "value1;value2", []string{"value1", "value2"}},
		{"mixed", "value1, value2;value3 value4", []string{"value1", "value2", "value3", "value4"}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			res := SplitMultiValueParam(tc.line)
			ast.Equal(tc.exp, res)
		})
	}
}
