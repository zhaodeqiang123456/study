package leetcode

func majorityElement(nums []int) int {

	var (
		count    int = 0
		majority int = 10e9 + 1
	)

	for _, num := range nums {
		if majority == num {
			count++
		} else {
			count--
		}

		if count < 0 {
			majority = num
			count = 1
		}
	}

	return majority
}
