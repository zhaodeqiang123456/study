package leetcode

import (
	"slices"
	"testing"
)

func TestSingleNumber(t *testing.T) {
	cases := []struct {
		name string
		nums []int
		want int
	}{
		{"题目给的例子", []int{2, 2, 1}, 1},
		{"唯一元素在开头", []int{4, 1, 1, 2, 2}, 4},
		{"只有一个元素", []int{1}, 1},
		{"负数", []int{-1, -1, -2}, -2},
		{"唯一元素在末尾", []int{7, 3, 7, 3, 9}, 9},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			orig := slices.Clone(c.nums)

			got := singleNumber(c.nums)

			if got != c.want {
				t.Fatalf("singleNumber(%v) = %d，期望 %d", orig, got, c.want)
			}
		})
	}
}
