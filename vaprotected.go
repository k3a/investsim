package main

import (
	"fmt"
	"math"
	"time"

	"github.com/jasonmerecki/gopriceoptions"
)

// VAProtected simulates value-averaging investment while also purchasing protective ATM PUT option every year
//
// Returns:
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
func VAProtected(points pricePoints, startTime time.Time, investmentInterval time.Duration, numOfSteps int, expectedValueIncreaseEveryStep float64,
	purchaseFixedFee, purchaseNPercentFee, intervalValueNPercentFee float64) (totalStockPcs int, totalMoneySpent, lowestInvestment, highestInvestment float64) {
	lowestInvestment = math.Pow(10, 300) // initialize lowest investment to some crazy-hight value initially

	lastOptionYear := 0
	lastOptionStrike := 0.0

	for step := 0; step < numOfSteps; step++ {
		now := startTime.Add(investmentInterval * time.Duration(step))
		price, err := points.PriceAt(now, investmentInterval/2)
		if err != nil {
			fmt.Printf("cannot get price - %s\n", err.Error())
			continue
		}

		// if we have an option contract purchased, let's expect that price can't go
		// lower than option strike. It is not 100% true due to option delta (time value)
		// but enough to do some estimates
		protectedPrice := price
		if lastOptionYear > 0 {
			if protectedPrice < lastOptionStrike {
				protectedPrice = lastOptionStrike
			}
		}

		// portfolio monteray value now taking protective options in account
		currentPortfolioValueValue := float64(totalStockPcs) * protectedPrice
		// expected portfolio value at this point
		expectedPortfolioValue := float64(step+1) * expectedValueIncreaseEveryStep
		// portfolio value monetary difference (monetary value to purchase/sell)
		expectedVsCurrentDiff := expectedPortfolioValue - currentPortfolioValueValue

		buyOrSellPcs := math.Round(expectedVsCurrentDiff / price) // number of stock to seel / buy this step
		spentThisStep := buyOrSellPcs * price                     // amount of money spent this step
		totalStockPcs += int(buyOrSellPcs)                        // amount of stock owned now

		// apply fees
		if step > 0 {
			spentThisStep += purchaseFixedFee + purchaseNPercentFee*spentThisStep + float64(totalStockPcs)*price*intervalValueNPercentFee
		} else {
			spentThisStep += purchaseFixedFee + purchaseNPercentFee*spentThisStep
		}

		// is time to purchase a new yearly option?
		if lastOptionYear != now.Year() {
			riskFreeInterest := 0.024 // ~ 2% now
			impliedVolatilityN := 0.3 // ~ 30% normally

			// compute const of a single option contract (covering 100 pcs of stock) at the current price
			// and store this price as option strike
			contractCost := 100.0 * gopriceoptions.PriceBlackScholes("p", price, price, 1.0 /*1year*/, impliedVolatilityN, riskFreeInterest, 0)
			lastOptionYear = now.Year()
			lastOptionStrike = price

			// buy enough contracts to cover all pcs owned so far (use round)
			numContractsToBuy := 1.0
			if totalStockPcs > 100 {
				numContractsToBuy = math.Round(float64(totalStockPcs) / 100)
			}

			// make this step more expensive due to option purchase
			spentOnContracts := contractCost * numContractsToBuy
			fmt.Printf("   (purchased %.0f pcs of $%.0f strike put options for a total of $%.0f)\n",
				numContractsToBuy, price, spentOnContracts)
			spentThisStep += spentOnContracts
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
