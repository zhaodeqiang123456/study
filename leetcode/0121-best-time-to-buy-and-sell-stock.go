package leetcode

func maxProfit(prices []int) int {
	var (
		inf       int = int(1e9)
		minprice  int = inf
		maxprofit int = 0
	)

	for _, price := range prices {
		maxprofit = max(price-minprice, maxprofit)
		minprice = min(price, minprice)
	}

	return maxprofit
}
