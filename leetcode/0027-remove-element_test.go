package leetcode

import (
	"slices"
	"testing"
)

func TestRemoveElement(t *testing.T) {
	cases := []struct {
		name     string
		nums     []int
		val      int
		wantK    int
		wantNums []int
	}{
		{"题目给的例子", []int{1, 1, 2}, 2, 2, []int{1, 1}},
		{"空数组", []int{}, 0, 0, []int{}},
		{"只有一个元素", []int{1}, 1, 0, []int{}},
		{"LeetCode 的长例子", []int{0, 0, 1, 1, 1, 2, 2, 3, 3, 4}, 3, 8, []int{0, 0, 1, 1, 1, 2, 2, 4}},
		{"整个数组都是目标值", []int{1, 1, 1, 1}, 1, 0, []int{}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			orig := slices.Clone(c.nums)

			got := removeElement(c.nums, c.val)

			if got != c.wantK {
				t.Fatalf("RemoveElement(%v) 返回 k=%d，期望 %d", orig, got, c.wantK)
			}
			if got > len(c.nums) {
				t.Fatalf("返回的 k=%d 超过了数组长度 %d", got, len(c.nums))
			}

			if !slices.Equal(c.nums[:got], c.wantNums) {
				t.Errorf("RemoveElement(%v) 之后前 %d 个元素是 %v，期望 %v",
					orig, got, c.nums[:got], c.wantNums)
			}
		})
	}
}
