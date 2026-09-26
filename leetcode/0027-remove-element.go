package leetcode

func removeElement(nums []int, val int) int {
	var (
		n      int = len(nums)
		insert int = 0
	)

	if n <= 0 {
		return 0
	}

	for _, p1 := range nums {
		if p1 != val {
			nums[insert] = p1
			insert++
		}
	}
	return insert
}
