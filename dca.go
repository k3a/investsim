package main

import (
	"fmt"
	"math"
	"time"
)

// DCA simulates dollar-cost averaging and returns number of stocks (items) at the end and total money spent
// Inputs:
// - startTime is time of the first investment
// - investmentInterval length of time between investments
// - numOfSteps number of total investment steps
// - toInvestEveryStep is amount to invest (spent) every investement period
// - purchaseFixedFee fixed amount of money to paid on every purchase (poplatky za nakupni transakci)
// - purchaseNPercentFee is floating-point (0=0%, 1.0=100%) pecent fee paid from money spent on each purchase (vstupni poplatky v %)
// - intervalValueNPercentFee is floating-point (0=0%, 1.0=100%) pecent fee paid from the account value
//   at the end of every investment period (zlomek rocniho spravcovkeho procentniho poplatku)
func DCA(points pricePoints, startTime time.Time, investmentInterval time.Duration, numOfSteps int, toInvestEveryStep float64,
	purchaseFixedFee, purchaseNPercentFee, intervalValueNPercentFee float64) (totalStockPcs int, totalMoneySpent float64) {
	for step := 0; step < numOfSteps; step++ {
		price, err := points.PriceAt(startTime.Add(investmentInterval*time.Duration(step)), investmentInterval/2)
		if err != nil {
			fmt.Printf("cannot get price - %s\n", err.Error())
			continue
		}

		buyPcsThisStep := math.Round(toInvestEveryStep / price) // number of shares we buy this step
		if buyPcsThisStep < 0 {
			fmt.Printf("cannot afford to buy at least 1 item this step\n")
			continue
		}
		totalStockPcs += int(buyPcsThisStep)    // increase amount of shares we now own
		spentThisStep := buyPcsThisStep * price // increase amount of money spent this step

		// apply fees
		if step > 0 {
			spentThisStep += purchaseFixedFee + purchaseNPercentFee*spentThisStep + float64(totalStockPcs)*price*intervalValueNPercentFee
		} else {
			spentThisStep += purchaseFixedFee + purchaseNPercentFee*spentThisStep
		}

		totalMoneySpent += spentThisStep
	}

	return
}
