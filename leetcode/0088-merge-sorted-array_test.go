package leetcode

import (
	"slices"
	"testing"
)

func TestMerge(t *testing.T) {
	cases := []struct {
		name  string
		nums1 []int
		m     int
		nums2 []int
		n     int
		want  []int
	}{
		{"题目给的例子", []int{1, 2, 3, 0, 0, 0}, 3, []int{2, 5, 6}, 3, []int{1, 2, 2, 3, 5, 6}},
		{"nums2 为空", []int{1}, 1, []int{}, 0, []int{1}},
		{"nums1 有效部分为空", []int{0}, 0, []int{1}, 1, []int{1}},
		{"nums2 全部小于 nums1", []int{4, 5, 6, 0, 0, 0}, 3, []int{1, 2, 3}, 3, []int{1, 2, 3, 4, 5, 6}},
		{"nums2 全部大于 nums1", []int{1, 2, 3, 0, 0, 0}, 3, []int{4, 5, 6}, 3, []int{1, 2, 3, 4, 5, 6}},
		{"有相等的元素", []int{1, 1, 2, 0, 0}, 3, []int{1, 1}, 2, []int{1, 1, 1, 1, 2}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			before := slices.Clone(c.nums1)

			merge(c.nums1, c.m, c.nums2, c.n)

			if !slices.Equal(c.nums1, c.want) {
				t.Errorf("merge(nums1=%v, m=%d, nums2=%v, n=%d) 之后 nums1=%v，期望 %v",
					before, c.m, c.nums2, c.n, c.nums1, c.want)
			}
		})
	}
}
