package leetcode

import (
	"reflect"
	"testing"
)

func TestTwoSum(t *testing.T) {
	tests := []struct {
		name   string
		nums   []int
		target int
		want   []int
	}{
		{name: "题目示例", nums: []int{2, 7, 11, 15}, target: 9, want: []int{0, 1}},
		{name: "重复元素", nums: []int{3, 3}, target: 6, want: []int{0, 1}},
		{name: "负数", nums: []int{-3, 4, 3, 90}, target: 0, want: []int{0, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := twoSum(tt.nums, tt.target)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("twoSum(%v, %d) = %v, want %v", tt.nums, tt.target, got, tt.want)
			}
		})
	}
}
