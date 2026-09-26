package leetcode

func plusOne(digits []int) []int {
	var (
		plus int = 1
		n    int = len(digits)
	)

	for index := n - 1; index >= 0; index-- {
		temp := plus + digits[index]

		if temp >= 10 {
			digits[index] = 0
			plus = 1
		} else {
			digits[index] = temp
			plus = 0
		}

	}

	if plus == 1 {
		digits = append([]int{1}, digits...)
	}

	return digits
}
