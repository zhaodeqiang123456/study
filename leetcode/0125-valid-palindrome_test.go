package leetcode

import "testing"

func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "字母数字混合回文", input: "A man, a plan, a canal: Panama", want: true},
		{name: "非回文", input: "race a car", want: false},
		{name: "只有标点", input: ",.!", want: true},
		{name: "单个字符", input: "a", want: true},
		{name: "空字符串", input: "", want: true},
		{name: "数字回文", input: "1a1", want: true},
		{name: "Unicode 字母", input: "上海自来水来自海上", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPalindrome(tt.input); got != tt.want {
				t.Fatalf("isPalindrome(%q) = %t, want %t", tt.input, got, tt.want)
			}
		})
	}
}
