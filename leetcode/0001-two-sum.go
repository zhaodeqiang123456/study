package leetcode

import (
	"log"
)

func twoSum(nums []int, target int) []int {
	hashTable := map[int]int{}
	for i, x := range nums {
		if p, ok := hashTable[target-x]; ok {
			return []int{p, i}
		}
		hashTable[x] = i
	}
	return nil
}

func main() {
	var nums []int = []int{3, 2, 4}

	log.Println(twoSum(nums, 6))
}
