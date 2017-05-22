package main

import (
	"fmt"
	"math"
	"time"

	"github.com/alexflint/go-arg"
)

type progArgs struct {
	FixedFee    float64 `arg:"-f,help:fixed fee per transaction (currency)"`
	PercentFee  float64 `arg:"-p,help:fee per transation (percent)"`
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

	fmt.Printf("oldest %s\n", oldest)
	fmt.Printf("newest %s\n", newest)

	// overwrite oldest point date by subtracting from newest
	if args.RecentYears > 0 {
		oldest = newest.AddDate(-args.RecentYears, 0, 0)
	}

	fmt.Printf("Investing or going to own %.0f\n", args.ToInvest)
	fmt.Printf("Investment period %s - %s\n", oldest, newest)
	fmt.Printf("Fixed fee %.2f | Percent fee %.2f | Timed fee %.2f\n",
		args.FixedFee, args.PercentFee, args.TimedFee)
	fmt.Println()

	// complete investment interval we are interested in
	completeInterval := newest.Sub(oldest)

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

	mostRecentPrice, err := points.PriceAt(newest, 24*time.Hour)
	if err != nil {
		fmt.Printf("cannot get most recent price - %s\n", err.Error())
		panic(err)
	}

	for _, interval := range intervalsToTest {
		toInvest := args.ToInvest

		numSteps := int(completeInterval / interval)
		toInvestInSingleStep := toInvest / float64(numSteps)
		intervalFee := float64(interval) / float64(12*30*24*time.Hour) * args.TimedFee / 100

		fmt.Printf("INTERVAL %s, %d steps\n", durationString(interval), numSteps)

		// DCA
		{
			var totalAmount, spent float64

			for step := 0; step < numSteps; step++ {
				price, err := points.PriceAt(oldest.Add(interval*time.Duration(step)), interval/2)
				if err != nil {
					fmt.Printf("cannot get price - %s\n", err.Error())
					continue
				}

				totalAmount += toInvestInSingleStep / price
				spentThisStep := toInvestInSingleStep

				if step > 0 {
					spentThisStep += args.FixedFee + args.PercentFee*toInvestInSingleStep + totalAmount*mostRecentPrice*intervalFee
				} else {
					spentThisStep += args.FixedFee + args.PercentFee*toInvestInSingleStep
				}

				spent += spentThisStep
			}

			finalInvestVal := totalAmount * mostRecentPrice

			fmt.Printf(" - DCA total stock # %.4f, spent %.4f, value at the end %.4f (%.4f %%), value/spent %.2f\n",
				totalAmount, spent, finalInvestVal, 100.0/toInvest*finalInvestVal, finalInvestVal/spent)
		}

		// VA
		{
			var totalAmount, spent, highestInvestment, lowestInvestment float64
			lowestInvestment = math.Pow(10, 300)

			for step := 0; step < numSteps; step++ {
				price, err := points.PriceAt(oldest.Add(interval*time.Duration(step)), interval)
				if err != nil {
					fmt.Printf("cannot get price - %s\n", err.Error())
					continue
				}

				currValue := totalAmount * price
				expectedValue := float64(step+1) * toInvestInSingleStep
				buySellValueThisStep := expectedValue - currValue
				spentThisStep := buySellValueThisStep

				totalAmount += buySellValueThisStep / price
				if step > 0 {
					spentThisStep += args.FixedFee + args.PercentFee*buySellValueThisStep + totalAmount*mostRecentPrice*intervalFee
				} else {
					spentThisStep += args.FixedFee + args.PercentFee*buySellValueThisStep
				}

				spent += spentThisStep

				// update min/max
				if spentThisStep > highestInvestment {
					highestInvestment = spentThisStep
				}
				if spentThisStep < lowestInvestment {
					lowestInvestment = spentThisStep
				}
			}

			finalInvestVal := totalAmount * mostRecentPrice

			fmt.Printf(" - VA total stock # %.4f, spent %.4f, value at the end %.4f (%.4f %%), value/spent %.2f\n",
				totalAmount, spent, finalInvestVal, 100.0/toInvest*finalInvestVal, finalInvestVal/spent)
			fmt.Printf("   Highest single investment %.4f (%.2f %% planned), lowest %.4f (%.2f %% planned)\n",
				highestInvestment, 100.0/toInvestInSingleStep*highestInvestment,
				lowestInvestment, 100.0/toInvestInSingleStep*lowestInvestment)
		}

	}
}
