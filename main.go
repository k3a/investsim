package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/alexflint/go-arg"
)

const dayInterval = 24 * time.Hour
const yearInterval = 365 * dayInterval

type progArgs struct {
	FixedFee    float64 `arg:"-f,help:fixed fee per transaction (currency)"`
	PercentFee  float64 `arg:"-p,help:fee per transaction (percent)"`
	TimedFee    float64 `arg:"-t,help:fee per year (percent)"`
	ToInvest    float64 `arg:"-i,help:amount (in currency) to invest or own at the end"`
	RecentYears int     `arg:"-r,help:specify starting date by subtracting this amout of years from the most recent CSV data point"`
	Years       int     `arg:"-y,help:alternative to -r (specify just the invesment horizon and return means)"`
	CSV         string  `arg:"positional,required,help:CSV file"`
}

func (a *progArgs) Description() string {
	return "Computes the best strategy for cost/value averaging using Yahoo CSV export of historical data"
}

type runResult struct {
	duration                                                time.Duration
	totalStockPcs                                           int
	totalMoneySpent, finalValueMoney                        float64
	lowestInvestment, highestInvestment                     float64
	lowestInvestmentPercentOff, highestInvestmentPercentOff float64
}

func (rr runResult) Print(title string, alsoLowestHighest bool) {
	profitLossPercentTotal := (rr.finalValueMoney/rr.totalMoneySpent - 1.0) * 100
	profitLossPercentAnual := profitLossPercentTotal / float64(rr.duration/yearInterval)

	if !alsoLowestHighest {
		fmt.Printf(" - %s total # of stocks %d, spent $%.0f, value at the end $%.0f, profit/loss %.0f %% (%.0f %% pa)\n",
			title, rr.totalStockPcs, rr.totalMoneySpent, rr.finalValueMoney, profitLossPercentTotal, profitLossPercentAnual)
	} else {
		fmt.Printf(" - %s total # of stocks %d, spent $%.0f, value at the end $%.0f, profit/loss %.0f %% (%.0f %% pa)\n",
			title, rr.totalStockPcs, rr.totalMoneySpent, rr.finalValueMoney, profitLossPercentTotal, profitLossPercentAnual)
		fmt.Printf("   Highest single investment $%.0f (%.0f %% planned), lowest $%.0f (%.0f %% planned)\n",
			rr.highestInvestment, rr.highestInvestmentPercentOff, rr.lowestInvestment, rr.lowestInvestmentPercentOff)
	}
}

func singleRun(points pricePoints, purchaseFeeFixed, puchaseFeeNPercent, timedFeeNPercent, toInvest float64,
	startTime, endTime time.Time, interval time.Duration) (dca, va runResult) {

	// price at the end of the investment horizon
	mostRecentPrice, err := points.PriceAt(endTime, 5*dayInterval)
	if err != nil {
		fmt.Printf("cannot get most recent price - %s\n", err.Error())
		panic(err)
	}

	// complete investment interval we are interested in
	completeInterval := endTime.Sub(startTime)

	// for a single interval compute:
	numSteps := int(completeInterval / interval) // steps this interval
	toInvestInSingleStep := toInvest / float64(numSteps)
	intervalFeeNPercent := float64(interval) / float64(yearInterval) * timedFeeNPercent

	// DCA - dollar-cost averaging
	{
		totalStockPcs, spent := DCA(points, startTime, interval, numSteps, toInvestInSingleStep,
			purchaseFeeFixed, puchaseFeeNPercent, intervalFeeNPercent)

		dca.duration = completeInterval
		dca.totalStockPcs = totalStockPcs
		dca.totalMoneySpent = spent
		dca.finalValueMoney = float64(totalStockPcs) * mostRecentPrice
		dca.lowestInvestment = toInvest
		dca.highestInvestment = toInvest
	}

	// VA - value averaging
	{
		totalStockPcs, spent, lowestInvestment, highestInvestment :=
			VA(points, startTime, interval, numSteps, toInvestInSingleStep,
				purchaseFeeFixed, puchaseFeeNPercent, intervalFeeNPercent)

		va.duration = completeInterval
		va.totalStockPcs = totalStockPcs
		va.totalMoneySpent = spent
		va.finalValueMoney = float64(totalStockPcs) * mostRecentPrice
		va.lowestInvestment = lowestInvestment
		va.lowestInvestmentPercentOff = 100.0 / toInvestInSingleStep * va.lowestInvestment
		va.highestInvestment = highestInvestment
		va.highestInvestmentPercentOff = 100.0 / toInvestInSingleStep * va.highestInvestment
	}

	// VAProtected - value averaging with option protection
	/*	{
		totalStockPcs, spent, lowestInvestment, highestInvestment :=
			VAProtected(points, oldest, interval, numSteps, toInvestInSingleStep,
				purchaseFeeFixed, puchaseFeeNPercent, intervalFeeNPercent)

		finalInvestVal := float64(totalStockPcs) * mostRecentPrice

		fmt.Printf(" - VAProtected total # of stocks %d, spent $%.0f, value at the end $%.0f, profit/loss %.0f %%\n",
			totalStockPcs, spent, finalInvestVal, (finalInvestVal/spent-1.0)*100)
		fmt.Printf("   Highest single investment $%.0f (%.0f %% planned), lowest $%.0f (%.0f %% planned)\n",
			highestInvestment, 100.0/toInvestInSingleStep*highestInvestment,
			lowestInvestment, 100.0/toInvestInSingleStep*lowestInvestment)
	}*/

	return
}

func simulateIntervalsInFixedRange(points pricePoints, purchaseFeeFixed, puchaseFeeNPercent, timedFeeNPercent, toInvest float64,
	startTime, endTime time.Time, intervalsToTest []time.Duration) {

	fmt.Printf("\nInvestment period %s - %s\n", startTime, endTime)

	for _, interval := range intervalsToTest {
		completeInterval := endTime.Sub(startTime)
		numSteps := int(completeInterval / interval) // steps this interval

		fmt.Printf("INTERVAL %s, %d steps\n", durationString(interval), numSteps)

		dca, va := singleRun(points, purchaseFeeFixed, puchaseFeeNPercent, timedFeeNPercent, toInvest, startTime, endTime, interval)

		dca.Print("DCA", false)
		va.Print("VA", true)
	}
}

type runResultArray []*runResult

func (s runResultArray) Len() int {
	return len(s)
}
func (s runResultArray) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s runResultArray) Less(i, j int) bool {
	return s[i].finalValueMoney < s[j].finalValueMoney
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

	fmt.Printf("Investing or going to own $%.0f\n", args.ToInvest)
	fmt.Printf("Fixed fee %.2f | Percent fee %.2f | Timed fee %.2f\n",
		args.FixedFee, args.PercentFee, args.TimedFee)

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

	if args.Years > 0 { // statistical computation with moving range
		investmentInterval := time.Duration(args.Years) * yearInterval
		availableInterval := newest.Sub(oldest)

		weeksToSlide := int((availableInterval - investmentInterval) / (7 * 24 * time.Hour))
		if weeksToSlide < 0 {
			fmt.Fprintf(os.Stderr, "ERROR: Provided data interval [%s - %s] (%s) does not fit investment horizon %s!",
				oldest, newest, availableInterval, investmentInterval)
			os.Exit(1)
		}
		fmt.Printf("Will slide investment horizon %d weeks inside available data\n", weeksToSlide)

		dcaResults := make(map[time.Duration]runResultArray)
		vaResults := make(map[time.Duration]runResultArray)

		for day := 0; day < weeksToSlide*7; day += 7 {
			for _, interval := range intervalsToTest {
				dca, va := singleRun(points, args.FixedFee, args.PercentFee/100, args.TimedFee/100, args.ToInvest,
					newest.Add(-time.Duration(day)*dayInterval-time.Duration(args.Years)*yearInterval),
					newest.Add(-time.Duration(day)*dayInterval),
					interval)

				dcaResults[interval] = append(dcaResults[interval], &dca)
				vaResults[interval] = append(vaResults[interval], &va)
			}
		}

		// for every interval
		for _, interval := range intervalsToTest {
			// sort results
			sort.Sort(dcaResults[interval])
			sort.Sort(vaResults[interval])

			// pick means (0.5-quantiles)
			dcaIntervalArr := dcaResults[interval]
			dca := dcaIntervalArr[len(dcaIntervalArr)/2]
			vaIntervalArr := vaResults[interval]
			va := vaIntervalArr[len(vaIntervalArr)/2]

			numSamples := len(dcaResults[interval])

			fmt.Printf("INTERVAL %s (%d samples)\n", durationString(interval), numSamples)
			dca.Print("DCA", false)
			va.Print("VA", true)

			var avgDCA, avgVA runResult
			for i := 0; i < numSamples; i++ {
				sample := dcaResults[interval][i]
				avgDCA.duration = sample.duration
				avgDCA.totalMoneySpent += sample.totalMoneySpent / float64(numSamples)
				avgDCA.finalValueMoney += sample.finalValueMoney / float64(numSamples)
				avgDCA.highestInvestment += sample.highestInvestment / float64(numSamples)
				avgDCA.lowestInvestment += sample.lowestInvestment / float64(numSamples)

				sample = vaResults[interval][i]
				avgVA.duration = sample.duration
				avgVA.totalMoneySpent += sample.totalMoneySpent / float64(numSamples)
				avgVA.finalValueMoney += sample.finalValueMoney / float64(numSamples)
				avgVA.highestInvestment += sample.highestInvestment / float64(numSamples)
				avgVA.lowestInvestment += sample.lowestInvestment / float64(numSamples)
			}
			avgDCA.Print("DCA AVG", false)
			avgVA.Print("VA AVG", true)
		}

	} else { // single range
		// overwrite oldest point date in range by subtracting from newest
		if args.RecentYears > 0 {
			oldest = newest.AddDate(-args.RecentYears, 0, 0)
		}

		simulateIntervalsInFixedRange(points, args.FixedFee, args.PercentFee/100, args.TimedFee/100, args.ToInvest, oldest, newest, intervalsToTest)
	}

}
