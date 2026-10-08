package leetcode

import "unicode"

func isPalindrome(s string) bool {
	// 新建一个切片，长度为0， 能力为s的长度
	good := make([]rune, 0, len(s))
	for _, ch := range s {
		// 遍历s字符串，如果是字母或者数组则添加进good切片
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) {
			good = append(good, unicode.ToLower(ch))
		}
	}
	// 双指针首尾遍历
	for i, j := 0, len(good)-1; i < j; i, j = i+1, j-1 {
		if good[i] != good[j] {
			return false
		}
	}
	return true
}
