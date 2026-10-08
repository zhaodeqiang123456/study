package main

import (
	"testing"
)

// 表驱动测试
func TestSliceCopy(t *testing.T) {
	tests := []struct {
		name      string
		build     func([]int) []int
		wantShare bool
	}{
		{
			name: "assignment",
			build: func(src []int) []int {
				return src
			},
			wantShare: true,
		},
		{
			name: "copy",
			build: func(src []int) []int {
				dst := make([]int, len(src))
				copy(dst, src)
				return dst
			},
			wantShare: false,
		},
		{
			name: "append",
			build: func(src []int) []int {
				return append([]int(nil), src...)
			},
			wantShare: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := []int{10, 20}
			dst := tt.build(src)
			dst[0] = 99

			shared := src[0] == 99
			if shared != tt.wantShare {
				t.Fatalf("shared=%v, want %v", shared, tt.wantShare)
			} else {
				t.Logf("shared=%v, want %v", shared, tt.wantShare)
			}
		})
	}
}

// 表驱动测试就是把多组输入、期望输出和测试名称放进一个切片，然后统一执行。
// 表驱动测试模板
func TestMinDistinctWindow(t *testing.T) {
	tests := []struct {
		name string
		nums []int
		k    int
		want int
	}{
		{"empty", nil, 2, 0},
		{"zero k", []int{1, 2}, 0, 0},
		{"one distinct", []int{1, 1, 1}, 1, 1},
		{"not enough distinct", []int{1, 2, 3}, 4, 0},
		{"normal", []int{1, 2, 3}, 2, 2},
		{"shrink then expand again", []int{1, 2, 1, 3, 2}, 3, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := minDistinctWindow(tt.nums, tt.k)
			if got != tt.want {
				t.Fatalf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func minDistinctWindow(nums []int, k int) int {
	// 窗口状态， 满足条件，收缩条件
	if k <= 0 || len(nums) == 0 {
		return 0
	}

	var (
		counts   = make(map[int]int)
		left     = 0
		minlen   = len(nums) + 1
		right    = 0
		distinct = 0
	)
	// {"special case", []int{1, 1, 3}, 2, 2},
	for right < len(nums) {

		counts[nums[right]]++

		if counts[nums[right]] == 1 {
			distinct++
		}

		for distinct >= k {
			if right-left+1 < minlen {
				minlen = right - left + 1
			}
			counts[nums[left]]--
			if counts[nums[left]] == 0 {
				distinct--
			}
			left++
		}

		right++
	}

	if minlen == len(nums)+1 {
		return 0
	}

	return minlen
}
