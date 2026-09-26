package leetcode

func moveZeroes(nums []int) {
	var n int = len(nums)
	if n <= 1 {
		return
	}
	insert := 0
	for _, p1 := range nums {
		if p1 != 0 {
			nums[insert] = p1
			insert++
		}
	}
	clear(nums[insert:])
}
