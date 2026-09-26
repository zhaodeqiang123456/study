package leetcode

// 空间复杂度为O(1),要求原地修改--双指针
// 初始插入位置为1，跳过了第一步，省去一些边界条件处理
func removeDuplicates(nums []int) int {
	if len(nums) <= 1 {
		return len(nums)
	}

	insert_index := 1
	prior := nums[0]
	for _, value := range nums {

		if prior != value {
			nums[insert_index] = value
			prior = value
			insert_index++
		}
	}

	return insert_index
}
