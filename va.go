package main

import (
	"fmt"
	"math"
	"time"
)

// VA simulates value-averaging investment and returns:
// - total number of shares owned at the end
// - total money spent to purchase them
// - lowest single investment money required
// - highest single investment money required
//
// Inputs:
// - startTime is time of the first investment
// - investmentInterval length of time between investments
// - numOfSteps number of total investment steps
// - expectedValueIncreaseEveryStep is expected increase of portfolio value (in money) every step
// - purchaseFixedFee fixed amount of money to paid on every purchase (poplatky za nakupni transakci)
// - purchaseNPercentFee is floating-point (0=0%, 1.0=100%) pecent fee paid from money spent on each purchase (vstupni poplatky v %)
// - intervalValueNPercentFee is floating-point (0=0%, 1.0=100%) pecent fee paid from the account value
//   at the end of every investment period (zlomek rocniho spravcovkeho procentniho poplatku)
func VA(points pricePoints, startTime time.Time, investmentInterval time.Duration, numOfSteps int, expectedValueIncreaseEveryStep float64,
	purchaseFixedFee, purchaseNPercentFee, intervalValueNPercentFee float64) (totalStockPcs int, totalMoneySpent, lowestInvestment, highestInvestment float64) {
	lowestInvestment = math.Pow(10, 300) // initialize lowest investment to some crazy-hight value initially

	for step := 0; step < numOfSteps; step++ {
		price, err := points.PriceAt(startTime.Add(investmentInterval*time.Duration(step)), investmentInterval/2)
		if err != nil {
			fmt.Printf("cannot get price - %s\n", err.Error())
			continue
		}

		currentPortfolioValueValue := float64(totalStockPcs) * price                 // portfolio monteray value now
		expectedPortfolioValue := float64(step+1) * expectedValueIncreaseEveryStep   // expected portfolio value at this point
		expectedVsCurrentDiff := expectedPortfolioValue - currentPortfolioValueValue // portfolio value monetary difference

		buyOrSellPcs := math.Round(expectedVsCurrentDiff / price) // number of stock to seel / buy this step
		spentThisStep := buyOrSellPcs * price                     // amount of money spent this step
		totalStockPcs += int(buyOrSellPcs)                        // amount of stock owned now

		// apply fees
		if step > 0 {
			spentThisStep += purchaseFixedFee + purchaseNPercentFee*spentThisStep + float64(totalStockPcs)*price*intervalValueNPercentFee
		} else {
			spentThisStep += purchaseFixedFee + purchaseNPercentFee*spentThisStep
		}

		totalMoneySpent += spentThisStep

		// update min/max
		if spentThisStep > highestInvestment {
			highestInvestment = spentThisStep
		}
		if spentThisStep < lowestInvestment {
			lowestInvestment = spentThisStep
		}
	}

	return totalStockPcs, totalMoneySpent, lowestInvestment, highestInvestment
}
