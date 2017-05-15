package main

import (
	"time"
	"sort"
	"fmt"
)

// pricePoint is a single price at a specific point in time
type pricePoint struct {
	Date  time.Time
	Price float64
}

// pricePoints stores and provides item prices at various points in time
type pricePoints struct {
	points []*pricePoint
	sorted bool
}

func (pp *pricePoints) ensureSorted() {
	if pp.sorted {
		return
	}
	pp.sorted = true

    sort.Slice(pp.points, func(i, j int) bool {
        return pp.points[i].Date.Before(pp.points[j].Date)
    })

}

// Oldest returns time value of the oldest data point
func (pp *pricePoints) Oldest() time.Time {
	pp.ensureSorted()

	if len(pp.points) > 0 {
		return pp.points[0].Date
	}

	return time.Time{}
}

// Newest returns time value of the newest (most recent) data point
func (pp *pricePoints) Newest() time.Time {
	pp.ensureSorted()

	if len(pp.points) > 0 {
		return pp.points[len(pp.points)-1].Date
	}

	return time.Time{}
}

// NumPoints returns number of stored points
func (pp *pricePoints) NumPoints() int {
	return len(pp.points)
}

// PriceAt returns item price closest to a specific time.
// maxDiff specifies the maximum allowable time duration between 
// requested time and recorded time.
func (pp *pricePoints) PriceAt(at time.Time, maxDiff time.Duration) (float64, error) {
	pp.ensureSorted()

	// return nearest point before requested time
	// points are sorted starting from oldest
	for _, p := range pp.points {
		durationAfterPoint := at.Sub(p.Date)
		if durationAfterPoint >= 0 && durationAfterPoint <= maxDiff {
			return p.Price, nil
		}
	}

	// test for when the requested point is before the first one (oldest)
	if len(pp.points) > 0 {
		firstp := pp.points[0]
		durationBeforeFirst := firstp.Date.Sub(at)
		if durationBeforeFirst >= 0 && durationBeforeFirst <= maxDiff {
			return firstp.Price, nil
		}
	}

	return 0, fmt.Errorf("requested point %s is outside of available data range", at)
}

