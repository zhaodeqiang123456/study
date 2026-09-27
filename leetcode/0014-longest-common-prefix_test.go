package leetcode

import "testing"

func TestLongestCommonPrefix(t *testing.T) {
	cases := []struct {
		name string
		strs []string
		want string
	}{
		{"题目给的例子", []string{"flower", "flow", "flight"}, "fl"},
		{"没有公共前缀", []string{"dog", "racecar", "car"}, ""},
		{"空字符串数组", []string{}, ""},
		{"只有一个字符串", []string{"interview"}, "interview"},
		{"一个字符串是另一个前缀", []string{"ab", "a"}, "a"},
		{"全部相同", []string{"test", "test", "test"}, "test"},
		{"包含空字符串", []string{"", "abc"}, ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := longestCommonPrefix(c.strs); got != c.want {
				t.Errorf("longestCommonPrefix(%v) = %q，期望 %q", c.strs, got, c.want)
			}
		})
	}
}
