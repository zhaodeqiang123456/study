package leetcode

import "testing"

func TestMajorityElement(t *testing.T) {
	tests := []struct {
		name string
		nums []int
		want int
	}{
		{name: "题目示例", nums: []int{3, 2, 3}, want: 3},
		{name: "多数元素分散出现", nums: []int{2, 2, 1, 1, 1, 2, 2}, want: 2},
		{name: "单个元素", nums: []int{7}, want: 7},
		{name: "全部相同", nums: []int{-4, -4, -4, -4}, want: -4},
		{name: "多数元素在末尾", nums: []int{1, 2, 2, 2}, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := majorityElement(tt.nums); got != tt.want {
				t.Fatalf("majorityElement(%v) = %d, want %d", tt.nums, got, tt.want)
			}
		})
	}
}
