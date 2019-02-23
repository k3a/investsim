package main

import (
	"fmt"
	"time"

	"github.com/alexflint/go-arg"
)

type progArgs struct {
	FixedFee    float64 `arg:"-f,help:fixed fee per transaction (currency)"`
	PercentFee  float64 `arg:"-p,help:fee per transaction (percent)"`
	TimedFee    float64 `arg:"-t,help:fee per year (percent)"`
	ToInvest    float64 `arg:"-i,help:amount (in currency) to invest or own at the end"`
	RecentYears int     `arg:"-r,help:specify starting date by subtracting this amout of years from the most recent CSV data point"`
	CSV         string  `arg:"positional,required,help:CSV file"`
}

func (a *progArgs) Description() string {
	return "Computes the best strategy for cost/value averaging using Yahoo CSV export of historical data"
}

func main() {
	var args progArgs
	args.FixedFee = 0
	args.PercentFee = 0
	args.TimedFee = 0
	args.ToInvest = 100
	args.RecentYears = 0
	arg.MustParse(&args)

	// load points
	points, err := parseCSV(args.CSV)
	if err != nil {
		panic(err)
	}

	// get dates of oldest and newest point
	oldest := points.Oldest()
	newest := points.Newest()

	fmt.Printf("oldest data point %s\n", oldest)
	fmt.Printf("newest data point %s\n", newest)
	fmt.Println()

	// overwrite oldest point date by subtracting from newest
	if args.RecentYears > 0 {
		oldest = newest.AddDate(-args.RecentYears, 0, 0)
	}

	fmt.Printf("Investing or going to own $%.0f\n", args.ToInvest)
	fmt.Printf("Investment period %s - %s\n", oldest, newest)
	fmt.Printf("Fixed fee %.2f | Percent fee %.2f | Timed fee %.2f\n",
		args.FixedFee, args.PercentFee, args.TimedFee)
	fmt.Println()

	// complete investment interval we are interested in
	completeInterval := newest.Sub(oldest)

	// step intervals to test
	intervalsToTest := []time.Duration{
		1 * 30 * 24 * time.Hour,
		2 * 30 * 24 * time.Hour,
		3 * 30 * 24 * time.Hour,
		6 * 30 * 24 * time.Hour,
		8 * 30 * 24 * time.Hour,
		12 * 30 * 24 * time.Hour,
		18 * 30 * 24 * time.Hour,
		24 * 30 * 24 * time.Hour,
		48 * 30 * 24 * time.Hour,
	}

	// price at the end (most recent price from the provided data)
	mostRecentPrice, err := points.PriceAt(newest, 24*time.Hour)
	if err != nil {
		fmt.Printf("cannot get most recent price - %s\n", err.Error())
		panic(err)
	}

	// for every interval length to test..
	for _, interval := range intervalsToTest {
		toInvest := args.ToInvest

		numSteps := int(completeInterval / interval) // steps this interval
		toInvestInSingleStep := toInvest / float64(numSteps)
		intervalFeeNPercent := float64(interval) / float64(12*30*24*time.Hour) * args.TimedFee / 100

		fmt.Printf("INTERVAL %s, %d steps\n", durationString(interval), numSteps)

		// DCA - dollar-cost averaging
		{
			totalStockPcs, spent := DCA(points, oldest, interval, numSteps, toInvestInSingleStep,
				args.FixedFee, args.PercentFee/100, intervalFeeNPercent)

			finalInvestVal := float64(totalStockPcs) * mostRecentPrice

			fmt.Printf(" - DCA total # of stocks %d, spent $%.0f, value at the end $%.0f, profit/loss %.0f %%\n",
				totalStockPcs, spent, finalInvestVal, (finalInvestVal/spent-1.0)*100)
		}

		// VA - value averaging
		{
			totalStockPcs, spent, lowestInvestment, highestInvestment :=
				VA(points, oldest, interval, numSteps, toInvestInSingleStep,
					args.FixedFee, args.PercentFee/100, intervalFeeNPercent)

			finalInvestVal := float64(totalStockPcs) * mostRecentPrice

			fmt.Printf(" - VA total # of stocks %d, spent $%.0f, value at the end $%.0f, profit/loss %.0f %%\n",
				totalStockPcs, spent, finalInvestVal, (finalInvestVal/spent-1.0)*100)
			fmt.Printf("   Highest single investment $%.0f (%.0f %% planned), lowest $%.0f (%.0f %% planned)\n",
				highestInvestment, 100.0/toInvestInSingleStep*highestInvestment,
				lowestInvestment, 100.0/toInvestInSingleStep*lowestInvestment)
		}

		// VAProtected - value averaging with option protection
		/*	{
			totalStockPcs, spent, lowestInvestment, highestInvestment :=
				VAProtected(points, oldest, interval, numSteps, toInvestInSingleStep,
					args.FixedFee, args.PercentFee/100, intervalFeeNPercent)

			finalInvestVal := float64(totalStockPcs) * mostRecentPrice

			fmt.Printf(" - VAProtected total # of stocks %d, spent $%.0f, value at the end $%.0f, profit/loss %.0f %%\n",
				totalStockPcs, spent, finalInvestVal, (finalInvestVal/spent-1.0)*100)
			fmt.Printf("   Highest single investment $%.0f (%.0f %% planned), lowest $%.0f (%.0f %% planned)\n",
				highestInvestment, 100.0/toInvestInSingleStep*highestInvestment,
				lowestInvestment, 100.0/toInvestInSingleStep*lowestInvestment)
		}*/

		// new line
		fmt.Println()

	}
}
