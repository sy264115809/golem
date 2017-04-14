package slice_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sy264115809/golem/utils/slice"
)

func TestReverse(t *testing.T) {
	strOrigin, strExpected := "string", "gnirts"
	slice.Reverse(&strOrigin)
	assert.Equal(t, strExpected, strOrigin)

	arrayOrigin, arrayExpected := [5]int{1, 2, 3, 4, 5}, [5]int{5, 4, 3, 2, 1}
	slice.Reverse(&arrayOrigin)
	assert.Equal(t, arrayExpected, arrayOrigin)

	sliceOrigin, sliceExpected := []int{1, 4, 2, 3, 5}, []int{5, 3, 2, 4, 1}
	slice.Reverse(&sliceOrigin)
	assert.Equal(t, sliceExpected, sliceOrigin)
}
