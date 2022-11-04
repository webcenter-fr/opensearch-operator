package helper

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kr/pretty"
)

func Diff(expected, current any) string {
	return cmp.Diff(expected, current, cmpopts.SortSlices(func(x, y any) bool {
		return pretty.Sprint(x) < pretty.Sprint(y)
	}))
}