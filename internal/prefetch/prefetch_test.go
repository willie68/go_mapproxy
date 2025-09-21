package prefetch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplit(t *testing.T) {
	ast := assert.New(t)
	tt := []struct {
		value string
		split []string
	}{
		{
			value: "1",
			split: []string{"1"},
		},
		{
			value: "1,2",
			split: []string{"1", "2"},
		},
		{
			value: "1,2,3",
			split: []string{"1", "2", "3"},
		},
		{
			value: "1, 2",
			split: []string{"1", "2"},
		},
		{
			value: " 1 , 2 ",
			split: []string{"1", "2"},
		},
		{
			value: "1 2",
			split: []string{"1", "2"},
		},
		{
			value: "1;2",
			split: []string{"1", "2"},
		},
		{
			value: "1; 2, 3",
			split: []string{"1", "2", "3"},
		},
		{
			value: "1.2",
			split: []string{"1.2"},
		},
	}

	for _, td := range tt {
		sd := SplitMultiValueParam(td.value)
		ast.EqualValues(td.split, sd)
	}
}
