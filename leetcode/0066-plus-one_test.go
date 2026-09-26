package leetcode

import (
	"slices"
	"testing"
)

func TestPlusOne(t *testing.T) {
	cases := []struct {
		name string
		nums []int
		want []int
	}{
		{"题目给的例子", []int{1, 2, 3}, []int{1, 2, 4}},
		{"中间进位", []int{1, 2, 9}, []int{1, 3, 0}},
		{"连续进位", []int{9, 9}, []int{1, 0, 0}},
		{"最高位有进位", []int{9}, []int{1, 0}},
		{"单个零", []int{0}, []int{1}},
		{"末位不是九", []int{4, 3, 2, 1}, []int{4, 3, 2, 2}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			before := slices.Clone(c.nums)

			got := plusOne(c.nums)

			if !slices.Equal(got, c.want) {
				t.Errorf("PlusOne(nums1=%v) 之后 nums=%v, 期望 %v",
					before, got, c.want)
			}
		})
	}
}
