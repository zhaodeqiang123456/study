package leetcode

import (
	"slices"
	"testing"
)

func TestMaxProfit(t *testing.T) {
	cases := []struct {
		name string
		prices []int
		want int
	}{
		{"题目给的例子", []int{7, 1, 5, 3, 6, 4}, 5},
		{"单调下降", []int{7, 6, 4, 3, 1}, 0},
		{"单调上升", []int{1, 2, 3, 4, 5}, 4},
		{"只有一天", []int{5}, 0},
		{"两天无收益", []int{2, 1}, 0},
		{"低点在中间", []int{3, 2, 6, 1, 4}, 4},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			orig := slices.Clone(c.prices)

			got := maxProfit(c.prices)

			if got != c.want {
				t.Fatalf("maxProfit(%v) = %d，期望 %d", orig, got, c.want)
			}
		})
	}
}
