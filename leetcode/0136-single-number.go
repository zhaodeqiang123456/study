package leetcode

func singleNumber(nums []int) int {
	sum := 0
	for _, num := range nums {
		sum = sum ^ num
	}

	return sum
}
