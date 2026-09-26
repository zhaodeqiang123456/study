package leetcode

import (
	"slices"
	"testing"
)

func TestMoveZeroes(t *testing.T) {
	cases := []struct {
		name string
		nums []int
		want []int
	}{
		{"题目给的例子", []int{0, 2, 3, 0, 0, 0}, []int{2, 3, 0, 0, 0, 0}},
		{"nums 为空", []int{}, []int{}},
		{"nums 只有一个元素", []int{0}, []int{0}},
		{"nums 元素都为零", []int{0, 0, 0}, []int{0, 0, 0}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			orig := slices.Clone(c.nums)

			moveZeroes(c.nums)

			if !slices.Equal(c.nums, c.want) {
				t.Errorf("TestMoveZeroes(%v) 之后元素是 %v, 期望元素是 %v",
					orig, c.nums, c.want)
			}
		})
	}
}
