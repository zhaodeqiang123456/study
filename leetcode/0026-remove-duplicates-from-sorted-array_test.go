package leetcode

import (
	"slices"
	"testing"
)

func TestRemoveDuplicates(t *testing.T) {
	cases := []struct {
		name     string
		nums     []int
		wantK    int
		wantNums []int // 前 k 个元素应当长什么样
	}{
		{"题目给的例子", []int{1, 1, 2}, 2, []int{1, 2}},
		{"空数组", []int{}, 0, []int{}},
		{"只有一个元素", []int{1}, 1, []int{1}},
		{"全都不重复", []int{1, 2, 3, 4, 5, 6}, 6, []int{1, 2, 3, 4, 5, 6}},
		{"全都重复", []int{2, 2, 2}, 1, []int{2}},
		{"LeetCode 的长例子", []int{0, 0, 1, 1, 1, 2, 2, 3, 3, 4}, 5, []int{0, 1, 2, 3, 4}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			orig := slices.Clone(c.nums)

			got := removeDuplicates(c.nums)

			if got != c.wantK {
				t.Fatalf("removeDuplicates(%v) 返回 k=%d，期望 %d", orig, got, c.wantK)
			}
			if got > len(c.nums) {
				t.Fatalf("返回的 k=%d 超过了数组长度 %d", got, len(c.nums))
			}
			if !slices.Equal(c.nums[:got], c.wantNums) {
				t.Errorf("removeDuplicates(%v) 之后前 %d 个元素是 %v，期望 %v",
					orig, got, c.nums[:got], c.wantNums)
			}
		})
	}
}
